package services

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/shared"
)

// MistralClient Mistral AI API 客戶端 (使用 OpenAI SDK)
type MistralClient struct {
	client      openai.Client
	model       string
	maxTokens   int
	temperature float64
	baseURL     string
}

// MistralRequest Mistral 請求結構 (相容 OpenAI 格式)
type MistralRequest struct {
	Model       string           `json:"model"`
	Messages    []MistralMessage `json:"messages"`
	MaxTokens   int              `json:"max_tokens"`
	Temperature float64          `json:"temperature"`
	User        string           `json:"user,omitempty"`
}

// MistralMessage Mistral 消息結構
type MistralMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// MistralResponse 使用官方 OpenAI SDK 的 ChatCompletion 作為響應類型
type MistralResponse = openai.ChatCompletion

// NewMistralClient 創建新的 Mistral 客戶端 (使用 OpenAI SDK)
func NewMistralClient() *MistralClient {
	// 確保環境變數已載入
	utils.LoadEnv()

	apiKey := utils.GetEnvWithDefault("MISTRAL_API_KEY", "")
	if apiKey == "" {
		utils.Logger.Fatal("MISTRAL_API_KEY is required but not set in environment")
	}

	// 從環境變數讀取配置
	modelName := utils.GetEnvWithDefault("MISTRAL_MODEL", "mistral-medium-latest")
	maxTokens := utils.GetEnvIntWithDefault("MISTRAL_MAX_TOKENS", 1200)
	temperature := utils.GetEnvFloatWithDefault("MISTRAL_TEMPERATURE", 0.8)

	// 獲取 Mistral API URL
	baseURL := utils.GetEnvWithDefault("MISTRAL_API_URL", "https://api.mistral.ai/v1")

	// 準備客戶端選項
	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	}

	// 創建 OpenAI 客戶端，使用 Mistral 端點
	client := openai.NewClient(options...)

	utils.Logger.WithField("base_url", baseURL).Info("Using Mistral API with OpenAI SDK")

	return &MistralClient{
		client:      client,
		model:       modelName,
		maxTokens:   maxTokens,
		temperature: temperature,
		baseURL:     baseURL,
	}
}

