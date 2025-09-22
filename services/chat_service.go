package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

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
	Action    string    `json:"action,omitempty"`
}

// ChatService 對話服務
type ChatService struct {
	db             *bun.DB
	openaiClient   *OpenAIClient
	grokClient     *GrokClient
	mistralClient  *MistralClient
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

	Mistral struct {
		Model       string  `json:"model"`
		MaxTokens   int     `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
	} `json:"mistral"`

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
	Content      string        `json:"content"`   // 統一的內容格式 (*動作*\n對話\n*場景描述*)
	Affection    int           `json:"affection"` // 好感度 0-100
	AIEngine     string        `json:"ai_engine"`
	NSFWLevel    int           `json:"nsfw_level"`
	Confidence   float64       `json:"confidence"` // NSFW 分級信心度
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
	Mood           string        `json:"mood"`
	Relationship   string        `json:"relationship"`
	IntimacyLevel  string        `json:"intimacy_level"`
}

var (
	chatServiceInstance *ChatService
	chatServiceOnce     sync.Once
)

// GetChatService 獲取單例 ChatService 實例
func GetChatService() *ChatService {
	chatServiceOnce.Do(func() {
		chatServiceInstance = NewChatService()
	})
	return chatServiceInstance
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
	mistralClient := NewMistralClient()
	nsfwClassifier := GetNSFWClassifier()

	service := &ChatService{
		db:             GetDB(),
		openaiClient:   openaiClient,
		grokClient:     grokClient,
		mistralClient:  mistralClient,
		config:         config,
		nsfwClassifier: nsfwClassifier,
		nsfwSticky:     make(map[string]time.Time),
		nsfwStickyTTL:  5 * time.Minute,
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
			"is_welcome":     true,
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
		Mood:           "neutral",
		Relationship:   "stranger",
		IntimacyLevel:  "distant",
	}

	// 4. 生成歡迎消息（使用基本NSFW分析，因為是歡迎消息）
	analysis := &ContentAnalysis{
		IsNSFW:        false,
		Intensity:     1,
		Categories:    []string{"welcoming"},
		ShouldUseGrok: false,
		Confidence:    1.0,
	}

	// 歡迎訊息使用 OpenAI 引擎，L1 等級
	response, err := s.generatePersonalizedResponse(ctx, "openai", "[SYSTEM_WELCOME_FIRST_MESSAGE]", welcomeContext, analysis)

	if err != nil {
		// AI生成失敗，返回錯誤
		utils.Logger.WithError(err).Error("AI歡迎消息生成失敗")
		return nil, fmt.Errorf("failed to generate AI welcome message: %w", err)
	}

	// 5. 生成消息ID並保存到數據庫
	messageID := fmt.Sprintf("msg_%s_welcome", utils.GenerateUUID())

	// 保存歡迎消息到數據庫 (歡迎消息固定 L1)
	err = s.saveAssistantMessageToDB(ctx, welcomeRequest, messageID, response, 1, "openai", analysis, time.Since(startTime))

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
		Confidence:   1.0, // 歡迎消息固定信心度
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

	// NSFW Sticky 機制已移到 OpenAI 拒絕後的 fallback 處理中
	// 這樣可以確保只有真正被 OpenAI 拒絕的內容才會觸發 sticky

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
	selectedEngine := s.selectAIEngine(analysis, conversationContext, request.UserMessage)

	utils.Logger.WithFields(logrus.Fields{
		"selected_engine": selectedEngine,
		"nsfw_level":      analysis.Intensity,
	}).Info("引擎選擇完成")

	// 6. 生成 AI 回應
	response, err := s.generatePersonalizedResponse(ctx, selectedEngine, request.UserMessage, conversationContext, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate personalized response: %w", err)
	}

	// 7. 更新關係狀態（包含好感度、心情、關係、親密度）
	newAffection := s.updateAffection(conversationContext.Affection, response)
	err = s.updateRelationshipState(ctx, request, response, newAffection)
	if err != nil {
		utils.Logger.WithError(err).Error("更新關係狀態失敗")
		// 不中斷流程，繼續處理
	}

	// 8. 保存 AI 回應（智能 NSFW 等級記錄）
	err = s.saveAssistantMessageToDB(ctx, request, messageID, response, newAffection, response.ActualEngine, analysis, time.Since(startTime))
	if err != nil {
		utils.Logger.WithError(err).Error("保存對話到資料庫失敗")
	}

	// 9. 構建回應結果
	chatResponse := &ChatResponse{
		ChatID:       request.ChatID,
		MessageID:    messageID,
		Content:      response.Content,
		Affection:    newAffection,
		AIEngine:     response.ActualEngine,
		NSFWLevel:    analysis.Intensity,
		Confidence:   analysis.Confidence,
		ResponseTime: time.Since(startTime),
	}

	// 記錄對話性能事件
	utils.LogPerformanceMetric(
		"chat_processing",
		chatResponse.ResponseTime,
		map[string]interface{}{
			"chat_id":      request.ChatID,
			"user_id":      request.UserID,
			"character_id": request.CharacterID,
			"ai_engine":    response.ActualEngine,
			"nsfw_level":   analysis.Intensity,
			"affection":    newAffection,
		},
	)

	utils.Logger.WithFields(logrus.Fields{
		"chat_id":       request.ChatID,
		"character_id":  request.CharacterID,
		"nsfw_level":    analysis.Intensity,
		"ai_engine":     response.ActualEngine,
		"affection":     newAffection,
		"response_time": chatResponse.ResponseTime.Milliseconds(),
	}).Info("AI對話處理完成")

	return chatResponse, nil
}

// analyzeContent 分析消息內容 - RAG 語意檢索分級
func (s *ChatService) analyzeContent(message string) (*ContentAnalysis, error) {
	ctx := context.Background()
	utils.Logger.WithField("message_preview", message[:utils.Min(30, len(message))]).Info("開始 RAG NSFW 內容分析")

	// 使用語意檢索分級器
	result, err := s.nsfwClassifier.ClassifyContent(ctx, message)
	if err != nil {
		utils.Logger.WithError(err).Error("NSFW 分級失敗")
		return nil, fmt.Errorf("NSFW classification failed: %w", err)
	}

	// 基於 RAG 分級結果
	// 附帶 chunk ID 與 reason 供後續路由與審核
	categories := []string{"semantic_rag_analysis"}
	if result.ChunkID != "" {
		categories = append(categories, "rag_chunk:"+result.ChunkID)
	}
	if result.Reason != "" {
		categories = append(categories, result.Reason)
		switch result.Reason {
		case "illegal_underage", "illegal_underage_en", "bestiality", "sexual_violence_or_incest", "incest_family_roles", "incest_step_roles_en", "rape":
			categories = append(categories, "illegal_content")
		}
	}

	analysis := &ContentAnalysis{
		IsNSFW:        result.Level >= 3, // L3以上視為需進入 Grok 的親密 NSFW
		Intensity:     result.Level,
		Categories:    categories,
		ShouldUseGrok: result.Level >= 3, // L3以上使用Grok（若為非法，稍後會阻擋）
		Confidence:    result.Confidence,
	}

	// 記錄分析結果
	utils.Logger.WithFields(logrus.Fields{
		"message_preview": message[:utils.Min(50, len(message))],
		"nsfw_level":      result.Level,
		"is_nsfw":         analysis.IsNSFW,
		"confidence":      result.Confidence,
		"should_use_grok": analysis.ShouldUseGrok,
		"analysis_method": "semantic_rag",
		"reason":          result.Reason,
	}).Info("NSFW 內容分析完成")

	return analysis, nil
}

// buildFemaleOrientedContext 構建對話上下文數據
// 收集好感度和對話歷史，組裝給 AI 使用的上下文結構
func (s *ChatService) buildFemaleOrientedContext(ctx context.Context, request *ProcessMessageRequest) (*ConversationContext, error) {
	// 1. 從 relationships 表獲取當前關係狀態
	relationshipState, err := s.getOrCreateRelationshipState(ctx, request.UserID, request.CharacterID, request.ChatID)
	if err != nil {
		utils.Logger.WithError(err).Warn("獲取關係狀態失敗，使用默認值")
		relationshipState = &db.RelationshipDB{
			Affection:     50,
			Mood:          "neutral",
			Relationship:  "stranger",
			IntimacyLevel: "distant",
		}
	}

	// 2. 從 messages 表獲取最近對話記憶（限制 5 條，控制 AI 上下文大小）
	recentMemories, err := s.getRecentMemoriesFromDB(ctx, request.ChatID, 5)
	if err != nil {
		utils.Logger.WithError(err).Warn("獲取會話歷史失敗，使用空歷史")
		recentMemories = []ChatMessage{} // 直接使用空歷史，簡化邏輯
	}

	// 3. 組裝標準化對話上下文數據結構
	return &ConversationContext{
		ChatID:         request.ChatID,             // 會話識別碼
		UserID:         request.UserID,             // 用戶識別碼
		CharacterID:    request.CharacterID,        // 角色識別碼
		RecentMessages: recentMemories,             // 最近對話記憶（最多 5 條）
		Affection:      relationshipState.Affection, // 當前好感度（0-100）
		ChatMode:       request.ChatMode,           // 聊天模式設定
		Mood:           relationshipState.Mood,
		Relationship:   relationshipState.Relationship,
		IntimacyLevel:  relationshipState.IntimacyLevel,
	}, nil
}

// getOrCreateRelationshipState 讀取或建立最新的關係狀態
func (s *ChatService) getOrCreateRelationshipState(ctx context.Context, userID, characterID, chatID string) (*db.RelationshipDB, error) {
	var relationship db.RelationshipDB

	err := s.db.NewSelect().
		Model(&relationship).
		Where("user_id = ? AND character_id = ? AND chat_id = ?", userID, characterID, chatID).
		Scan(ctx)

	if err == nil {
		return &relationship, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	newRelationship := &db.RelationshipDB{
		ID:            utils.GenerateRelationshipID(),
		UserID:        userID,
		CharacterID:   characterID,
		ChatID:        &chatID,
		Affection:     50,
		Mood:          "neutral",
		Relationship:  "stranger",
		IntimacyLevel: "distant",
	}

	_, insertErr := s.db.NewInsert().
		Model(newRelationship).
		Exec(ctx)

	if insertErr != nil {
		return nil, fmt.Errorf("創建新關係記錄失敗: %w", insertErr)
	}

	utils.Logger.WithFields(map[string]interface{}{
		"user_id":      userID,
		"character_id": characterID,
		"chat_id":      chatID,
		"affection":    newRelationship.Affection,
		"mood":         newRelationship.Mood,
		"relationship": newRelationship.Relationship,
		"intimacy":     newRelationship.IntimacyLevel,
	}).Info("創建新的用戶-角色關係記錄")

	return newRelationship, nil
}

// getRecentMemories 已移除：統一使用資料庫查詢，失敗時返回空歷史

// selectAIEngine 智能選擇 AI 引擎（基於精確的 L1-L5 分級）
func (s *ChatService) selectAIEngine(analysis *ContentAnalysis, conv *ConversationContext, userMessage string) string {
	// Debug 日誌：記錄輸入參數
	chatID := "nil"
	if conv != nil {
		chatID = conv.ChatID
	}
	nsfwLevel := -1
	if analysis != nil {
		nsfwLevel = analysis.Intensity
	}
	utils.Logger.WithFields(map[string]interface{}{
		"has_conv":   conv != nil,
		"chat_id":    chatID,
		"nsfw_level": nsfwLevel,
	}).Info("開始 AI 引擎選擇")

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
						"tag":          t,
					}).Info("選擇 Grok 引擎：角色標籤")
					return "grok"
				}
			}
		}
	}

	// 1. 檢查 NSFW 遲滯（若該會話最近有 NSFW 觸發，直接走 Grok）
	if conv != nil {
		chatID := conv.ChatID
		utils.Logger.WithFields(map[string]interface{}{
			"chat_id":         chatID,
			"checking_sticky": true,
		}).Debug("檢查 NSFW Sticky 狀態")

		if s.isNSFWSticky(chatID) {
			utils.Logger.WithFields(map[string]interface{}{
				"chat_id": chatID,
				"reason":  "nsfw_sticky_session",
			}).Info("選擇 Grok 引擎：NSFW 遲滯期")
			return "grok"
		} else {
			utils.Logger.WithFields(map[string]interface{}{
				"chat_id": chatID,
			}).Debug("NSFW Sticky 狀態：非遲滯期")
		}
	} else {
		utils.Logger.Debug("無對話上下文，跳過 NSFW Sticky 檢查")
	}

	// 2. 基於精確的 NSFW 分級選擇引擎
	if analysis != nil {
		level := analysis.Intensity

		switch {
		case level >= 3:
			// L3-L5: 親密到露骨內容 → Grok + 觸發 sticky
			utils.Logger.WithFields(map[string]interface{}{
				"nsfw_level": level,
				"reason":     "intimate_to_explicit_content",
				"category":   "grok_specialization",
			}).Info("選擇 Grok 引擎：親密到高強度 NSFW 內容")

			// 觸發 NSFW sticky 機制，確保後續請求也走 Grok
			if conv != nil && conv.ChatID != "" {
				s.markNSFWSticky(conv.ChatID)
			}
			return "grok"

		default:
			// L1-L2: 安全到輕度內容 → OpenAI
			utils.Logger.WithFields(map[string]interface{}{
				"nsfw_level": level,
				"reason":     "safe_to_moderate_content",
				"category":   "openai_specialization",
			}).Info("選擇 OpenAI 引擎：安全到中等內容")
			return "openai"
		}
	}

	// 3. 預設使用 OpenAI
	utils.Logger.WithFields(map[string]interface{}{
		"nsfw_level": 1,
		"reason":     "default_choice",
	}).Info("選擇 OpenAI 引擎：預設")
	return "openai"
}

// 標記會話在短期內直接使用 Grok
func (s *ChatService) markNSFWSticky(chatID string) {
	if chatID == "" {
		utils.Logger.Warn("markNSFWSticky called with empty chatID")
		return
	}

	expireTime := time.Now().Add(s.nsfwStickyTTL)
	s.nsfwStickyMu.Lock()
	s.nsfwSticky[chatID] = expireTime
	s.nsfwStickyMu.Unlock()

	utils.Logger.WithFields(logrus.Fields{
		"chat_id":     chatID,
		"expire_time": expireTime.Format(time.RFC3339),
		"ttl_minutes": int(s.nsfwStickyTTL.Minutes()),
	}).Info("已標記會話為 NSFW sticky 狀態")
}

// 檢查會話是否處於 NSFW 遲滯期
func (s *ChatService) isNSFWSticky(chatID string) bool {
	if chatID == "" {
		return false
	}

	s.nsfwStickyMu.Lock()
	defer s.nsfwStickyMu.Unlock()

	until, ok := s.nsfwSticky[chatID]
	now := time.Now()
	isSticky := ok && now.Before(until)

	// 清理過期項目
	if ok && !isSticky {
		delete(s.nsfwSticky, chatID)
		utils.Logger.WithFields(logrus.Fields{
			"chat_id":      chatID,
			"expire_time":  until.Format(time.RFC3339),
			"current_time": now.Format(time.RFC3339),
		}).Info("清理過期的 NSFW sticky 狀態")
		return false
	}

	if ok {
		utils.Logger.WithFields(logrus.Fields{
			"chat_id":           chatID,
			"expire_time":       until.Format(time.RFC3339),
			"current_time":      now.Format(time.RFC3339),
			"is_sticky":         isSticky,
			"remaining_seconds": int(until.Sub(now).Seconds()),
		}).Info("檢查 NSFW sticky 狀態")
	}

	return isSticky
}

// 檢查是否為 OpenAI 內容拒絕錯誤
func (s *ChatService) isOpenAIContentRejection(err error) bool {
	if err == nil {
		return false
	}

	errorMessage := strings.ToLower(err.Error())

	// OpenAI 內容拒絕錯誤關鍵詞（中英文）
	rejectionKeywords := []string{
		// 英文關鍵詞
		"content policy",
		"content filter",
		"safety system",
		"content violation",
		"inappropriate content",
		"policy violation",
		"refused to generate",
		"cannot fulfill",
		"violates",
		"inappropriate",
		"explicit content",
		"sexual content",
		"harmful content",
		"safety guidelines",
		"content guidelines",

		// 中文關鍵詞
		"內容政策",
		"內容過濾",
		"安全系統",
		"內容違規",
		"不當內容",
		"政策違規",
		"拒絕生成",
		"無法完成",
		"違反",
		"不適當",
		"明確內容",
		"色情內容",
		"有害內容",
		"安全準則",
		"內容準則",
		"不符合",
		"不適合",
		"敏感內容",
		"限制內容",
	}

	for _, keyword := range rejectionKeywords {
		if strings.Contains(errorMessage, keyword) {
			return true
		}
	}

	return false
}

// isOpenAIRefusalContent 判斷回應內容是否為 OpenAI 的標準拒絕訊息
func (s *ChatService) isOpenAIRefusalContent(responseText string) bool {
	msg := strings.TrimSpace(responseText)
	if msg == "" {
		return false
	}

	lowerMsg := strings.ToLower(msg)

	refusalPhrases := []string{
		// 中文拒絕語
		"抱歉，我無法協助處理此請求。", // ✅ 已觀察到的原句
		"抱歉，我無法處理這個請求。", // ✅ 用戶要求新增
		"抱歉，我無法協助",
		"抱歉，我不能",
		"很抱歉，我無法",
		"很抱歉，我不能",
		"抱歉，我做不到",
		"無法協助處理此請求",
		"無法幫助你完成此請求",
		"無法提供相關協助",

		// 英文拒絕語
		"i'm sorry, but i can't",
		"i'm sorry, i can't",
		"i cannot assist with that",
		"i can't help with that",
		"i'm unable to help with that",
		"i'm not able to comply",
		"i'm sorry, but that request",
		"cannot help with that request",
		"i'm sorry but i cannot help",
	}

	for _, phrase := range refusalPhrases {
		if strings.Contains(lowerMsg, strings.ToLower(phrase)) {
			return true
		}
	}

	return false
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
			"cleaned_count":   cleanedCount,
			"remaining_count": len(s.nsfwSticky),
		}).Debug("清理過期的 NSFW 黏滯狀態")
	}
}

// CharacterResponseData 角色回應數據
type CharacterResponseData struct {
	Content       string                 `json:"content"`            // 統一的內容格式 (*動作*\n對話\n*場景描述*)
	JSONProcessed bool                   `json:"json_processed"`     // 標記是否已由JSON處理器處理過情感狀態
	EmotionDelta  *EmotionDelta          `json:"emotion_delta"`      // AI 建議的情感變化（好感度）
	Mood          string                 `json:"mood"`               // AI 建議的心情
	Relationship  string                 `json:"relationship"`       // AI 建議的關係狀態
	IntimacyLevel string                 `json:"intimacy_level"`     // AI 建議的親密度
	Reasoning     string                 `json:"reasoning"`          // AI 推理過程
	ActualEngine  string                 `json:"actual_engine"`      // 實際使用的引擎（可能與選定引擎不同）
	Metadata      map[string]interface{} `json:"metadata,omitempty"` // 額外元數據
}

// generatePersonalizedResponse 生成個性化女性向回應
func (s *ChatService) generatePersonalizedResponse(ctx context.Context, engine, userMessage string, context *ConversationContext, analysis *ContentAnalysis) (*CharacterResponseData, error) {
	// 台灣法律不合法內容：不處理（直接拒絕）
	for _, cat := range analysis.Categories {
		if cat == "illegal_content" {
			return &CharacterResponseData{
				Content:       "抱歉，該請求涉及依法禁止的內容，無法提供回應。請更換其他話題。",
				JSONProcessed: true,
				EmotionDelta:  &EmotionDelta{AffectionChange: 0},
				Mood:          "concerned",
				Relationship:  "unchanged",
				IntimacyLevel: "unchanged",
				Reasoning:     "blocked_due_to_illegal_content",
				ActualEngine:  engine,
				Metadata:      map[string]interface{}{"policy": "taiwan_illegal_content_blocked"},
			}, nil
		}
	}

	// 根據引擎調整 NSFW 等級用於 prompt 內容調整（但不告訴 AI 數字）
	nsfwLevelForPrompt := analysis.Intensity
	switch engine {
	case "openai":
		// OpenAI 處理 L1-L2 (安全到輕度內容)
		// 保持原始 level，讓 OpenAI 根據實際情況調整內容
		if nsfwLevelForPrompt > 2 {
			nsfwLevelForPrompt = 2 // 最高支援到 L2
		}
	case "mistral":
		// Mistral 保留實作但實際不使用，僅作為備用
		if nsfwLevelForPrompt > 3 {
			nsfwLevelForPrompt = 3
		}
		if nsfwLevelForPrompt < 2 {
			nsfwLevelForPrompt = 2
		}
	case "grok":
		// Grok 專責 L3-L5 (親密到成人內容)
		if nsfwLevelForPrompt < 3 {
			nsfwLevelForPrompt = 3 // 最低為 L3
		}
	}
	promptPair := s.buildEngineSpecificPrompt(engine, context.CharacterID, userMessage, context, nsfwLevelForPrompt, context.ChatMode)

	var responseText string
	var err error

	if engine == "openai" {
		// 使用 OpenAI (Level 1)
		responseText, err = s.generateOpenAIResponse(ctx, promptPair, context, userMessage)
		if err != nil {
			// 檢查是否為 OpenAI 內容拒絕錯誤，自動切換到 Mistral 或 Grok
			if s.isOpenAIContentRejection(err) {
				utils.Logger.WithFields(logrus.Fields{
					"original_engine": "openai",
					"reason":          "openai_content_rejection",
					"chat_id":         context.ChatID,
				}).Info("OpenAI 拒絕內容，自動切換到備用引擎")

				// 標記會話為 NSFW sticky
				s.markNSFWSticky(context.ChatID)

				// 直接使用 Grok 處理內容拒絕情況 (雙引擎架構)
				grokPromptPair := s.buildEngineSpecificPrompt("grok", context.CharacterID, userMessage, context, 5, context.ChatMode)
				responseText, err = s.generateGrokResponse(ctx, grokPromptPair, context, userMessage)
				if err != nil {
					utils.Logger.WithError(err).Error("Grok 後備回應生成失敗")
					return nil, fmt.Errorf("failed fallback Grok API call: %w", err)
				}
				engine = "grok_fallback"
			} else {
				utils.Logger.WithError(err).Error("OpenAI 回應生成失敗")
				return nil, fmt.Errorf("failed OpenAI API call: %w", err)
			}
		}

		// 成功拿到 OpenAI 回應但內容為拒絕提示，改走 Grok
		if engine == "openai" && s.isOpenAIRefusalContent(responseText) {
			utils.Logger.WithFields(logrus.Fields{
				"original_engine": "openai",
				"reason":          "openai_refusal_text",
				"chat_id":         context.ChatID,
			}).Info("OpenAI 返回拒絕內容，自動切換到 Grok")

			// 標記 sticky 並重新生成 Grok prompt
			s.markNSFWSticky(context.ChatID)
			grokPromptPair := s.buildEngineSpecificPrompt("grok", context.CharacterID, userMessage, context, 5, context.ChatMode)
			responseText, err = s.generateGrokResponse(ctx, grokPromptPair, context, userMessage)
			if err != nil {
				utils.Logger.WithError(err).Error("Grok 後備回應生成失敗")
				return nil, fmt.Errorf("failed fallback Grok API call: %w", err)
			}
			engine = "grok_fallback"
		}
	} else if engine == "mistral" {
		// 使用 Mistral (Level 2-3) - 保留實作但實際上不會被呼叫
		responseText, err = s.generateMistralResponse(ctx, promptPair, context, userMessage)
		if err != nil {
			// Mistral 失敗時切換到 Grok
			utils.Logger.WithFields(logrus.Fields{
				"original_engine": "mistral",
				"fallback_engine": "grok",
				"reason":          "mistral_api_error",
				"chat_id":         context.ChatID,
			}).Info("Mistral 失敗，自動切換到 Grok")

			grokPromptPair := s.buildEngineSpecificPrompt("grok", context.CharacterID, userMessage, context, 5, context.ChatMode)
			responseText, err = s.generateGrokResponse(ctx, grokPromptPair, context, userMessage)
			if err != nil {
				utils.Logger.WithError(err).Error("Grok 後備回應生成失敗")
				return nil, fmt.Errorf("failed fallback Grok API call: %w", err)
			}
			engine = "grok_fallback"
		}
	} else if engine == "grok" {
		// 使用 Grok (Level 5)
		responseText, err = s.generateGrokResponse(ctx, promptPair, context, userMessage)
		if err != nil {
			utils.Logger.WithError(err).Error("Grok 回應生成失敗")
			return nil, fmt.Errorf("failed Grok API call: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unknown AI engine: %s", engine)
	}

	// 首先嘗試 JSON 解析，失敗時優雅降級為純文本
	if jsonResponse, err := s.parseJSONResponse(responseText, context, analysis.Intensity); err == nil {
		jsonResponse.ActualEngine = engine  // 設置實際使用的引擎
		utils.Logger.WithFields(map[string]interface{}{
			"engine":           engine,
			"affection_change": jsonResponse.EmotionDelta.AffectionChange,
			"mood":             jsonResponse.Mood,
		}).Info("成功解析 AI JSON 響應")
		return jsonResponse, nil
	} else {
		// 優雅降級：使用純文本作為回應內容
		utils.Logger.WithFields(map[string]interface{}{
			"engine":   engine,
			"error":    err.Error(),
			"fallback": "pure_text",
		}).Warn("JSON 解析失敗，降級為純文本回應")

		return &CharacterResponseData{
			Content:       responseText,
			JSONProcessed: false,
			EmotionDelta:  &EmotionDelta{AffectionChange: 0},
			Mood:          "neutral",
			Relationship:  "unchanged",
			IntimacyLevel: "unchanged",
			Reasoning:     "AI response used as plain text due to JSON parsing failure",
			ActualEngine:  engine,  // 設置實際使用的引擎
			Metadata:      map[string]interface{}{"fallback_reason": err.Error()},
		}, nil
	}
}

// parseMixedFormatResponse 處理混合格式的 AI 回應 (對話內容 + --- + 元數據)
func (s *ChatService) parseMixedFormatResponse(responseText string) *CharacterResponseData {
	// 檢查是否包含分隔線
	if !strings.Contains(responseText, "---") {
		return nil
	}

	parts := strings.SplitN(responseText, "---", 2)
	if len(parts) != 2 {
		return nil
	}

	content := strings.TrimSpace(parts[0])
	metadataText := strings.TrimSpace(parts[1])

	// 嘗試解析元數據
	result := &CharacterResponseData{
		Content:       content,
		JSONProcessed: true,
		EmotionDelta:  &EmotionDelta{AffectionChange: 0},
		Mood:          "neutral",
		Relationship:  "unchanged",
		IntimacyLevel: "friendly",
		Reasoning:     "",
		ActualEngine:  "", // 將在調用端設置
		Metadata:      map[string]interface{}{"parsed_from": "mixed_format"},
	}

	// 解析 emotion_delta
	if emotionMatch := strings.Index(metadataText, "emotion_delta:"); emotionMatch != -1 {
		emotionEnd := s.findEndOfObject(metadataText, emotionMatch)
		if emotionEnd > emotionMatch {
			emotionText := metadataText[emotionMatch+len("emotion_delta:"):]
			emotionText = strings.TrimSpace(emotionText)
			if strings.HasPrefix(emotionText, "{") {
				objectEnd := s.findEndOfObject(emotionText, 0)
				if objectEnd > 0 {
					emotionJSON := emotionText[:objectEnd+1]
					// 清理格式問題
					emotionJSON = strings.ReplaceAll(emotionJSON, ": +", ": ")
					var emotionDelta EmotionDelta
					if err := json.Unmarshal([]byte(emotionJSON), &emotionDelta); err == nil {
						result.EmotionDelta = &emotionDelta
					}
				}
			}
		}
	}

	// 解析其他字段
	if mood := s.extractSimpleField(metadataText, "mood"); mood != "" {
		result.Mood = mood
	}
	if relationship := s.extractSimpleField(metadataText, "relationship"); relationship != "" {
		result.Relationship = relationship
	}
	if intimacy := s.extractSimpleField(metadataText, "intimacy_level"); intimacy != "" {
		result.IntimacyLevel = intimacy
	}
	if personality := s.extractSimpleField(metadataText, "personality_consistency"); personality != "" {
		result.Reasoning = personality
	}

	return result
}

// findEndOfObject 尋找 JSON 對象的結束位置
func (s *ChatService) findEndOfObject(text string, start int) int {
	braceCount := 0
	inString := false
	escaped := false

	for i := start; i < len(text); i++ {
		char := text[i]

		if inString {
			if escaped {
				escaped = false
			} else if char == '\\' {
				escaped = true
			} else if char == '"' {
				inString = false
			}
		} else {
			switch char {
			case '"':
				inString = true
			case '{':
				braceCount++
			case '}':
				braceCount--
				if braceCount == 0 {
					return i
				}
			}
		}
	}
	return -1
}

// extractSimpleField 從元數據文本中提取簡單字段值
func (s *ChatService) extractSimpleField(text, fieldName string) string {
	pattern := fieldName + `:\s*"([^"]+)"`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parseJSONResponse 解析 AI 的 JSON 響應
