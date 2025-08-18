package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Character 角色模型
type Character struct {
	bun.BaseModel `bun:"table:characters,alias:c"`

	ID          string                 `bun:"id,pk" json:"id"`
	Name        string                 `bun:"name,notnull" json:"name"`
	Type        string                 `bun:"type,notnull" json:"type"`
	Description string                 `bun:"description" json:"description,omitempty"`
	AvatarURL   string                 `bun:"avatar_url" json:"avatar_url,omitempty"`
	Popularity  int                    `bun:"popularity,default:0" json:"popularity"`
	Tags        []string               `bun:"tags,array" json:"tags,omitempty"`
	Appearance  map[string]interface{} `bun:"appearance,type:jsonb" json:"appearance,omitempty"`
	Personality map[string]interface{} `bun:"personality,type:jsonb" json:"personality,omitempty"`
	Background  string                 `bun:"background" json:"background,omitempty"`
	IsActive    bool                   `bun:"is_active,default:true" json:"is_active"`
	CreatedAt   time.Time              `bun:"created_at,nullzero,default:now()" json:"created_at"`
	UpdatedAt   time.Time              `bun:"updated_at,nullzero,default:now()" json:"updated_at"`

	// 關聯
	Sessions []*ChatSession `bun:"rel:has-many,join:id=character_id" json:"sessions,omitempty"`
	Scenes   []*Scene       `bun:"rel:has-many,join:id=character_id" json:"scenes,omitempty"`
}

// TableName 返回數據庫表名
func (c *Character) TableName() string {
	return "characters"
}

// BeforeAppendModel 在模型操作前執行
func (c *Character) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		c.UpdatedAt = time.Now()
	}
	return nil
}

// CharacterResponse 角色響應格式
type CharacterResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	AvatarURL   string                 `json:"avatar_url,omitempty"`
	Popularity  int                    `json:"popularity"`
	Tags        []string               `json:"tags,omitempty"`
	Appearance  map[string]interface{} `json:"appearance,omitempty"`
	Personality map[string]interface{} `json:"personality,omitempty"`
	Background  string                 `json:"background,omitempty"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ToResponse 轉換為響應格式
func (c *Character) ToResponse() *CharacterResponse {
	return &CharacterResponse{
		ID:          c.ID,
		Name:        c.Name,
		Type:        c.Type,
		Description: c.Description,
		AvatarURL:   c.AvatarURL,
		Popularity:  c.Popularity,
		Tags:        c.Tags,
		Appearance:  c.Appearance,
		Personality: c.Personality,
		Background:  c.Background,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// CharacterListResponse 角色列表響應 (Bun 版本)
type CharacterListResponse struct {
	Characters []*CharacterResponse `json:"characters"`
	Pagination PaginationResponse   `json:"pagination"`
}

// CharacterStatsResponse 角色統計響應
type CharacterStatsResponse struct {
	CharacterID    string                 `json:"character_id"`
	BasicInfo      CharacterBasicInfo     `json:"basic_info"`
	InteractionStats CharacterInteractionStats `json:"interaction_stats"`
	RelationshipStats CharacterRelationshipStats `json:"relationship_stats"`
	ContentStats   CharacterContentStats  `json:"content_stats"`
	UserPreferences CharacterUserPreferences `json:"user_preferences"`
	GeneratedAt    time.Time              `json:"generated_at"`
}

// CharacterBasicInfo 角色基本信息
type CharacterBasicInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Popularity  int      `json:"popularity"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// CharacterInteractionStats 互動統計
type CharacterInteractionStats struct {
	TotalConversations int           `json:"total_conversations"`
	TotalMessages      int           `json:"total_messages"`
	TotalUsers         int           `json:"total_users"`
	AvgSessionLength   int64         `json:"avg_session_length"` // Duration in seconds
	LastInteraction    *time.Time    `json:"last_interaction"`
	ActiveDays         int           `json:"active_days"`
	MessagesByRole     map[string]int `json:"messages_by_role"`
	EngineUsage        map[string]int `json:"engine_usage"`
}