// GenerateResponse 生成對話回應 (使用 OpenAI SDK + Structured Output)
func (c *MistralClient) GenerateResponse(ctx context.Context, request *MistralRequest) (*MistralResponse, error) {
	startTime := time.Now()

	// 記錄請求開始
	utils.Logger.WithFields(map[string]interface{}{
		"service":        "mistral",
		"base_url":       c.baseURL,
		"model":          request.Model,
		"max_tokens":     request.MaxTokens,
		"temperature":    request.Temperature,
		"user":           request.User,
		"messages_count": len(request.Messages),
	}).Info("Mistral API request started")

	// 開發模式下詳細記錄 prompt 內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "mistral",
			"model":   request.Model,
			"user":    request.User,
		}).Info("🤖 Mistral Request Details")

		for i, msg := range request.Messages {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "mistral",
				"message_index":  i,
				"role":           msg.Role,
				"content_length": len(msg.Content),
			}).Info(fmt.Sprintf("📝 Prompt [%s]: %s", msg.Role, msg.Content))
		}
	}

	// 設置默認值
	if request.Model == "" {
		request.Model = c.model
	}
	if request.MaxTokens == 0 {
		request.MaxTokens = c.maxTokens
	}
	if request.Temperature <= 0 {
		request.Temperature = c.temperature
	}

	// 轉換為 OpenAI SDK 格式
	messages := make([]openai.ChatCompletionMessageParamUnion, len(request.Messages))
	for i, msg := range request.Messages {
		messages[i] = openai.UserMessage(msg.Content)
		switch msg.Role {
		case "system":
			messages[i] = openai.SystemMessage(msg.Content)
		case "assistant":
			messages[i] = openai.AssistantMessage(msg.Content)
		}
	}

	// 構建請求參數
	params := openai.ChatCompletionNewParams{
		Model:       openai.ChatModel(request.Model),
		Messages:    messages,
		MaxTokens:   openai.Int(int64(request.MaxTokens)),
		Temperature: openai.Float(request.Temperature),
	}

	if request.User != "" {
		params.User = openai.String(request.User)
	}

	// 設置 Mistral 的 JSON Schema (與 OpenAI 相同格式)
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"content": map[string]interface{}{
				"type":        "string",
				"description": "角色回應內容，包含動作描述和對話",
			},
			"emotion_delta": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"affection_change": map[string]interface{}{
						"type":        "integer",
						"description": "好感度變化，必須是整數",
						"minimum":     -5,
						"maximum":     5,
					},
				},
				"required":             []string{"affection_change"},
				"additionalProperties": false,
			},
			"mood": map[string]interface{}{
				"type": "string",
				"enum": []string{
					"neutral", "happy", "excited", "shy", "romantic",
					"passionate", "pleased", "loving", "friendly",
					"polite", "concerned", "annoyed", "upset", "disappointed",
				},
				"description": "角色當前情緒狀態",
			},
			"relationship": map[string]interface{}{
				"type": "string",
				"enum": []string{"stranger", "friend", "close_friend", "lover", "soulmate"},
				"description": "角色與用戶的關係狀態",
			},
			"intimacy_level": map[string]interface{}{
				"type": "string",
				"enum": []string{"distant", "friendly", "close", "intimate", "deeply_intimate"},
				"description": "親密度層級",
			},
			"reasoning": map[string]interface{}{
				"type":        "string",
				"description": "決策推理說明",
			},
		},
		"required":             []string{"content", "emotion_delta", "mood", "relationship", "intimacy_level", "reasoning"},
		"additionalProperties": false,
	}

	jsonSchemaParam := shared.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "character_response",
		Description: openai.String("角色對話回應格式"),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}

	params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
			Type:       "json_schema",
			JSONSchema: jsonSchemaParam,
		},
	}

	// 發送請求
	utils.Logger.WithFields(map[string]interface{}{
		"service": "mistral",
		"model":   request.Model,
	}).Info("Sending Mistral API request via OpenAI SDK")

	resp, err := c.client.Chat.Completions.New(ctx, params, option.WithRequestTimeout(60*time.Second))
	if err != nil {
		// 記錄詳細的錯誤信息用於診斷
		utils.Logger.WithFields(map[string]interface{}{
			"service":        "mistral",
			"error":          err.Error(),
			"error_type":     fmt.Sprintf("%T", err),
			"model":          request.Model,
			"base_url":       c.baseURL,
			"max_tokens":     request.MaxTokens,
			"temperature":    request.Temperature,
			"messages_count": len(request.Messages),
			"request_time":   time.Since(startTime),
		}).Error("Mistral API call failed")

		// 檢查是否是超時錯誤
		if ctx.Err() == context.DeadlineExceeded {
			utils.Logger.WithFields(map[string]interface{}{
				"service":      "mistral",
				"timeout_type": "context_deadline",
				"elapsed":      time.Since(startTime),
			}).Error("Mistral API 請求超時")
		}

		return nil, fmt.Errorf("failed Mistral API call: %w", err)
	}

	// 計算響應時間
	duration := time.Since(startTime)

	// 計算 Mistral API 成本 (多模型支援)
	promptTokens := int(resp.Usage.PromptTokens)
	completionTokens := int(resp.Usage.CompletionTokens)

	// Mistral 定價系統 (per 1M tokens)
	var inputCostPer1M, outputCostPer1M float64

	switch string(resp.Model) {
	// Mistral Small series
	case "mistral-small-latest", "mistral-small-3.2", "mistral-small":
		inputCostPer1M = 0.10  // $0.10 per 1M input tokens
		outputCostPer1M = 0.30 // $0.30 per 1M output tokens

	// Mistral Medium series
	case "mistral-medium-latest", "mistral-medium-3", "mistral-medium":
		inputCostPer1M = 0.40  // $0.40 per 1M input tokens
		outputCostPer1M = 2.00 // $2.00 per 1M output tokens

	// Mistral Large series
	case "mistral-large-latest", "mistral-large", "mistral-large-2":
		inputCostPer1M = 2.00  // $2.00 per 1M input tokens
		outputCostPer1M = 6.00 // $6.00 per 1M output tokens

	// Magistral series (thinking models)
	case "magistral-small-latest", "magistral-small":
		inputCostPer1M = 0.50  // $0.50 per 1M input tokens
		outputCostPer1M = 1.50 // $1.50 per 1M output tokens

	case "magistral-medium-latest", "magistral-medium":
		inputCostPer1M = 2.00  // $2.00 per 1M input tokens
		outputCostPer1M = 5.00 // $5.00 per 1M output tokens

	// Legacy models
	case "mistral-7b-instruct", "mistral-8x7b-instruct":
		inputCostPer1M = 0.25  // Legacy pricing
		outputCostPer1M = 0.25

	default:
		// Default to Small pricing for unknown models
		inputCostPer1M = 0.10
		outputCostPer1M = 0.30
		utils.Logger.WithField("model", resp.Model).Warn("Unknown Mistral model, using Small pricing")
	}

	inputCost := float64(promptTokens) * inputCostPer1M / 1000000
	outputCost := float64(completionTokens) * outputCostPer1M / 1000000
	totalCost := inputCost + outputCost

	// 記錄成功響應，包含詳細成本資訊
	utils.Logger.WithFields(map[string]interface{}{
		"service":            "mistral",
		"response_id":        resp.ID,
		"model":              resp.Model,
		"prompt_tokens":      resp.Usage.PromptTokens,
		"completion_tokens":  resp.Usage.CompletionTokens,
		"total_tokens":       resp.Usage.TotalTokens,
		"input_cost_usd":     fmt.Sprintf("$%.6f", inputCost),
		"output_cost_usd":    fmt.Sprintf("$%.6f", outputCost),
		"total_cost_usd":     fmt.Sprintf("$%.6f", totalCost),
		"input_rate_per_1m":  fmt.Sprintf("$%.2f", inputCostPer1M),
		"output_rate_per_1m": fmt.Sprintf("$%.2f", outputCostPer1M),
		"choices_count":      len(resp.Choices),
		"duration_ms":        duration.Milliseconds(),
	}).Info("Mistral API response received")

	// 開發模式下詳細記錄響應內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service":     "mistral",
			"response_id": resp.ID,
			"model":       resp.Model,
		}).Info("🎯 Mistral Response Details")

		for i, choice := range resp.Choices {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "mistral",
				"choice_index":   i,
				"finish_reason":  string(choice.FinishReason),
				"content_length": len(choice.Message.Content),
			}).Info(fmt.Sprintf("💬 Response [%d]: %s", i, choice.Message.Content))
		}
	}

	return resp, nil
}