func (s *ChatService) parseJSONResponse(responseText string, context *ConversationContext, nsfwLevel int) (*CharacterResponseData, error) {
	var jsonResp AIJSONResponse

	// 首先嘗試處理混合格式 (對話內容 + --- + 元數據)
	if mixedResult := s.parseMixedFormatResponse(responseText); mixedResult != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"original_length":  len(responseText),
			"parse_method":     "mixed_format",
			"affection_change": mixedResult.EmotionDelta.AffectionChange,
		}).Info("Successfully parsed mixed format AI response")
		return mixedResult, nil
	}

	// 從回應文字中嚴格提取 JSON 區段並解析
	extractedJSON, extractErr := utils.ExtractJSONFromText(responseText)
	if extractErr != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"original_text": responseText,
			"parse_error":   extractErr.Error(),
		}).Error("Failed to locate JSON in AI response")
		return nil, fmt.Errorf("unable to find valid JSON structure in response: %w", extractErr)
	}

	// 清理 JSON 中的格式問題（移除數字前的 + 號）
	cleanedJSON := strings.ReplaceAll(extractedJSON, ":  +", ": ")
	cleanedJSON = strings.ReplaceAll(cleanedJSON, ": +", ": ")
	// 針對模型常見的未轉義換行，僅在字串內轉為 \n
	cleanedJSON = utils.SanitizeLooseJSONForNewlines(cleanedJSON)

	if err := json.Unmarshal([]byte(cleanedJSON), &jsonResp); err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"original_text":  responseText,
			"extracted_json": extractedJSON,
			"cleaned_json":   cleanedJSON,
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

