package services

import (
    "context"
    "fmt"
    "math/rand"
    "strings"
    "time"

    "github.com/clarencetw/thewavess-ai-core/models"
    "github.com/clarencetw/thewavess-ai-core/models/db"
    "github.com/clarencetw/thewavess-ai-core/utils"
    "github.com/sirupsen/logrus"
    "github.com/uptrace/bun"
)


// ChatMessage 簡化的聊天消息類型（內部使用）
type ChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatService 對話服務
type ChatService struct {
	db           *bun.DB
	openaiClient *OpenAIClient
	grokClient   *GrokClient
	config       *ChatConfig
	memoryManager  *MemoryManager
	emotionManager *EmotionManager
	nsfwAnalyzer   *NSFWAnalyzer
}

// ChatConfig 對話配置
type ChatConfig struct {
	OpenAI struct {
		Model       string  `json:"model"`
		MaxTokens   int     `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
	} `json:"openai"`

	Grok struct {
		Model       string  `json:"model"`
		MaxTokens   int     `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
	} `json:"grok"`

	NSFW struct {
		DetectionThreshold float64 `json:"detection_threshold"`
		MaxIntensityLevel  int     `json:"max_intensity_level"`
	} `json:"nsfw"`

	Scene struct {
		EnableDescriptions   bool `json:"enable_descriptions"`
		MaxDescriptionLength int  `json:"max_description_length"`
		UpdateFrequency      int  `json:"update_frequency"`
	} `json:"scene"`
}

