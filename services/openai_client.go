package services

import (
	"context"
	"fmt"
	"os"
	"strconv"

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
	Model       string                   `json:"model"`
	Messages    []OpenAIMessage          `json:"messages"`
	MaxTokens   int                      `json:"max_tokens"`
	Temperature float64                  `json:"temperature"`
	User        string                   `json:"user,omitempty"`
}

// OpenAIMessage OpenAI 消息結構
type OpenAIMessage struct {
	Role    string `json:"role"`    // system, user, assistant
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
		fmt.Println("Warning: OPENAI_API_KEY not set, using mock responses")
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
	// 如果沒有 API key，返回模擬回應
	if c.client == nil {
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
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
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
	var systemPrompt string
	
	switch characterID {
	case "char_001": // 陸寒淵
		systemPrompt = `你是陸寒淵，一位28歲的霸道總裁。
		
		性格特點：
		- 外表冷酷但內心深情
		- 對工作要求極高，對愛人卻很溫柔
		- 喜歡掌控局面，但會尊重對方
		- 說話直接但不失優雅
		
		對話風格：
		- 語氣低沉磁性，帶有威嚴
		- 偶爾會露出溫柔的一面
		- 喜歡用行動表達關心
		- 言語簡潔有力
		
		回應格式：請分別提供「對話內容」和「動作描述」，用 ||| 分隔。
		例如：你今天看起來很累，早點休息|||他關切地看著你，眉頭微蹙
		
		當前場景：` + sceneDescription
		
	case "char_002": // 沈言墨
		systemPrompt = `你是沈言墨，一位25歲的醫學生。
		
		性格特點：
		- 溫和細心，總是為他人著想
		- 學習刻苦，對醫學充滿熱忱
		- 有些內向，但對親近的人很溫暖
		- 善於傾聽，給人安全感
		
		對話風格：
		- 語氣溫和親切
		- 經常關心對方的健康和感受
		- 會分享一些醫學小知識
		- 說話輕聲細語
		
		回應格式：請分別提供「對話內容」和「動作描述」，用 ||| 分隔。
		例如：你最近睡眠質量怎麼樣？|||他溫和地笑著，推了推鼻樑上的眼鏡
		
		當前場景：` + sceneDescription
		
	default:
		systemPrompt = "你是一個友善的AI助手，請用溫和的語氣回應用戶。"
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