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

	db "github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// GrokClient Grok 客戶端
type GrokClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
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
		utils.Logger.Fatal("GROK_API_KEY is required but not set in environment")
	}

	// 獲取 API URL
	baseURL := utils.GetEnvWithDefault("GROK_API_URL", "https://api.x.ai/v1")

	return &GrokClient{
		apiKey:  apiKey,
		baseURL: baseURL,
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
	url := c.baseURL + "/chat/completions"
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
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

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

	// 獲取 character 物件
	character, err := characterService.GetCharacter(ctx, characterID)
	if err != nil {
		utils.Logger.WithError(err).Error("獲取角色失敗")
		// 使用基本 prompt 作為 fallback
		systemPrompt := "請以創意的方式回應用戶。"
		return []GrokMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		}
	}

	// 轉換為 db.CharacterDB 類型
	dbCharacter := &db.CharacterDB{
		ID:   character.ID,
		Name: character.GetName(),
		Type: string(character.Type),
		Tags: character.Metadata.Tags,
		UserDescription: character.UserDescription,
	}

	promptBuilder.WithCharacter(dbCharacter)
	promptBuilder.WithContext(conversationContext)
	promptBuilder.WithNSFWLevel(nsfwLevel)
	promptBuilder.WithUserMessage(userMessage)
	promptBuilder.WithChatMode(chatMode)
	systemPrompt := promptBuilder.Build()

	messages := []GrokMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// 添加對話歷史
    if conversationContext != nil {
        // 僅保留最近2則歷史（舊 -> 新）
        count := len(conversationContext.RecentMessages)
        if count > 2 { count = 2 }
        for i := count - 1; i >= 0; i-- {
            msg := conversationContext.RecentMessages[i]
            messages = append(messages, GrokMessage{Role: msg.Role, Content: msg.Content})
        }
    }

    // 添加當前用戶消息（避免與歷史最後一則重複）
    shouldAppendUser := true
    if conversationContext != nil && len(conversationContext.RecentMessages) > 0 {
        latest := conversationContext.RecentMessages[0]
        if latest.Role == "user" && strings.TrimSpace(latest.Content) == strings.TrimSpace(userMessage) {
            shouldAppendUser = false
        }
    }
    if shouldAppendUser {
        messages = append(messages, GrokMessage{Role: "user", Content: userMessage})
    }

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