// ProcessMessageRequest 處理消息請求
type ProcessMessageRequest struct {
	SessionID   string                 `json:"session_id"`
	UserMessage string                 `json:"user_message"`
	CharacterID string                 `json:"character_id"`
	UserID      string                 `json:"user_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse 對話回應
type ChatResponse struct {
	SessionID         string        `json:"session_id"`
	MessageID         string        `json:"message_id"`
	SceneDescription  string        `json:"scene_description"`
	CharacterDialogue string        `json:"character_dialogue"`
	CharacterAction   string        `json:"character_action"`
	EmotionState      *EmotionState `json:"emotion_state"`
	AIEngine          string        `json:"ai_engine"`
	NSFWLevel         int           `json:"nsfw_level"`
	ResponseTime      time.Duration `json:"response_time"`
	NovelChoices      []NovelChoice `json:"novel_choices,omitempty"`
	SpecialEvent      *SpecialEvent `json:"special_event,omitempty"`
}

// EmotionState 情感狀態
type EmotionState struct {
	Affection     int    `json:"affection"`      // 好感度 0-100
	Mood          string `json:"mood"`           // happy, sad, shy, excited, concerned
	Relationship  string `json:"relationship"`   // stranger, friend, ambiguous, lover
	IntimacyLevel string `json:"intimacy_level"` // distant, friendly, close, intimate
}

// NovelChoice 小說選項
type NovelChoice struct {
	ID          string `json:"id"`
	Text        string `json:"text"`
	Consequence string `json:"consequence"`
}

// SpecialEvent 特殊事件
type SpecialEvent struct {
	Triggered   bool   `json:"triggered"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// SceneDescriptor 場景描述器
type SceneDescriptor struct {
	Location       string `json:"location"`        // 地點
	TimeOfDay      string `json:"time_of_day"`     // 時間
	Weather        string `json:"weather"`         // 天氣/氛圍
	Mood           string `json:"mood"`            // 當前氣氛
	CharacterState string `json:"character_state"` // 角色狀態
}

// ContentAnalysis 內容分析結果
type ContentAnalysis struct {
	IsNSFW        bool     `json:"is_nsfw"`
	Intensity     int      `json:"intensity"`  // 1-5 級
	Categories    []string `json:"categories"` // romantic, suggestive, explicit
	ShouldUseGrok bool     `json:"should_use_grok"`
	Confidence    float64  `json:"confidence"`
}

// ConversationContext 對話上下文
type ConversationContext struct {
	SessionID       string                 `json:"session_id"`
	UserID          string                 `json:"user_id"`
	CharacterID     string                 `json:"character_id"`
	RecentMessages  []ChatMessage          `json:"recent_messages"`
	EmotionState    *EmotionState          `json:"emotion_state"`
	SceneState      *SceneDescriptor       `json:"scene_state"`
	UserPreferences map[string]interface{} `json:"user_preferences"`
	MemoryPrompt    string                 `json:"memory_prompt"` // 記憶提示詞
}

// NewChatService 創建新的對話服務
func NewChatService() *ChatService {
    // 載入環境變數（非 production 會載入 .env）
    utils.LoadEnv()

    config := &ChatConfig{
        OpenAI: struct {
            Model       string  `json:"model"`
            MaxTokens   int     `json:"max_tokens"`
            Temperature float64 `json:"temperature"`
        }{
            Model:       utils.GetEnvWithDefault("OPENAI_MODEL", "gpt-4o"),
            MaxTokens:   utils.GetEnvIntWithDefault("OPENAI_MAX_TOKENS", 800),
            Temperature: utils.GetEnvFloatWithDefault("OPENAI_TEMPERATURE", 0.8),
        },
        Grok: struct {
            Model       string  `json:"model"`
            MaxTokens   int     `json:"max_tokens"`
            Temperature float64 `json:"temperature"`
        }{
            // 預設使用 grok-3；若需回退可在環境改為 grok-beta
            Model:       utils.GetEnvWithDefault("GROK_MODEL", "grok-3"),
            MaxTokens:   utils.GetEnvIntWithDefault("GROK_MAX_TOKENS", 1000),
            Temperature: utils.GetEnvFloatWithDefault("GROK_TEMPERATURE", 0.9),
        },
		NSFW: struct {
			DetectionThreshold float64 `json:"detection_threshold"`
			MaxIntensityLevel  int     `json:"max_intensity_level"`
		}{
			DetectionThreshold: 0.7,
			MaxIntensityLevel:  5,
		},
		Scene: struct {
			EnableDescriptions   bool `json:"enable_descriptions"`
			MaxDescriptionLength int  `json:"max_description_length"`
			UpdateFrequency      int  `json:"update_frequency"`
		}{
			EnableDescriptions:   true,
			MaxDescriptionLength: 120,
			UpdateFrequency:      5,
		},
	}

	return &ChatService{
		db:           GetDB(),
		openaiClient: NewOpenAIClient(),
		grokClient:   NewGrokClient(),
		config:       config,
		memoryManager:  NewMemoryManager(),
		emotionManager: NewEmotionManager(),
		nsfwAnalyzer:   NewNSFWAnalyzer(),
	}
}

// ProcessMessage 處理用戶消息並生成回應 - 女性向AI聊天系統
func (s *ChatService) ProcessMessage(ctx context.Context, request *ProcessMessageRequest) (*ChatResponse, error) {
	startTime := time.Now()

	utils.Logger.WithFields(logrus.Fields{
		"session_id":   request.SessionID,
		"user_id":      request.UserID,
		"character_id": request.CharacterID,
		"message_len":  len(request.UserMessage),
	}).Info("開始處理AI對話請求")


	// 1. NSFW內容智能分析（5級分級系統）
	analysis, err := s.analyzeContent(request.UserMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze content: %w", err)
	}

	// 預先生成成對的 Message ID（先持久化用戶訊息，確保上下文包含當前輪）
	messageID := generateMessageID()                   // 助手訊息 ID
	userMessageID := fmt.Sprintf("user_%s", messageID) // 用戶訊息 ID 與本輪綁定

	// 1.5 保存用戶消息（確保稍後載入歷史能包含本輪用戶訊息）
	if err := s.saveUserMessageToDB(ctx, request, userMessageID, analysis); err != nil {
		utils.Logger.WithError(err).Error("保存用戶消息失敗：將降級為臨時上下文")
		// 不中斷：稍後上下文將降級為僅使用內存與提示詞
	}


	// 2. 構建女性向對話上下文（已包含本次用戶訊息，若上一步失敗則只會包含歷史）
	conversationContext, err := s.buildFemaleOrientedContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to build female-oriented context: %w", err)
	}

	// 3. 智能引擎選擇（OpenAI vs Grok）
	engine := s.selectAIEngine(analysis, conversationContext.UserPreferences)

	// 4. 動態場景生成（女性喜愛的沉浸感）
	sceneDescription := ""
	if s.config.Scene.EnableDescriptions {
		sceneDescription = s.generateRomanticScene(conversationContext, analysis.Intensity)
	}

	// 5. 角色個性化回應生成
	response, err := s.generatePersonalizedResponse(ctx, engine, request.UserMessage, conversationContext, sceneDescription, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate personalized response: %w", err)
	}

	// 5.1 角色一致性檢查和優化
	consistencyResult := s.checkAndOptimizeCharacterConsistency(request.CharacterID, response.Dialogue, conversationContext)
	if !consistencyResult.IsConsistent {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": request.CharacterID,
			"score":        consistencyResult.Score,
			"violations":   len(consistencyResult.Violations),
		}).Warn("角色一致性檢查發現問題")

		// 記錄違規詳情用於改進
		for _, violation := range consistencyResult.Violations {
			utils.Logger.WithFields(logrus.Fields{
				"type":        violation.Type,
				"severity":    violation.Severity,
				"description": violation.Description,
			}).Debug("一致性違規詳情")
		}
	}

	// 6. 情感狀態智能更新（好感度、關係發展）
	newEmotionState := s.updateEmotionStateAdvanced(conversationContext.EmotionState, request.UserID, request.CharacterID, request.UserMessage, response, analysis)

	// 7. 記憶系統更新（長期關係發展）
	s.updateMemorySystem(request.UserID, request.CharacterID, request.SessionID, request.UserMessage, response.Dialogue, newEmotionState)

	// 8. 特殊事件檢測（關係里程碑等）
	specialEvent := s.detectSpecialEvents(newEmotionState, conversationContext.EmotionState)


	// 9. 保存 AI 回應到資料庫（用先前生成的 messageID）
	err = s.saveAssistantMessageToDB(ctx, request, messageID, response, sceneDescription, newEmotionState, engine, analysis, time.Since(startTime))
	if err != nil {
		utils.Logger.WithError(err).Error("保存對話到資料庫失敗")
		// 不中斷流程，但記錄錯誤
	}

	// 10. 構建完整回應
	chatResponse := &ChatResponse{
		SessionID:         request.SessionID,
		MessageID:         messageID,
		SceneDescription:  sceneDescription,
		CharacterDialogue: response.Dialogue,
		CharacterAction:   response.Action,
		EmotionState:      newEmotionState,
		AIEngine:          engine,
		NSFWLevel:         analysis.Intensity,
		ResponseTime:      time.Since(startTime),
		SpecialEvent:      specialEvent,
	}

	utils.Logger.WithFields(logrus.Fields{
		"session_id":    request.SessionID,
		"character_id":  request.CharacterID,
		"nsfw_level":    analysis.Intensity,
		"ai_engine":     engine,
		"affection":     newEmotionState.Affection,
		"relationship":  newEmotionState.Relationship,
		"response_time": chatResponse.ResponseTime.Milliseconds(),
	}).Info("AI對話處理完成")


	return chatResponse, nil
}

