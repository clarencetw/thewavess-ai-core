package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// GrokClient Grok 客戶端 (使用 OpenAI SDK)
type GrokClient struct {
	client      openai.Client
	model       string // Grok 模型名稱 (string 類型以支援自定義模型)
	maxTokens   int
	temperature float64
	baseURL     string
}

// GrokRequest Grok 請求結構 (相容 OpenAI 格式)
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

// GrokResponse 使用官方 OpenAI SDK 的 ChatCompletion 作為響應類型
type GrokResponse = openai.ChatCompletion

// NewGrokClient 創建新的 Grok 客戶端 (使用 OpenAI SDK)
func NewGrokClient() *GrokClient {
	// 確保環境變數已載入
	utils.LoadEnv()

	apiKey := utils.GetEnvWithDefault("GROK_API_KEY", "")
	if apiKey == "" {
		utils.Logger.Fatal("GROK_API_KEY is required but not set in environment")
	}

	// 從環境變數讀取配置
	modelName := utils.GetEnvWithDefault("GROK_MODEL", "grok-beta")
	maxTokens := utils.GetEnvIntWithDefault("GROK_MAX_TOKENS", 2000)
	temperature := utils.GetEnvFloatWithDefault("GROK_TEMPERATURE", 0.7)

	// 獲取 API URL
	baseURL := utils.GetEnvWithDefault("GROK_API_URL", "https://api.x.ai/v1")

	// 創建 OpenAI 客戶端，使用 xAI 端點
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	utils.Logger.WithField("base_url", baseURL).Info("Using xAI Grok API with OpenAI SDK")

	return &GrokClient{
		client:      client,
		model:       modelName,
		maxTokens:   maxTokens,
		temperature: temperature,
		baseURL:     baseURL,
	}
}

// GenerateResponse 生成對話回應 (使用 OpenAI SDK)
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

	// 構建請求參數 (Grok 使用自定義模型名稱)
	params := openai.ChatCompletionNewParams{
		Model:       openai.ChatModel(request.Model),
		Messages:    messages,
		MaxTokens:   openai.Int(int64(request.MaxTokens)),
		Temperature: openai.Float(request.Temperature),
	}

	if request.User != "" {
		params.User = openai.String(request.User)
	}

	// 發送請求
	utils.Logger.WithFields(map[string]interface{}{
		"service": "grok",
		"model":   request.Model,
	}).Info("Sending Grok API request via OpenAI SDK")

	resp, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "grok",
			"error":   err.Error(),
			"model":   request.Model,
		}).Error("Grok API call failed")
		return nil, fmt.Errorf("failed Grok API call: %w", err)
	}

	// 計算響應時間
	duration := time.Since(startTime)

	// 計算 Grok API 成本 (保留現有邏輯)
	promptTokens := int(resp.Usage.PromptTokens)
	completionTokens := int(resp.Usage.CompletionTokens)

	var inputCostPer1M, outputCostPer1M float64
	switch resp.Model {
	case "grok-4-0709":
		inputCostPer1M = 3.00   // $3.00 per 1M input tokens
		outputCostPer1M = 15.00 // $15.00 per 1M output tokens
	case "grok-4-fast-reasoning", "grok-4-fast", "grok-4-fast-reasoning-latest":
		inputCostPer1M = 0.20  // $0.20 per 1M input tokens
		outputCostPer1M = 0.50 // $0.50 per 1M output tokens
	case "grok-4-fast-non-reasoning", "grok-4-fast-non-reasoning-latest", "grok-4-mini-non-reasoning-latest":
		inputCostPer1M = 0.20  // $0.20 per 1M input tokens
		outputCostPer1M = 0.50 // $0.50 per 1M output tokens
	case "grok-3", "grok-3-latest", "grok-3-beta", "grok-3-fast", "grok-3-fast-latest", "grok-3-fast-beta":
		inputCostPer1M = 3.00   // $3.00 per 1M input tokens
		outputCostPer1M = 15.00 // $15.00 per 1M output tokens
	case "grok-3-mini":
		inputCostPer1M = 0.30  // $0.30 per 1M input tokens
		outputCostPer1M = 0.50 // $0.50 per 1M output tokens
	case "grok-2-vision-1212":
		inputCostPer1M = 2.00   // $2.00 per 1M input tokens
		outputCostPer1M = 10.00 // $10.00 per 1M output tokens
	case "grok-code-fast-1":
		inputCostPer1M = 0.20  // $0.20 per 1M input tokens
		outputCostPer1M = 1.50 // $1.50 per 1M output tokens
	default:
		// Default to grok-3 pricing for unknown models
		inputCostPer1M = 3.00
		outputCostPer1M = 15.00
	}

	inputCost := float64(promptTokens) * inputCostPer1M / 1000000
	outputCost := float64(completionTokens) * outputCostPer1M / 1000000
	totalCost := inputCost + outputCost

	// 記錄成功響應，包含詳細成本資訊
	utils.Logger.WithFields(map[string]interface{}{
		"service":            "grok",
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
	}).Info("Grok API response received")

	// 開發模式下詳細記錄響應內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service":     "grok",
			"response_id": resp.ID,
			"model":       resp.Model,
		}).Info("🎯 Grok Response Details")

		for i, choice := range resp.Choices {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "grok",
				"choice_index":   i,
				"finish_reason":  string(choice.FinishReason),
				"content_length": len(choice.Message.Content),
			}).Info(fmt.Sprintf("💬 Response [%d]: %s", i, choice.Message.Content))
		}
	}

	return resp, nil
}