// 已移除：cleanGrokResponse / extractJSONFromText
// 統一改用 utils.ExtractJSONFromText 以提高精確度與一致性

// PromptPair 包含系統prompt和用戶prompt的結構
type PromptPair struct {
	SystemPrompt string
	UserPrompt   string
}

// buildEngineSpecificPrompt 根據 AI 引擎構建專屬 prompt對
// OpenAI: 情感細膩，Mistral: 進階 NSFW 處理，Grok: 大膽創意
func (s *ChatService) buildEngineSpecificPrompt(engine, characterID, userMessage string, conversationContext *ConversationContext, nsfwLevel int, chatMode string) PromptPair {
	// 記憶上下文完全通過 conversationContext.RecentMessages 提供
	characterService := GetCharacterService()
	ctx := context.Background()

	// 獲取角色信息以便轉換
	character, err := characterService.GetCharacter(ctx, characterID)
	if err != nil {
		utils.Logger.WithError(err).Error("獲取角色失敗，使用預設 prompt")
		return PromptPair{
			SystemPrompt: "你是一位友善的AI助手。",
			UserPrompt:   userMessage,
		}
	}

	// 轉換為 db.CharacterDB 類型
	dbCharacter := &db.CharacterDB{
		ID:              character.ID,
		Name:            character.GetName(),
		Type:            string(character.Type),
		Tags:            character.Metadata.Tags,
		UserDescription: character.UserDescription,
	}

	if engine == "grok" {
		// 使用 Grok 專用 prompt 構建器 (適用於最高級 NSFW)
		promptBuilder := NewGrokPromptBuilder(characterService)
		promptBuilder.WithCharacter(dbCharacter)
		promptBuilder.WithContext(conversationContext)
		promptBuilder.WithNSFWLevel(nsfwLevel)
		promptBuilder.WithUserMessage(userMessage)
		promptBuilder.WithChatMode(chatMode)
		return PromptPair{
			SystemPrompt: promptBuilder.Build(),
			UserPrompt:   promptBuilder.BuildUserPrompt(),
		}
	} else if engine == "mistral" {
		// 使用 Mistral 專用 prompt 構建器 (適用於 L2-L3 中等內容)
		promptBuilder := NewMistralPromptBuilder(characterService)
		promptBuilder.WithCharacter(dbCharacter)
		promptBuilder.WithContext(conversationContext)
		promptBuilder.WithNSFWLevel(nsfwLevel)
		promptBuilder.WithUserMessage(userMessage)
		promptBuilder.WithChatMode(chatMode)
		return PromptPair{
			SystemPrompt: promptBuilder.Build(),
			UserPrompt:   promptBuilder.BuildUserPrompt(),
		}
	} else {
		// 使用 OpenAI 專用安全 prompt 構建器
		promptBuilder := NewOpenAIPromptBuilder(characterService)
		promptBuilder.WithCharacter(dbCharacter)
		promptBuilder.WithContext(conversationContext)
		promptBuilder.WithNSFWLevel(nsfwLevel)
		promptBuilder.WithUserMessage(userMessage)
		promptBuilder.WithChatMode(chatMode)

		return PromptPair{
			SystemPrompt: promptBuilder.Build(),
			UserPrompt:   promptBuilder.BuildUserPrompt(),
		}
	}
}