// analyzeContent 分析消息內容
func (s *ChatService) analyzeContent(message string) (*ContentAnalysis, error) {
	// 使用專門的 NSFW 分析器
	level, analysis := s.nsfwAnalyzer.AnalyzeContent(message)

	// 記錄分析結果
	utils.Logger.WithFields(logrus.Fields{
		"message_preview": message[:min(50, len(message))],
		"nsfw_level":      level,
		"is_nsfw":         analysis.IsNSFW,
		"confidence":      analysis.Confidence,
		"should_use_grok": analysis.ShouldUseGrok,
	}).Debug("內容分析完成")

	return analysis, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildFemaleOrientedContext 構建女性向對話上下文
func (s *ChatService) buildFemaleOrientedContext(ctx context.Context, request *ProcessMessageRequest) (*ConversationContext, error) {
	// 從數據庫獲取實際的會話歷史和情感狀態
	emotionState, err := s.getOrCreateEmotionStateFromDB(ctx, request.UserID, request.CharacterID)
	if err != nil {
		utils.Logger.WithError(err).Warn("獲取情感狀態失敗，使用默認值")
		emotionState = s.getOrCreateEmotionState(request.UserID, request.CharacterID)
	}

	// 獲取最近的對話記憶（短期記憶：最近 5-10 條訊息）
	recentMemories, err := s.getRecentMemoriesFromDB(ctx, request.SessionID, 5)
	if err != nil {
		utils.Logger.WithError(err).Warn("獲取會話歷史失敗，使用內存數據")
		recentMemories = s.getRecentMemories(request.SessionID, request.UserID, request.CharacterID, 5)
	}

	// 獲取用戶偏好設置
	userPreferences, err := s.getUserPreferencesFromDB(ctx, request.UserID)
	if err != nil {
		utils.Logger.WithError(err).Warn("獲取用戶偏好失敗，使用默認值")
		userPreferences = s.getUserPreferences(request.UserID)
	}

	// 生成優化的記憶提示詞（長期記憶：偏好/稱呼/里程碑/禁忌）
	memoryPrompt := s.memoryManager.GenerateOptimizedMemoryPrompt(request.UserID, request.CharacterID)

	// 確保會話存在
	err = s.ensureSessionExists(ctx, request.SessionID, request.UserID, request.CharacterID)
	if err != nil {
		utils.Logger.WithError(err).Error("確保會話存在失敗")
	}

	return &ConversationContext{
		SessionID:       request.SessionID,
		UserID:          request.UserID,
		CharacterID:     request.CharacterID,
		RecentMessages:  recentMemories,
		EmotionState:    emotionState,
		SceneState:      s.generateSceneState(request.CharacterID, emotionState),
		UserPreferences: userPreferences,
		MemoryPrompt:    memoryPrompt,
	}, nil
}

// getOrCreateEmotionState 獲取或創建情感狀態
func (s *ChatService) getOrCreateEmotionState(userID, characterID string) *EmotionState {
	// 使用情感管理器獲取或創建情感狀態
	return s.emotionManager.GetEmotionState(userID, characterID)
}




// getRecentMemories 獲取最近的對話記憶
func (s *ChatService) getRecentMemories(sessionID, userID, characterID string, limit int) []ChatMessage {
	// 從記憶管理器獲取短期記憶
	shortTermMemory := s.memoryManager.GetShortTermMemory(sessionID)
	if shortTermMemory == nil || len(shortTermMemory.RecentMessages) == 0 {
		return []ChatMessage{}
	}

	// 轉換為 ChatMessage 格式
	messages := make([]ChatMessage, 0, len(shortTermMemory.RecentMessages))
	for _, msg := range shortTermMemory.RecentMessages {
		// 檢查Summary是否為空，避免產生空內容的消息
		if strings.TrimSpace(msg.Summary) == "" {
			continue // 跳過空Summary的消息
		}
		
		messages = append(messages, ChatMessage{
			Role:      msg.Role,
			Content:   msg.Summary,
			CreatedAt: msg.Timestamp,
		})
	}

	// 限制返回數量
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	return messages
}

// getUserPreferences 獲取用戶偏好
func (s *ChatService) getUserPreferences(userID string) map[string]interface{} {
	// 使用默認偏好
	return map[string]interface{}{
		"scene_style":      "romantic",
		"response_length":  "medium",
		"emotion_tracking": true,
	}
}

// generateSceneState 生成場景狀態
func (s *ChatService) generateSceneState(characterID string, emotion *EmotionState) *SceneDescriptor {
	switch characterID {
	case "character_01": // 沈宸
		return &SceneDescriptor{
			Location:       "豪華辦公室",
			TimeOfDay:      s.getCurrentTimeOfDay(),
			Weather:        "城市夜景透過落地窗映入室內",
			Mood:           s.getSceneMood(emotion),
			CharacterState: s.getCharacterState(characterID, emotion),
		}
	case "character_02": // 林知遠
		return &SceneDescriptor{
			Location:       "溫馨診療室",
			TimeOfDay:      s.getCurrentTimeOfDay(),
			Weather:        "柔和的燈光營造出安心的氛圍",
			Mood:           s.getSceneMood(emotion),
			CharacterState: s.getCharacterState(characterID, emotion),
		}
	case "character_03": // 周曜
		return &SceneDescriptor{
			Location:       "溫馨直播室",
			TimeOfDay:      s.getCurrentTimeOfDay(),
			Weather:        "音樂與柔和燈光交織的暖心空間",
			Mood:           s.getSceneMood(emotion),
			CharacterState: s.getCharacterState(characterID, emotion),
		}
	default:
		return &SceneDescriptor{
			Location:       "舒適房間",
			TimeOfDay:      s.getCurrentTimeOfDay(),
			Weather:        "溫暖的氛圍",
			Mood:           "comfortable",
			CharacterState: "放鬆狀態",
		}
	}
}

// getCurrentTimeOfDay 獲取當前時間段
func (s *ChatService) getCurrentTimeOfDay() string {
	hour := time.Now().Hour()
	if hour < 6 {
		return "深夜"
	} else if hour < 12 {
		return "上午"
	} else if hour < 18 {
		return "下午"
	} else {
		return "晚上"
	}
}

// getSceneMood 獲取場景情緒
func (s *ChatService) getSceneMood(emotion *EmotionState) string {
	if emotion.Affection >= 70 {
		return "romantic"
	} else if emotion.Affection >= 50 {
		return "warm"
	} else {
		return "professional"
	}
}

// getCharacterState 獲取角色狀態
func (s *ChatService) getCharacterState(characterID string, emotion *EmotionState) string {
	baseStates := map[string][]string{
		"character_01": {"專注工作中", "思考中", "等待你的回應", "專注地看著你"},
		"character_02": {"溫和地笑著", "關心地看著你", "耐心等待", "溫柔地注視"},
	}

	if states, exists := baseStates[characterID]; exists {
		index := emotion.Affection % len(states)
		return states[index]
	}
	return "等待中"
}

// selectAIEngine 選擇 AI 引擎
func (s *ChatService) selectAIEngine(analysis *ContentAnalysis, userPrefs map[string]interface{}) string {
	// NSFW 功能永久開啟，根據內容分析決定使用哪個引擎
	if analysis.ShouldUseGrok {
		return "grok"
	}
	return "openai"
}




// CharacterResponseData 角色回應數據
type CharacterResponseData struct {
	Dialogue string `json:"dialogue"`
	Action   string `json:"action"`
}

// generatePersonalizedResponse 生成個性化女性向回應
func (s *ChatService) generatePersonalizedResponse(ctx context.Context, engine, userMessage string, context *ConversationContext, sceneDescription string, analysis *ContentAnalysis) (*CharacterResponseData, error) {

	// 構建女性向角色提示詞
	prompt := s.buildFemaleOrientedPrompt(context.CharacterID, userMessage, context, sceneDescription, analysis.Intensity)

	var dialogue string
	var err error

	if engine == "openai" {
		// 使用 OpenAI (Level 1-4)
		dialogue, err = s.generateOpenAIResponse(ctx, prompt, context)
		if err != nil {
			utils.Logger.WithError(err).Error("OpenAI 回應生成失敗")
			return nil, fmt.Errorf("OpenAI API failed: %w", err)
		}
	} else if engine == "grok" {
		// 使用 Grok (Level 5)
		dialogue, err = s.generateGrokResponse(ctx, prompt, context)
		if err != nil {
			utils.Logger.WithError(err).Error("Grok 回應生成失敗")
			return nil, fmt.Errorf("Grok API failed: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unknown AI engine: %s", engine)
	}

	// 解析AI回應中的對話和動作
	dialogue, action := s.parseDialogueAndAction(dialogue, context.CharacterID, context.EmotionState, analysis.Intensity)

	return &CharacterResponseData{
		Dialogue: dialogue,
		Action:   action,
	}, nil
}

// buildFemaleOrientedPrompt 構建女性向角色提示詞（使用統一模板）
func (s *ChatService) buildFemaleOrientedPrompt(characterID, userMessage string, conversationContext *ConversationContext, sceneDescription string, nsfwLevel int) string {
	// 獲取記憶提示詞
	memoryPrompt := s.memoryManager.GetMemoryPrompt(conversationContext.SessionID, conversationContext.UserID, conversationContext.CharacterID)

	// 使用統一的prompt模板確保一致性
	characterService := GetCharacterService()
	promptBuilder := NewPromptBuilder(characterService)
	ctx := context.Background()
	return promptBuilder.
		WithCharacter(ctx, characterID).
		WithContext(conversationContext).
		WithNSFWLevel(nsfwLevel).
		WithUserMessage(userMessage).
		WithSceneDescription(sceneDescription).
		WithMemory(memoryPrompt).
		Build(ctx)
}




// generateGrokResponse 生成Grok回應
func (s *ChatService) generateGrokResponse(ctx context.Context, prompt string, context *ConversationContext) (string, error) {
	// 構建 Grok 請求
	messages := []GrokMessage{
		{
			Role:    "system",
			Content: prompt,
		},
	}

	// 添加最近的對話歷史作為上下文
	if context.RecentMessages != nil && len(context.RecentMessages) > 0 {
		for _, msg := range context.RecentMessages {
			// 檢查內容是否為空，避免傳遞空內容到API
			if strings.TrimSpace(msg.Content) == "" {
				continue // 跳過空內容的消息
			}
			
			role := "user"
			if msg.Role == "assistant" {
				role = "assistant"
			}
			messages = append(messages, GrokMessage{
				Role:    role,
				Content: msg.Content,
			})
		}
	}

	// 創建 Grok 請求
	request := &GrokRequest{
		Model:       s.config.Grok.Model,
		Messages:    messages,
		MaxTokens:   s.config.Grok.MaxTokens,
		Temperature: s.config.Grok.Temperature,
		User:        context.UserID,
	}

	// 調用 Grok API
	utils.Logger.WithFields(map[string]interface{}{
		"session_id":   context.SessionID,
		"character_id": context.CharacterID,
		"user_id":      context.UserID,
	}).Info("調用 Grok API")

	response, err := s.grokClient.GenerateResponse(ctx, request)
	if err != nil {
		utils.Logger.WithError(err).Error("Grok API 調用失敗")
		return "", fmt.Errorf("Grok API call failed: %w", err)
	}

	// 從回應中提取對話內容
	if len(response.Choices) > 0 {
		dialogue := response.Choices[0].Message.Content

		utils.Logger.WithFields(map[string]interface{}{
			"session_id":   context.SessionID,
			"response_len": len(dialogue),
			"tokens_used":  response.Usage.TotalTokens,
		}).Info("Grok API 響應成功")

		return dialogue, nil
	}

	// 如果沒有回應內容，返回錯誤
	utils.Logger.Warn("Grok API 返回空回應")
	return "", fmt.Errorf("Grok API returned empty response")
}

// generateOpenAIResponse 生成OpenAI回應
func (s *ChatService) generateOpenAIResponse(ctx context.Context, prompt string, context *ConversationContext) (string, error) {
	// 構建 OpenAI 請求
	messages := []OpenAIMessage{
		{
			Role:    "system",
			Content: prompt,
		},
	}

	// 添加最近的對話歷史作為上下文
	if context.RecentMessages != nil && len(context.RecentMessages) > 0 {
		for _, msg := range context.RecentMessages {
			// 檢查內容是否為空，避免傳遞空內容到API
			if strings.TrimSpace(msg.Content) == "" {
				continue // 跳過空內容的消息
			}
			
			role := "user"
			if msg.Role == "assistant" {
				role = "assistant"
			}
			messages = append(messages, OpenAIMessage{
				Role:    role,
				Content: msg.Content,
			})
		}
	}

	// 創建 OpenAI 請求
	request := &OpenAIRequest{
		Model:       s.config.OpenAI.Model,
		Messages:    messages,
		MaxTokens:   s.config.OpenAI.MaxTokens,
		Temperature: s.config.OpenAI.Temperature,
		User:        context.UserID,
	}

	// 調用 OpenAI API
	utils.Logger.WithFields(map[string]interface{}{
		"session_id":   context.SessionID,
		"character_id": context.CharacterID,
		"user_id":      context.UserID,
	}).Info("調用 OpenAI API")

	response, err := s.openaiClient.GenerateResponse(ctx, request)
	if err != nil {
		utils.Logger.WithError(err).Error("OpenAI API 調用失敗")
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// 從回應中提取對話內容
	if len(response.Choices) > 0 {
		dialogue := response.Choices[0].Message.Content

		utils.Logger.WithFields(map[string]interface{}{
			"session_id":   context.SessionID,
			"response_len": len(dialogue),
			"tokens_used":  response.Usage.TotalTokens,
		}).Info("OpenAI API 響應成功")

		return dialogue, nil
	}

	// 如果沒有回應內容，返回錯誤
	utils.Logger.Warn("OpenAI API 返回空回應")
	return "", fmt.Errorf("OpenAI API returned empty response")
}




// parseDialogueAndAction 解析AI回應中的對話和動作
func (s *ChatService) parseDialogueAndAction(response, characterID string, emotion *EmotionState, nsfwLevel int) (string, string) {
	// 嘗試解析 "對話|||動作" 格式
	if strings.Contains(response, "|||") {
		parts := strings.SplitN(response, "|||", 2)
		if len(parts) == 2 {
			dialogue := strings.TrimSpace(parts[0])
			action := strings.TrimSpace(parts[1])
			
			// 驗證解析結果不為空
			if dialogue != "" && action != "" {
				return dialogue, action
			}
		}
	}

	// 如果無法解析或格式不正確，使用AI完整回應作為對話，並生成備用動作
	dialogue := strings.TrimSpace(response)
	action := s.generateFallbackAction(characterID, dialogue, emotion, nsfwLevel)
	
	utils.Logger.WithFields(logrus.Fields{
		"character_id": characterID,
		"response_len": len(response),
		"parsed_format": false,
	}).Debug("AI回應格式解析失敗，使用備用動作")
	
	return dialogue, action
}

// generateFallbackAction 生成備用動作（當AI未按格式回應時）
func (s *ChatService) generateFallbackAction(characterID, dialogue string, emotion *EmotionState, nsfwLevel int) string {
	// 基於對話內容智能生成動作
	dialogue = strings.ToLower(dialogue)
	
	// 根據角色特質生成基礎動作
	var baseActions []string
	switch characterID {
	case "character_01": // 沈宸
		baseActions = []string{
			"他深邃的眼眸注視著你", 
			"他的聲音低沉磁性", 
			"他優雅而威嚴地看著你",
		}
	case "character_02": // 林知遠  
		baseActions = []string{
			"他溫和地觀察著，眼神專業而深層", 
			"他理解地看著你", 
			"他平穩而溫柔地回應",
		}
	case "character_03": // 周曜
		baseActions = []string{
			"他開朗地笑著，眼中閃爍著陽光", 
			"他熱情地看著你", 
			"他溫暖而親切地回應",
		}
	default:
		baseActions = []string{"他溫柔地看著你"}
	}

	// 基於對話內容調整動作
	if strings.Contains(dialogue, "累") || strings.Contains(dialogue, "疲憊") {
		switch characterID {
		case "character_01":
			return "他關切地看著你，眉頭微蹙"
		case "character_02":
			return "他專業地觀察著，聲音溫柔關切"
		case "character_03":
			return "他關心地靠近，眼中满含溫暖"
		default:
			return "他露出關心的表情，聲音輕柔"
		}
	}
	
	if strings.Contains(dialogue, "休息") || strings.Contains(dialogue, "睡") {
		switch characterID {
		case "character_01":
			return "他溫柔地伸手，想要撫摸你的頭"
		case "character_02":
			return "他輕聲建議，聲音帶著專業的關懷"
		case "character_03":
			return "他柔聲提醒，笑容溫暖如阳光"
		default:
			return "他輕聲提醒，眼中滿含關懷"
		}
	}

	// 根據NSFW級別調整親密度
	if nsfwLevel >= 3 && emotion.Affection >= 60 {
		switch characterID {
		case "character_01":
			return "他步向你，眼神中閃爍著溫柔的光芒"
		case "character_02":
			return "他温和地靠近，動作輕柔而專業"
		case "character_03":
			return "他熱情地靠近，眼中滿含溫暖的光芒"
		default:
			return "他輕柔地靠近，動作充滿溫暖"
		}
	}

	// 默認返回基礎動作
	index := rand.Intn(len(baseActions))
	return baseActions[index]
}

// updateEmotionStateAdvanced 高級情感狀態更新
func (s *ChatService) updateEmotionStateAdvanced(currentState *EmotionState, userID, characterID, userMessage string, response *CharacterResponseData, analysis *ContentAnalysis) *EmotionState {
	// 使用情感管理器更新情感狀態
	newState := s.emotionManager.UpdateEmotion(currentState, userMessage, analysis)

	// 保存情感快照到歷史記錄
	if currentState != nil && userID != "" && characterID != "" {
		trigger := "user_message"
		if analysis.IsNSFW {
			trigger = fmt.Sprintf("nsfw_level_%d", analysis.Intensity)
		}

		// 構建上下文信息
		context := fmt.Sprintf("用戶消息: %s", userMessage)
		if response != nil {
			context += fmt.Sprintf(" | AI回應: %s", response.Dialogue)
		}

		s.emotionManager.SaveEmotionSnapshot(
			userID,
			characterID,
			trigger,
			context,
			currentState,
			newState,
		)

		utils.Logger.WithFields(logrus.Fields{
			"user_id":       userID,
			"character_id":  characterID,
			"old_affection": currentState.Affection,
			"new_affection": newState.Affection,
			"trigger":       trigger,
		}).Info("情感狀態快照已保存")
	}

	return newState
}



// generateRomanticScene 生成浪漫場景 - 使用記憶體配置系統
func (s *ChatService) generateRomanticScene(convContext *ConversationContext, nsfwLevel int) string {
	characterID := convContext.CharacterID
	affection := convContext.EmotionState.Affection

	// 使用角色服務獲取角色場景
	configService := GetCharacterService()
	scenes, err := configService.GetCharacterScenes(context.Background(), characterID, "romantic", "evening", affection, nsfwLevel)
	
	if err != nil || len(scenes) == 0 {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"affection":    affection,
			"nsfw_level":   nsfwLevel,
			"error":        err,
		}).Error("無法獲取角色場景，返回空場景")

		// 返回空場景，讓上層處理
		return ""
	}

	// 使用第一個匹配的場景（按權重排序）
	scene := scenes[0]
	baseScene := scene.Description

	utils.Logger.WithFields(logrus.Fields{
		"character_id": characterID,
		"scene_id":     scene.ID,
		"affection":    affection,
		"nsfw_level":   nsfwLevel,
		"scene_type":   scene.SceneType,
		"base_scene":   baseScene,
	}).Debug("成功生成角色場景描述")

	return stringValue(baseScene)
}


// updateMemorySystem 更新記憶系統
// checkAndOptimizeCharacterConsistency 檢查並優化角色一致性
func (s *ChatService) checkAndOptimizeCharacterConsistency(characterID, response string, context *ConversationContext) *ConsistencyCheckResult {
	consistencyChecker := GetConsistencyChecker()
	result := consistencyChecker.CheckCharacterConsistency(characterID, response, context)

	// 記錄一致性檢查結果到日誌，用於持續改進
	utils.Logger.WithFields(logrus.Fields{
		"character_id":      characterID,
		"consistency_score": result.Score,
		"is_consistent":     result.IsConsistent,
		"violations_count":  len(result.Violations),
		"suggestions_count": len(result.Suggestions),
	}).Info("角色一致性檢查完成")

	// 如果一致性分數太低，可以考慮在未來版本中重新生成回應
	if result.Score < 0.5 {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"score":        result.Score,
			"suggestions":  result.Suggestions,
		}).Warn("角色一致性分數過低，建議改進")
	}

	return result
}

