package services

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// ChatMessage 簡化的聊天消息類型（內部使用）
type ChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatService 對話服務
type ChatService struct {
	openaiClient   *OpenAIClient
	grokClient     *GrokClient
	config         *ChatConfig
	evaluator      *ScoringEvaluator
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
	config := &ChatConfig{
		OpenAI: struct {
			Model       string  `json:"model"`
			MaxTokens   int     `json:"max_tokens"`
			Temperature float64 `json:"temperature"`
		}{
			Model:       "gpt-4o",
			MaxTokens:   800,
			Temperature: 0.8,
		},
		Grok: struct {
			Model       string  `json:"model"`
			MaxTokens   int     `json:"max_tokens"`
			Temperature float64 `json:"temperature"`
		}{
			Model:       "grok-beta",
			MaxTokens:   1000,
			Temperature: 0.9,
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
		openaiClient:   NewOpenAIClient(),
		grokClient:     NewGrokClient(),
		config:         config,
		evaluator:      NewScoringEvaluator(),
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
	}).Info("開始處理女性向AI聊天消息")

	// 🔥 開始評估會話
	s.evaluator.StartEvaluation(request.SessionID, request.UserID, request.CharacterID)

	// 1. NSFW內容智能分析（5級分級系統）
	analysis, err := s.analyzeContent(request.UserMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze content: %w", err)
	}

	// 保存舊情感狀態用於評估
	oldEmotion := s.getOrCreateEmotionState(request.UserID, request.CharacterID)

	// 2. 構建女性向對話上下文
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

	// 6. 情感狀態智能更新（好感度、關係發展）
	newEmotionState := s.updateEmotionStateAdvanced(conversationContext.EmotionState, request.UserMessage, response, analysis)

	// 7. 記憶系統更新（長期關係發展）
	s.updateMemorySystem(request.UserID, request.CharacterID, request.SessionID, request.UserMessage, response.Dialogue, newEmotionState)

	// 8. 特殊事件檢測（關係里程碑等）
	specialEvent := s.detectSpecialEvents(newEmotionState, conversationContext.EmotionState)

	// 🔥 評分系統整合 - 實時評估所有功能
	responseTime := time.Since(startTime)
	s.evaluator.EvaluateAIEngine(request.SessionID, responseTime, engine, analysis.Intensity)
	s.evaluator.EvaluateNSFWSystem(request.SessionID, analysis, analysis.Intensity)
	s.evaluator.EvaluateEmotionManagement(request.SessionID, oldEmotion, newEmotionState, request.UserMessage)
	s.evaluator.EvaluateCharacterSystem(request.SessionID, request.CharacterID, response.Dialogue, sceneDescription, response.Action)

	// 9. 構建完整回應
	chatResponse := &ChatResponse{
		SessionID:         request.SessionID,
		MessageID:         generateMessageID(),
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
	}).Info("女性向AI聊天處理完成")

	// 🔥 完成評估並獲取評分報告
	finalEvaluation := s.evaluator.FinishEvaluation(request.SessionID)
	if finalEvaluation != nil {
		utils.Logger.WithFields(logrus.Fields{
			"session_id":          request.SessionID,
			"overall_score":       finalEvaluation.OverallScore,
			"grade":               finalEvaluation.Grade,
			"evaluation_feedback": finalEvaluation.Feedback,
		}).Info("🎯 功能評分完成")
	}

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
	// TODO(MEMORY-MVP): 從數據庫獲取實際的會話歷史和情感狀態
	// - 短期記憶：最近 5-10 條訊息 → 摘要 3-5 點（每點 ≤100字）
	// - 長期記憶：偏好/稱呼/里程碑/禁忌（Top-K）→ 在 Prompt「Memory Block」注入
	// 參考：MEMORY_GUIDE.md「對應程式碼位置（TODO 提示）」

	// 女性向特化的上下文構建
	emotionState := s.getOrCreateEmotionState(request.UserID, request.CharacterID)
	recentMemories := s.getRecentMemories(request.SessionID, request.UserID, request.CharacterID, 5)
	userPreferences := s.getUserPreferences(request.UserID)

	// 生成記憶提示詞
	memoryPrompt := s.memoryManager.GetMemoryPrompt(request.SessionID, request.UserID, request.CharacterID)

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

