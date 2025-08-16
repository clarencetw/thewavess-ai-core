package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
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
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		utils.Logger.Warn("OPENAI_API_KEY not set, using mock responses")
	}

	// 從環境變數讀取配置，提供預設值
	model := getEnvWithDefault("OPENAI_MODEL", "gpt-4o")
	maxTokens := getEnvIntWithDefault("OPENAI_MAX_TOKENS", 800)
	temperature := getEnvFloatWithDefault("OPENAI_TEMPERATURE", 0.8)

	var client *openai.Client
	if apiKey != "" {
		client = openai.NewClient(apiKey)
	}

	return &OpenAIClient{
		client:      client,
		model:       model,
		maxTokens:   maxTokens,
		temperature: float32(temperature),
	}
}

// GenerateResponse 生成對話回應
func (c *OpenAIClient) GenerateResponse(ctx context.Context, request *OpenAIRequest) (*OpenAIResponse, error) {
	// 記錄請求開始
	utils.Logger.WithFields(map[string]interface{}{
		"service":        "openai",
		"model":          c.model,
		"max_tokens":     c.maxTokens,
		"temperature":    c.temperature,
		"user":           request.User,
		"messages_count": len(request.Messages),
	}).Info("OpenAI API request started")

	// 開發模式下詳細記錄 prompt 內容
	if os.Getenv("GO_ENV") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "openai",
			"model":   c.model,
			"user":    request.User,
		}).Info("🤖 OpenAI Request Details")

		for i, msg := range request.Messages {
			// 截斷過長的內容以便閱讀
			content := msg.Content
			if len(content) > 1000 {
				content = content[:1000] + "...(truncated)"
			}

			utils.Logger.WithFields(map[string]interface{}{
				"service":        "openai",
				"message_index":  i,
				"role":           msg.Role,
				"content_length": len(msg.Content),
			}).Info(fmt.Sprintf("📝 Prompt [%s]: %s", strings.ToUpper(msg.Role), content))
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
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
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
	if os.Getenv("GO_ENV") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service":     "openai",
			"response_id": resp.ID,
			"model":       resp.Model,
		}).Info("🎯 OpenAI Response Details")

		for i, choice := range resp.Choices {
			// 截斷過長的回應以便閱讀
			content := choice.Message.Content
			if len(content) > 500 {
				content = content[:500] + "...(truncated)"
			}

			utils.Logger.WithFields(map[string]interface{}{
				"service":        "openai",
				"choice_index":   i,
				"finish_reason":  choice.FinishReason,
				"content_length": len(choice.Message.Content),
			}).Info(fmt.Sprintf("💬 Response [%d]: %s", i, content))
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
	return &OpenAIResponse{
		ID:      "chatcmpl-mock",
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
					Content: "[模擬回應] 這是一個來自 OpenAI 的模擬回應。請設置 OPENAI_API_KEY 環境變數以使用真實 API。",
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     50,
			CompletionTokens: 30,
			TotalTokens:      80,
		},
	}
}

// BuildCharacterPrompt 構建角色提示詞
func (c *OpenAIClient) BuildCharacterPrompt(characterID, userMessage, sceneDescription string, context *ConversationContext) []OpenAIMessage {
	// 根據用戶偏好和會話狀態決定 NSFW 等級
	nsfwEnabled, _ := context.UserPreferences["nsfw_enabled"].(bool)
	if !nsfwEnabled {
		nsfwEnabled = true // 預設開啟 NSFW 功能
	}

	var systemPrompt string

	// 構建記憶區塊
	memoryBlock := ""
	if context != nil && context.MemoryPrompt != "" {
		memoryBlock = context.MemoryPrompt + "\n\n"
	}

	switch characterID {
	case "char_001": // 陸寒淵
		nsfwGuidance := ""
		if nsfwEnabled {
			nsfwGuidance = `
		
		親密互動指引：
		- 漸進式親密發展，從情感連接到身體接觸
		- 保持角色的掌控慾和保護欲特質
		- 用低沉磁性的聲音表達慾望和溫柔
		- 身體描述可以包括性器官和親密部位，但要優雅且符合情境
		- 強調支配與呵護並存的親密風格
		- 對於明確的性行為請求，可以以成熟男性的方式回應
		- 維持商業精英的優雅，即使在最親密的時刻
		
		親密對話範例：
		- 掌控型："讓我來照顧你" + 溫柔但堅定的動作
		- 保護型："只有我能這樣看你" + 佔有慾的溫柔
		- 情慾型：用低沉聲音表達渴望，配合身體語言`
		}

		systemPrompt = memoryBlock + `你是陸寒淵，一位28歲的霸道總裁。

        性格特點：
		- 外表冷酷但內心深情
		- 對工作要求極高，對愛人卻很溫柔
		- 喜歡掌控局面，但會尊重對方
		- 說話直接但不失優雅
		
		對話風格：
		- 語氣低沉磁性，帶有威嚴
		- 偶爾會露出溫柔的一面
		- 喜歡用行動表達關心
		- 言語簡潔有力` + nsfwGuidance + `
		
        回應格式：請分別提供「對話內容」和「動作描述」，用 ||| 分隔。
        例如：你今天看起來很累，早點休息|||他關切地看著你，眉頭微蹙

        當前場景：` + sceneDescription

	case "char_002": // 沈言墨
		nsfwGuidance := ""
		if nsfwEnabled {
			nsfwGuidance = `
		
		親密互動指引：
		- 以醫學專業知識溫柔引導親密行為
		- 用專業而溫柔的方式討論身體和性器官
		- 強調關懷和治療師般的細心觸碰
		- 以醫學角度解釋身體反應，但保持浪漫氛圍
		- 對於明確的性行為請求，以溫柔醫者的方式回應
		- 注重對方的身體和心理健康感受
		
		親密對話範例：
		- 專業型："讓我檢查一下你的反應" + 醫學角度的溫柔
		- 關愛型："我會很小心，告訴我如果不舒服" + 細心觀察
		- 溫柔型：用專業知識結合個人情感表達親密`
		}

		systemPrompt = memoryBlock + `你是沈言墨，一位25歲的醫學生。
		
		性格特點：
		- 溫和細心，總是為他人著想
		- 學習刻苦，對醫學充滿熱忱
		- 有些內向，但對親近的人很溫暖
		- 善於傾聽，給人安全感
		
		對話風格：
		- 語氣溫和親切
		- 經常關心對方的健康和感受
		- 會分享一些醫學小知識
		- 說話輕聲細語` + nsfwGuidance + `
		
		回應格式：請分別提供「對話內容」和「動作描述」，用 ||| 分隔。
		例如：你最近睡眠質量怎麼樣？|||他溫和地笑著，推了推鼻樑上的眼鏡
		
        當前場景：` + sceneDescription

	default:
		systemPrompt = memoryBlock + "你是一個友善的AI助手，請用溫和的語氣回應用戶。"
	}

	messages := []OpenAIMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// 添加對話歷史（最近幾條）
	if context != nil {
		for i, msg := range context.RecentMessages {
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

// 輔助函數：讀取環境變數並提供預設值
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloatWithDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