func (s *ChatService) updateMemorySystem(userID, characterID, sessionID, userMessage, aiResponse string, emotion *EmotionState) {
	// 更新短期記憶
	messages := []ChatMessage{
		{
			Role:      "user",
			Content:   userMessage,
			CreatedAt: time.Now(),
		},
		{
			Role:      "assistant",
			Content:   aiResponse,
			CreatedAt: time.Now(),
		},
	}
	s.memoryManager.UpdateShortTermMemory(sessionID, userID, characterID, messages)

	// 更新長期記憶
	s.memoryManager.ExtractAndUpdateLongTermMemory(userID, characterID, userMessage, aiResponse, emotion)

	utils.Logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"character_id": characterID,
		"session_id":   sessionID,
		"affection":    emotion.Affection,
		"relationship": emotion.Relationship,
	}).Info("記憶系統更新完成")
}

// detectSpecialEvents 檢測特殊事件
func (s *ChatService) detectSpecialEvents(newEmotion, oldEmotion *EmotionState) *SpecialEvent {
	// 檢測關係里程碑
	if oldEmotion.Relationship != newEmotion.Relationship {
		return &SpecialEvent{
			Triggered:   true,
			Type:        "relationship_milestone",
			Description: fmt.Sprintf("關係狀態從 %s 發展到 %s", oldEmotion.Relationship, newEmotion.Relationship),
		}
	}

	// 檢測好感度重大變化
	if newEmotion.Affection >= 80 && oldEmotion.Affection < 80 {
		return &SpecialEvent{
			Triggered:   true,
			Type:        "affection_milestone",
			Description: "好感度達到80，關係進入新階段",
		}
	}

	return nil
}