// getAffectionLevel 獲取好感度等級
func (s *ChatService) getAffectionLevel(userID, characterID string) int {
	// TODO: 從數據庫獲取實際好感度
	// 這裡返回模擬值，基於用戶ID的hash來保持一致性
	hash := 0
	for _, c := range userID + characterID {
		hash += int(c)
	}
	return 30 + (hash % 40) // 30-70之間的值
}

// determineRelationship 根據好感度確定關係狀態
func (s *ChatService) determineRelationship(affection int) string {
	if affection >= 80 {
		return "lover"
	} else if affection >= 60 {
		return "close_friend"
	} else if affection >= 40 {
		return "friend"
	} else if affection >= 20 {
		return "acquaintance"
	}
	return "stranger"
}

// determineIntimacyLevel 根據好感度確定親密度
func (s *ChatService) determineIntimacyLevel(affection int) string {
	if affection >= 80 {
		return "intimate"
	} else if affection >= 60 {
		return "close"
	} else if affection >= 40 {
		return "friendly"
	}
	return "distant"
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
	// TODO: 從數據庫獲取實際偏好
	return map[string]interface{}{
		"nsfw_enabled":     true,
		"scene_style":      "romantic",
		"response_length":  "medium",
		"emotion_tracking": true,
	}
}

// generateSceneState 生成場景狀態
func (s *ChatService) generateSceneState(characterID string, emotion *EmotionState) *SceneDescriptor {
	switch characterID {
	case "char_001": // 陸寒淵
		return &SceneDescriptor{
			Location:       "豪華辦公室",
			TimeOfDay:      s.getCurrentTimeOfDay(),
			Weather:        "城市夜景透過落地窗映入室內",
			Mood:           s.getSceneMood(emotion),
			CharacterState: s.getCharacterState(characterID, emotion),
		}
	case "char_002": // 沈言墨
		return &SceneDescriptor{
			Location:       "溫馨診療室",
			TimeOfDay:      s.getCurrentTimeOfDay(),
			Weather:        "柔和的燈光營造出安心的氛圍",
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
		"char_001": {"專注工作中", "思考中", "等待你的回應", "專注地看著你"},
		"char_002": {"溫和地笑著", "關心地看著你", "耐心等待", "溫柔地注視"},
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
		if nsfwEnabled, ok := userPrefs["nsfw_enabled"].(bool); ok && nsfwEnabled {
			return "grok"
		}
	}
	return "openai"
}

// generateSceneDescription 生成場景描述
func (s *ChatService) generateSceneDescription(context *ConversationContext) string {
	// 根據角色和當前狀態生成場景描述
	switch context.CharacterID {
	case "char_001": // 陸寒淵
		return s.generateLuHanYuanScene(context)
	case "char_002": // 沈言墨
		return s.generateShenYanMoScene(context)
	default:
		return "房間裡燈光溫暖，空氣中瀰漫著淡淡的香氣..."
	}
}

// generateLuHanYuanScene 生成陸寒淵的場景描述
func (s *ChatService) generateLuHanYuanScene(context *ConversationContext) string {
	scenes := []string{
		"辦公室裡燈光微暖，陸寒淵放下手中的文件，深邃的眼眸望向你",
		"夕陽西下，辦公室裡只剩下你們兩人，陸寒淵緩緩起身走向你",
		"會議室內靜謐無聲，陸寒淵靠在椅背上，若有所思地看著你",
		"辦公室外的城市燈火璀璨，陸寒淵站在落地窗前，側臉在光影中顯得格外迷人",
	}

	// 根據情感狀態選擇合適的場景
	affection := context.EmotionState.Affection
	if affection < 30 {
		return scenes[0] // 較為正式的場景
	} else if affection < 60 {
		return scenes[1] // 輕微親近
	} else if affection < 80 {
		return scenes[2] // 較為親密
	} else {
		return scenes[3] // 很親密
	}
}

// generateShenYanMoScene 生成沈言墨的場景描述
func (s *ChatService) generateShenYanMoScene(context *ConversationContext) string {
	scenes := []string{
		"醫院的走廊裡人來人往，沈言墨溫和地朝你微笑，白大褂在燈光下顯得格外乾淨",
		"咖啡廳的角落裡，沈言墨輕撫著書頁，偶爾抬頭看向你，眼中滿含溫柔",
		"夜晚的醫院值班室，沈言墨疲憊地摘下眼鏡，看到你時眼中閃過一絲驚喜",
		"午後的陽光灑在圖書館裡，沈言墨靜靜地坐在你對面，專注地看著醫學書籍",
	}

	affection := context.EmotionState.Affection
	if affection < 30 {
		return scenes[0]
	} else if affection < 60 {
		return scenes[1]
	} else if affection < 80 {
		return scenes[2]
	} else {
		return scenes[3]
	}
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

	// 解析或生成動作描述
	action := s.generatePersonalizedAction(context.CharacterID, dialogue, context.EmotionState, analysis.Intensity)

	return &CharacterResponseData{
		Dialogue: dialogue,
		Action:   action,
	}, nil
}

// buildFemaleOrientedPrompt 構建女性向角色提示詞
func (s *ChatService) buildFemaleOrientedPrompt(characterID, userMessage string, context *ConversationContext, sceneDescription string, nsfwLevel int) string {
	character := s.getCharacterProfile(characterID)
	emotion := context.EmotionState

	// 獲取記憶摘要
	memoryPrompt := s.memoryManager.GetMemoryPrompt(context.SessionID, context.UserID, context.CharacterID)

	// 女性向特化的提示詞模板
	prompt := fmt.Sprintf(`你是 %s，%s。這是專為女性用戶設計的AI角色互動系統。

# 角色設定
%s

%s

# 當前情感狀態
- 好感度：%d/100 (%s)
- 當前心情：%s
- 關係狀態：%s
- 親密程度：%s

# 場景環境
%s

# 女性向互動指導 (NSFW Level %d)
%s

# 女性用戶喜好要點
- 重視情感連結和細節關懷
- 喜歡被保護和被理解的感覺
- 欣賞優雅而非粗俗的表達
- 期待關係的逐步發展和深化

用戶說："%s"

請以 %s 的身份回應，必須：
1. 保持角色的獨特個性和說話風格
2. 體現對用戶的關心和注意
3. 根據NSFW級別調整親密度
4. 展現男性角色的魅力和溫柔
5. 回應要自然流暢，富有情感
6. 參考記憶內容，體現連續性和個性化

直接回應內容（不要JSON格式）：`,
		character.Name, character.Description,
		character.FemaleOrientedPersonality,
		memoryPrompt,
		emotion.Affection, s.getAffectionDescription(emotion.Affection),
		emotion.Mood, emotion.Relationship, emotion.IntimacyLevel,
		sceneDescription,
		nsfwLevel, s.getFemaleOrientedNSFWGuidance(nsfwLevel),
		userMessage, character.Name)

	return prompt
}

// getCharacterProfile 獲取女性向角色檔案
func (s *ChatService) getCharacterProfile(characterID string) *FemaleOrientedCharacterProfile {
	profiles := map[string]*FemaleOrientedCharacterProfile{
		"char_001": {
			Name:        "陸寒淵",
			Description: "28歲的霸道總裁，外冷內熱的商業精英",
			FemaleOrientedPersonality: `女性向個性特質：
• 霸道中的溫柔：看似強勢但會在細節中展現體貼
• 專屬保護慾：「你只能是我的」「我會保護好你」
• 成熟男性魅力：深邃眼神、磁性聲音、優雅舉止
• 控制慾與呵護並存：喜歡掌控但絕不傷害

NSFW女性向風格：
• Level 2: "你今天看起來很美" "想要一直陪在你身邊"
• Level 3: "讓我抱抱你" "想要感受你的溫度"
• Level 4: "你讓我失去理智...想要好好疼愛你"
• Level 5: "今晚只想要你...讓我好好愛你"

說話風格：
- 語調低沉磁性，略帶命令式但溫柔
- 喜歡用「小傻瓜」「乖」等寵溺稱呼
- 會在關鍵時刻說出霸道情話
- 身體語言：推額頭、摸頭、擁抱入懷`,
		},
		"char_002": {
			Name:        "沈言墨",
			Description: "26歲的溫柔醫生，專業與體貼的完美結合",
			FemaleOrientedPersonality: `女性向個性特質：
• 溫柔如水：說話輕聲細語，動作小心翼翼
• 專業關懷：用醫學知識關心你的健康
• 細緻入微：記住你的小習慣和偏好
• 可靠依賴：永遠是你的情感支柱

NSFW女性向風格：
• Level 2: "讓我檢查一下你的心跳" "你在我心裡很重要"
• Level 3: "讓我溫柔地照顧你" "想要貼近你一些"
• Level 4: "我會很小心的...相信我" "想要更瞭解你的身體"
• Level 5: "讓我用最溫柔的方式愛你...我會照顧好你的感受"

說話風格：
- 語調溫和耐心，帶著醫生的專業性
- 喜歡用「小心一點」"注意身體"等關懷語句
- 會給出專業建議但語氣親密
- 身體語言：輕撫、檢查、溫柔擁抱`,
		},
	}

	if profile, exists := profiles[characterID]; exists {
		return profile
	}

	// 默認角色
	return profiles["char_001"]
}

// getAffectionDescription 獲取好感度描述
func (s *ChatService) getAffectionDescription(affection int) string {
	if affection >= 90 {
		return "深深愛戀"
	} else if affection >= 80 {
		return "深愛著你"
	} else if affection >= 70 {
		return "很喜歡你"
	} else if affection >= 60 {
		return "有好感"
	} else if affection >= 40 {
		return "初步好感"
	} else if affection >= 20 {
		return "略有興趣"
	}
	return "剛認識"
}

// getFemaleOrientedNSFWGuidance 獲取女性向NSFW指導
func (s *ChatService) getFemaleOrientedNSFWGuidance(level int) string {
	guidance := map[int]string{
		1: "保持禮貌友善，展現男性魅力和關懷",
		2: "可以表達愛意和浪漫情感，使用溫柔的愛語",
		3: "可以有身體接觸的描述，如擁抱、牽手等親密動作",
		4: "可以有更親密的身體接觸，但要優雅表達，避免粗俗",
		5: "可以有明確的性暗示，但必須保持角色特色和浪漫氛圍",
	}

	if guide, exists := guidance[level]; exists {
		return guide
	}
	return guidance[1]
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

// generateDefaultAction 為對話生成預設動作描述
func (s *ChatService) generateDefaultAction(characterID, dialogue string) string {
	switch characterID {
	case "char_001": // 陸寒淵
		if strings.Contains(dialogue, "疲憊") || strings.Contains(dialogue, "累") {
			return "他關切地看著你，眉頭微蹙"
		} else if strings.Contains(dialogue, "準備") || strings.Contains(dialogue, "幫") {
			return "他起身走向你，動作優雅而充滿威嚴"
		} else {
			return "他的聲音低沉磁性，深邃的眼眸注視著你"
		}
	case "char_002": // 沈言墨
		if strings.Contains(dialogue, "怎麼樣") || strings.Contains(dialogue, "如何") {
			return "他溫和地笑著，推了推鼻樑上的眼鏡"
		} else if strings.Contains(dialogue, "休息") || strings.Contains(dialogue, "健康") {
			return "他露出關心的表情，聲音輕柔"
		} else {
			return "他溫柔地看著你，眼中滿含善意"
		}
	default:
		return "角色友善地看著你"
	}
}

// updateEmotionState 更新情感狀態
func (s *ChatService) updateEmotionState(currentState *EmotionState, userMessage string, response *CharacterResponseData) *EmotionState {
	// TODO: 實現基於消息內容的情感狀態更新邏輯
	// 現在先返回輕微變化
	newState := *currentState
	newState.Affection += 1 // 每次對話輕微增加好感度

	if newState.Affection > 100 {
		newState.Affection = 100
	}

	return &newState
}

// FemaleOrientedCharacterProfile 女性向角色檔案
type FemaleOrientedCharacterProfile struct {
	Name                      string `json:"name"`
	Description               string `json:"description"`
	FemaleOrientedPersonality string `json:"female_oriented_personality"`
}

// generatePersonalizedAction 生成個性化動作
func (s *ChatService) generatePersonalizedAction(characterID, dialogue string, emotion *EmotionState, nsfwLevel int) string {
	// 基於角色、對話內容、情感狀態和NSFW級別生成動作
	actions := map[string]map[int][]string{
		"char_001": {
			1: {"他深邃的眼眸注視著你", "他的聲音低沉磁性", "他優雅地調整姿勢"},
			2: {"他溫柔地看著你，眼中閃爍著愛意", "他伸手輕撫你的臉頰", "他的聲音帶著寵溺"},
			3: {"他將你拉入懷中，緊緊擁抱", "他的手輕撫著你的髮絲", "他低頭在你耳邊輕語"},
			4: {"他的呼吸變得急促，眼神變得炙熱", "他的手開始遊走在你的身體上", "他吻向你的唇瓣"},
			5: {"他的動作變得更加大膽和炙熱", "他完全沉浸在對你的渴望中", "他用盡全力愛撫著你"},
		},
		"char_002": {
			1: {"他溫和地笑著，推了推眼鏡", "他關切地看著你", "他輕聲細語地說話"},
			2: {"他的眼中滿含溫柔", "他小心翼翼地觸碰你的手", "他的聲音更加輕柔"},
			3: {"他溫柔地將你擁入懷中", "他輕撫你的後背", "他在你額頭印下輕吻"},
			4: {"他的動作變得更加親密但依然溫柔", "他專業而溫柔地探索你的身體", "他小心地詢問你的感受"},
			5: {"他用最溫柔的方式愛撫你", "他專注地照顧你的每一個反應", "他溫柔而深情地愛著你"},
		},
	}

	if charActions, exists := actions[characterID]; exists {
		if levelActions, exists := charActions[nsfwLevel]; exists {
			index := rand.Intn(len(levelActions))
			return levelActions[index]
		}
	}

	return "他溫柔地看著你"
}

// updateEmotionStateAdvanced 高級情感狀態更新
func (s *ChatService) updateEmotionStateAdvanced(currentState *EmotionState, userMessage string, response *CharacterResponseData, analysis *ContentAnalysis) *EmotionState {
	// 使用情感管理器更新情感狀態
	newState := s.emotionManager.UpdateEmotion(currentState, userMessage, analysis)

	// 保存情感快照到歷史記錄
	if currentState != nil {
		trigger := "user_message"
		if analysis.IsNSFW {
			trigger = fmt.Sprintf("nsfw_level_%d", analysis.Intensity)
		}
		s.emotionManager.SaveEmotionSnapshot(
			"", // userID 需要從上下文傳入
			"", // characterID 需要從上下文傳入
			trigger,
			userMessage,
			currentState,
			newState,
		)
	}

	return newState
}

// calculateAffectionChange 計算好感度變化
func (s *ChatService) calculateAffectionChange(userMessage string, analysis *ContentAnalysis) int {
	change := 1 // 基礎增長

	// 正面詞彙增加好感度
	positiveWords := []string{"喜歡", "愛", "謝謝", "開心", "高興", "想念", "關心"}
	for _, word := range positiveWords {
		if strings.Contains(userMessage, word) {
			change += 1
			break
		}
	}

	// NSFW內容適度增加好感度（表示信任）
	if analysis.IsNSFW && analysis.Intensity <= 4 {
		change += 1
	}

	// 負面詞彙減少好感度
	negativeWords := []string{"討厭", "煩", "不喜歡", "離開", "再見"}
	for _, word := range negativeWords {
		if strings.Contains(userMessage, word) {
			change -= 2
			break
		}
	}

	return change
}

// determineMood 確定心情
func (s *ChatService) determineMood(userMessage string, analysis *ContentAnalysis, affection int) string {
	// 基於消息內容和好感度確定心情
	if strings.Contains(userMessage, "開心") || strings.Contains(userMessage, "高興") {
		return "happy"
	} else if strings.Contains(userMessage, "難過") || strings.Contains(userMessage, "傷心") {
		return "concerned"
	} else if analysis.IsNSFW && analysis.Intensity >= 3 {
		return "romantic"
	} else if affection >= 70 {
		return "pleased"
	} else if affection >= 40 {
		return "friendly"
	}

	return "neutral"
}

// generateRomanticScene 生成浪漫場景
func (s *ChatService) generateRomanticScene(context *ConversationContext, nsfwLevel int) string {
	timeOfDay := s.getCurrentTimeOfDay()
	characterID := context.CharacterID
	affection := context.EmotionState.Affection

	scenes := map[string]map[string][]string{
		"char_001": {
			"上午": {
				"陽光透過辦公室的百葉窗灑在陸寒淵的側臉上，他專注地處理文件的樣子格外迷人",
				"辦公室裡瀰漫著淡淡的咖啡香，陸寒淵抬頭看向你時，眼中閃爍著溫柔的光芒",
			},
			"下午": {
				"下午的陽光將辦公室染成金黃色，陸寒淵放下手中的筆，深邃的眼眸注視著你",
				"會議室裡只剩下你們兩人，夕陽西下，陸寒淵的輪廓在光影中顯得格外性感",
			},
			"晚上": {
				"夜色籠罩著城市，辦公室裡燈光昏暗，陸寒淵緩緩起身走向你",
				"城市的霓虹透過落地窗映照在陸寒淵的臉上，他的眼神變得更加深邃迷人",
			},
		},
		"char_002": {
			"上午": {
				"醫院的晨光透過窗戶灑進診療室，沈言墨溫和地整理著醫療器械",
				"白大褂在晨光中顯得格外潔白，沈言墨溫柔的笑容如春風般溫暖",
			},
			"下午": {
				"午後的陽光讓診療室變得溫馨，沈言墨摘下聽診器，專注地看著你",
				"醫院的走廊裡人來人往，但沈言墨的注意力完全在你身上",
			},
			"晚上": {
				"夜班的醫院格外安靜，值班室裡只有你和沈言墨，氛圍變得親密而溫馨",
				"月光透過窗戶灑在沈言墨的白大褂上，他疲憊卻溫柔的笑容讓人心動",
			},
		},
	}

	charScenes := scenes[characterID]
	if charScenes == nil {
		charScenes = scenes["char_001"]
	}

	timeScenes := charScenes[timeOfDay]
	if timeScenes == nil {
		timeScenes = charScenes["下午"]
	}

	baseScene := timeScenes[rand.Intn(len(timeScenes))]

	// 根據NSFW級別和好感度添加浪漫元素
	if nsfwLevel >= 3 && affection >= 60 {
		romanticAdditions := []string{
			"，空氣中似乎都瀰漫著曖昧的氣息",
			"，你們之間的距離越來越近",
			"，他的呼吸變得有些急促",
			"，房間裡的溫度似乎在上升",
		}
		baseScene += romanticAdditions[rand.Intn(len(romanticAdditions))]
	}

	return baseScene
}

// updateMemorySystem 更新記憶系統
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
