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
	isAzure     bool
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
		utils.Logger.Warn("OPENAI_API_KEY not set, using mock responses")
	}

	// 從環境變數讀取配置，提供預設值
	model := utils.GetEnvWithDefault("OPENAI_MODEL", "gpt-4o")
	maxTokens := utils.GetEnvIntWithDefault("OPENAI_MAX_TOKENS", 1200)
	temperature := utils.GetEnvFloatWithDefault("OPENAI_TEMPERATURE", 0.8)

	// 獲取自定義 API URL，支援 Azure 或其他端點
	baseURL := utils.GetEnvWithDefault("OPENAI_API_URL", "https://api.openai.com/v1")
	isAzure := false

	// 檢查是否為 Azure OpenAI
	if strings.Contains(baseURL, "azure.com") {
		isAzure = true
		// Azure OpenAI 需要特殊的 URL 和配置處理
		// 保持原始 baseURL，讓 go-openai 庫處理具體的端點路徑
	}

	var client *openai.Client
	if apiKey != "" {
		if isAzure {
			// Azure OpenAI 需要特殊配置 - 使用 DefaultAzureConfig
			config := openai.DefaultAzureConfig(apiKey, baseURL)
			// Azure 需要部署名稱，通常就是模型名稱
			config.AzureModelMapperFunc = func(model string) string {
				return model // 使用模型名稱作為部署名稱
			}
			client = openai.NewClientWithConfig(config)

			utils.Logger.WithFields(map[string]interface{}{
				"base_url":    baseURL,
				"api_type":    "azure",
				"api_version": config.APIVersion,
			}).Info("Using Azure OpenAI API")
		} else if baseURL != "https://api.openai.com/v1" {
			// 其他自定義端點
			config := openai.DefaultConfig(apiKey)
			config.BaseURL = baseURL
			client = openai.NewClientWithConfig(config)

			utils.Logger.WithField("base_url", baseURL).Info("Using custom OpenAI API URL")
		} else {
			// 使用默認 OpenAI API
			client = openai.NewClient(apiKey)
		}
	}

	return &OpenAIClient{
		client:      client,
		model:       model,
		maxTokens:   maxTokens,
		temperature: float32(temperature),
		baseURL:     baseURL,
		isAzure:     isAzure,
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

	// 如果沒有 API key，返回模擬回應
	if c.client == nil {
		utils.Logger.WithField("service", "openai").Info("Using mock response (API key not set)")
		return c.generateMockResponse(request), nil
	}

	// 轉換消息格式
	messages := make([]openai.ChatCompletionMessage, len(request.Messages))
	for i, msg := range request.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 調用 OpenAI API
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
		User:        request.User,
	})

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

// generateMockResponse 生成模擬回應（當 API key 未設置時）
func (c *OpenAIClient) generateMockResponse(request *OpenAIRequest) *OpenAIResponse {
	// 分析整個對話上下文生成智能回應
	var mockContent string
	userMessage := ""
	systemPrompt := ""
	
	// 獲取用戶消息和system prompt
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
		
		// 分析角色和場景
		isNSFW := strings.Contains(systemPrompt, "level") && (strings.Contains(systemPrompt, "3") || strings.Contains(systemPrompt, "4"))
		
		// 根據關鍵詞和場景生成符合女性向風格的回應
		if strings.Contains(userMessage, "你好") || strings.Contains(userMessage, "嗨") {
			if isNSFW {
				mockContent = "你好...很高興又見到你了。今天想要怎麼度過呢？"
			} else {
				mockContent = "你好呢～很高興見到你。今天過得怎麼樣？"
			}
		} else if strings.Contains(userMessage, "累") || strings.Contains(userMessage, "疲憊") {
			mockContent = "辛苦了...來我這裡休息一下吧。我會一直陪在你身邊的。"
		} else if strings.Contains(userMessage, "開心") || strings.Contains(userMessage, "高興") {
			mockContent = "看到你這麼開心，我也跟著開心起來了呢～能分享一下是什麼好事嗎？"
		} else if strings.Contains(userMessage, "愛") {
			if isNSFW {
				mockContent = "我也愛你...讓我用行動證明我的心意吧。"
			} else {
				mockContent = "我的心裡也有著同樣溫暖的感受...你對我來說很特別。"
			}
		} else {
			// 默認根據場景回應
			if isNSFW {
				mockContent = "我明白你的想法...在這個只屬於我們的空間裡，我會好好照顧你。"
			} else {
				mockContent = "我明白你想說的...無論何時，我都會認真聆聽你的心聲。"
			}
		}
	} else {
		mockContent = "很高興能與你對話...有什麼想聊的嗎？"
	}
	
	return &OpenAIResponse{
		ID:      fmt.Sprintf("chatcmpl-mock-%d", 1234567890),
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   c.model,
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
			PromptTokens:     len(userMessage) / 4,
			CompletionTokens: len(mockContent) / 4,
			TotalTokens:      (len(userMessage) + len(mockContent)) / 4,
		},
	}
}

// BuildCharacterPrompt 構建角色提示詞（使用統一模板）
func (c *OpenAIClient) BuildCharacterPrompt(characterID, userMessage string, conversationContext *ConversationContext, nsfwLevel int, chatMode string) []OpenAIMessage {

	// 使用OpenAI專屬的prompt構建器
	characterService := GetCharacterService()
	promptBuilder := NewOpenAIPromptBuilder(characterService)
	ctx := context.Background()
	systemPrompt := promptBuilder.
		WithCharacter(ctx, characterID).
		WithContext(conversationContext).
		WithNSFWLevel(nsfwLevel).
		WithUserMessage(userMessage).
		WithChatMode(chatMode).
		Build(ctx)

	messages := []OpenAIMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// 添加對話歷史（最近幾條）
	if conversationContext != nil {
		for i, msg := range conversationContext.RecentMessages {
			if i >= 5 { // 只保留最近5條消息
				break
			}
			messages = append(messages, OpenAIMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 添加當前用戶消息
	messages = append(messages, OpenAIMessage{
		Role:    "user",
		Content: userMessage,
	})

	return messages
}
