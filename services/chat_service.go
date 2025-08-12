package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
)

// ChatService 對話服務
type ChatService struct {
	openaiClient *OpenAIClient
	grokClient   *GrokClient
	config       *ChatConfig
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
	SessionID         string         `json:"session_id"`
	MessageID         string         `json:"message_id"`
	SceneDescription  string         `json:"scene_description"`
	CharacterDialogue string         `json:"character_dialogue"`
	CharacterAction   string         `json:"character_action"`
	EmotionState      *EmotionState  `json:"emotion_state"`
	AIEngine          string         `json:"ai_engine"`
	NSFWLevel         int            `json:"nsfw_level"`
	ResponseTime      time.Duration  `json:"response_time"`
	NovelChoices      []NovelChoice  `json:"novel_choices,omitempty"`
	SpecialEvent      *SpecialEvent  `json:"special_event,omitempty"`
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
	Intensity     int      `json:"intensity"`      // 1-5 級
	Categories    []string `json:"categories"`     // romantic, suggestive, explicit
	ShouldUseGrok bool     `json:"should_use_grok"`
	Confidence    float64  `json:"confidence"`
}

// ConversationContext 對話上下文
type ConversationContext struct {
	SessionID       string                 `json:"session_id"`
	UserID          string                 `json:"user_id"`
	CharacterID     string                 `json:"character_id"`
	RecentMessages  []models.ChatMessage   `json:"recent_messages"`
	EmotionState    *EmotionState          `json:"emotion_state"`
	SceneState      *SceneDescriptor       `json:"scene_state"`
	UserPreferences map[string]interface{} `json:"user_preferences"`
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
		openaiClient: NewOpenAIClient(),
		grokClient:   NewGrokClient(),
		config:       config,
	}
}

// ProcessMessage 處理用戶消息並生成回應
func (s *ChatService) ProcessMessage(ctx context.Context, request *ProcessMessageRequest) (*ChatResponse, error) {
	startTime := time.Now()

	// 1. 內容分析與分類
	analysis, err := s.analyzeContent(request.UserMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze content: %w", err)
	}

	// 2. 構建對話上下文
	context, err := s.buildConversationContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to build context: %w", err)
	}

	// 3. 選擇 AI 引擎
	engine := s.selectAIEngine(analysis, context.UserPreferences)

	// 4. 生成場景敘述
	sceneDescription := ""
	if s.config.Scene.EnableDescriptions {
		sceneDescription = s.generateSceneDescription(context)
	}

	// 5. 生成角色回應
	response, err := s.generateCharacterResponse(ctx, engine, request.UserMessage, context, sceneDescription)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// 6. 更新情感狀態
	newEmotionState := s.updateEmotionState(context.EmotionState, request.UserMessage, response)

	// 7. 構建最終回應
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
	}

	return chatResponse, nil
}

// analyzeContent 分析消息內容
func (s *ChatService) analyzeContent(message string) (*ContentAnalysis, error) {
	// TODO: 實現內容分析邏輯
	// 這裡先返回基本分析，後續會實現 NSFW 檢測
	return &ContentAnalysis{
		IsNSFW:        false,
		Intensity:     1,
		Categories:    []string{"normal"},
		ShouldUseGrok: false,
		Confidence:    0.95,
	}, nil
}

// buildConversationContext 構建對話上下文
func (s *ChatService) buildConversationContext(ctx context.Context, request *ProcessMessageRequest) (*ConversationContext, error) {
	// TODO: 從數據庫獲取會話歷史和用戶偏好
	
	// 模擬數據，後續會從數據庫獲取
	return &ConversationContext{
		SessionID:   request.SessionID,
		UserID:      request.UserID,
		CharacterID: request.CharacterID,
		EmotionState: &EmotionState{
			Affection:     50,
			Mood:          "neutral",
			Relationship:  "stranger",
			IntimacyLevel: "friendly",
		},
		SceneState: &SceneDescriptor{
			Location:       "辦公室",
			TimeOfDay:      "下午",
			Weather:        "陽光透過百葉窗灑進室內",
			Mood:           "professional",
			CharacterState: "專注工作中",
		},
		UserPreferences: map[string]interface{}{
			"nsfw_enabled": true,
			"scene_style":  "romantic",
		},
	}, nil
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

// generateCharacterResponse 生成角色回應
func (s *ChatService) generateCharacterResponse(ctx context.Context, engine, userMessage string, context *ConversationContext, sceneDescription string) (*CharacterResponseData, error) {
	var response *OpenAIResponse
	var err error
	
	if engine == "openai" {
		// 構建 OpenAI 請求
		messages := s.openaiClient.BuildCharacterPrompt(context.CharacterID, userMessage, sceneDescription, context)
		
		request := &OpenAIRequest{
			Model:       s.config.OpenAI.Model,
			Messages:    messages,
			MaxTokens:   s.config.OpenAI.MaxTokens,
			Temperature: s.config.OpenAI.Temperature,
			User:        context.UserID,
		}
		
		response, err = s.openaiClient.GenerateResponse(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("OpenAI API call failed: %w", err)
		}
	} else {
		// TODO: 實現 Grok 調用
		return s.generateFallbackResponse(context.CharacterID), nil
	}
	
	// 解析回應
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}
	
	content := response.Choices[0].Message.Content
	
	// 嘗試解析 "對話|||動作" 格式
	parts := strings.Split(content, "|||")
	if len(parts) >= 2 {
		return &CharacterResponseData{
			Dialogue: strings.TrimSpace(parts[0]),
			Action:   strings.TrimSpace(parts[1]),
		}, nil
	}
	
	// 如果沒有分隔符，將整個內容作為對話，生成預設動作
	dialogue := strings.TrimSpace(content)
	action := s.generateDefaultAction(context.CharacterID, dialogue)
	
	return &CharacterResponseData{
		Dialogue: dialogue,
		Action:   action,
	}, nil
}

// generateFallbackResponse 生成備用回應（當 API 不可用時）
func (s *ChatService) generateFallbackResponse(characterID string) *CharacterResponseData {
	switch characterID {
	case "char_001": // 陸寒淵
		return &CharacterResponseData{
			Dialogue: "你今天看起來有些疲憊，需要我為你準備什麼嗎？",
			Action:   "他的聲音低沉磁性，眼中帶著一絲不易察覺的關切",
		}
	case "char_002": // 沈言墨
		return &CharacterResponseData{
			Dialogue: "你好，很高興見到你。今天過得怎麼樣？",
			Action:   "他溫和地笑著，推了推鼻樑上的眼鏡",
		}
	default:
		return &CharacterResponseData{
			Dialogue: "你好，很高興與你對話。",
			Action:   "角色友善地看著你",
		}
	}
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

// generateMessageID 生成消息 ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}