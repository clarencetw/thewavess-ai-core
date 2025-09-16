package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gage-technologies/mistral-go"
	"github.com/sirupsen/logrus"
)

// MistralClient Mistral AI API 客戶端
type MistralClient struct {
    client *mistral.MistralClient
    config *MistralConfig
}

// MistralConfig Mistral 配置
type MistralConfig struct {
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
}

// MistralResponse Mistral 回應結構
type MistralResponse struct {
	Content   string                 `json:"content"`
	Usage     *MistralUsage          `json:"usage,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// MistralUsage 使用統計
type MistralUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewMistralClient 創建新的 Mistral 客戶端
func NewMistralClient() *MistralClient {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		utils.Logger.Warn("MISTRAL_API_KEY not found, Mistral client will be disabled")
		return nil
	}

	// 預設配置
	config := &MistralConfig{
		Model:       "mistral-medium-latest", // 中等模型，平衡性能與成本的 NSFW 處理
		MaxTokens:   1200,
		Temperature: 0.8,
		TopP:        0.9,
	}

	client := mistral.NewMistralClientDefault(apiKey)


	return &MistralClient{
		client: client,
		config: config,
	}
}

// GenerateResponse 使用 Mistral 生成回應
func (mc *MistralClient) GenerateResponse(ctx context.Context, systemPrompt, userMessage string, userID string) (*MistralResponse, error) {
	if mc == nil || mc.client == nil {
		return nil, fmt.Errorf("Mistral client not initialized")
	}

	startTime := time.Now()

    utils.Logger.WithFields(logrus.Fields{
        "service":          "mistral",
        "model":            mc.config.Model,
        "max_tokens":       mc.config.MaxTokens,
        "temperature":      mc.config.Temperature,
        "user":             userID,
        "messages_count":   2,
        "system_length":    len(systemPrompt),
        "user_length":      len(userMessage),
    }).Info("Mistral API request started")

	// 開發模式下詳細記錄 prompt 內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(logrus.Fields{
			"service": "mistral",
			"model":   mc.config.Model,
			"user":    userID,
		}).Info("🤖 Mistral Request Details")

		utils.Logger.WithFields(logrus.Fields{
			"service":        "mistral",
			"message_index":  0,
			"role":           "system",
			"content_length": len(systemPrompt),
		}).Info(fmt.Sprintf("📝 Prompt [SYSTEM]: %s", systemPrompt))

		utils.Logger.WithFields(logrus.Fields{
			"service":        "mistral",
			"message_index":  1,
			"role":           "user",
			"content_length": len(userMessage),
		}).Info(fmt.Sprintf("📝 Prompt [USER]: %s", userMessage))
	} else {
		// 生產環境只記錄基本信息
		utils.Logger.WithFields(logrus.Fields{
			"service":        "mistral",
			"message_index":  0,
			"role":           "system",
			"content_length": len(systemPrompt),
		}).Debug("Mistral request message")

		utils.Logger.WithFields(logrus.Fields{
			"service":        "mistral",
			"message_index":  1,
			"role":           "user",
			"content_length": len(userMessage),
		}).Debug("Mistral request message")
	}

	// 構建消息
	messages := []mistral.ChatMessage{
		{
			Role:    mistral.RoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    mistral.RoleUser,
			Content: userMessage,
		},
	}

	// 調用 Mistral API (使用簡化的 API 調用方式)
	response, err := mc.client.Chat(mc.config.Model, messages, nil)
	if err != nil {
		utils.Logger.WithError(err).WithFields(logrus.Fields{
			"model":   mc.config.Model,
			"user_id": userID,
		}).Error("Mistral API 調用失敗")
		return nil, fmt.Errorf("Mistral API call failed: %w", err)
	}

	duration := time.Since(startTime)

	// 提取回應內容
	var content string
	if len(response.Choices) > 0 && response.Choices[0].Message.Content != "" {
		content = response.Choices[0].Message.Content
	} else {
		return nil, fmt.Errorf("Mistral API returned empty content")
	}

	// 構建回應
	mistralResponse := &MistralResponse{
		Content:   content,
		RequestID: response.ID,
		Metadata: map[string]interface{}{
			"model":        response.Model,
			"created":      response.Created,
			"duration_ms":  duration.Milliseconds(),
			"finish_reason": func() string {
				if len(response.Choices) > 0 {
					return string(response.Choices[0].FinishReason)
				}
				return ""
			}(),
		},
	}

	// 添加使用統計
	if response.Usage.PromptTokens > 0 {
		mistralResponse.Usage = &MistralUsage{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		}
	}

	utils.Logger.WithFields(logrus.Fields{
		"service":           "mistral",
		"response_id":       response.ID,
		"model":             response.Model,
		"prompt_tokens":     mistralResponse.Usage.PromptTokens,
		"completion_tokens": mistralResponse.Usage.CompletionTokens,
		"total_tokens":      mistralResponse.Usage.TotalTokens,
		"choices_count":     len(response.Choices),
		"duration_ms":       duration.Milliseconds(),
	}).Info("Mistral API response received")

	// 開發模式下詳細記錄響應內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(logrus.Fields{
			"service":     "mistral",
			"response_id": response.ID,
			"model":       response.Model,
		}).Info("🎯 Mistral Response Details")

		for i, choice := range response.Choices {
			utils.Logger.WithFields(logrus.Fields{
				"service":        "mistral",
				"choice_index":   i,
				"finish_reason":  string(choice.FinishReason),
				"content_length": len(choice.Message.Content),
			}).Info(fmt.Sprintf("💬 Response [%d]: %s", i, choice.Message.Content))
		}
	} else {
		// 生產環境只記錄基本信息
		for i, choice := range response.Choices {
			utils.Logger.WithFields(logrus.Fields{
				"service":        "mistral",
				"choice_index":   i,
				"finish_reason":  string(choice.FinishReason),
				"content_length": len(choice.Message.Content),
			}).Debug("Mistral response choice")
		}
	}

	return mistralResponse, nil
}


// IsContentRejection 檢查是否為 Mistral 內容拒絕錯誤
func (mc *MistralClient) IsContentRejection(err error) bool {
	if err == nil {
		return false
	}

	errorMessage := strings.ToLower(err.Error())

	// Mistral 內容拒絕錯誤關鍵詞
	rejectionKeywords := []string{
		"content policy",
		"safety filter",
		"content filter",
		"inappropriate",
		"cannot generate",
		"unable to provide",
		"content guidelines",
		"safety guidelines",
		"moderation",
		"违反",
		"不当",
		"安全",
		"内容政策",
	}

	for _, keyword := range rejectionKeywords {
		if strings.Contains(errorMessage, keyword) {
			return true
		}
	}

	return false
}

// GetModelInfo 獲取模型信息
func (mc *MistralClient) GetModelInfo() map[string]interface{} {
	if mc == nil || mc.config == nil {
		return map[string]interface{}{
			"available": false,
			"reason":    "client_not_initialized",
		}
	}

	return map[string]interface{}{
		"available":    true,
		"model":        mc.config.Model,
		"max_tokens":   mc.config.MaxTokens,
		"temperature":  mc.config.Temperature,
		"top_p":        mc.config.TopP,
		"supports":     []string{"chat", "moderate_nsfw", "multilingual"},
		"description":  "Mistral AI 中等模型 - 適合處理進階 NSFW 內容",
	}
}

// ValidateConnection 驗證 Mistral 連接
func (mc *MistralClient) ValidateConnection(ctx context.Context) error {
	if mc == nil || mc.client == nil {
		return fmt.Errorf("Mistral client not initialized")
	}

	// 發送測試請求
	testResponse, err := mc.GenerateResponse(ctx,
		"You are a helpful assistant.",
		"Hello, please respond with a simple greeting.",
		"test_user")

	if err != nil {
		return fmt.Errorf("Mistral connection validation failed: %w", err)
	}

	if testResponse.Content == "" {
		return fmt.Errorf("Mistral returned empty response")
	}

	utils.Logger.Info("Mistral 連接驗證成功")
	return nil
}
