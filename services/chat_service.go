package services

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"
    "sync"

    "github.com/clarencetw/thewavess-ai-core/models"
    "github.com/clarencetw/thewavess-ai-core/models/db"
    "github.com/clarencetw/thewavess-ai-core/utils"
    "github.com/sirupsen/logrus"
    "github.com/uptrace/bun"
)

// ChatMessage 聊天消息類型（內部使用）
type ChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatService 對話服務
type ChatService struct {
    db             *bun.DB
    openaiClient   *OpenAIClient
    grokClient     *GrokClient
    config         *ChatConfig
    nsfwClassifier *NSFWClassifier
    // 簡單的 NSFW 遲滯（會話內短期內直接走 Grok）
    nsfwSticky    map[string]time.Time
    nsfwStickyMu  sync.RWMutex
    nsfwStickyTTL time.Duration
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
}

// ProcessMessageRequest 處理消息請求
type ProcessMessageRequest struct {
	ChatID      string                 `json:"chat_id"`
	UserMessage string                 `json:"user_message"`
	CharacterID string                 `json:"character_id"`
	UserID      string                 `json:"user_id"`
	ChatMode    string                 `json:"chat_mode,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse 對話回應
type ChatResponse struct {
	ChatID       string        `json:"chat_id"`
	MessageID    string        `json:"message_id"`
	Content      string        `json:"content"`         // 統一的內容格式 (*動作*\n對話\n*場景描述*)
	Affection    int           `json:"affection"`       // 好感度 0-100
	AIEngine     string        `json:"ai_engine"`
	NSFWLevel    int           `json:"nsfw_level"`
	Confidence   float64       `json:"confidence"`      // NSFW 分級信心度
	ResponseTime time.Duration `json:"response_time"`
}


// NSFWAnalysis NSFW 分析結果 - 純AI分析，無關鍵字檢測
type NSFWAnalysis struct {
	Level     int     `json:"level"`     // NSFW 等級 1-5
	Intensity float64 `json:"intensity"` // 強度 0.0-1.0
	Reason    string  `json:"reason"`    // AI分析理由
}


// ContentAnalysis 內容分析結果
type ContentAnalysis struct {
	IsNSFW        bool     `json:"is_nsfw"`
	Intensity     int      `json:"intensity"`  // 1-5 級
	Categories    []string `json:"categories"` // romantic, suggestive, explicit
	ShouldUseGrok bool     `json:"should_use_grok"`
	Confidence    float64  `json:"confidence"`
}

// AIJSONResponse AI JSON 響應結構 - AI 直接輸出的 JSON 格式  
type AIJSONResponse struct {
	Content       string                 `json:"content"`            // 統一的內容格式 (*動作*\n對話\n*場景描述*)
	EmotionDelta  *EmotionDelta          `json:"emotion_delta"`      // 情感變化建議（好感度）
	Mood          string                 `json:"mood"`               // 當前心情
	Relationship  string                 `json:"relationship"`       // 當前關係狀態
	IntimacyLevel string                 `json:"intimacy_level"`     // 當前親密度
	Reasoning     string                 `json:"reasoning"`          // AI 的推理過程（可選）
	Metadata      map[string]interface{} `json:"metadata,omitempty"` // 額外元數據
}

// EmotionDelta 情感變化建議
type EmotionDelta struct {
	AffectionChange int `json:"affection_change"` // 好感度變化 (-10 to +10)
}


// ConversationContext 對話上下文
type ConversationContext struct {
	ChatID         string        `json:"chat_id"`
	UserID         string        `json:"user_id"`
	CharacterID    string        `json:"character_id"`
	RecentMessages []ChatMessage `json:"recent_messages"` // 統一的記憶來源
	Affection      int           `json:"affection"`       // 好感度 0-100
	ChatMode       string        `json:"chat_mode"`       // 聊天模式
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
			MaxTokens:   utils.GetEnvIntWithDefault("OPENAI_MAX_TOKENS", 1200),
			Temperature: utils.GetEnvFloatWithDefault("OPENAI_TEMPERATURE", 0.8),
		},
		Grok: struct {
			Model       string  `json:"model"`
			MaxTokens   int     `json:"max_tokens"`
			Temperature float64 `json:"temperature"`
		}{
			// 預設使用 grok-3；若需回退可在環境改為 grok-beta
			Model:       utils.GetEnvWithDefault("GROK_MODEL", "grok-3"),
			MaxTokens:   utils.GetEnvIntWithDefault("GROK_MAX_TOKENS", 2000),
			Temperature: utils.GetEnvFloatWithDefault("GROK_TEMPERATURE", 0.9),
		},
		NSFW: struct {
			DetectionThreshold float64 `json:"detection_threshold"`
			MaxIntensityLevel  int     `json:"max_intensity_level"`
		}{
			DetectionThreshold: 0.7,
			MaxIntensityLevel:  5,
		},
	}

	openaiClient := NewOpenAIClient()
	grokClient := NewGrokClient()
	nsfwClassifier := NewNSFWClassifier()

    service := &ChatService{
        db:             GetDB(),
        openaiClient:   openaiClient,
        grokClient:     grokClient,
        config:         config,
        nsfwClassifier: nsfwClassifier,
        nsfwSticky:     make(map[string]time.Time),
        nsfwStickyTTL:  3 * time.Minute,
    }
    
    // 啟動 NSFW 黏滯清理程序，防止記憶體洩漏
    go service.startNSFWStickyCleanup()
    
    return service
}

// GenerateWelcomeMessage 生成吸引人的歡迎消息
func (s *ChatService) GenerateWelcomeMessage(ctx context.Context, userID, characterID, chatID string) (*ChatResponse, error) {
	startTime := time.Now()

	utils.Logger.WithFields(logrus.Fields{
		"chat_id":      chatID,
		"user_id":      userID,
		"character_id": characterID,
	}).Info("生成歡迎消息")

	// 1. 獲取角色信息
	characterService := GetCharacterService()
	character, err := characterService.GetCharacter(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// 2. 構建歡迎消息請求（使用特殊的歡迎消息標識）
	welcomeRequest := &ProcessMessageRequest{
		ChatID:      chatID,
		UserMessage: "[SYSTEM_WELCOME_FIRST_MESSAGE]", // 系統標識，用於生成歡迎消息
		CharacterID: characterID,
		UserID:      userID,
		ChatMode:    "casual",
		Metadata: map[string]interface{}{
			"is_welcome":    true,
			"character_name": character.Name,
			"character_type": character.Type,
		},
	}

	// 3. 構建歡迎消息專用的對話上下文
	welcomeContext := &ConversationContext{
		ChatID:         chatID,
		UserID:         userID,
		CharacterID:    characterID,
		RecentMessages: []ChatMessage{}, // 新會話沒有歷史消息
		Affection:      50,              // 預設好感度
		ChatMode:       "casual",
	}

	// 4. 生成歡迎消息（使用基本NSFW分析，因為是歡迎消息）
	analysis := &ContentAnalysis{
		IsNSFW:        false,
		Intensity:     1,
		Categories:    []string{"welcoming"},
		ShouldUseGrok: false,
		Confidence:    1.0,
	}
	
	response, err := s.generatePersonalizedResponse(ctx, "openai", "[SYSTEM_WELCOME_FIRST_MESSAGE]", welcomeContext, analysis)
	
	if err != nil {
		// AI生成失敗，返回錯誤
		utils.Logger.WithError(err).Error("AI歡迎消息生成失敗")
		return nil, fmt.Errorf("failed to generate AI welcome message: %w", err)
	}

	// 5. 生成消息ID並保存到數據庫
	messageID := fmt.Sprintf("msg_%s_welcome", utils.GenerateUUID())
	
	// 保存歡迎消息到數據庫
	err = s.saveAssistantMessageToDB(ctx, welcomeRequest, messageID, response, 50, "openai", analysis, time.Since(startTime))

	if err != nil {
		utils.Logger.WithError(err).Error("保存歡迎消息失敗")
		// 不返回錯誤，因為消息已經生成了
	}

	// 6. 返回歡迎消息回應
	return &ChatResponse{
		ChatID:       chatID,
		MessageID:    messageID,
		Content:      response.Content,
		Affection:    50,
		AIEngine:     "openai",
		NSFWLevel:    1,
		ResponseTime: time.Since(startTime),
	}, nil
}


// ProcessMessage 處理用戶消息並生成回應 - 女性向AI聊天系統
func (s *ChatService) ProcessMessage(ctx context.Context, request *ProcessMessageRequest) (*ChatResponse, error) {
	startTime := time.Now()

	utils.Logger.WithFields(logrus.Fields{
		"chat_id":      request.ChatID,
		"user_id":      request.UserID,
		"character_id": request.CharacterID,
		"message_len":  len(request.UserMessage),
	}).Info("開始處理AI對話請求")

	// 1. NSFW 內容分析
	utils.Logger.WithField("user_message", request.UserMessage[:utils.Min(20, len(request.UserMessage))]).Info("即將調用NSFW分析函數")
    analysis, err := s.analyzeContent(request.UserMessage)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze content: %w", err)
    }
    utils.Logger.WithField("nsfw_level", analysis.Intensity).Info("NSFW分析完成，返回等級")

    // 命中 NSFW 時，對該會話施加短期遲滯，後續直接走 Grok
    if analysis.IsNSFW {
        s.markNSFWSticky(request.ChatID)
    }

	// 2. 生成訊息 ID
	conversationTurnID := utils.GenerateUUID()
	messageID := fmt.Sprintf("msg_%s_ai", conversationTurnID)
	userMessageID := fmt.Sprintf("msg_%s_user", conversationTurnID)

	// 3. 保存用戶訊息
	if err := s.saveUserMessageToDB(ctx, request, userMessageID, analysis); err != nil {
		utils.Logger.WithError(err).Error("保存用戶消息失敗：將降級為臨時上下文")
	}

	// 4. 構建對話上下文
	conversationContext, err := s.buildFemaleOrientedContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to build female-oriented context: %w", err)
	}

	// 5. 選擇 AI 引擎
	engine := s.selectAIEngine(analysis, conversationContext, nil)
	utils.Logger.WithFields(logrus.Fields{
		"selected_engine": engine,
	}).Info("引擎選擇完成")

	// 6. 生成 AI 回應
	response, err := s.generatePersonalizedResponse(ctx, engine, request.UserMessage, conversationContext, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate personalized response: %w", err)
	}

	// 7. 更新好感度
	newAffection := s.updateAffection(conversationContext.Affection, response)

	// 8. 保存 AI 回應
	err = s.saveAssistantMessageToDB(ctx, request, messageID, response, newAffection, engine, analysis, time.Since(startTime))
	if err != nil {
		utils.Logger.WithError(err).Error("保存對話到資料庫失敗")
	}

	// 9. 構建回應結果
	chatResponse := &ChatResponse{
		ChatID:       request.ChatID,
		MessageID:    messageID,
		Content:      response.Content,
		Affection:    newAffection,
		AIEngine:     engine,
		NSFWLevel:    analysis.Intensity,
		Confidence:   analysis.Confidence,
		ResponseTime: time.Since(startTime),
	}

	// 記錄對話性能事件
	utils.LogPerformanceEvent(
		"chat_processing",
		chatResponse.ResponseTime.Milliseconds(),
		logrus.Fields{
			"chat_id":      request.ChatID,
			"user_id":      request.UserID,
			"character_id": request.CharacterID,
			"ai_engine":    engine,
			"nsfw_level":   analysis.Intensity,
			"affection":    newAffection,
		},
	)

	utils.Logger.WithFields(logrus.Fields{
		"chat_id":       request.ChatID,
		"character_id":  request.CharacterID,
		"nsfw_level":    analysis.Intensity,
		"ai_engine":     engine,
		"affection":     newAffection,
		"response_time": chatResponse.ResponseTime.Milliseconds(),
	}).Info("AI對話處理完成")

	return chatResponse, nil
}

// analyzeContent 分析消息內容 - 關鍵字布林門分析（極簡高效）
func (s *ChatService) analyzeContent(message string) (*ContentAnalysis, error) {
    ctx := context.Background()
    utils.Logger.WithField("message_preview", message[:utils.Min(30, len(message))]).Info("開始關鍵字 NSFW 內容分析")

    // 使用極簡關鍵字分級器
    result, err := s.nsfwClassifier.ClassifyContent(ctx, message)
    if err != nil {
        utils.Logger.WithError(err).Error("NSFW 分級失敗")
        return nil, fmt.Errorf("NSFW classification failed: %w", err)
    }

    // 基於分級結果
    analysis := &ContentAnalysis{
        IsNSFW:        result.Level >= 5,
        Intensity:     result.Level,
        Categories:    []string{"keyword_gate"},
        ShouldUseGrok: result.Level >= 5,
        Confidence:    result.Confidence,
    }

	// 記錄分析結果
    utils.Logger.WithFields(logrus.Fields{
        "message_preview": message[:utils.Min(50, len(message))],
        "nsfw_level":      result.Level,
        "is_nsfw":         analysis.IsNSFW,
        "confidence":      result.Confidence,
        "should_use_grok": analysis.ShouldUseGrok,
        "analysis_method": "keyword_gate",
        "reason":          result.Reason,
    }).Info("NSFW 內容分析完成")

	return analysis, nil
}


// buildFemaleOrientedContext 構建對話上下文數據
// 收集好感度和對話歷史，組裝給 AI 使用的上下文結構
func (s *ChatService) buildFemaleOrientedContext(ctx context.Context, request *ProcessMessageRequest) (*ConversationContext, error) {
	// 1. 從 relationships 表獲取好感度數值（0-100）
	affection, err := s.getAffectionFromDB(ctx, request.UserID, request.CharacterID, request.ChatID)
	if err != nil {
		utils.Logger.WithError(err).Warn("獲取好感度失敗，使用默認值")
		affection = 50 // 默認中性好感度，確保系統穩定性
	}

	// 2. 從 messages 表獲取最近對話記憶（限制 5 條，控制 AI 上下文大小）
	recentMemories, err := s.getRecentMemoriesFromDB(ctx, request.ChatID, 5)
	if err != nil {
		utils.Logger.WithError(err).Warn("獲取會話歷史失敗，使用內存數據")
		recentMemories = s.getRecentMemories(request.ChatID, request.UserID, request.CharacterID, 5)
	}

	// 3. 組裝標準化對話上下文數據結構
	return &ConversationContext{
		ChatID:         request.ChatID,        // 會話識別碼
		UserID:         request.UserID,        // 用戶識別碼
		CharacterID:    request.CharacterID,   // 角色識別碼
		RecentMessages: recentMemories,        // 最近對話記憶（最多 5 條）
		Affection:      affection,             // 當前好感度（0-100）
		ChatMode:       request.ChatMode,      // 聊天模式設定
	}, nil
}

// getAffectionFromDB 從 relationships 表獲取好感度
func (s *ChatService) getAffectionFromDB(ctx context.Context, userID, characterID, chatID string) (int, error) {
	var relationship db.RelationshipDB
	
	err := s.db.NewSelect().
		Model(&relationship).
		Where("user_id = ? AND character_id = ? AND chat_id = ?", userID, characterID, chatID).
		Scan(ctx)
	
	if err != nil {
		// 如果關係記錄不存在，創建一個新的
		newRelationship := &db.RelationshipDB{
			ID:          utils.GenerateRelationshipID(),
			UserID:      userID,
			CharacterID: characterID,
			ChatID:      &chatID, // 直接設置 ChatID 指標
			Affection:   50, // 默認好感度
			Mood:        "neutral",
			Relationship: "stranger",
			IntimacyLevel: "distant",
		}
		
		_, insertErr := s.db.NewInsert().
			Model(newRelationship).
			Exec(ctx)
		
		if insertErr != nil {
			utils.Logger.WithError(insertErr).Error("創建新關係記錄失敗")
			return 50, nil
		}
		
		utils.Logger.WithFields(map[string]interface{}{
			"user_id":      userID,
			"character_id": characterID,
			"chat_id":      chatID,
			"affection":    50,
		}).Info("創建新的用戶-角色關係記錄")
		
		return 50, nil
	}
	
	return relationship.Affection, nil
}

// getRecentMemories 獲取最近的對話記憶 - fallback版本（當資料庫查詢失敗時使用）
func (s *ChatService) getRecentMemories(chatID, userID, characterID string, limit int) []ChatMessage {
	// 作為 getRecentMemoriesFromDB 失敗時的 fallback
	// 在生產環境中，這裡可以嘗試從快取或其他資料源獲取
	utils.Logger.WithFields(logrus.Fields{
		"chat_id":      chatID,
		"user_id":      userID,
		"character_id": characterID,
		"limit":        limit,
	}).Debug("使用 fallback 記憶獲取（返回空歷史）")

	return []ChatMessage{}
}

// selectAIEngine 以極簡布林門選擇 AI 引擎（OpenAI 預設，明確 NSFW → Grok）
func (s *ChatService) selectAIEngine(analysis *ContentAnalysis, conv *ConversationContext, _ map[string]interface{}) string {
    // 0. 角色標籤預分流（含 nsfw 標籤直接 Grok）
    if conv != nil {
        characterService := GetCharacterService()
        if character, err := characterService.GetCharacter(context.Background(), conv.CharacterID); err == nil {
            for _, tag := range character.Metadata.Tags {
                t := strings.ToLower(tag)
                if t == "nsfw" || t == "adult" {
                    utils.Logger.WithFields(map[string]interface{}{
                        "character_id": conv.CharacterID,
                        "reason":       "character_tag",
                    }).Info("選擇 Grok 引擎：角色標籤預分流")
                    return "grok"
                }
            }
        }
    }

    // 1. 會話遲滯：近期命中 NSFW 的會話短期內直接 Grok
    if conv != nil && s.isNSFWSticky(conv.ChatID) {
        utils.Logger.WithFields(map[string]interface{}{
            "chat_id": conv.ChatID,
            "reason":  "nsfw_sticky",
        }).Info("選擇 Grok 引擎：NSFW 遲滯生效")
        return "grok"
    }
    
    // 2. 布林 NSFW 決策：只要是 NSFW 就 Grok，否則 OpenAI
    if analysis.IsNSFW || analysis.ShouldUseGrok {
        utils.Logger.WithFields(map[string]interface{}{
            "nsfw":       analysis.IsNSFW,
            "categories": analysis.Categories,
            "confidence": analysis.Confidence,
            "reason":     "nsfw_boolean_gate",
        }).Info("選擇 Grok 引擎：露骨內容")
        return "grok"
    }

    // 預設使用 OpenAI（安全/非露骨）
    utils.Logger.WithFields(map[string]interface{}{
        "nsfw":       analysis.IsNSFW,
        "categories": analysis.Categories,
    }).Info("選擇 OpenAI 引擎：預設")
    return "openai"
}

// 標記會話在短期內直接使用 Grok
func (s *ChatService) markNSFWSticky(chatID string) {
    if chatID == "" {
        return
    }
    s.nsfwStickyMu.Lock()
    s.nsfwSticky[chatID] = time.Now().Add(s.nsfwStickyTTL)
    s.nsfwStickyMu.Unlock()
}

// 檢查會話是否處於 NSFW 遲滯期
func (s *ChatService) isNSFWSticky(chatID string) bool {
    if chatID == "" {
        return false
    }
    s.nsfwStickyMu.RLock()
    until, ok := s.nsfwSticky[chatID]
    s.nsfwStickyMu.RUnlock()
    return ok && time.Now().Before(until)
}

// startNSFWStickyCleanup 定期清理過期的 NSFW 黏滯狀態，防止記憶體洩漏
func (s *ChatService) startNSFWStickyCleanup() {
    ticker := time.NewTicker(1 * time.Minute) // 每 1 分鐘清理一次，提高效能
    defer ticker.Stop()
    
    for range ticker.C {
        s.cleanupExpiredNSFWSticky()
    }
}

// cleanupExpiredNSFWSticky 清理過期的 NSFW 黏滯狀態
func (s *ChatService) cleanupExpiredNSFWSticky() {
    now := time.Now()
    s.nsfwStickyMu.Lock()
    defer s.nsfwStickyMu.Unlock()
    
    cleanedCount := 0
    for chatID, until := range s.nsfwSticky {
        if now.After(until) {
            delete(s.nsfwSticky, chatID)
            cleanedCount++
        }
    }
    
    if cleanedCount > 0 {
        utils.Logger.WithFields(map[string]interface{}{
            "cleaned_count":    cleanedCount,
            "remaining_count":  len(s.nsfwSticky),
        }).Debug("清理過期的 NSFW 黏滯狀態")
    }
}

// CharacterResponseData 角色回應數據
type CharacterResponseData struct {
	Content        string                 `json:"content"`            // 統一的內容格式 (*動作*\n對話\n*場景描述*)
	JSONProcessed  bool                   `json:"json_processed"`     // 標記是否已由JSON處理器處理過情感狀態
	EmotionDelta   *EmotionDelta          `json:"emotion_delta"`      // AI 建議的情感變化（好感度）
	Mood           string                 `json:"mood"`               // AI 建議的心情
	Relationship   string                 `json:"relationship"`       // AI 建議的關係狀態
	IntimacyLevel  string                 `json:"intimacy_level"`     // AI 建議的親密度
	Reasoning      string                 `json:"reasoning"`          // AI 推理過程
	Metadata       map[string]interface{} `json:"metadata,omitempty"` // 額外元數據
}

// generatePersonalizedResponse 生成個性化女性向回應
func (s *ChatService) generatePersonalizedResponse(ctx context.Context, engine, userMessage string, context *ConversationContext, analysis *ContentAnalysis) (*CharacterResponseData, error) {

    // 根據引擎構建專屬prompt（OpenAI 固定 L1、Grok 固定 L5）
    nsfwLevelForPrompt := 1
    if engine == "grok" {
        nsfwLevelForPrompt = 5
    }
    prompt := s.buildEngineSpecificPrompt(engine, context.CharacterID, userMessage, context, nsfwLevelForPrompt, context.ChatMode)

	var responseText string
	var err error

	if engine == "openai" {
		// 使用 OpenAI (Level 1-4)
		responseText, err = s.generateOpenAIResponse(ctx, prompt, context, userMessage)
		if err != nil {
			utils.Logger.WithError(err).Error("OpenAI 回應生成失敗")
			return nil, fmt.Errorf("failed OpenAI API call: %w", err)
		}
	} else if engine == "grok" {
		// 使用 Grok (Level 5)
		responseText, err = s.generateGrokResponse(ctx, prompt, context, userMessage)
		if err != nil {
			utils.Logger.WithError(err).Error("Grok 回應生成失敗")
			return nil, fmt.Errorf("failed Grok API call: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unknown AI engine: %s", engine)
	}

	// 首先嘗試 JSON 解析
	if jsonResponse, err := s.parseJSONResponse(responseText, context, analysis.Intensity); err == nil {
		utils.Logger.WithFields(map[string]interface{}{
			"engine":           engine,
			"affection_change": jsonResponse.EmotionDelta.AffectionChange,
			"mood":             jsonResponse.Mood,
		}).Info("成功解析 AI JSON 響應")
		return jsonResponse, nil
	} else {
		utils.Logger.WithError(err).Error("JSON 解析失敗")
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w", err)
	}
}

// parseJSONResponse 解析 AI 的 JSON 響應
func (s *ChatService) parseJSONResponse(responseText string, context *ConversationContext, nsfwLevel int) (*CharacterResponseData, error) {
	var jsonResp AIJSONResponse

    // 從回應文字中嚴格提取 JSON 區段並解析
    extractedJSON, extractErr := utils.ExtractJSONFromText(responseText)
    if extractErr != nil {
        utils.Logger.WithFields(map[string]interface{}{
            "original_text": responseText,
            "parse_error":   extractErr.Error(),
        }).Error("Failed to locate JSON in AI response")
        return nil, fmt.Errorf("unable to find valid JSON structure in response: %w", extractErr)
    }

    if err := json.Unmarshal([]byte(extractedJSON), &jsonResp); err != nil {
        utils.Logger.WithFields(map[string]interface{}{
            "original_text":  responseText,
            "extracted_json": extractedJSON,
            "parse_error":    err.Error(),
        }).Error("Failed to parse JSON response from AI")
        return nil, fmt.Errorf("JSON parsing failed: %w", err)
    }

	// 驗證必要字段
	if jsonResp.Content == "" {
		return nil, fmt.Errorf("JSON response missing content field")
	}

	// 構建 CharacterResponseData
	response := &CharacterResponseData{
		Content:       jsonResp.Content,
		JSONProcessed: true,
		EmotionDelta:  jsonResp.EmotionDelta,
		Mood:          jsonResp.Mood,
		Relationship:  jsonResp.Relationship,
		IntimacyLevel: jsonResp.IntimacyLevel,
		Reasoning:     jsonResp.Reasoning,
		Metadata:      jsonResp.Metadata,
	}

	// 驗證情感變化數據
	if response.EmotionDelta == nil {
		response.EmotionDelta = &EmotionDelta{
			AffectionChange: 1, // 默認小幅正面變化
		}
	}

	return response, nil
}

// SaveFallbackAssistantMessage 將 fallback 文字回應保存到資料庫（AI 失敗保底）
func (s *ChatService) SaveFallbackAssistantMessage(ctx context.Context, request *ProcessMessageRequest, messageID string, content string) (int, error) {
    defaultAffection := 50

    // 超級簡化：直接保存 fallback 訊息，不需要複雜結構
    message := &models.Message{
        ID:                 messageID,
        ChatID:             request.ChatID,
        Role:               "assistant",
        Dialogue:           content,
        NSFWLevel:          1,         // fallback 固定為安全等級
        AIEngine:           "fallback",
        ResponseTimeMs:     0,         // fallback 回應時間固定為 0
        CreatedAt:          time.Now(),
    }

    _, err := s.db.NewInsert().Model(message).Exec(ctx)
    if err != nil {
        utils.Logger.WithError(err).Error("保存 fallback 訊息失敗")
        return defaultAffection, err
    }

    return defaultAffection, nil
}

// 已移除：cleanGrokResponse / extractJSONFromText
// 統一改用 utils.ExtractJSONFromText 以提高精確度與一致性

// buildEngineSpecificPrompt 根據 AI 引擎構建專屬 prompt
// OpenAI: 情感細膩，Grok: 大膽創意
func (s *ChatService) buildEngineSpecificPrompt(engine, characterID, userMessage string, conversationContext *ConversationContext, nsfwLevel int, chatMode string) string {
	// 記憶上下文完全通過 conversationContext.RecentMessages 提供
	characterService := GetCharacterService()
	ctx := context.Background()

	if engine == "grok" {
		// 使用Grok專屬prompt構建器
		promptBuilder := NewGrokPromptBuilder(characterService)
		return promptBuilder.
			WithCharacter(ctx, characterID).
			WithContext(conversationContext).
			WithNSFWLevel(nsfwLevel).
			WithUserMessage(userMessage).
			WithChatMode(chatMode).
			Build(ctx)
	} else {
		// 使用OpenAI專屬prompt構建器
		promptBuilder := NewOpenAIPromptBuilder(characterService)
		return promptBuilder.
			WithCharacter(ctx, characterID).
			WithContext(conversationContext).
			WithNSFWLevel(nsfwLevel).
			WithUserMessage(userMessage).
			WithChatMode(chatMode).
			Build(ctx)
	}
}

// generateGrokResponse 生成Grok回應
func (s *ChatService) generateGrokResponse(ctx context.Context, prompt string, context *ConversationContext, currentUserMessage string) (string, error) {
	// 構建 Grok 請求
	messages := []GrokMessage{
		{
			Role:    "system",
			Content: prompt,
		},
	}

	// 添加最近的對話歷史作為上下文
	if len(context.RecentMessages) > 0 {
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

	// 重要修復：加入當前用戶消息
	if strings.TrimSpace(currentUserMessage) != "" {
		messages = append(messages, GrokMessage{
			Role:    "user",
			Content: currentUserMessage,
		})
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
		"chat_id":      context.ChatID,
		"character_id": context.CharacterID,
		"user_id":      context.UserID,
	}).Info("調用 Grok API")

	response, err := s.grokClient.GenerateResponse(ctx, request)
	if err != nil {
		utils.Logger.WithError(err).Error("Grok API 調用失敗")
		return "", fmt.Errorf("failed Grok API call: %w", err)
	}

	// 從回應中提取對話內容
	if len(response.Choices) > 0 {
		dialogue := response.Choices[0].Message.Content

		utils.Logger.WithFields(map[string]interface{}{
			"chat_id":      context.ChatID,
			"response_len": len(dialogue),
			"tokens_used":  response.Usage.TotalTokens,
		}).Info("Grok API 響應成功")

		return dialogue, nil
	}

	// 如果沒有回應內容，返回錯誤
	utils.Logger.Warn("Grok API 返回空回應")
	return "", fmt.Errorf("empty response from Grok API")
}

// generateOpenAIResponse 生成OpenAI回應
func (s *ChatService) generateOpenAIResponse(ctx context.Context, prompt string, context *ConversationContext, currentUserMessage string) (string, error) {
	// 構建 OpenAI 請求
	messages := []OpenAIMessage{
		{
			Role:    "system",
			Content: prompt,
		},
	}

	// 添加最近的對話歷史作為上下文
	if len(context.RecentMessages) > 0 {
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

	// 重要修復：加入當前用戶消息
	if strings.TrimSpace(currentUserMessage) != "" {
		messages = append(messages, OpenAIMessage{
			Role:    "user",
			Content: currentUserMessage,
		})
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
		"chat_id":      context.ChatID,
		"character_id": context.CharacterID,
		"user_id":      context.UserID,
	}).Info("調用 OpenAI API")

	response, err := s.openaiClient.GenerateResponse(ctx, request)
	if err != nil {
		utils.Logger.WithError(err).Error("OpenAI API 調用失敗")
		return "", fmt.Errorf("failed OpenAI API call: %w", err)
	}

	// 從回應中提取對話內容
	if len(response.Choices) > 0 {
		dialogue := response.Choices[0].Message.Content

		utils.Logger.WithFields(map[string]interface{}{
			"chat_id":      context.ChatID,
			"response_len": len(dialogue),
			"tokens_used":  response.Usage.TotalTokens,
		}).Info("OpenAI API 響應成功")

		return dialogue, nil
	}

	// 如果沒有回應內容，返回錯誤
	utils.Logger.Warn("OpenAI API 返回空回應")
	return "", fmt.Errorf("empty response from OpenAI API")
}

// updateAffection 好感度更新
func (s *ChatService) updateAffection(currentAffection int, response *CharacterResponseData) int {
	newAffection := currentAffection
	
	// 從AI回應中獲取好感度變化
	if response != nil && response.JSONProcessed && response.EmotionDelta != nil {
		newAffection += response.EmotionDelta.AffectionChange
	}

	// 確保好感度在有效範圍內 (0-100)
	if newAffection > 100 {
		newAffection = 100
	} else if newAffection < 0 {
		newAffection = 0
	}

	return newAffection
}




// saveUserMessageToDB 先保存用戶消息（以便上下文讀取包含本輪）
func (s *ChatService) saveUserMessageToDB(ctx context.Context, request *ProcessMessageRequest, userMessageID string, analysis *ContentAnalysis) error {
	// 使用 RunInTx 處理事務
	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		userMessage := &db.MessageDB{
			ID:        userMessageID,
			ChatID:    request.ChatID,
			Role:      "user",
			Dialogue:  request.UserMessage,
			NSFWLevel: analysis.Intensity,
			CreatedAt: time.Now(),
		}

		if _, err := tx.NewInsert().Model(userMessage).Exec(ctx); err != nil {
			return fmt.Errorf("failed to save user message: %w", err)
		}

		// 更新會話統計（只加用戶部分）
		if _, err := tx.NewUpdate().
			Model((*db.ChatDB)(nil)).
			Set("message_count = message_count + 1").
			Set("total_characters = total_characters + ?", len(request.UserMessage)).
			Set("last_message_at = ?", time.Now()).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", request.ChatID).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to update session stats (user): %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// saveAssistantMessageToDB 保存 AI 回應（第二步）
func (s *ChatService) saveAssistantMessageToDB(ctx context.Context, request *ProcessMessageRequest, messageID string, response *CharacterResponseData, affection int, engine string, analysis *ContentAnalysis, responseTime time.Duration) error {
	// 使用 RunInTx 處理事務
	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 1. 保存 AI 消息（記錄情感狀態變化）
		aiMessage := &db.MessageDB{
			ID:               messageID,
			ChatID:           request.ChatID,
			Role:             "assistant",
			Dialogue:         response.Content, // 統一內容存儲在 Dialogue 字段
			SceneDescription: nil,              // 不再單獨存儲
			Action:           nil,              // 不再單獨存儲
			EmotionalState: map[string]interface{}{
				"mood_change": response.Mood,
				"reasoning":   response.Reasoning,
				"trigger":     "AI response generation",
			},
			AIEngine:       engine,
			ResponseTimeMs: int(responseTime.Milliseconds()),
			NSFWLevel:      analysis.Intensity,
			CreatedAt:      time.Now(),
		}

		if _, err := tx.NewInsert().Model(aiMessage).Exec(ctx); err != nil {
			return fmt.Errorf("failed to save AI message: %w", err)
		}

		// 2. 更新 relationships 表的持久狀態 (使用AI建議的所有關係狀態)
		updateQuery := tx.NewUpdate().
			Model((*db.RelationshipDB)(nil)).
			Set("affection = ?", affection).
			Set("last_interaction = ?", time.Now()).
			Set("total_interactions = total_interactions + 1").
			Set("updated_at = ?", time.Now()).
			Where("user_id = ? AND character_id = ? AND chat_id = ?", request.UserID, request.CharacterID, request.ChatID)
		
		// 如果AI提供了 mood，使用AI建議的mood
		if response.Mood != "" {
			updateQuery = updateQuery.Set("mood = ?", response.Mood)
		}
		
		// 如果AI提供了 relationship，使用AI建議的relationship
		if response.Relationship != "" {
			updateQuery = updateQuery.Set("relationship = ?", response.Relationship)
		}
		
		// 如果AI提供了 intimacy_level，使用AI建議的intimacy_level
		if response.IntimacyLevel != "" {
			updateQuery = updateQuery.Set("intimacy_level = ?", response.IntimacyLevel)
		}
		
		if _, err := updateQuery.Exec(ctx); err != nil {
			return fmt.Errorf("failed to update relationship: %w", err)
		}

		// 3. 更新會話統計
		if _, err := tx.NewUpdate().
			Model((*db.ChatDB)(nil)).
			Set("message_count = message_count + 1").
			Set("total_characters = total_characters + ?", len(response.Content)).
			Set("last_message_at = ?", time.Now()).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", request.ChatID).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to update session stats (AI): %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	utils.Logger.WithFields(logrus.Fields{
		"chat_id":    request.ChatID,
		"message_id": messageID,
		"ai_msg_len": len(response.Content),
		"nsfw_level": analysis.Intensity,
		"ai_engine":  engine,
	}).Info("AI 消息已保存到資料庫")

	return nil
}

// getRecentMemoriesFromDB 從資料庫獲取最近對話記憶，新消息在前
func (s *ChatService) getRecentMemoriesFromDB(ctx context.Context, chatID string, limit int) ([]ChatMessage, error) {
	var messages []db.MessageDB

	// 查詢最近的消息，使用 limit*2 確保獲取足夠的用戶和AI消息對
	err := s.db.NewSelect().
		Model(&messages).
		Where("chat_id = ?", chatID).
		Order("created_at DESC").
		Limit(limit * 2). // 擴大查詢範圍，考慮用戶和AI消息交替出現
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query chat history from database: %w", err)
	}

	// 轉換為 ChatMessage 格式，保持新消息在前的順序
	chatMessages := make([]ChatMessage, 0, len(messages))
	for _, msg := range messages { // 保持資料庫查詢順序：新的消息在前
		// 跳過空消息內容
		if strings.TrimSpace(msg.Dialogue) == "" {
			continue
		}

		chatMessages = append(chatMessages, ChatMessage{
			Role:      msg.Role,
			Content:   msg.Dialogue,
			CreatedAt: msg.CreatedAt,
		})
	}

	// 如果消息過多，保留最新的對話
	if len(chatMessages) > limit {
		chatMessages = chatMessages[:limit]
	}

	utils.Logger.WithFields(logrus.Fields{
		"chat_id":         chatID,
		"requested_limit": limit,
		"actual_count":    len(chatMessages),
		"db_query_limit":  limit * 2,
	}).Debug("成功從資料庫獲取會話歷史")

	return chatMessages, nil
}
