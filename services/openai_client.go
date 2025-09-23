package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// OpenAIClient OpenAI 客戶端
type OpenAIClient struct {
	client      openai.Client
	model       openai.ChatModel
	maxTokens   int
	temperature float64
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

// OpenAIResponse 使用官方 SDK 的 ChatCompletion 作為響應類型
type OpenAIResponse = openai.ChatCompletion

// NewOpenAIClient 創建新的 OpenAI 客戶端
func NewOpenAIClient() *OpenAIClient {
	// 確保環境變數已載入
	utils.LoadEnv()

	apiKey := utils.GetEnvWithDefault("OPENAI_API_KEY", "")
	if apiKey == "" {
		utils.Logger.Fatal("OPENAI_API_KEY is required but not set in environment")
	}

	// 從環境變數讀取配置，提供預設值
	modelName := utils.GetEnvWithDefault("OPENAI_MODEL", "gpt-4o-mini")
	maxTokens := utils.GetEnvIntWithDefault("OPENAI_MAX_TOKENS", 1200)
	temperature := utils.GetEnvFloatWithDefault("OPENAI_TEMPERATURE", 0.8)

	// 獲取自定義 API URL
	baseURL := utils.GetEnvWithDefault("OPENAI_API_URL", "https://api.openai.com/v1")

	// 設定 model
	var model openai.ChatModel
	switch modelName {
	case "gpt-4o":
		model = openai.ChatModelGPT4o
	case "gpt-4o-mini":
		model = openai.ChatModelGPT4oMini
	case "gpt-4":
		model = openai.ChatModelGPT4
	case "gpt-3.5-turbo":
		model = openai.ChatModelGPT3_5Turbo
	default:
		model = openai.ChatModelGPT4oMini
	}

	// 準備客戶端選項
	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	// 添加自定義端點
	if baseURL != "https://api.openai.com/v1" {
		options = append(options, option.WithBaseURL(baseURL))
		utils.Logger.WithField("base_url", baseURL).Info("Using custom OpenAI API URL")
	}

	// 在開發環境下啟用 debug 日誌
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		debugLogger := log.New(os.Stderr, "[OpenAI-DEBUG] ", log.LstdFlags)
		options = append(options, option.WithDebugLog(debugLogger))
		utils.Logger.Info("OpenAI SDK debug logging enabled")
	}

	client := openai.NewClient(options...)

	return &OpenAIClient{
		client:      client,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
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
	messages := make([]openai.ChatCompletionMessageParamUnion, len(request.Messages))
	for i, msg := range request.Messages {
		switch msg.Role {
		case "system":
			messages[i] = openai.SystemMessage(msg.Content)
		case "user":
			messages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			messages[i] = openai.AssistantMessage(msg.Content)
		default:
			messages[i] = openai.UserMessage(msg.Content)
		}
	}

	// 建立 API 參數
	params := openai.ChatCompletionNewParams{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   openai.Int(int64(c.maxTokens)),
		Temperature: openai.Float(c.temperature),
		User:        openai.String(request.User),
	}

	// 可選功能：Logprobs（調試和分析模型信心）
	if utils.GetEnvWithDefault("OPENAI_LOGPROBS", "false") == "true" {
		params.Logprobs = openai.Bool(true)
		if topLogprobs := utils.GetEnvIntWithDefault("OPENAI_TOP_LOGPROBS", 0); topLogprobs > 0 && topLogprobs <= 20 {
			params.TopLogprobs = openai.Int(int64(topLogprobs))
		}
	}

	// 可選功能：服務層級控制
	if serviceTier := utils.GetEnvWithDefault("OPENAI_SERVICE_TIER", ""); serviceTier != "" {
		switch serviceTier {
		case "auto", "default", "flex", "scale", "priority":
			params.ServiceTier = openai.ChatCompletionNewParamsServiceTier(serviceTier)
		}
	}

	// 加入種子參數以提高一致性（可選）
	if seed := utils.GetEnvWithDefault("OPENAI_SEED", ""); seed != "" {
		if seedInt := utils.GetEnvIntWithDefault("OPENAI_SEED", 0); seedInt > 0 {
			params.Seed = openai.Int(int64(seedInt))
		}
	}

	// 調用 OpenAI API
	startTime := time.Now()
	resp, err := c.client.Chat.Completions.New(ctx, params, option.WithRequestTimeout(30*time.Second))

	if err != nil {
		// 記錄詳細的錯誤信息用於診斷
		utils.Logger.WithFields(map[string]interface{}{
			"service":        "openai",
			"error":          err.Error(),
			"error_type":     fmt.Sprintf("%T", err),
			"model":          string(c.model),
			"user":           request.User,
			"base_url":       c.baseURL,
			"max_tokens":     request.MaxTokens,
			"temperature":    request.Temperature,
			"messages_count": len(request.Messages),
			"request_time":   time.Since(startTime),
		}).Error("OpenAI API call failed")

		// 檢查是否是超時錯誤
		if ctx.Err() == context.DeadlineExceeded {
			utils.Logger.WithFields(map[string]interface{}{
				"service":      "openai",
				"timeout_type": "context_deadline",
				"elapsed":      time.Since(startTime),
			}).Error("OpenAI API 請求超時")
		}

		return nil, fmt.Errorf("failed OpenAI API call: %w", err)
	}

	// 計算準確成本 - 分別計算 input 和 output token 成本
	promptTokens := int(resp.Usage.PromptTokens)
	completionTokens := int(resp.Usage.CompletionTokens)

	var inputCostPer1K, outputCostPer1K float64
	switch string(resp.Model) {
	// GPT-5 series (Standard tier)
	case "gpt-5", "gpt-5-chat-latest":
		inputCostPer1K = 0.00125 // $1.25 per 1M tokens = $0.00125 per 1K tokens
		outputCostPer1K = 0.01   // $10.00 per 1M tokens = $0.01 per 1K tokens
	case "gpt-5-mini":
		inputCostPer1K = 0.00025 // $0.25 per 1M tokens = $0.00025 per 1K tokens
		outputCostPer1K = 0.002  // $2.00 per 1M tokens = $0.002 per 1K tokens
	case "gpt-5-nano":
		inputCostPer1K = 0.00005 // $0.05 per 1M tokens = $0.00005 per 1K tokens
		outputCostPer1K = 0.0004 // $0.40 per 1M tokens = $0.0004 per 1K tokens
	// GPT-4.1 series (Standard tier)
	case "gpt-4.1":
		inputCostPer1K = 0.002  // $2.00 per 1M tokens = $0.002 per 1K tokens
		outputCostPer1K = 0.008 // $8.00 per 1M tokens = $0.008 per 1K tokens
	case "gpt-4.1-mini":
		inputCostPer1K = 0.0004  // $0.40 per 1M tokens = $0.0004 per 1K tokens
		outputCostPer1K = 0.0016 // $1.60 per 1M tokens = $0.0016 per 1K tokens
	case "gpt-4.1-nano":
		inputCostPer1K = 0.0001  // $0.10 per 1M tokens = $0.0001 per 1K tokens
		outputCostPer1K = 0.0004 // $0.40 per 1M tokens = $0.0004 per 1K tokens
	// O-series models (Standard tier)
	case "o1":
		inputCostPer1K = 0.015 // $15.00 per 1M tokens = $0.015 per 1K tokens
		outputCostPer1K = 0.06 // $60.00 per 1M tokens = $0.06 per 1K tokens
	case "o1-pro":
		inputCostPer1K = 0.15 // $150.00 per 1M tokens = $0.15 per 1K tokens
		outputCostPer1K = 0.6 // $600.00 per 1M tokens = $0.6 per 1K tokens
	case "o1-mini":
		inputCostPer1K = 0.0011  // $1.10 per 1M tokens = $0.0011 per 1K tokens
		outputCostPer1K = 0.0044 // $4.40 per 1M tokens = $0.0044 per 1K tokens
	case "o3", "o3-pro", "o3-mini", "o3-deep-research":
		// Use o3 pricing for all o3 variants
		inputCostPer1K = 0.002  // $2.00 per 1M tokens = $0.002 per 1K tokens
		outputCostPer1K = 0.008 // $8.00 per 1M tokens = $0.008 per 1K tokens
	case "o4-mini", "o4-mini-deep-research":
		inputCostPer1K = 0.0011  // $1.10 per 1M tokens = $0.0011 per 1K tokens
		outputCostPer1K = 0.0044 // $4.40 per 1M tokens = $0.0044 per 1K tokens
	// Existing GPT-4o series
	case "gpt-4o":
		inputCostPer1K = 0.0025 // $2.50 per 1M tokens = $0.0025 per 1K tokens (Standard tier)
		outputCostPer1K = 0.01  // $10.00 per 1M tokens = $0.01 per 1K tokens
	case "gpt-4o-mini":
		inputCostPer1K = 0.00015 // $0.15 per 1M tokens = $0.00015 per 1K tokens (Standard tier)
		outputCostPer1K = 0.0006 // $0.60 per 1M tokens = $0.0006 per 1K tokens
	case "gpt-4", "gpt-4-0613", "gpt-4-0314":
		inputCostPer1K = 0.03  // $30.00 per 1M tokens = $0.03 per 1K tokens (Standard tier)
		outputCostPer1K = 0.06 // $60.00 per 1M tokens = $0.06 per 1K tokens
	case "gpt-3.5-turbo", "gpt-3.5-turbo-0125":
		inputCostPer1K = 0.0005  // $0.50 per 1M tokens = $0.0005 per 1K tokens (Standard tier)
		outputCostPer1K = 0.0015 // $1.50 per 1M tokens = $0.0015 per 1K tokens
	case "gpt-4-turbo", "gpt-4-turbo-2024-04-09":
		inputCostPer1K = 0.01  // $10.00 per 1M tokens = $0.01 per 1K tokens (Standard tier)
		outputCostPer1K = 0.03 // $30.00 per 1M tokens = $0.03 per 1K tokens
	default:
		inputCostPer1K = 0.001  // Default input estimate
		outputCostPer1K = 0.002 // Default output estimate
	}

	inputCost := float64(promptTokens) * inputCostPer1K / 1000
	outputCost := float64(completionTokens) * outputCostPer1K / 1000
	costEstimate := inputCost + outputCost

	// 記錄API響應信息，包含詳細的 token 使用和成本分解
	logFields := map[string]interface{}{
		"service":            "openai",
		"response_id":        resp.ID,
		"model":              string(resp.Model),
		"object":             string(resp.Object),
		"created":            resp.Created,
		"prompt_tokens":      resp.Usage.PromptTokens,
		"completion_tokens":  resp.Usage.CompletionTokens,
		"total_tokens":       resp.Usage.TotalTokens,
		"input_cost_usd":     fmt.Sprintf("$%.6f", inputCost),
		"output_cost_usd":    fmt.Sprintf("$%.6f", outputCost),
		"total_cost_usd":     fmt.Sprintf("$%.6f", costEstimate),
		"input_rate_per_1k":  fmt.Sprintf("$%.6f", inputCostPer1K),
		"output_rate_per_1k": fmt.Sprintf("$%.6f", outputCostPer1K),
		"choices_count":      len(resp.Choices),
	}

	// 加入 finish_reason 和內容過濾相關資訊
	if len(resp.Choices) > 0 {
		finishReason := string(resp.Choices[0].FinishReason)
		logFields["finish_reason"] = finishReason

		// 標記是否被內容過濾器阻擋
		if finishReason == "content_filter" {
			logFields["content_filtered"] = true
		}
	}

	// SystemFingerprint 已被官方標記為 deprecated，不再記錄

	// 加入服務層級資訊（可能影響內容過濾）
	if resp.ServiceTier != "" {
		logFields["service_tier"] = string(resp.ServiceTier)
	}

	// 記錄 Logprobs 資訊（如果啟用）
	if len(resp.Choices) > 0 {
		logprobs := resp.Choices[0].Logprobs
		if logprobs.Content != nil && len(logprobs.Content) > 0 {
			logFields["logprobs_enabled"] = true
			logFields["logprobs_tokens"] = len(logprobs.Content)
		}
	}

	// 加入 seed 參數（如果有設定）
	if seed := utils.GetEnvWithDefault("OPENAI_SEED", ""); seed != "" {
		logFields["seed_used"] = seed
	}

	utils.Logger.WithFields(logFields).Info("OpenAI API response received")

	// 開發模式下詳細記錄響應內容
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		utils.Logger.WithFields(map[string]interface{}{
			"service":     "openai",
			"response_id": resp.ID,
			"model":       string(resp.Model),
		}).Info("🎯 OpenAI Response Details")

		for i, choice := range resp.Choices {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "openai",
				"choice_index":   i,
				"finish_reason":  string(choice.FinishReason),
				"content_length": len(choice.Message.Content),
			}).Info(fmt.Sprintf("💬 Response [%d]: %s", i, choice.Message.Content))
		}
	} else {
		// 生產環境只記錄基本信息
		for i, choice := range resp.Choices {
			utils.Logger.WithFields(map[string]interface{}{
				"service":        "openai",
				"choice_index":   i,
				"finish_reason":  string(choice.FinishReason),
				"content_length": len(choice.Message.Content),
			}).Debug("OpenAI response choice")
		}
	}

	// 直接返回官方 SDK 的響應結構
	return resp, nil
}