// generateMessageID 生成消息 ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// ==================== 資料庫操作方法 ====================

// ensureSessionExists 確保會話存在 - 簡化為一對一架構
func (s *ChatService) ensureSessionExists(ctx context.Context, sessionID, userID, characterID string) error {
    // 首先嘗試找到該用戶與角色的現有會話
    var existingSession db.ChatSessionDB
    err := s.db.NewSelect().
        Model(&existingSession).
        Where("user_id = ? AND character_id = ?", userID, characterID).
        Scan(ctx)

	if err == nil {
		// 會話已存在，更新為活躍狀態並使用現有ID
        _, updateErr := s.db.NewUpdate().
            Model((*db.ChatSessionDB)(nil)).
            Set("status = ?", "active").
            Set("updated_at = ?", time.Now()).
            Where("id = ?", existingSession.ID).
            Exec(ctx)

		if updateErr != nil {
			utils.Logger.WithError(updateErr).Warn("更新現有會話狀態失敗")
		}

		utils.Logger.WithFields(logrus.Fields{
			"existing_session_id":  existingSession.ID,
			"requested_session_id": sessionID,
			"user_id":              userID,
			"character_id":         characterID,
		}).Info("使用現有的用戶-角色對話會話")

		return nil
	}

    // 會話不存在，創建新的
    newSession := &db.ChatSessionDB{
        ID:          sessionID,
        UserID:      userID,
        CharacterID: characterID,
        Title:       s.generateSessionTitle(characterID),
        Status:      "active",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    _, err = s.db.NewInsert().
        Model(newSession).
        Exec(ctx)

	if err != nil {
		return fmt.Errorf("創建會話失敗: %w", err)
	}

	utils.Logger.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"user_id":      userID,
		"character_id": characterID,
	}).Info("創建新的用戶-角色對話會話")

	return nil
}

