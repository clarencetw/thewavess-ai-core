package models

import (
	"fmt"
	"strings"
	"time"
)

// CharacterType 角色類型枚舉
type CharacterType string

const (
	CharacterTypeDominant CharacterType = "dominant"   // 霸道型（如霸道總裁）
	CharacterTypeGentle   CharacterType = "gentle"     // 溫柔型（如溫柔醫生）
	CharacterTypePlayful  CharacterType = "playful"    // 活潑型
	CharacterTypeMystery  CharacterType = "mystery"    // 神秘型
	CharacterTypeReliable CharacterType = "reliable"   // 可靠型
)

// Locale 語言區域枚舉
type Locale string

const (
	LocaleTraditionalChinese Locale = "zh-Hant" // 繁體中文
	LocaleSimplifiedChinese  Locale = "zh-Hans" // 簡體中文
	LocaleJapanese          Locale = "ja"       // 日語
	LocaleEnglish           Locale = "en"       // 英語
)

// StyleType 對話風格類型枚舉
type StyleType string

const (
	StyleTypeStandard StyleType = "standard"  // 標準風格
	StyleTypeRomantic StyleType = "romantic"  // 浪漫風格
	StyleTypeIntimate StyleType = "intimate"  // 親密風格
	StyleTypePlayful  StyleType = "playful"   // 俏皮風格
	StyleTypeFormal   StyleType = "formal"    // 正式風格
	StyleTypeCasual   StyleType = "casual"    // 休閒風格
)

// EngineType AI引擎類型枚舉
type EngineType string

const (
	EngineOpenAI EngineType = "openai" // OpenAI GPT
	EngineGrok   EngineType = "grok"   // Grok
)

// NSFWLevel NSFW等級枚舉
type NSFWLevel int

const (
	NSFWLevelSafe     NSFWLevel = 1 // 安全內容
	NSFWLevelRomantic NSFWLevel = 2 // 浪漫內容
	NSFWLevelIntimate NSFWLevel = 3 // 親密內容
	NSFWLevelAdult    NSFWLevel = 4 // 成人內容
	NSFWLevelExplicit NSFWLevel = 5 // 明確成人內容
)

// Character 核心角色模型
type Character struct {
	// 基本身份信息
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Type     CharacterType `json:"type"`
	Locale   Locale        `json:"locale"`
	IsActive bool          `json:"is_active"`

	// 元數據信息
	Metadata CharacterMetadata `json:"metadata"`

	// 行為配置
	Behavior CharacterBehavior `json:"behavior"`

	// 內容配置
	Content CharacterContent `json:"content"`

	// 時間戳記
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CharacterMetadata 角色元數據
type CharacterMetadata struct {
	Description *string               `json:"description,omitempty"`
	AvatarURL   *string               `json:"avatar_url,omitempty"`
	Tags        []string              `json:"tags"`
	Popularity  int                   `json:"popularity"`
	Appearance  CharacterAppearance   `json:"appearance"`
	Background  *string               `json:"background,omitempty"`
	Personality CharacterPersonality  `json:"personality"`
}

// CharacterAppearance 角色外觀描述
type CharacterAppearance struct {
	Height      string `json:"height"`
	Build       string `json:"build"`
	EyeColor    string `json:"eye_color"`
	HairColor   string `json:"hair_color"`
	HairStyle   string `json:"hair_style"`
	SkinTone    string `json:"skin_tone"`
	Style       string `json:"style"`
	Distinctive string `json:"distinctive"` // 特徵描述
}

// CharacterPersonality 角色性格特質
type CharacterPersonality struct {
	Traits           []string                 `json:"traits"`
	CoreValues       []string                 `json:"core_values"`
	Strengths        []string                 `json:"strengths"`
	Weaknesses       []string                 `json:"weaknesses"`
	PersonalityScore CharacterPersonalityScore `json:"personality_score"`
	BehaviorPatterns []string                 `json:"behavior_patterns"`
}

// CharacterPersonalityScore 性格評分（1-10分）
type CharacterPersonalityScore struct {
	Extroversion  int `json:"extroversion"`  // 外向性
	Agreeableness int `json:"agreeableness"` // 親和性
	Dominance     int `json:"dominance"`     // 支配性
	Emotional     int `json:"emotional"`     // 情緒穩定性
	Openness      int `json:"openness"`      // 開放性
	Reliability   int `json:"reliability"`   // 可靠性
}

// CharacterBehavior 角色行為配置
type CharacterBehavior struct {
	SpeechStyles      []CharacterSpeechStyle   `json:"speech_styles"`
	NSFWConfig        CharacterNSFWConfig      `json:"nsfw_config"`
	EmotionalConfig   CharacterEmotionalConfig `json:"emotional_config"`
	InteractionRules  []string                 `json:"interaction_rules"`
	PreferredTopics   []string                 `json:"preferred_topics"`
	AvoidedTopics     []string                 `json:"avoided_topics"`
}

// CharacterSpeechStyle 對話風格配置
type CharacterSpeechStyle struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	StyleType        StyleType  `json:"style_type"`
	Tone             *string    `json:"tone,omitempty"`
	Description      *string    `json:"description,omitempty"`
	MinLength        int        `json:"min_length"`
	MaxLength        int        `json:"max_length"`
	PositiveKeywords []string   `json:"positive_keywords"`
	NegativeKeywords []string   `json:"negative_keywords"`
	Templates        []string   `json:"templates"`
	Weight           int        `json:"weight"`
	IsActive         bool       `json:"is_active"`
	AffectionRange   [2]int     `json:"affection_range"` // [min, max]
	NSFWRange        [2]int     `json:"nsfw_range"`      // [min, max]
}

