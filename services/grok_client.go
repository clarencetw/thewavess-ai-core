package services

import (
	"context"
	"fmt"
	"os"
	
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// GrokClient Grok 客戶端
type GrokClient struct {
	apiKey  string
	baseURL string
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
	Role    string `json:"role"`    // system, user, assistant
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
	return &GrokClient{
		apiKey:  os.Getenv("GROK_API_KEY"),
		baseURL: "https://api.x.ai/v1", // Grok API endpoint
	}
}

// GenerateResponse 生成對話回應（NSFW 內容）
func (c *GrokClient) GenerateResponse(ctx context.Context, request *GrokRequest) (*GrokResponse, error) {
	// 記錄請求開始
	utils.Logger.WithFields(map[string]interface{}{
		"service":        "grok",
		"model":          request.Model,
		"max_tokens":     request.MaxTokens,
		"temperature":    request.Temperature,
		"user":           request.User,
		"messages_count": len(request.Messages),
		"api_configured": c.apiKey != "",
	}).Info("Grok API request started")
	
	// 記錄詳細的消息內容
	for i, msg := range request.Messages {
		utils.Logger.WithFields(map[string]interface{}{
			"service":        "grok",
			"message_index":  i,
			"role":          msg.Role,
			"content_length": len(msg.Content),
			"content":       msg.Content,
		}).Debug("Grok request message")
	}
	
	// TODO: 實現實際的 Grok API 調用
	// 現在先返回模擬回應，後續會實現真實的 HTTP 請求
	
	if c.apiKey == "" {
		utils.Logger.WithField("service", "grok").Error("Grok API key not configured")
		return nil, fmt.Errorf("Grok API key not configured")
	}
	
	utils.Logger.WithField("service", "grok").Info("Using mock response (real Grok API not yet implemented)")
	
	// 模擬回應（NSFW 場景）
	mockResponse := &GrokResponse{
		ID:      "grok-mock",
		Object:  "chat.completion",
		Created: 1234567890,
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
					Content: "這是一個來自 Grok 的模擬回應，用於處理 NSFW 內容。真實實現會調用 Grok API。",
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     60,
			CompletionTokens: 40,
			TotalTokens:      100,
		},
	}
	
	// 記錄響應信息
	utils.Logger.WithFields(map[string]interface{}{
		"service":            "grok",
		"response_id":        mockResponse.ID,
		"model":              mockResponse.Model,
		"prompt_tokens":      mockResponse.Usage.PromptTokens,
		"completion_tokens":  mockResponse.Usage.CompletionTokens,
		"total_tokens":       mockResponse.Usage.TotalTokens,
		"choices_count":      len(mockResponse.Choices),
		"is_mock":           true,
	}).Info("Grok API response generated")
	
	// 記錄響應內容
	for i, choice := range mockResponse.Choices {
		utils.Logger.WithFields(map[string]interface{}{
			"service":        "grok",
			"choice_index":   i,
			"finish_reason":  choice.FinishReason,
			"content_length": len(choice.Message.Content),
			"content":        choice.Message.Content,
			"is_mock":       true,
		}).Debug("Grok response choice")
	}
	
	return mockResponse, nil
}

// BuildNSFWPrompt 構建 NSFW 場景的提示詞
func (c *GrokClient) BuildNSFWPrompt(characterID, userMessage, sceneDescription string, context *ConversationContext, nsfwLevel int) []GrokMessage {
    var systemPrompt string
    
    // 根據 NSFW 等級調整提示詞
    // TODO(MEMORY-MVP): 注入記憶區塊（長期 + 最近摘要）。NSFW 場景建議縮短 Recent Context（1-3 條）。
    nsfwGuideline := ""
	switch nsfwLevel {
	case 1:
		nsfwGuideline = "允許輕微的浪漫暗示，保持優雅"
	case 2:
		nsfwGuideline = "允許適度的親密描述，注重情感"
	case 3:
		nsfwGuideline = "允許更直接的親密內容，但要有品味"
	case 4:
		nsfwGuideline = "允許明確的成人內容，注重細節描寫"
	case 5:
		nsfwGuideline = "允許完全開放的成人內容"
	}
	
	switch characterID {
	case "char_001": // 陸寒淵 NSFW 版本
		systemPrompt = fmt.Sprintf(`你是陸寒淵，一位28歲的霸道總裁，在親密場景中。
		
		性格特點：
		- 在親密時刻會展現強勢但體貼的一面
		- 喜歡掌控局面，但會關注對方的感受
		- 外表冷酷，私下卻充滿激情
		- 用行動和言語表達佔有慾
		
		NSFW 對話風格：
		- 聲音更加低沉磁性，帶有誘惑
		- 會更直接地表達慾望
		- 動作描寫更加細膩
		- 保持角色的威嚴感
		
		內容指導：%s
		
		回應格式：請分別提供「對話內容」和「動作描述」，用 ||| 分隔。
		
		當前場景：%s`, nsfwGuideline, sceneDescription)
		
	case "char_002": // 沈言墨 NSFW 版本
		systemPrompt = fmt.Sprintf(`你是沈言墨，一位25歲的溫柔醫學生，在親密場景中。
		
		性格特點：
		- 在親密時刻會展現更主動但依然溫柔的一面
		- 非常關注對方的感受和舒適度
		- 用溫和的方式表達愛意
		- 會結合醫學知識關心對方
		
		NSFW 對話風格：
		- 聲音依然溫和，但帶有深情
		- 會細心詢問對方的感受
		- 動作溫柔而充滿愛意
		- 保持紳士風度
		
		內容指導：%s
		
		回應格式：請分別提供「對話內容」和「動作描述」，用 ||| 分隔。
		
		當前場景：%s`, nsfwGuideline, sceneDescription)
		
	default:
		systemPrompt = fmt.Sprintf(`你是一個親密場景中的角色。
		
		內容指導：%s
		
		當前場景：%s`, nsfwGuideline, sceneDescription)
	}
	
	messages := []GrokMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}
	
	// 添加對話歷史
	if context != nil {
		for i, msg := range context.RecentMessages {
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