// saveUserMessageToDB 先保存用戶消息（以便上下文讀取包含本輪）
func (s *ChatService) saveUserMessageToDB(ctx context.Context, request *ProcessMessageRequest, userMessageID string, analysis *ContentAnalysis) error {
    // 使用 RunInTx 處理事務
    err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
        userMessage := &db.MessageDB{
            ID:        userMessageID,
            SessionID: request.SessionID,
            Role:      "user",
            Content:   request.UserMessage,
            NSFWLevel: analysis.Intensity,
            CreatedAt: time.Now(),
        }

        if _, err := tx.NewInsert().Model(userMessage).Exec(ctx); err != nil {
            return fmt.Errorf("保存用戶消息失敗: %w", err)
        }

        // 更新會話統計（只加用戶部分）
        if _, err := tx.NewUpdate().
            Model((*db.ChatSessionDB)(nil)).
            Set("message_count = message_count + 1").
            Set("total_characters = total_characters + ?", len(request.UserMessage)).
            Set("last_message_at = ?", time.Now()).
            Set("updated_at = ?", time.Now()).
            Where("id = ?", request.SessionID).
            Exec(ctx); err != nil {
            return fmt.Errorf("更新會話統計(用戶)失敗: %w", err)
        }

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// saveAssistantMessageToDB 保存 AI 回應（第二步）
func (s *ChatService) saveAssistantMessageToDB(ctx context.Context, request *ProcessMessageRequest, messageID string, response *CharacterResponseData, sceneDescription string, emotionState *EmotionState, engine string, analysis *ContentAnalysis, responseTime time.Duration) error {
    // 使用 RunInTx 處理事務
    err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
        aiMessage := &db.MessageDB{
            ID:               messageID,
            SessionID:        request.SessionID,
            Role:             "assistant",
            Content:          response.Dialogue,
            SceneDescription: &sceneDescription,
            CharacterAction:  &response.Action,
            EmotionalState: map[string]interface{}{
                "affection":      emotionState.Affection,
                "mood":           emotionState.Mood,
                "relationship":   emotionState.Relationship,
                "intimacy_level": emotionState.IntimacyLevel,
            },
            AIEngine:       &engine,
            ResponseTimeMs: func() *int { ms := int(responseTime.Milliseconds()); return &ms }(),
            NSFWLevel:      analysis.Intensity,
            CreatedAt:      time.Now(),
        }

        if _, err := tx.NewInsert().Model(aiMessage).Exec(ctx); err != nil {
            return fmt.Errorf("保存AI消息失敗: %w", err)
        }

        // 更新會話統計（再加助手部分）
        if _, err := tx.NewUpdate().
            Model((*db.ChatSessionDB)(nil)).
            Set("message_count = message_count + 1").
            Set("total_characters = total_characters + ?", len(response.Dialogue)).
            Set("last_message_at = ?", time.Now()).
            Set("updated_at = ?", time.Now()).
            Where("id = ?", request.SessionID).
            Exec(ctx); err != nil {
            return fmt.Errorf("更新會話統計(AI)失敗: %w", err)
        }

		return nil
	})
	if err != nil {
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"session_id": request.SessionID,
		"message_id": messageID,
		"ai_msg_len": len(response.Dialogue),
		"nsfw_level": analysis.Intensity,
		"ai_engine":  engine,
	}).Info("AI 消息已保存到資料庫")

	return nil
}

