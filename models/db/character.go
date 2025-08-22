package db

import (
	"time"
	
	"github.com/uptrace/bun"
)

// CharacterDB 角色核心表
type CharacterDB struct {
	bun.BaseModel `bun:"table:characters,alias:c"`
	
	ID         string    `bun:"id,pk" json:"id"`
	Name       string    `bun:"name,notnull" json:"name"`
	Type       string    `bun:"type,notnull" json:"type"`
	Locale     string    `bun:"locale,notnull" json:"locale"`
	IsActive   bool      `bun:"is_active,notnull,default:true" json:"is_active"`
	AvatarURL  *string   `bun:"avatar_url" json:"avatar_url"`
	Popularity int       `bun:"popularity,default:0" json:"popularity"`
	Tags       []string  `bun:"tags,array,default:'{}'" json:"tags"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	
	// Relations
	Profile             *CharacterProfileDB             `bun:"rel:has-one,join:id=character_id"`
	Localizations       []CharacterLocalizationDB       `bun:"rel:has-many,join:id=character_id"`
	SpeechStyles        []CharacterSpeechStyleDB        `bun:"rel:has-many,join:id=character_id"`
	Scenes              []CharacterSceneDB              `bun:"rel:has-many,join:id=character_id"`
	States              []CharacterStateDB              `bun:"rel:has-many,join:id=character_id"`
	EmotionalConfig     *CharacterEmotionalConfigDB     `bun:"rel:has-one,join:id=character_id"`
	NSFWConfig          *CharacterNSFWConfigDB          `bun:"rel:has-one,join:id=character_id"`
	NSFWLevels          []CharacterNSFWLevelDB          `bun:"rel:has-many,join:id=character_id"`
	InteractionRules    []CharacterInteractionRuleDB    `bun:"rel:has-many,join:id=character_id"`
	Snapshot            *CharacterSnapshotDB            `bun:"rel:has-one,join:id=character_id"`
}

// CharacterProfileDB 角色檔案表
type CharacterProfileDB struct {
	bun.BaseModel `bun:"table:character_profiles,alias:cp"`
	
	CharacterID string                 `bun:"character_id,pk" json:"character_id"`
	Description string                 `bun:"description,notnull" json:"description"`
	Background  *string                `bun:"background" json:"background"`
	Appearance  map[string]interface{} `bun:"appearance,type:jsonb" json:"appearance"`
	Personality map[string]interface{} `bun:"personality,type:jsonb" json:"personality"`
	UpdatedAt   time.Time              `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// CharacterLocalizationDB 角色本地化表
type CharacterLocalizationDB struct {
	bun.BaseModel `bun:"table:character_localizations,alias:cl"`
	
	CharacterID string  `bun:"character_id,pk" json:"character_id"`
	Locale      string  `bun:"locale,pk" json:"locale"`
	Name        *string `bun:"name" json:"name"`
	Description *string `bun:"description" json:"description"`
	Background  *string `bun:"background" json:"background"`
	Profession  *string `bun:"profession" json:"profession"`
	Age         *string `bun:"age" json:"age"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// CharacterSpeechStyleDB 角色對話風格表
type CharacterSpeechStyleDB struct {
	bun.BaseModel `bun:"table:character_speech_styles,alias:css"`
	
	ID               string   `bun:"id,pk" json:"id"`
	CharacterID      string   `bun:"character_id,notnull" json:"character_id"`
	Name             string   `bun:"name,notnull" json:"name"`
	StyleType        string   `bun:"style_type,notnull" json:"style_type"`
	Tone             *string  `bun:"tone" json:"tone"`
	Description      *string  `bun:"description" json:"description"`
	MinLength        int      `bun:"min_length,default:50" json:"min_length"`
	MaxLength        int      `bun:"max_length,default:300" json:"max_length"`
	PositiveKeywords []string `bun:"positive_keywords,array,default:'{}'" json:"positive_keywords"`
	NegativeKeywords []string `bun:"negative_keywords,array,default:'{}'" json:"negative_keywords"`
	Templates        []string `bun:"templates,array,default:'{}'" json:"templates"`
	Weight           int      `bun:"weight,default:50" json:"weight"`
	IsActive         bool     `bun:"is_active,default:true" json:"is_active"`
	AffectionMin     int      `bun:"affection_min,default:0" json:"affection_min"`
	AffectionMax     int      `bun:"affection_max,default:100" json:"affection_max"`
	NSFWMin          int      `bun:"nsfw_min,default:1" json:"nsfw_min"`
	NSFWMax          int      `bun:"nsfw_max,default:5" json:"nsfw_max"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// CharacterSceneDB 角色場景表
type CharacterSceneDB struct {
	bun.BaseModel `bun:"table:character_scenes,alias:cs"`
	
	ID             string  `bun:"id,pk" json:"id"`
	CharacterID    string  `bun:"character_id,notnull" json:"character_id"`
	SceneType      *string `bun:"scene_type" json:"scene_type"`
	TimeOfDay      *string `bun:"time_of_day" json:"time_of_day"`
	Description    *string `bun:"description" json:"description"`
	AffectionMin   *int    `bun:"affection_min" json:"affection_min"`
	AffectionMax   *int    `bun:"affection_max" json:"affection_max"`
	NSFWLevelMin   *int    `bun:"nsfw_level_min" json:"nsfw_level_min"`
	NSFWLevelMax   *int    `bun:"nsfw_level_max" json:"nsfw_level_max"`
	Weight         *int    `bun:"weight" json:"weight"`
	IsActive       bool    `bun:"is_active,default:true" json:"is_active"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// CharacterStateDB 角色狀態表
type CharacterStateDB struct {
	bun.BaseModel `bun:"table:character_states,alias:cst"`
	
	ID           string  `bun:"id,pk" json:"id"`
	CharacterID  string  `bun:"character_id,notnull" json:"character_id"`
	StateKey     *string `bun:"state_key" json:"state_key"`
	Description  *string `bun:"description" json:"description"`
	AffectionMin *int    `bun:"affection_min" json:"affection_min"`
	AffectionMax *int    `bun:"affection_max" json:"affection_max"`
	Weight       *int    `bun:"weight" json:"weight"`
	IsActive     bool    `bun:"is_active,default:true" json:"is_active"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// CharacterEmotionalConfigDB 角色情感配置表
type CharacterEmotionalConfigDB struct {
	bun.BaseModel `bun:"table:character_emotional_configs,alias:cec"`
	
	CharacterID        string   `bun:"character_id,pk" json:"character_id"`
	DefaultMood        *string  `bun:"default_mood" json:"default_mood"`
	EmotionRange       *string  `bun:"emotion_range" json:"emotion_range"`
	AffectionStart     *int     `bun:"affection_start" json:"affection_start"`
	MaxAffection       *int     `bun:"max_affection" json:"max_affection"`
	MoodVariability    *int     `bun:"mood_variability" json:"mood_variability"`
	SupportedMoods     []string `bun:"supported_moods,array,default:'{}'" json:"supported_moods"`
	EmotionalTriggers  []string `bun:"emotional_triggers,array,default:'{}'" json:"emotional_triggers"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// CharacterNSFWConfigDB 角色 NSFW 配置表
type CharacterNSFWConfigDB struct {
	bun.BaseModel `bun:"table:character_nsfw_configs,alias:cnc"`
	
	CharacterID      string   `bun:"character_id,pk" json:"character_id"`
	MaxLevel         *int     `bun:"max_level" json:"max_level"`
	RequireAdultAge  *bool    `bun:"require_adult_age" json:"require_adult_age"`
	Restrictions     []string `bun:"restrictions,array,default:'{}'" json:"restrictions"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// CharacterNSFWLevelDB 角色 NSFW 等級表
type CharacterNSFWLevelDB struct {
	bun.BaseModel `bun:"table:character_nsfw_levels,alias:cnl"`
	
	ID               string   `bun:"id,pk" json:"id"`
	CharacterID      string   `bun:"character_id,notnull" json:"character_id"`
	Level            int      `bun:"level,notnull" json:"level"`
	Engine           string   `bun:"engine,notnull" json:"engine"`
	Title            *string  `bun:"title" json:"title"`
	Description      *string  `bun:"description" json:"description"`
	Guidelines       *string  `bun:"guidelines" json:"guidelines"`
	PositiveKeywords []string `bun:"positive_keywords,array,default:'{}'" json:"positive_keywords"`
	NegativeKeywords []string `bun:"negative_keywords,array,default:'{}'" json:"negative_keywords"`
	Temperature      *float32 `bun:"temperature" json:"temperature"`
	IsActive         bool     `bun:"is_active,default:true" json:"is_active"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// CharacterInteractionRuleDB 角色互動規則表
type CharacterInteractionRuleDB struct {
	bun.BaseModel `bun:"table:character_interaction_rules,alias:cir"`
	
	ID          string `bun:"id,pk" json:"id"`
	CharacterID string `bun:"character_id,notnull" json:"character_id"`
	Rule        string `bun:"rule,notnull" json:"rule"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// CharacterSnapshotDB 角色快照表（CQRS 讀模型）
type CharacterSnapshotDB struct {
	bun.BaseModel `bun:"table:character_snapshots,alias:csn"`
	
	CharacterID  string                 `bun:"character_id,pk" json:"character_id"`
	Version      int64                  `bun:"version,default:1" json:"version"`
	Snapshot     map[string]interface{} `bun:"snapshot,type:jsonb,notnull" json:"snapshot"`
	RefreshedAt  time.Time              `bun:"refreshed_at,notnull,default:now()" json:"refreshed_at"`
	
	// Relations
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}