package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
)

// GrokClient Grok 客戶端
type GrokClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	isAzure    bool
}

// GrokRequest Grok 請求結構（類似 OpenAI 格式）
type GrokRequest struct {
	Model       string        `json:"model"`
	Messages    []GrokMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
	User        string        `json:"user,omitempty"`
}

// GrokMessage Grok 消息結構
type GrokMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// GrokResponse Grok 回應結構
type GrokResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewGrokClient 創建新的 Grok 客戶端
func NewGrokClient() *GrokClient {
	// 確保環境變數已載入
	utils.LoadEnv()
	
	// 獲取 API URL，支援 Azure 或其他自定義端點
	baseURL := utils.GetEnvWithDefault("GROK_API_URL", "https://api.x.ai/v1")
	isAzure := false
	
	// 檢查是否為 Azure AI Foundry
	if strings.Contains(baseURL, "azure.com") {
		isAzure = true
	}
	
	return &GrokClient{
		apiKey:  utils.GetEnvWithDefault("GROK_API_KEY", ""),
		baseURL: baseURL,
		isAzure: isAzure,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateResponse 生成對話回應（NSFW 內容）
func (c *GrokClient) GenerateResponse(ctx context.Context, request *GrokRequest) (*GrokResponse, error) {
	startTime := time.Now()
	
	// 記錄請求開始
	utils.Logger.WithFields(map[string]interface{}{
		"service":        "grok",
		"base_url":       c.baseURL,
		"model":          request.Model,
		"max_tokens":     request.MaxTokens,
		"temperature":    request.Temperature,
		"user":           request.User,
		"messages_count": len(request.Messages),
		"api_configured": c.apiKey != "",
	}).Info("Grok API request started")

	// 檢查 API Key - 如果未配置，使用模擬響應
	if c.apiKey == "" {
		utils.Logger.WithField("service", "grok").Warn("Grok API key not configured, using mock response")
		return c.generateMockResponse(request), nil
	}

	// 開發模式下詳細記錄 prompt 內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "grok",
			"model":   request.Model,
			"user":    request.User,
		}).Info("🔥 Grok Request Details")

		for i, msg := range request.Messages {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "grok",
				"message_index":  i,
				"role":           msg.Role,
				"content_length": len(msg.Content),
			}).Info(fmt.Sprintf("📝 Prompt [%s]: %s", strings.ToUpper(msg.Role), msg.Content))
		}
	}

	// 設置默認值
	if request.Model == "" {
		request.Model = getGrokModel()
	}
	if request.MaxTokens == 0 {
		request.MaxTokens = getGrokMaxTokens()
	}
	// 動態調整溫度：若未顯式設定，依據 prompt 中的 Level 推斷
	if request.Temperature <= 0 {
		lvl := inferNSFWLevelFromMessages(request.Messages)
		request.Temperature = temperatureForLevel(lvl)
	}

	// 準備 HTTP 請求
	requestBody, err := json.Marshal(request)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "grok",
			"error":   err.Error(),
		}).Error("Failed to marshal Grok request")
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 創建 HTTP 請求
	var url string
	if c.isAzure {
		// Azure AI Foundry 使用與 OpenAI 相同的端點結構
		url = c.baseURL + "/models/chat/completions?api-version=2024-05-01-preview"
	} else {
		url = c.baseURL + "/chat/completions"
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "grok",
			"url":     url,
			"error":   err.Error(),
		}).Error("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 設置請求標頭
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "thewavess-ai-core/1.0")
	
	// Azure 需要不同的認證方式
	if c.isAzure {
		httpReq.Header.Set("api-key", c.apiKey)
	} else {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// 發送 HTTP 請求
	utils.Logger.WithFields(map[string]interface{}{
		"service":        "grok",
		"url":            url,
		"content_length": len(requestBody),
	}).Info("Sending Grok API request")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "grok",
			"url":     url,
			"error":   err.Error(),
		}).Error("Failed to send Grok API request")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 讀取響應
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service":     "grok",
			"status_code": resp.StatusCode,
			"error":       err.Error(),
		}).Error("Failed to read Grok API response")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 檢查 HTTP 狀態碼
	if resp.StatusCode != http.StatusOK {
		utils.Logger.WithFields(map[string]interface{}{
			"service":         "grok",
			"status_code":     resp.StatusCode,
			"response_body":   string(responseBody),
			"content_length":  len(responseBody),
		}).Error("Grok API returned non-200 status")
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	// 解析響應
	var grokResponse GrokResponse
	if err := json.Unmarshal(responseBody, &grokResponse); err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service":        "grok",
			"error":          err.Error(),
			"response_body":  string(responseBody),
			"content_length": len(responseBody),
		}).Error("Failed to unmarshal Grok API response")
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 計算響應時間
	duration := time.Since(startTime)

	// 記錄成功響應
	utils.Logger.WithFields(map[string]interface{}{
		"service":           "grok",
		"response_id":       grokResponse.ID,
		"model":             grokResponse.Model,
		"prompt_tokens":     grokResponse.Usage.PromptTokens,
		"completion_tokens": grokResponse.Usage.CompletionTokens,
		"total_tokens":      grokResponse.Usage.TotalTokens,
		"choices_count":     len(grokResponse.Choices),
		"duration_ms":       duration.Milliseconds(),
		"is_mock":           false,
	}).Info("Grok API response received")

	// 開發模式下詳細記錄響應內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service":     "grok",
			"response_id": grokResponse.ID,
			"model":       grokResponse.Model,
		}).Info("🎯 Grok Response Details")

		for i, choice := range grokResponse.Choices {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "grok",
				"choice_index":   i,
				"finish_reason":  choice.FinishReason,
				"content_length": len(choice.Message.Content),
				"is_mock":        false,
			}).Info(fmt.Sprintf("💬 Response [%d]: %s", i, choice.Message.Content))
		}
	}

	return &grokResponse, nil
}