// getRecentMemoriesFromDB 從資料庫獲取最近的對話記憶
func (s *ChatService) getRecentMemoriesFromDB(ctx context.Context, sessionID string, limit int) ([]ChatMessage, error) {
    var messages []db.MessageDB

    err := s.db.NewSelect().
        Model(&messages).
        Where("session_id = ?", sessionID).
        Order("created_at DESC").
        Limit(limit * 2). // 獲取用戶和AI的消息
        Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("查詢會話歷史失敗: %w", err)
	}

	// 轉換為 ChatMessage 格式
	chatMessages := make([]ChatMessage, 0, len(messages))
    for i := len(messages) - 1; i >= 0; i-- { // 反轉順序，最舊的在前
        msg := messages[i]
        chatMessages = append(chatMessages, ChatMessage{
            Role:      msg.Role,
            Content:   msg.Content,
            CreatedAt: msg.CreatedAt,
        })
    }

	// 限制返回數量
	if len(chatMessages) > limit {
		chatMessages = chatMessages[len(chatMessages)-limit:]
	}

	utils.Logger.WithFields(logrus.Fields{
		"session_id":     sessionID,
		"messages_count": len(chatMessages),
	}).Debug("從資料庫獲取會話歷史成功")

	return chatMessages, nil
}

// getOrCreateEmotionStateFromDB 從資料庫獲取或創建情感狀態
func (s *ChatService) getOrCreateEmotionStateFromDB(ctx context.Context, userID, characterID string) (*EmotionState, error) {
	// 從最近的消息中獲取情感狀態
	var message models.Message
	err := s.db.NewSelect().
		Model(&message).
		Join("JOIN chat_sessions cs ON cs.id = m.session_id").
		Where("cs.user_id = ? AND cs.character_id = ?", userID, characterID).
		Where("m.role = 'assistant'").
		Where("m.emotional_state IS NOT NULL").
		Order("m.created_at DESC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		// 如果沒有找到歷史情感狀態，創建新的
		return &EmotionState{
			Affection:     30, // 初始好感度
			Mood:          "neutral",
			Relationship:  "stranger",
			IntimacyLevel: "distant",
		}, nil
	}

	// 解析情感狀態
	emotionalState := message.EmotionalState

	affection := 30
	if val, ok := emotionalState["affection"].(float64); ok {
		affection = int(val)
	}

	mood := "neutral"
	if val, ok := emotionalState["mood"].(string); ok {
		mood = val
	}

	relationship := "stranger"
	if val, ok := emotionalState["relationship"].(string); ok {
		relationship = val
	}

	intimacyLevel := "distant"
	if val, ok := emotionalState["intimacy_level"].(string); ok {
		intimacyLevel = val
	}

	return &EmotionState{
		Affection:     affection,
		Mood:          mood,
		Relationship:  relationship,
		IntimacyLevel: intimacyLevel,
	}, nil
}

