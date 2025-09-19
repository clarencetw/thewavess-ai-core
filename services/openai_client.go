package services

import (
	"context"
	"fmt"
	"strings"

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

	var client openai.Client
	if baseURL != "https://api.openai.com/v1" {
		// 自定義端點
		client = openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(baseURL),
		)
		utils.Logger.WithField("base_url", baseURL).Info("Using custom OpenAI API URL")
	} else {
		// 使用默認 OpenAI API
		client = openai.NewClient(
			option.WithAPIKey(apiKey),
		)
	}

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
	resp, err := c.client.Chat.Completions.New(ctx, params)

	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "openai",
			"error":   err.Error(),
			"model":   string(c.model),
			"user":    request.User,
		}).Error("OpenAI API call failed")
		return nil, fmt.Errorf("failed OpenAI API call: %w", err)
	}

	// 計算簡單成本估算
	totalTokens := int(resp.Usage.TotalTokens)
	var costEstimate float64
	switch string(resp.Model) {
	case "gpt-4o":
		costEstimate = float64(totalTokens) * 0.000005 // $0.005 per 1K tokens
	case "gpt-4o-mini":
		costEstimate = float64(totalTokens) * 0.00000015 // $0.00015 per 1K tokens
	case "gpt-4":
		costEstimate = float64(totalTokens) * 0.00003 // $0.03 per 1K tokens
	case "gpt-3.5-turbo":
		costEstimate = float64(totalTokens) * 0.0000015 // $0.0015 per 1K tokens
	default:
		costEstimate = float64(totalTokens) * 0.000002 // Default estimate
	}

	// 記錄API響應信息，包含 token 使用和成本
	logFields := map[string]interface{}{
		"service":           "openai",
		"response_id":       resp.ID,
		"model":             string(resp.Model),
		"object":            string(resp.Object),
		"created":           resp.Created,
		"prompt_tokens":     resp.Usage.PromptTokens,
		"completion_tokens": resp.Usage.CompletionTokens,
		"total_tokens":      resp.Usage.TotalTokens,
		"cost_usd":          fmt.Sprintf("$%.6f", costEstimate),
		"choices_count":     len(resp.Choices),
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
