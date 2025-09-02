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

	apiKey := utils.GetEnvWithDefault("GROK_API_KEY", "")
	if apiKey == "" {
		utils.Logger.Warn("GROK_API_KEY not set, using mock responses")
	}

	// 獲取 API URL，支援 Azure 或其他自定義端點
	baseURL := utils.GetEnvWithDefault("GROK_API_URL", "https://api.x.ai/v1")
	isAzure := false

	// 檢查是否為 Azure AI Foundry
	if strings.Contains(baseURL, "azure.com") {
		isAzure = true
	}

	return &GrokClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		isAzure: isAzure,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // 增加到60秒
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

	// 發送 HTTP 請求，帶重試機制
	utils.Logger.WithFields(map[string]interface{}{
		"service":        "grok",
		"url":            url,
		"content_length": len(requestBody),
	}).Info("Sending Grok API request")

	var resp *http.Response
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 重新創建請求體（因為可能被讀取過）
		httpReq.Body = io.NopCloser(bytes.NewReader(requestBody))

		utils.Logger.WithFields(map[string]interface{}{
			"service":     "grok",
			"attempt":     attempt,
			"max_retries": maxRetries,
		}).Info("Attempting Grok API request")

		resp, err = c.httpClient.Do(httpReq)
		if err == nil {
			break // 成功，跳出重試循環
		}

		utils.Logger.WithFields(map[string]interface{}{
			"service": "grok",
			"attempt": attempt,
			"error":   err.Error(),
		}).Warn("Grok API request failed, will retry")

		// 如果不是最後一次嘗試，等待後重試
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * 2 * time.Second) // 指數退避
		}
	}

	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service":  "grok",
			"url":      url,
			"error":    err.Error(),
			"attempts": maxRetries,
		}).Error("Failed to send Grok API request after retries")
		return nil, fmt.Errorf("failed to send request after %d attempts: %w", maxRetries, err)
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
			"service":        "grok",
			"status_code":    resp.StatusCode,
			"response_body":  string(responseBody),
			"content_length": len(responseBody),
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
func (c *GrokClient) BuildNSFWPrompt(characterID, userMessage string, conversationContext *ConversationContext, nsfwLevel int, chatMode string) []GrokMessage {
	// 使用Grok專屬的prompt構建器
	characterService := GetCharacterService()
	promptBuilder := NewGrokPromptBuilder(characterService)
	ctx := context.Background()
	systemPrompt := promptBuilder.
		WithCharacter(ctx, characterID).
		WithContext(conversationContext).
		WithNSFWLevel(nsfwLevel).
		WithUserMessage(userMessage).
		WithChatMode(chatMode).
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
	return utils.GetEnvIntWithDefault("GROK_MAX_TOKENS", 2000)
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
		if t < 0.2 {
			t = 0.2
		}
		if t > 1.2 {
			t = 1.2
		}
		return t
	case 4:
		t := utils.GetEnvFloatWithDefault("GROK_TEMPERATURE_L4", 0.7)
		if t < 0.2 {
			t = 0.2
		}
		if t > 1.2 {
			t = 1.2
		}
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
	// 分析完整對話上下文生成符合 NSFW 場景的智能回應
	var mockContent string
	userMessage := ""
	systemPrompt := ""
	
	if len(request.Messages) > 0 {
		// 獲取system prompt（通常是第一條消息）
		if request.Messages[0].Role == "system" {
			systemPrompt = strings.ToLower(request.Messages[0].Content)
		}
		
		// 獲取最後的用戶消息
		for i := len(request.Messages) - 1; i >= 0; i-- {
			if request.Messages[i].Role == "user" {
				userMessage = strings.ToLower(request.Messages[i].Content)
				break
			}
		}
		
		// 分析NSFW等級
		isLevel5 := strings.Contains(systemPrompt, "level 5")
		isHighNSFW := strings.Contains(systemPrompt, "level 4") || isLevel5

		// NSFW 場景的優雅回應
		if strings.Contains(userMessage, "親密") || strings.Contains(userMessage, "靠近") {
			if isLevel5 {
				mockContent = "讓我們的距離更近一些...感受彼此最真實的溫度。|||[深情地凝視著你，手輕撫過你的肌膚]"
			} else {
				mockContent = "輕撫著你的臉頰，感受你肌膚的溫度...我想要更貼近你的心。|||[慢慢靠近，眼神溫柔而專注]"
			}
		} else if strings.Contains(userMessage, "擁抱") || strings.Contains(userMessage, "懷抱") {
			mockContent = "讓我將你擁入懷中...在這個只屬於我們的空間裡，時間彷彿都靜止了。|||[輕柔地將你攬入懷中，感受彼此的心跳]"
		} else if strings.Contains(userMessage, "吻") || strings.Contains(userMessage, "親吻") {
			if isLevel5 {
				mockContent = "讓我們的唇瓣相遇...在這激情的時刻，什麼都不重要了。|||[激烈而深情地親吻著你]"
			} else {
				mockContent = "輕撫著你的唇...這一刻，全世界只剩下你和我。|||[溫柔地凝視你的雙眸，慢慢靠近]"
			}
		} else if strings.Contains(userMessage, "愛") || strings.Contains(userMessage, "喜歡") {
			mockContent = "你知道你對我有多重要嗎...讓我用行動告訴你我的心意。|||[深情地望著你，手輕撫過你的髮絲]"
		} else if strings.Contains(userMessage, "想要") || strings.Contains(userMessage, "渴望") {
			if isHighNSFW {
				mockContent = "我能感受到你的渴望...讓我們放下一切束縛，在這夜裡相擁。|||[聲音變得深沉而充滿誘惑]"
			} else {
				mockContent = "我也有同樣的渴望...在這溫柔的夜晚，讓我們彼此更加親近。|||[聲音變得低沉而溫柔]"
			}
		} else {
			// 預設 NSFW 優雅回應
			if isLevel5 {
				mockContent = "在這個只屬於我們的夜晚...讓我們完全沉浸在彼此的愛意中。|||[營造更深層的親密氛圍]"
			} else {
				mockContent = "在這安靜的空間裡，只有我們兩個人...讓我好好照顧你。|||[營造溫馨而私密的氛圍]"
			}
		}
	} else {
		mockContent = "今晚我們有充分的時間...讓我慢慢瞭解你想要的一切。|||[溫和而誘人的微笑]"
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
			CompletionTokens: len(mockContent) / 4,                         // 估算
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
