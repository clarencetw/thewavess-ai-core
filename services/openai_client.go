package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sashabaranov/go-openai"
)

// OpenAIClient OpenAI 客戶端
type OpenAIClient struct {
	client      *openai.Client
	model       string
	maxTokens   int
	temperature float32
	baseURL     string
}

// OpenAIRequest OpenAI 請求結構
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
	User        string          `json:"user,omitempty"`
}

// OpenAIMessage OpenAI 消息結構
type OpenAIMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// OpenAIResponse OpenAI 回應結構
type OpenAIResponse struct {
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

// NewOpenAIClient 創建新的 OpenAI 客戶端
func NewOpenAIClient() *OpenAIClient {
	// 確保環境變數已載入
	utils.LoadEnv()

	apiKey := utils.GetEnvWithDefault("OPENAI_API_KEY", "")
	if apiKey == "" {
		utils.Logger.Fatal("OPENAI_API_KEY is required but not set in environment")
	}

	// 從環境變數讀取配置，提供預設值
	model := utils.GetEnvWithDefault("OPENAI_MODEL", "gpt-4o")
	maxTokens := utils.GetEnvIntWithDefault("OPENAI_MAX_TOKENS", 1200)
	temperature := utils.GetEnvFloatWithDefault("OPENAI_TEMPERATURE", 0.8)

	// 獲取自定義 API URL
	baseURL := utils.GetEnvWithDefault("OPENAI_API_URL", "https://api.openai.com/v1")

	var client *openai.Client
	if baseURL != "https://api.openai.com/v1" {
		// 自定義端點
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = baseURL
		client = openai.NewClientWithConfig(config)

		utils.Logger.WithField("base_url", baseURL).Info("Using custom OpenAI API URL")
	} else {
		// 使用默認 OpenAI API
		client = openai.NewClient(apiKey)
	}

	return &OpenAIClient{
		client:      client,
		model:       model,
		maxTokens:   maxTokens,
		temperature: float32(temperature),
		baseURL:     baseURL,
	}
}

// GenerateResponse 生成對話回應
func (c *OpenAIClient) GenerateResponse(ctx context.Context, request *OpenAIRequest) (*OpenAIResponse, error) {
	// 記錄請求開始
	utils.Logger.WithFields(map[string]interface{}{
		"service":        "openai",
		"base_url":       c.baseURL,
		"model":          c.model,
		"max_tokens":     c.maxTokens,
		"temperature":    c.temperature,
		"user":           request.User,
		"messages_count": len(request.Messages),
	}).Info("OpenAI API request started")

	// 開發模式下詳細記錄 prompt 內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "openai",
			"model":   c.model,
			"user":    request.User,
		}).Info("🤖 OpenAI Request Details")

		for i, msg := range request.Messages {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "openai",
				"message_index":  i,
				"role":           msg.Role,
				"content_length": len(msg.Content),
			}).Info(fmt.Sprintf("📝 Prompt [%s]: %s", strings.ToUpper(msg.Role), msg.Content))
		}
	} else {
		// 生產環境只記錄基本信息
		for i, msg := range request.Messages {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "openai",
				"message_index":  i,
				"role":           msg.Role,
				"content_length": len(msg.Content),
			}).Debug("OpenAI request message")
		}
	}

	// 轉換消息格式
	messages := make([]openai.ChatCompletionMessage, len(request.Messages))
	for i, msg := range request.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 構建請求
	chatRequest := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
		User:        request.User,
	}

	// 調用 OpenAI API
	resp, err := c.client.CreateChatCompletion(ctx, chatRequest)

	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "openai",
			"error":   err.Error(),
			"model":   c.model,
			"user":    request.User,
		}).Error("OpenAI API call failed")
		return nil, fmt.Errorf("failed OpenAI API call: %w", err)
	}

	// 記錄API響應信息
	utils.Logger.WithFields(map[string]interface{}{
		"service":           "openai",
		"response_id":       resp.ID,
		"model":             resp.Model,
		"prompt_tokens":     resp.Usage.PromptTokens,
		"completion_tokens": resp.Usage.CompletionTokens,
		"total_tokens":      resp.Usage.TotalTokens,
		"choices_count":     len(resp.Choices),
	}).Info("OpenAI API response received")

	// 開發模式下詳細記錄響應內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service":     "openai",
			"response_id": resp.ID,
			"model":       resp.Model,
		}).Info("🎯 OpenAI Response Details")

		for i, choice := range resp.Choices {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "openai",
				"choice_index":   i,
				"finish_reason":  choice.FinishReason,
				"content_length": len(choice.Message.Content),
			}).Info(fmt.Sprintf("💬 Response [%d]: %s", i, choice.Message.Content))
		}
	} else {
		// 生產環境只記錄基本信息
		for i, choice := range resp.Choices {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "openai",
				"choice_index":   i,
				"finish_reason":  choice.FinishReason,
				"content_length": len(choice.Message.Content),
			}).Debug("OpenAI response choice")
		}
	}

	// 轉換回應格式
	response := &OpenAIResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	// 轉換選項
	for _, choice := range resp.Choices {
		response.Choices = append(response.Choices, struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			Index: choice.Index,
			Message: struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: string(choice.FinishReason),
		})
	}

	return response, nil
}

// BuildCharacterPrompt 構建角色提示詞
func (c *OpenAIClient) BuildCharacterPrompt(characterID, userMessage string, conversationContext *ConversationContext, nsfwLevel int, chatMode string) []OpenAIMessage {

	// 獲取角色資料
	characterService := GetCharacterService()
	ctx := context.Background()
	dbCharacter, err := characterService.GetCharacterDB(ctx, characterID)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to get character for prompt building")
		return nil
	}

	// 使用OpenAI專屬的prompt構建器
	promptBuilder := NewOpenAIPromptBuilder(characterService)
	promptBuilder.WithCharacter(dbCharacter)
	promptBuilder.WithContext(conversationContext)
	promptBuilder.WithNSFWLevel(nsfwLevel)
	promptBuilder.WithUserMessage(userMessage)
	promptBuilder.WithChatMode(chatMode)
	systemPrompt := promptBuilder.Build()

	messages := []OpenAIMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// 添加對話歷史（最近幾條）
    if conversationContext != nil {
        // 僅保留最近2則歷史（舊 -> 新）
        count := len(conversationContext.RecentMessages)
        if count > 2 { count = 2 }
        for i := count - 1; i >= 0; i-- {
            msg := conversationContext.RecentMessages[i]
            messages = append(messages, OpenAIMessage{Role: msg.Role, Content: msg.Content})
        }
    }

    // 添加當前用戶消息（避免與歷史的最後一則用戶訊息重複）
    shouldAppendUser := true
    if conversationContext != nil && len(conversationContext.RecentMessages) > 0 {
        latest := conversationContext.RecentMessages[0] // 最新在前
        if latest.Role == "user" && strings.TrimSpace(latest.Content) == strings.TrimSpace(userMessage) {
            shouldAppendUser = false
        }
    }
    if shouldAppendUser {
        messages = append(messages, OpenAIMessage{Role: "user", Content: userMessage})
    }

	return messages
}