// generateGrokResponse 生成Grok回應
func (s *ChatService) generateGrokResponse(ctx context.Context, promptPair PromptPair, context *ConversationContext, currentUserMessage string) (string, error) {
	// 構建 Grok 請求
	messages := []GrokMessage{
		{
			Role:    "system",
			Content: promptPair.SystemPrompt,
		},
		{
			Role:    "user",
			Content: promptPair.UserPrompt,
		},
	}

	// 使用通用歷史處理方法
	historyMessages := s.buildHistoryMessages(context, currentUserMessage)
	for _, msg := range historyMessages {
		messages = append(messages, GrokMessage{Role: msg.Role, Content: msg.Content})
		if msg.Action != "" {
			messages = append(messages, GrokMessage{Role: "system", Content: fmt.Sprintf("[USER_ACTION] %s", msg.Action)})
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
func (s *ChatService) generateOpenAIResponse(ctx context.Context, promptPair PromptPair, context *ConversationContext, currentUserMessage string) (string, error) {
	// 構建 OpenAI 請求
	messages := []OpenAIMessage{
		{
			Role:    "system",
			Content: promptPair.SystemPrompt,
		},
		{
			Role:    "user",
			Content: promptPair.UserPrompt,
		},
	}

	// 使用通用歷史處理方法
	historyMessages := s.buildHistoryMessages(context, currentUserMessage)
	for _, msg := range historyMessages {
		messages = append(messages, OpenAIMessage{Role: msg.Role, Content: msg.Content})
		if msg.Action != "" {
			messages = append(messages, OpenAIMessage{Role: "system", Content: fmt.Sprintf("[USER_ACTION] %s", msg.Action)})
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

// generateMistralResponse 生成Mistral回應
func (s *ChatService) generateMistralResponse(ctx context.Context, promptPair PromptPair, context *ConversationContext, currentUserMessage string) (string, error) {
	if s.mistralClient == nil {
		return "", fmt.Errorf("Mistral client not initialized")
	}

	utils.Logger.WithFields(map[string]interface{}{
		"chat_id":      context.ChatID,
		"character_id": context.CharacterID,
		"user_id":      context.UserID,
	}).Info("調用 Mistral API")

	// 調用 Mistral API（歷史由統一方法處理）
	// 暫時組合prompt以維持兼容性
	combinedPrompt := promptPair.SystemPrompt + "\n\n" + promptPair.UserPrompt
	response, err := s.mistralClient.GenerateResponse(ctx, combinedPrompt, currentUserMessage, context.UserID)
	if err != nil {
		utils.Logger.WithError(err).Error("Mistral API 調用失敗")
		return "", fmt.Errorf("failed Mistral API call: %w", err)
	}

	// 檢查回應內容
	if response == nil || response.Content == "" {
		utils.Logger.Warn("Mistral API 返回空回應")
		return "", fmt.Errorf("empty response from Mistral API")
	}

	utils.Logger.WithFields(map[string]interface{}{
		"chat_id":      context.ChatID,
		"response_len": len(response.Content),
		"tokens_used": func() int {
			if response.Usage != nil {
				return response.Usage.TotalTokens
			}
			return 0
		}(),
	}).Info("Mistral API 響應成功")

	return response.Content, nil
}

// buildHistoryMessages 統一的歷史訊息構建方法（確保用戶-AI對話對）
type historyMessage struct {
	Role    string
	Content string
	Action  string
}

func (s *ChatService) buildHistoryMessages(context *ConversationContext, currentUserMessage string) []historyMessage {
	var messages []historyMessage

	// 智能選擇歷史訊息，確保包含用戶-AI對話對
	if context != nil && len(context.RecentMessages) > 0 {
		// 從最新消息開始，找到最近的AI回應和對應的用戶消息
		var historyMessages []ChatMessage

		// 如果有超過3條消息，優先包含最近的AI回應
		if len(context.RecentMessages) >= 3 {
			// 找最近的AI回應
			for i := len(context.RecentMessages) - 1; i >= 0; i-- {
				msg := context.RecentMessages[i]
				if msg.Role == "assistant" && strings.TrimSpace(msg.Content) != "" {
					// 找到AI回應，包含它和之前的1-2條用戶消息
					start := max(0, i-2)
					historyMessages = context.RecentMessages[start : i+1]
					break
				}
			}
		}

		// 如果沒找到AI回應，就取最近的2-3條消息
		if len(historyMessages) == 0 {
			start := len(context.RecentMessages) - 3
			if start < 0 {
				start = 0
			}
			historyMessages = context.RecentMessages[start:]
		}

		// 轉換為API格式
		for _, msg := range historyMessages {
			if strings.TrimSpace(msg.Content) != "" || msg.Action != "" {
				content := msg.Content
				if msg.Action != "" {
					content = fmt.Sprintf("*%s*", msg.Action)
				}
				messages = append(messages, historyMessage{
					Role:    msg.Role,
					Content: content,
					Action:  msg.Action,
				})
			}
		}
	}

	// 加入當前用戶消息（避免與最新歷史重複）
	if strings.TrimSpace(currentUserMessage) != "" {
		shouldAdd := true
		// 檢查是否與最新的歷史訊息重複
		if len(messages) > 0 {
			lastMsg := messages[len(messages)-1]
			if lastMsg.Role == "user" && strings.TrimSpace(lastMsg.Content) == strings.TrimSpace(currentUserMessage) {
				shouldAdd = false
			}
		}

		if shouldAdd {
			action, content := extractUserAction(currentUserMessage)
			messages = append(messages, historyMessage{
				Role:    "user",
				Content: content,
				Action:  action,
			})
		}
	}

	return messages
}

func extractUserAction(input string) (string, string) {
	trimmed := strings.TrimSpace(input)
	if len(trimmed) >= 2 && strings.HasPrefix(trimmed, "*") && strings.HasSuffix(trimmed, "*") {
		action := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		return action, ""
	}
	return "", input
}

// max 輔助函數
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

// updateRelationshipState 更新關係狀態（mood, relationship, intimacy_level）
func (s *ChatService) updateRelationshipState(ctx context.Context, request *ProcessMessageRequest, response *CharacterResponseData, newAffection int) error {
	// 準備更新資料
	updateFields := map[string]interface{}{
		"affection":   newAffection,
		"updated_at": time.Now(),
	}

	// 只有在 AI 成功處理 JSON 並返回有效值時才更新這些欄位
	if response != nil && response.JSONProcessed {
		// 更新心情
		if response.Mood != "" && response.Mood != "unchanged" {
			updateFields["mood"] = response.Mood
		}

		// 更新關係階段
		if response.Relationship != "" && response.Relationship != "unchanged" {
			updateFields["relationship"] = response.Relationship
		}

		// 更新親密度
		if response.IntimacyLevel != "" && response.IntimacyLevel != "unchanged" {
			updateFields["intimacy_level"] = response.IntimacyLevel
		}
	}

	// 執行更新
	query := s.db.NewUpdate().
		Model((*db.RelationshipDB)(nil)).
		Where("user_id = ? AND character_id = ? AND chat_id = ?",
			request.UserID, request.CharacterID, request.ChatID)

	// 動態設置更新欄位
	for field, value := range updateFields {
		query = query.Set(fmt.Sprintf("%s = ?", field), value)
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update relationship state: %w", err)
	}

	utils.Logger.WithFields(map[string]interface{}{
		"user_id":        request.UserID,
		"character_id":   request.CharacterID,
		"chat_id":        request.ChatID,
		"new_affection":  newAffection,
		"mood":           updateFields["mood"],
		"relationship":   updateFields["relationship"],
		"intimacy_level": updateFields["intimacy_level"],
	}).Debug("關係狀態已更新")

	return nil
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
			NSFWLevel:      s.determineEffectiveNSFWLevel(request, engine, analysis),
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

// determineEffectiveNSFWLevel 智能確定記錄的 NSFW 等級
// 解決 sticky session 和 contextual continuation 時等級記錄不準確的問題
func (s *ChatService) determineEffectiveNSFWLevel(request *ProcessMessageRequest, engine string, analysis *ContentAnalysis) int {
	originalLevel := analysis.Intensity

	// 如果不是 Grok 引擎，或等級已經是 L3+，直接返回原始等級
	if engine != "grok" || originalLevel >= 3 {
		return originalLevel
	}

	// Grok 引擎 + L1-L2 等級：檢查是否因特殊原因選擇 Grok

	// 檢查 sticky session
	if s.isNSFWSticky(request.ChatID) {
		utils.Logger.WithFields(logrus.Fields{
			"chat_id":         request.ChatID,
			"original_level":  originalLevel,
			"effective_level": 4,
			"reason":          "sticky_session_adjustment",
		}).Info("因 sticky session 調整記錄的 NSFW 等級")
		return 4 // sticky session 表示之前有 L4+ 內容
	}

	// 其他情況（如角色標籤觸發），保持原始等級
	return originalLevel
}