// CharacterNSFWConfig NSFW配置
type CharacterNSFWConfig struct {
	MaxLevel        NSFWLevel              `json:"max_level"`
	RequireAdultAge bool                   `json:"require_adult_age"`
	Restrictions    []string               `json:"restrictions"`
	LevelConfigs    []CharacterNSFWLevel   `json:"level_configs"`
}

// CharacterNSFWLevel 特定等級NSFW配置
type CharacterNSFWLevel struct {
	Level            NSFWLevel    `json:"level"`
	Engine           EngineType   `json:"engine"`
	Title            *string      `json:"title,omitempty"`
	Description      *string      `json:"description,omitempty"`
	Guidelines       *string      `json:"guidelines,omitempty"`
	PositiveKeywords []string     `json:"positive_keywords"`
	NegativeKeywords []string     `json:"negative_keywords"`
	Temperature      *float64     `json:"temperature,omitempty"`
	IsActive         bool         `json:"is_active"`
}

// CharacterEmotionalConfig 情感配置
type CharacterEmotionalConfig struct {
	DefaultMood       *string  `json:"default_mood,omitempty"`
	EmotionRange      *string  `json:"emotion_range,omitempty"`
	AffectionStart    int      `json:"affection_start"`
	MaxAffection      int      `json:"max_affection"`
	MoodVariability   int      `json:"mood_variability"`
	SupportedMoods    []string `json:"supported_moods"`
	EmotionalTriggers []string `json:"emotional_triggers"`
}

// CharacterContent 角色內容配置
type CharacterContent struct {
	Scenes         []CharacterScene         `json:"scenes"`
	States         []CharacterState         `json:"states"`
	Localizations  map[Locale]CharacterL10N `json:"localizations"`
}

// CharacterScene 角色場景描述
type CharacterScene struct {
	ID           string     `json:"id"`
	SceneType    *string    `json:"scene_type,omitempty"`
	TimeOfDay    *string    `json:"time_of_day,omitempty"`
	Description  *string    `json:"description,omitempty"`
	AffectionMin int        `json:"affection_min"`
	AffectionMax int        `json:"affection_max"`
	NSFWLevelMin NSFWLevel  `json:"nsfw_level_min"`
	NSFWLevelMax NSFWLevel  `json:"nsfw_level_max"`
	Weight       int        `json:"weight"`
	IsActive     bool       `json:"is_active"`
}

// CharacterState 角色狀態描述
type CharacterState struct {
	ID           string  `json:"id"`
	StateKey     *string `json:"state_key,omitempty"`
	Description  *string `json:"description,omitempty"`
	AffectionMin int     `json:"affection_min"`
	AffectionMax int     `json:"affection_max"`
	Weight       int     `json:"weight"`
	IsActive     bool    `json:"is_active"`
}

// CharacterL10N 角色本地化配置
type CharacterL10N struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Background  *string `json:"background,omitempty"`
	Profession  *string `json:"profession,omitempty"`
	Age         *string `json:"age,omitempty"`
}

// CharacterStats 角色統計數據（分離的統計模型）
type CharacterStats struct {
	CharacterID       string                    `json:"character_id"`
	BasicInfo         CharacterBasicInfo        `json:"basic_info"`
	InteractionStats  CharacterInteractionStats `json:"interaction_stats"`
	RelationshipStats CharacterRelationshipStats `json:"relationship_stats"`
	ContentStats      CharacterContentStats     `json:"content_stats"`
	UserPreferences   CharacterUserPreferences  `json:"user_preferences"`
	PerformanceStats  CharacterPerformanceStats `json:"performance_stats"`
	GeneratedAt       time.Time                 `json:"generated_at"`
}