// getUserPreferencesFromDB 從資料庫獲取用戶偏好
func (s *ChatService) getUserPreferencesFromDB(ctx context.Context, userID string) (map[string]interface{}, error) {
	var user models.User
	err := s.db.NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("查詢用戶偏好失敗: %w", err)
	}

	preferences := map[string]interface{}{
		"scene_style":      "romantic",
		"response_length":  "medium",
		"emotion_tracking": true,
	}

	// 合併用戶偏好
	if user.Preferences != nil {
		for key, value := range user.Preferences {
			preferences[key] = value
		}
	}

	return preferences, nil
}

// generateSessionTitle 根據角色生成會話標題
func (s *ChatService) generateSessionTitle(characterID string) string {
	characterNames := map[string]string{
		"character_01": "與沈宸的對話",
		"character_02": "與林知遠的對話",
		"character_03": "與周曜的對話",
	}

	if title, exists := characterNames[characterID]; exists {
		return title
	}
	return "AI對話會話"
}

// GetOrCreateUserCharacterSession 獲取或創建用戶與角色的唯一會話
func (s *ChatService) GetOrCreateUserCharacterSession(ctx context.Context, userID, characterID string) (*models.ChatSession, error) {
	// 查找現有會話
	var session models.ChatSession
	err := s.db.NewSelect().
		Model(&session).
		Where("user_id = ? AND character_id = ?", userID, characterID).
		Scan(ctx)

	if err == nil {
		// 會話存在，更新活躍時間
		session.UpdatedAt = time.Now()
		_, updateErr := s.db.NewUpdate().
			Model(&session).
			Where("id = ?", session.ID).
			Exec(ctx)

		if updateErr != nil {
			utils.Logger.WithError(updateErr).Warn("更新會話活躍時間失敗")
		}

		return &session, nil
	}

	// 會話不存在，創建新的
	newSession := &models.ChatSession{
		ID:          utils.GenerateUUID(),
		UserID:      userID,
		CharacterID: characterID,
		Title:       s.generateSessionTitle(characterID),
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = s.db.NewInsert().
		Model(newSession).
		Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("創建用戶-角色會話失敗: %w", err)
	}

	utils.Logger.WithFields(logrus.Fields{
		"session_id":   newSession.ID,
		"user_id":      userID,
		"character_id": characterID,
	}).Info("創建新的用戶-角色專屬會話")

	return newSession, nil
}

// GetUserCharacterSessions 獲取用戶的所有角色對話會話
func (s *ChatService) GetUserCharacterSessions(ctx context.Context, userID string) ([]*models.ChatSession, error) {
	var sessions []*models.ChatSession

	err := s.db.NewSelect().
		Model(&sessions).
		Relation("Character").
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("查詢用戶會話失敗: %w", err)
	}

	utils.Logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"sessions_count": len(sessions),
	}).Debug("獲取用戶角色對話會話成功")

	return sessions, nil
}

// GetSessionStatistics 獲取會話統計信息
func (s *ChatService) GetSessionStatistics(ctx context.Context, sessionID string) (*SessionStatistics, error) {
	var session models.ChatSession
	err := s.db.NewSelect().
		Model(&session).
		Where("id = ?", sessionID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("查詢會話失敗: %w", err)
	}

	// 查詢消息統計
	var messageStats struct {
		TotalMessages   int `bun:"total_messages"`
		UserMessages    int `bun:"user_messages"`
		AIMessages      int `bun:"ai_messages"`
		AvgResponseTime int `bun:"avg_response_time"`
	}

	err = s.db.NewSelect().
		Model((*models.Message)(nil)).
		ColumnExpr("COUNT(*) as total_messages").
		ColumnExpr("COUNT(*) FILTER (WHERE role = 'user') as user_messages").
		ColumnExpr("COUNT(*) FILTER (WHERE role = 'assistant') as ai_messages").
		ColumnExpr("AVG(response_time_ms) FILTER (WHERE role = 'assistant') as avg_response_time").
		Where("session_id = ?", sessionID).
		Scan(ctx, &messageStats)

	if err != nil {
		return nil, fmt.Errorf("查詢消息統計失敗: %w", err)
	}

	return &SessionStatistics{
		SessionID:       sessionID,
		TotalMessages:   messageStats.TotalMessages,
		UserMessages:    messageStats.UserMessages,
		AIMessages:      messageStats.AIMessages,
		AvgResponseTime: time.Duration(messageStats.AvgResponseTime) * time.Millisecond,
		CreatedAt:       session.CreatedAt,
		LastMessageAt:   session.LastMessageAt,
	}, nil
}

// SessionStatistics 會話統計結構
type SessionStatistics struct {
	SessionID       string        `json:"session_id"`
	TotalMessages   int           `json:"total_messages"`
	UserMessages    int           `json:"user_messages"`
	AIMessages      int           `json:"ai_messages"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	CreatedAt       time.Time     `json:"created_at"`
	LastMessageAt   *time.Time    `json:"last_message_at"`
}