// BuildNSFWPrompt 構建 NSFW 場景的提示詞（使用統一模板）
func (c *GrokClient) BuildNSFWPrompt(characterID, userMessage, sceneDescription string, conversationContext *ConversationContext, nsfwLevel int) []GrokMessage {
	// 構建記憶提示詞（NSFW 場景使用縮短版本以節省 token）
	memoryPrompt := ""
	if conversationContext != nil && conversationContext.MemoryPrompt != "" {
		// 對 NSFW 場景，截短記憶內容
		lines := strings.Split(conversationContext.MemoryPrompt, "\n")
		var shortMemory []string
		for i, line := range lines {
			if i >= 8 { // 限制最多 8 行記憶內容
				break
			}
			shortMemory = append(shortMemory, line)
		}
		if len(shortMemory) > 0 {
			memoryPrompt = strings.Join(shortMemory, "\n")
		}
	}

	// 使用統一的prompt模板
	characterService := GetCharacterService()
	promptBuilder := NewPromptBuilder(characterService)
	ctx := context.Background()
	systemPrompt := promptBuilder.
		WithCharacter(ctx, characterID).
		WithContext(conversationContext).
		WithNSFWLevel(nsfwLevel).
		WithUserMessage(userMessage).
		WithSceneDescription(sceneDescription).
		WithMemory(memoryPrompt).
		Build(ctx)

	messages := []GrokMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// 添加對話歷史
	if conversationContext != nil {
		for i, msg := range conversationContext.RecentMessages {
			if i >= 3 { // NSFW 場景保留較少歷史
				break
			}
			messages = append(messages, GrokMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 添加當前用戶消息
	messages = append(messages, GrokMessage{
		Role:    "user",
		Content: userMessage,
	})

	return messages
}

// getGrokModel 獲取 Grok 模型配置
func getGrokModel() string {
	return utils.GetEnvWithDefault("GROK_MODEL", "grok-beta")
}

// getGrokMaxTokens 獲取 Grok 最大 Token 數配置
func getGrokMaxTokens() int {
	return utils.GetEnvIntWithDefault("GROK_MAX_TOKENS", 1000)
}

// getGrokTemperature 獲取 Grok 溫度配置
func getGrokTemperature() float64 {
	return utils.GetEnvFloatWithDefault("GROK_TEMPERATURE", 0.7)
}

// inferNSFWLevelFromMessages 從 system prompt 內判斷 Level 4/5
func inferNSFWLevelFromMessages(msgs []GrokMessage) int {
	for _, m := range msgs {
		if m.Role != "system" {
			continue
		}
		s := m.Content
		if strings.Contains(s, "Level 5") {
			return 5
		}
		if strings.Contains(s, "Level 4") {
			return 4
		}
	}
	return 3
}

// temperatureForLevel 根據 NSFW 等級動態調整溫度
func temperatureForLevel(level int) float64 {
	// 預設（可被環境變數覆蓋）
	switch level {
	case 5:
		t := utils.GetEnvFloatWithDefault("GROK_TEMPERATURE_L5", 0.6)
		if t < 0.2 { t = 0.2 }
		if t > 1.2 { t = 1.2 }
		return t
	case 4:
		t := utils.GetEnvFloatWithDefault("GROK_TEMPERATURE_L4", 0.7)
		if t < 0.2 { t = 0.2 }
		if t > 1.2 { t = 1.2 }
		return t
	case 3:
		return 0.75
	case 1, 2:
		return 0.60
	default:
		return getGrokTemperature() // fallback: GROK_TEMPERATURE or 0.7
	}
}

// generateMockResponse 生成模擬響應（用於 API key 未配置或測試場景）
func (c *GrokClient) generateMockResponse(request *GrokRequest) *GrokResponse {
	// 根據用戶消息生成更智能的模擬響應
	var mockContent string
	if len(request.Messages) > 0 {
		userMessage := request.Messages[len(request.Messages)-1].Content
		
		// 簡單的關鍵詞響應映射
		if strings.Contains(strings.ToLower(userMessage), "親密") || 
		   strings.Contains(strings.ToLower(userMessage), "擁抱") {
			mockContent = "輕輕地將你擁入懷中，感受彼此的溫度...這是一個來自 Grok 的模擬回應，用於處理親密內容。真實實現會調用 Grok API。"
		} else if strings.Contains(strings.ToLower(userMessage), "愛") {
			mockContent = "我也愛你...這是一個來自 Grok 的模擬回應，用於處理情感內容。真實實現會調用 Grok API。"
		} else {
			mockContent = "這是一個來自 Grok 的模擬回應，用於處理 NSFW 內容。真實實現會調用 Grok API。"
		}
	} else {
		mockContent = "這是一個來自 Grok 的模擬回應，用於處理 NSFW 內容。真實實現會調用 Grok API。"
	}

	mockResponse := &GrokResponse{
		ID:      fmt.Sprintf("grok-mock-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   request.Model,
		Choices: []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: mockContent,
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     len(fmt.Sprintf("%v", request.Messages)) / 4, // 估算
			CompletionTokens: len(mockContent) / 4,                        // 估算
			TotalTokens:      (len(fmt.Sprintf("%v", request.Messages)) + len(mockContent)) / 4,
		},
	}

	utils.Logger.WithFields(map[string]interface{}{
		"service":           "grok",
		"response_id":       mockResponse.ID,
		"model":             mockResponse.Model,
		"prompt_tokens":     mockResponse.Usage.PromptTokens,
		"completion_tokens": mockResponse.Usage.CompletionTokens,
		"total_tokens":      mockResponse.Usage.TotalTokens,
		"is_mock":           true,
	}).Info("Generated Grok mock response")

	return mockResponse
}