// CharacterRelationshipStats 關係統計
type CharacterRelationshipStats struct {
	AvgAffectionLevel   float64            `json:"avg_affection_level"`
	RelationshipStages  map[string]int     `json:"relationship_stages"`
	MoodDistribution    map[string]int     `json:"mood_distribution"`
	IntimacyLevels      map[string]int     `json:"intimacy_levels"`
	KeyMoments          int                `json:"key_moments"`
	SpecialEvents       int                `json:"special_events"`
	EmotionalProgression []EmotionalMilestone `json:"emotional_progression"`
}

// CharacterContentStats 內容統計
type CharacterContentStats struct {
	RomanticScenes     int            `json:"romantic_scenes"`
	DailyConversations int            `json:"daily_conversations"`
	NSFWLevelDistribution map[string]int `json:"nsfw_level_distribution"`
	SceneTypes         map[string]int `json:"scene_types"`
	MemorableQuotes    int            `json:"memorable_quotes"`
	RegeneratedMessages int           `json:"regenerated_messages"`
}

// CharacterUserPreferences 用戶偏好統計
type CharacterUserPreferences struct {
	FavoriteScenarios  []string `json:"favorite_scenarios"`
	PreferredMoods     []string `json:"preferred_moods"`
	InteractionStyles  []string `json:"interaction_styles"`
	PopularTags        []string `json:"popular_tags"`
	SessionModes       map[string]int `json:"session_modes"`
}

// EmotionalMilestone 情感里程碑
type EmotionalMilestone struct {
	Date        time.Time `json:"date"`
	Event       string    `json:"event"`
	Affection   int       `json:"affection"`
	Relationship string   `json:"relationship"`
	UsersCount  int       `json:"users_count"`
}

// Scene 場景模型
type Scene struct {
	bun.BaseModel `bun:"table:scenes,alias:s"`

	ID               string    `bun:"id,pk" json:"id"`
	CharacterID      string    `bun:"character_id,notnull" json:"character_id"`
	TimeOfDay        string    `bun:"time_of_day,notnull" json:"time_of_day"`
	AffectionMin     int       `bun:"affection_min,default:0" json:"affection_min"`
	AffectionMax     int       `bun:"affection_max,default:100" json:"affection_max"`
	NSFWLevelMin     int       `bun:"nsfw_level_min,default:1" json:"nsfw_level_min"`
	NSFWLevelMax     int       `bun:"nsfw_level_max,default:5" json:"nsfw_level_max"`
	Description      string    `bun:"description,notnull" json:"description"`
	RomanticAddition string    `bun:"romantic_addition" json:"romantic_addition,omitempty"`
	Weight           int       `bun:"weight,default:1" json:"weight"`
	IsActive         bool      `bun:"is_active,default:true" json:"is_active"`
	CreatedAt        time.Time `bun:"created_at,nullzero,default:now()" json:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at,nullzero,default:now()" json:"updated_at"`

	// 關聯
	Character *Character `bun:"rel:belongs-to,join:character_id=id" json:"character,omitempty"`
}

// TableName 返回數據庫表名
func (s *Scene) TableName() string {
	return "scenes"
}

// BeforeAppendModel 在模型操作前執行
func (s *Scene) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		s.UpdatedAt = time.Now()
	}
	return nil
}

// SceneResponse 場景響應格式
type SceneResponse struct {
	ID               string    `json:"id"`
	CharacterID      string    `json:"character_id"`
	TimeOfDay        string    `json:"time_of_day"`
	AffectionMin     int       `json:"affection_min"`
	AffectionMax     int       `json:"affection_max"`
	NSFWLevelMin     int       `json:"nsfw_level_min"`
	NSFWLevelMax     int       `json:"nsfw_level_max"`
	Description      string    `json:"description"`
	RomanticAddition string    `json:"romantic_addition,omitempty"`
	Weight           int       `json:"weight"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ToResponse 轉換為響應格式
func (s *Scene) ToResponse() *SceneResponse {
	return &SceneResponse{
		ID:               s.ID,
		CharacterID:      s.CharacterID,
		TimeOfDay:        s.TimeOfDay,
		AffectionMin:     s.AffectionMin,
		AffectionMax:     s.AffectionMax,
		NSFWLevelMin:     s.NSFWLevelMin,
		NSFWLevelMax:     s.NSFWLevelMax,
		Description:      s.Description,
		RomanticAddition: s.RomanticAddition,
		Weight:           s.Weight,
		IsActive:         s.IsActive,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}