// CharacterBasicInfo 基本信息統計
type CharacterBasicInfo struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Popularity  int       `json:"popularity"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	Version     string    `json:"version"`
}

// CharacterInteractionStats 互動統計
type CharacterInteractionStats struct {
	TotalConversations int                 `json:"total_conversations"`
	TotalMessages      int                 `json:"total_messages"`
	TotalUsers         int                 `json:"total_users"`
	AvgSessionLength   int64               `json:"avg_session_length_seconds"`
	LastInteraction    *time.Time          `json:"last_interaction"`
	ActiveDays         int                 `json:"active_days"`
	MessagesByRole     map[string]int      `json:"messages_by_role"`
	EngineUsage        map[string]int      `json:"engine_usage"`
	HourlyDistribution map[string]int      `json:"hourly_distribution"`
	DailyDistribution  map[string]int      `json:"daily_distribution"`
}

// CharacterRelationshipStats 關係統計
type CharacterRelationshipStats struct {
	AvgAffectionLevel     float64                `json:"avg_affection_level"`
	RelationshipStages    map[string]int         `json:"relationship_stages"`
	MoodDistribution      map[string]int         `json:"mood_distribution"`
	IntimacyLevels        map[string]int         `json:"intimacy_levels"`
	KeyMoments            int                    `json:"key_moments"`
	SpecialEvents         int                    `json:"special_events"`
	EmotionalProgression  []EmotionalMilestone   `json:"emotional_progression"`
	AffectionProgression  []AffectionMilestone   `json:"affection_progression"`
}

// CharacterContentStats 內容統計
type CharacterContentStats struct {
	RomanticScenes          int            `json:"romantic_scenes"`
	DailyConversations      int            `json:"daily_conversations"`
	NSFWLevelDistribution   map[string]int `json:"nsfw_level_distribution"`
	SceneTypes              map[string]int `json:"scene_types"`
	SpeechStyleUsage        map[string]int `json:"speech_style_usage"`
	MemorableQuotes         int            `json:"memorable_quotes"`
	RegeneratedMessages     int            `json:"regenerated_messages"`
	UserSatisfactionRating  float64        `json:"user_satisfaction_rating"`
}

// CharacterUserPreferences 用戶偏好統計
type CharacterUserPreferences struct {
	FavoriteScenarios   []string       `json:"favorite_scenarios"`
	PreferredMoods      []string       `json:"preferred_moods"`
	InteractionStyles   []string       `json:"interaction_styles"`
	PopularTags         []string       `json:"popular_tags"`
	SessionModes        map[string]int `json:"session_modes"`
	NSFWPreferences     map[string]int `json:"nsfw_preferences"`
	TimePreferences     map[string]int `json:"time_preferences"`
	DevicePreferences   map[string]int `json:"device_preferences"`
}

// CharacterPerformanceStats 性能統計
type CharacterPerformanceStats struct {
	AvgResponseTime       int64   `json:"avg_response_time_ms"`
	SuccessRate           float64 `json:"success_rate"`
	ErrorRate             float64 `json:"error_rate"`
	TimeoutRate           float64 `json:"timeout_rate"`
	TokenUsage            int64   `json:"token_usage"`
	CostPerInteraction    float64 `json:"cost_per_interaction"`
	EnginePerformance     map[string]EngineStats `json:"engine_performance"`
}

// EmotionalMilestone 情感里程碑
type EmotionalMilestone struct {
	Date         time.Time `json:"date"`
	Event        string    `json:"event"`
	Affection    int       `json:"affection"`
	Relationship string    `json:"relationship"`
	UsersCount   int       `json:"users_count"`
	Significance string    `json:"significance"`
}

// AffectionMilestone 好感度里程碑
type AffectionMilestone struct {
	Date        time.Time `json:"date"`
	Level       int       `json:"level"`
	Description string    `json:"description"`
	UsersCount  int       `json:"users_count"`
}

// EngineStats 引擎性能統計
type EngineStats struct {
	RequestCount    int64   `json:"request_count"`
	AvgResponseTime int64   `json:"avg_response_time_ms"`
	SuccessRate     float64 `json:"success_rate"`
	ErrorRate       float64 `json:"error_rate"`
	TokenUsage      int64   `json:"token_usage"`
}

// 請求和響應 DTOs

// CharacterCreateRequest 創建角色請求
type CharacterCreateRequest struct {
	Name     string        `json:"name" binding:"required,min=1,max=50"`
	Type     CharacterType `json:"type" binding:"required"`
	Locale   Locale        `json:"locale" binding:"required"`
	Metadata CharacterMetadataRequest `json:"metadata" binding:"required"`
	Behavior CharacterBehaviorRequest `json:"behavior" binding:"required"`
	Content  CharacterContentRequest  `json:"content"`
}

// CharacterUpdateRequest 更新角色請求
type CharacterUpdateRequest struct {
	Name     *string                   `json:"name,omitempty" binding:"omitempty,min=1,max=50"`
	Type     *CharacterType            `json:"type,omitempty"`
	Locale   *Locale                   `json:"locale,omitempty"`
	IsActive *bool                     `json:"is_active,omitempty"`
	Metadata *CharacterMetadataRequest `json:"metadata,omitempty"`
	Behavior *CharacterBehaviorRequest `json:"behavior,omitempty"`
	Content  *CharacterContentRequest  `json:"content,omitempty"`
}

// CharacterMetadataRequest 元數據請求
type CharacterMetadataRequest struct {
	Description *string                    `json:"description,omitempty"`
	AvatarURL   *string                    `json:"avatar_url,omitempty"`
	Tags        []string                   `json:"tags"`
	Popularity  *int                       `json:"popularity,omitempty" binding:"omitempty,min=0,max=100"`
	Appearance  *CharacterAppearance       `json:"appearance,omitempty"`
	Background  *string                    `json:"background,omitempty"`
	Personality *CharacterPersonality      `json:"personality,omitempty"`
}

// CharacterBehaviorRequest 行為配置請求
type CharacterBehaviorRequest struct {
	SpeechStyles     []CharacterSpeechStyleRequest `json:"speech_styles"`
	NSFWConfig       *CharacterNSFWConfig          `json:"nsfw_config,omitempty"`
	EmotionalConfig  *CharacterEmotionalConfig     `json:"emotional_config,omitempty"`
	InteractionRules []string                      `json:"interaction_rules"`
	PreferredTopics  []string                      `json:"preferred_topics"`
	AvoidedTopics    []string                      `json:"avoided_topics"`
}

// CharacterSpeechStyleRequest 對話風格請求
type CharacterSpeechStyleRequest struct {
	Name             string    `json:"name" binding:"required,min=1,max=50"`
	StyleType        StyleType `json:"style_type" binding:"required"`
	Tone             string    `json:"tone" binding:"required,min=1,max=200"`
	Description      *string   `json:"description,omitempty"`
	MinLength        *int      `json:"min_length,omitempty" binding:"omitempty,min=10,max=1000"`
	MaxLength        *int      `json:"max_length,omitempty" binding:"omitempty,min=50,max=2000"`
	PositiveKeywords []string  `json:"positive_keywords"`
	NegativeKeywords []string  `json:"negative_keywords"`
	Templates        []string  `json:"templates"`
	Weight           *int      `json:"weight,omitempty" binding:"omitempty,min=1,max=100"`
	IsActive         *bool     `json:"is_active,omitempty"`
	AffectionRange   *[2]int   `json:"affection_range,omitempty"`
	NSFWRange        *[2]int   `json:"nsfw_range,omitempty"`
}

// CharacterContentRequest 內容配置請求
type CharacterContentRequest struct {
	Scenes        []CharacterSceneRequest         `json:"scenes"`
	States        []CharacterStateRequest         `json:"states"`
	Localizations map[Locale]CharacterL10NRequest `json:"localizations"`
}

// CharacterSceneRequest 場景請求
type CharacterSceneRequest struct {
	SceneType    string     `json:"scene_type" binding:"required"`
	TimeOfDay    string     `json:"time_of_day" binding:"required"`
	Description  string     `json:"description" binding:"required,min=1,max=500"`
	AffectionMin *int       `json:"affection_min,omitempty" binding:"omitempty,min=0,max=100"`
	AffectionMax *int       `json:"affection_max,omitempty" binding:"omitempty,min=0,max=100"`
	NSFWLevelMin *NSFWLevel `json:"nsfw_level_min,omitempty" binding:"omitempty,min=1,max=5"`
	NSFWLevelMax *NSFWLevel `json:"nsfw_level_max,omitempty" binding:"omitempty,min=1,max=5"`
	Weight       *int       `json:"weight,omitempty" binding:"omitempty,min=1,max=100"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

// CharacterStateRequest 狀態請求
type CharacterStateRequest struct {
	StateKey     string `json:"state_key" binding:"required"`
	Description  string `json:"description" binding:"required,min=1,max=200"`
	AffectionMin *int   `json:"affection_min,omitempty" binding:"omitempty,min=0,max=100"`
	AffectionMax *int   `json:"affection_max,omitempty" binding:"omitempty,min=0,max=100"`
	Weight       *int   `json:"weight,omitempty" binding:"omitempty,min=1,max=100"`
	IsActive     *bool  `json:"is_active,omitempty"`
}

// CharacterL10NRequest 本地化請求
type CharacterL10NRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Background  *string `json:"background,omitempty"`
	Profession  *string `json:"profession,omitempty"`
	Age         *string `json:"age,omitempty"`
}

// CharacterResponse 角色響應
type CharacterResponse struct {
	ID       string                    `json:"id"`
	Name     string                    `json:"name"`
	Type     CharacterType             `json:"type"`
	Locale   Locale                    `json:"locale"`
	IsActive bool                      `json:"is_active"`
	Metadata CharacterMetadataResponse `json:"metadata"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// CharacterDetailResponse 角色詳細響應
type CharacterDetailResponse struct {
	CharacterResponse
	Behavior CharacterBehaviorResponse `json:"behavior"`
	Content  CharacterContentResponse  `json:"content"`
}

// CharacterMetadataResponse 元數據響應
type CharacterMetadataResponse struct {
	Description *string               `json:"description,omitempty"`
	AvatarURL   *string               `json:"avatar_url,omitempty"`
	Tags        []string              `json:"tags"`
	Popularity  int                   `json:"popularity"`
	Appearance  CharacterAppearance   `json:"appearance"`
	Background  *string               `json:"background,omitempty"`
	Personality CharacterPersonality  `json:"personality"`
}

// CharacterBehaviorResponse 行為配置響應
type CharacterBehaviorResponse struct {
	SpeechStyles     []CharacterSpeechStyle    `json:"speech_styles"`
	NSFWConfig       CharacterNSFWConfig       `json:"nsfw_config"`
	EmotionalConfig  CharacterEmotionalConfig  `json:"emotional_config"`
	InteractionRules []string                  `json:"interaction_rules"`
	PreferredTopics  []string                  `json:"preferred_topics"`
	AvoidedTopics    []string                  `json:"avoided_topics"`
}

// CharacterContentResponse 內容配置響應
type CharacterContentResponse struct {
	Scenes        []CharacterScene         `json:"scenes"`
	States        []CharacterState         `json:"states"`
	Localizations map[Locale]CharacterL10N `json:"localizations"`
}

// CharacterListResponse 角色列表響應
type CharacterListResponse struct {
	Characters []*CharacterResponse `json:"characters"`
	Pagination PaginationResponse   `json:"pagination"`
}

// CharacterListQuery 角色列表查詢參數
type CharacterListQuery struct {
	Page       int           `form:"page,default=1" binding:"min=1"`
	PageSize   int           `form:"page_size,default=20" binding:"min=1,max=100"`
	Type       CharacterType `form:"type"`
	Locale     Locale        `form:"locale"`
	IsActive   *bool         `form:"is_active"`
	Tags       []string      `form:"tags"`
	Search     string        `form:"search"`
	SortBy     string        `form:"sort_by,default=created_at"`
	SortOrder  string        `form:"sort_order,default=desc" binding:"oneof=asc desc"`
}

// Validation methods

// IsValid 驗證角色類型
func (ct CharacterType) IsValid() bool {
	switch ct {
	case CharacterTypeDominant, CharacterTypeGentle, CharacterTypePlayful, CharacterTypeMystery, CharacterTypeReliable:
		return true
	}
	return false
}

// IsValid 驗證語言區域
func (l Locale) IsValid() bool {
	switch l {
	case LocaleTraditionalChinese, LocaleSimplifiedChinese, LocaleJapanese, LocaleEnglish:
		return true
	}
	return false
}

// IsValid 驗證對話風格類型
func (st StyleType) IsValid() bool {
	switch st {
	case StyleTypeStandard, StyleTypeRomantic, StyleTypeIntimate, StyleTypePlayful, StyleTypeFormal, StyleTypeCasual:
		return true
	}
	return false
}

// IsValid 驗證引擎類型
func (et EngineType) IsValid() bool {
	switch et {
	case EngineOpenAI, EngineGrok:
		return true
	}
	return false
}

// IsValid 驗證NSFW等級
func (nl NSFWLevel) IsValid() bool {
	return nl >= NSFWLevelSafe && nl <= NSFWLevelExplicit
}

// RequiresAdult 檢查是否需要成人驗證
func (nl NSFWLevel) RequiresAdult() bool {
	return nl >= NSFWLevelAdult
}

// IsValidPersonalityScore 驗證性格評分
func (ps CharacterPersonalityScore) IsValid() error {
	scores := []int{ps.Extroversion, ps.Agreeableness, ps.Dominance, ps.Emotional, ps.Openness, ps.Reliability}
	for i, score := range scores {
		if score < 1 || score > 10 {
			return fmt.Errorf("personality score %d is invalid: %d (must be 1-10)", i, score)
		}
	}
	return nil
}

// Error types for domain-specific error handling

// CharacterError 角色相關錯誤
type CharacterError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e CharacterError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Type, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Common error constructors
func NewCharacterNotFoundError(id string) CharacterError {
	return CharacterError{
		Type:    "CHARACTER_NOT_FOUND",
		Message: fmt.Sprintf("角色 %s 不存在", id),
		Field:   "id",
	}
}

func NewCharacterValidationError(field, message string) CharacterError {
	return CharacterError{
		Type:    "CHARACTER_VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

func NewCharacterConfigurationError(message string) CharacterError {
	return CharacterError{
		Type:    "CHARACTER_CONFIGURATION_ERROR",
		Message: message,
	}
}

// Model conversion methods

// ToResponse 轉換為基本響應格式
func (c *Character) ToResponse() *CharacterResponse {
	return &CharacterResponse{
		ID:       c.ID,
		Name:     c.Name,
		Type:     c.Type,
		Locale:   c.Locale,
		IsActive: c.IsActive,
		Metadata: CharacterMetadataResponse{
			Description: c.Metadata.Description,
			AvatarURL:   c.Metadata.AvatarURL,
			Tags:        c.Metadata.Tags,
			Popularity:  c.Metadata.Popularity,
			Appearance:  c.Metadata.Appearance,
			Background:  c.Metadata.Background,
			Personality: c.Metadata.Personality,
		},
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// ToDetailResponse 轉換為詳細響應格式
func (c *Character) ToDetailResponse() *CharacterDetailResponse {
	return &CharacterDetailResponse{
		CharacterResponse: *c.ToResponse(),
		Behavior: CharacterBehaviorResponse{
			SpeechStyles:     c.Behavior.SpeechStyles,
			NSFWConfig:       c.Behavior.NSFWConfig,
			EmotionalConfig:  c.Behavior.EmotionalConfig,
			InteractionRules: c.Behavior.InteractionRules,
			PreferredTopics:  c.Behavior.PreferredTopics,
			AvoidedTopics:    c.Behavior.AvoidedTopics,
		},
		Content: CharacterContentResponse{
			Scenes:        c.Content.Scenes,
			States:        c.Content.States,
			Localizations: c.Content.Localizations,
		},
	}
}

// FromCreateRequest 從創建請求構建角色
func (c *Character) FromCreateRequest(req *CharacterCreateRequest) {
	c.Name = req.Name
	c.Type = req.Type
	c.Locale = req.Locale
	c.IsActive = true

	// 設置元數據
	if req.Metadata.Description != nil {
		c.Metadata.Description = req.Metadata.Description
	}
	if req.Metadata.AvatarURL != nil {
		c.Metadata.AvatarURL = req.Metadata.AvatarURL
	}
	if req.Metadata.Tags != nil {
		c.Metadata.Tags = req.Metadata.Tags
	}
	if req.Metadata.Popularity != nil {
		c.Metadata.Popularity = *req.Metadata.Popularity
	}
	if req.Metadata.Appearance != nil {
		c.Metadata.Appearance = *req.Metadata.Appearance
	}
	if req.Metadata.Background != nil {
		c.Metadata.Background = req.Metadata.Background
	}
	if req.Metadata.Personality != nil {
		c.Metadata.Personality = *req.Metadata.Personality
	}

	// 設置行為配置
	c.Behavior.SpeechStyles = make([]CharacterSpeechStyle, len(req.Behavior.SpeechStyles))
	for i, styleReq := range req.Behavior.SpeechStyles {
		style := &c.Behavior.SpeechStyles[i]
		style.ID = fmt.Sprintf("style_%d", i+1)
		style.Name = styleReq.Name
		style.StyleType = styleReq.StyleType
		style.Tone = &styleReq.Tone
		if styleReq.Description != nil {
			style.Description = styleReq.Description
		}
		if styleReq.MinLength != nil {
			style.MinLength = *styleReq.MinLength
		} else {
			style.MinLength = 50
		}
		if styleReq.MaxLength != nil {
			style.MaxLength = *styleReq.MaxLength
		} else {
			style.MaxLength = 300
		}
		if styleReq.Weight != nil {
			style.Weight = *styleReq.Weight
		} else {
			style.Weight = 50
		}
		if styleReq.IsActive != nil {
			style.IsActive = *styleReq.IsActive
		} else {
			style.IsActive = true
		}
		if styleReq.AffectionRange != nil {
			style.AffectionRange = *styleReq.AffectionRange
		}
		if styleReq.NSFWRange != nil {
			style.NSFWRange = *styleReq.NSFWRange
		}
		style.PositiveKeywords = styleReq.PositiveKeywords
		style.NegativeKeywords = styleReq.NegativeKeywords
		style.Templates = styleReq.Templates
	}

	if req.Behavior.NSFWConfig != nil {
		c.Behavior.NSFWConfig = *req.Behavior.NSFWConfig
	}
	if req.Behavior.EmotionalConfig != nil {
		c.Behavior.EmotionalConfig = *req.Behavior.EmotionalConfig
	}
	c.Behavior.InteractionRules = req.Behavior.InteractionRules
	c.Behavior.PreferredTopics = req.Behavior.PreferredTopics
	c.Behavior.AvoidedTopics = req.Behavior.AvoidedTopics

	// 設置內容配置
	if len(req.Content.Scenes) > 0 || len(req.Content.States) > 0 || len(req.Content.Localizations) > 0 {
		// 處理場景
		c.Content.Scenes = make([]CharacterScene, len(req.Content.Scenes))
		for i, sceneReq := range req.Content.Scenes {
			scene := &c.Content.Scenes[i]
			scene.ID = fmt.Sprintf("scene_%d", i+1)
			scene.SceneType = &sceneReq.SceneType
			scene.TimeOfDay = &sceneReq.TimeOfDay
			scene.Description = &sceneReq.Description
			if sceneReq.AffectionMin != nil {
				scene.AffectionMin = *sceneReq.AffectionMin
			}
			if sceneReq.AffectionMax != nil {
				scene.AffectionMax = *sceneReq.AffectionMax
			} else {
				scene.AffectionMax = 100
			}
			if sceneReq.NSFWLevelMin != nil {
				scene.NSFWLevelMin = *sceneReq.NSFWLevelMin
			} else {
				scene.NSFWLevelMin = NSFWLevelSafe
			}
			if sceneReq.NSFWLevelMax != nil {
				scene.NSFWLevelMax = *sceneReq.NSFWLevelMax
			} else {
				scene.NSFWLevelMax = NSFWLevelExplicit
			}
			if sceneReq.Weight != nil {
				scene.Weight = *sceneReq.Weight
			} else {
				scene.Weight = 50
			}
			if sceneReq.IsActive != nil {
				scene.IsActive = *sceneReq.IsActive
			} else {
				scene.IsActive = true
			}
		}

		// 處理狀態
		c.Content.States = make([]CharacterState, len(req.Content.States))
		for i, stateReq := range req.Content.States {
			state := &c.Content.States[i]
			state.ID = fmt.Sprintf("state_%d", i+1)
			state.StateKey = &stateReq.StateKey
			state.Description = &stateReq.Description
			if stateReq.AffectionMin != nil {
				state.AffectionMin = *stateReq.AffectionMin
			}
			if stateReq.AffectionMax != nil {
				state.AffectionMax = *stateReq.AffectionMax
			} else {
				state.AffectionMax = 100
			}
			if stateReq.Weight != nil {
				state.Weight = *stateReq.Weight
			} else {
				state.Weight = 50
			}
			if stateReq.IsActive != nil {
				state.IsActive = *stateReq.IsActive
			} else {
				state.IsActive = true
			}
		}

		// 處理本地化
		if len(req.Content.Localizations) > 0 {
			c.Content.Localizations = make(map[Locale]CharacterL10N)
			for locale, l10nReq := range req.Content.Localizations {
				l10n := CharacterL10N{}
				if l10nReq.Name != nil {
					l10n.Name = l10nReq.Name
				}
				if l10nReq.Description != nil {
					l10n.Description = l10nReq.Description
				}
				if l10nReq.Background != nil {
					l10n.Background = l10nReq.Background
				}
				if l10nReq.Profession != nil {
					l10n.Profession = l10nReq.Profession
				}
				if l10nReq.Age != nil {
					l10n.Age = l10nReq.Age
				}
				c.Content.Localizations[locale] = l10n
			}
		}
	}

	// 設置時間戳記
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
}

// ApplyUpdateRequest 應用更新請求
func (c *Character) ApplyUpdateRequest(req *CharacterUpdateRequest) {
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Type != nil {
		c.Type = *req.Type
	}
	if req.Locale != nil {
		c.Locale = *req.Locale
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}

	// 更新元數據
	if req.Metadata != nil {
		if req.Metadata.Description != nil {
			c.Metadata.Description = req.Metadata.Description
		}
		if req.Metadata.AvatarURL != nil {
			c.Metadata.AvatarURL = req.Metadata.AvatarURL
		}
		if req.Metadata.Tags != nil {
			c.Metadata.Tags = req.Metadata.Tags
		}
		if req.Metadata.Popularity != nil {
			c.Metadata.Popularity = *req.Metadata.Popularity
		}
		if req.Metadata.Appearance != nil {
			c.Metadata.Appearance = *req.Metadata.Appearance
		}
		if req.Metadata.Background != nil {
			c.Metadata.Background = req.Metadata.Background
		}
		if req.Metadata.Personality != nil {
			c.Metadata.Personality = *req.Metadata.Personality
		}
	}

	// 更新行為配置
	if req.Behavior != nil {
		if req.Behavior.NSFWConfig != nil {
			c.Behavior.NSFWConfig = *req.Behavior.NSFWConfig
		}
		if req.Behavior.EmotionalConfig != nil {
			c.Behavior.EmotionalConfig = *req.Behavior.EmotionalConfig
		}
		if req.Behavior.InteractionRules != nil {
			c.Behavior.InteractionRules = req.Behavior.InteractionRules
		}
		if req.Behavior.PreferredTopics != nil {
			c.Behavior.PreferredTopics = req.Behavior.PreferredTopics
		}
		if req.Behavior.AvoidedTopics != nil {
			c.Behavior.AvoidedTopics = req.Behavior.AvoidedTopics
		}
	}

	// 更新內容配置
	if len(req.Content.Localizations) > 0 {
		if c.Content.Localizations == nil {
			c.Content.Localizations = make(map[Locale]CharacterL10N)
		}
		for locale, l10nReq := range req.Content.Localizations {
			l10n := c.Content.Localizations[locale]
			if l10nReq.Name != nil {
				l10n.Name = l10nReq.Name
			}
			if l10nReq.Description != nil {
				l10n.Description = l10nReq.Description
			}
			if l10nReq.Background != nil {
				l10n.Background = l10nReq.Background
			}
			if l10nReq.Profession != nil {
				l10n.Profession = l10nReq.Profession
			}
			if l10nReq.Age != nil {
				l10n.Age = l10nReq.Age
			}
			c.Content.Localizations[locale] = l10n
		}
	}

	// 更新時間戳記
	c.UpdatedAt = time.Now()
}

// Utility methods

// GetName 獲取角色名稱（支持本地化）
func (c *Character) GetName(locale Locale) string {
	if c.Content.Localizations != nil {
		if l10n, exists := c.Content.Localizations[locale]; exists && l10n.Name != nil && *l10n.Name != "" {
			return *l10n.Name
		}
	}
	return c.Name
}

// GetDescription 獲取角色描述（支持本地化）
func (c *Character) GetDescription(locale Locale) string {
	if c.Content.Localizations != nil {
		if l10n, exists := c.Content.Localizations[locale]; exists && l10n.Description != nil && *l10n.Description != "" {
			return *l10n.Description
		}
	}
	if c.Metadata.Description != nil {
		return *c.Metadata.Description
	}
	return ""
}

// GetBestSpeechStyle 根據上下文獲取最適合的對話風格
func (c *Character) GetBestSpeechStyle(nsfwLevel NSFWLevel, affection int) *CharacterSpeechStyle {
	var bestStyle *CharacterSpeechStyle
	bestWeight := -1

	for i := range c.Behavior.SpeechStyles {
		style := &c.Behavior.SpeechStyles[i]
		if !style.IsActive {
			continue
		}

		// 檢查NSFW等級範圍
		if style.NSFWRange != [2]int{} {
			if int(nsfwLevel) < style.NSFWRange[0] || int(nsfwLevel) > style.NSFWRange[1] {
				continue
			}
		}

		// 檢查好感度範圍
		if style.AffectionRange != [2]int{} {
			if affection < style.AffectionRange[0] || affection > style.AffectionRange[1] {
				continue
			}
		}

		// 選擇權重最高的風格
		if style.Weight > bestWeight {
			bestStyle = style
			bestWeight = style.Weight
		}
	}

	return bestStyle
}

// GetNSFWConfig 獲取指定等級的NSFW配置
func (c *Character) GetNSFWConfig(level NSFWLevel) *CharacterNSFWLevel {
	for i := range c.Behavior.NSFWConfig.LevelConfigs {
		config := &c.Behavior.NSFWConfig.LevelConfigs[i]
		if config.Level == level && config.IsActive {
			return config
		}
	}
	return nil
}

// GetActiveScenes 獲取符合條件的活躍場景
func (c *Character) GetActiveScenes(sceneType, timeOfDay string, affection int, nsfwLevel NSFWLevel) []CharacterScene {
	var scenes []CharacterScene
	
	for _, scene := range c.Content.Scenes {
		if !scene.IsActive {
			continue
		}
		
		if sceneType != "" && (scene.SceneType == nil || *scene.SceneType != sceneType) {
			continue
		}
		
		if timeOfDay != "" && (scene.TimeOfDay == nil || *scene.TimeOfDay != timeOfDay) {
			continue
		}
		
		if affection < scene.AffectionMin || affection > scene.AffectionMax {
			continue
		}
		
		if nsfwLevel < scene.NSFWLevelMin || nsfwLevel > scene.NSFWLevelMax {
			continue
		}
		
		scenes = append(scenes, scene)
	}
	
	return scenes
}

// MatchesQuery 檢查角色是否符合查詢條件
func (c *Character) MatchesQuery(query *CharacterListQuery) bool {
	// 類型過濾
	if query.Type != "" && c.Type != query.Type {
		return false
	}

	// 語言過濾
	if query.Locale != "" && c.Locale != query.Locale {
		return false
	}

	// 狀態過濾
	if query.IsActive != nil && c.IsActive != *query.IsActive {
		return false
	}

	// 標籤過濾
	if len(query.Tags) > 0 {
		hasTag := false
		for _, queryTag := range query.Tags {
			for _, charTag := range c.Metadata.Tags {
				if strings.EqualFold(charTag, queryTag) {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	// 搜索過濾
	if query.Search != "" {
		search := strings.ToLower(query.Search)
		description := ""
		if c.Metadata.Description != nil {
			description = *c.Metadata.Description
		}
		background := ""
		if c.Metadata.Background != nil {
			background = *c.Metadata.Background
		}
		if !strings.Contains(strings.ToLower(c.Name), search) &&
			!strings.Contains(strings.ToLower(description), search) &&
			!strings.Contains(strings.ToLower(background), search) {
			return false
		}
	}

	return true
}

// Validate 驗證角色數據
func (c *Character) Validate() error {
	if c.ID == "" {
		return NewCharacterValidationError("id", "角色ID不能為空")
	}
	if c.Name == "" {
		return NewCharacterValidationError("name", "角色名稱不能為空")
	}
	if !c.Type.IsValid() {
		return NewCharacterValidationError("type", "無效的角色類型")
	}
	if !c.Locale.IsValid() {
		return NewCharacterValidationError("locale", "無效的語言區域")
	}

	// 驗證性格評分
	if err := c.Metadata.Personality.PersonalityScore.IsValid(); err != nil {
		return NewCharacterValidationError("personality_score", err.Error())
	}

	// 驗證對話風格
	for i, style := range c.Behavior.SpeechStyles {
		if style.Name == "" {
			return NewCharacterValidationError(fmt.Sprintf("behavior.speech_styles[%d].name", i), "對話風格名稱不能為空")
		}
		if !style.StyleType.IsValid() {
			return NewCharacterValidationError(fmt.Sprintf("behavior.speech_styles[%d].style_type", i), "無效的對話風格類型")
		}
	}

	// 驗證NSFW配置
	for i, config := range c.Behavior.NSFWConfig.LevelConfigs {
		if !config.Level.IsValid() {
			return NewCharacterValidationError(fmt.Sprintf("behavior.nsfw_config.level_configs[%d].level", i), "無效的NSFW等級")
		}
		if !config.Engine.IsValid() {
			return NewCharacterValidationError(fmt.Sprintf("behavior.nsfw_config.level_configs[%d].engine", i), "無效的AI引擎類型")
		}
	}

	return nil
}