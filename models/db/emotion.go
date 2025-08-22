package db

import (
	"time"
	"github.com/uptrace/bun"
)

// EmotionStateDB 情感狀態資料庫模型
type EmotionStateDB struct {
	bun.BaseModel `bun:"table:emotion_states,alias:es"`
	
	ID                string    `bun:"id,pk" json:"id"`
	UserID            string    `bun:"user_id,notnull" json:"user_id"`
	CharacterID       string    `bun:"character_id,notnull" json:"character_id"`
	Affection         int       `bun:"affection,notnull,default:30" json:"affection"`
	Mood              string    `bun:"mood,notnull,default:'neutral'" json:"mood"`
	Relationship      string    `bun:"relationship,notnull,default:'stranger'" json:"relationship"`
	IntimacyLevel     string    `bun:"intimacy_level,notnull,default:'distant'" json:"intimacy_level"`
	TotalInteractions int       `bun:"total_interactions,default:0" json:"total_interactions"`
	LastInteraction   time.Time `bun:"last_interaction,notnull,default:now()" json:"last_interaction"`
	CreatedAt         time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt         time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	
	// Relations
	User      *UserDB      `bun:"rel:belongs-to,join:user_id=id"`
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// EmotionHistoryDB 情感歷史資料庫模型
type EmotionHistoryDB struct {
	bun.BaseModel `bun:"table:emotion_history,alias:eh"`
	
	ID                string    `bun:"id,pk" json:"id"`
	UserID            string    `bun:"user_id,notnull" json:"user_id"`
	CharacterID       string    `bun:"character_id,notnull" json:"character_id"`
	OldAffection      int       `bun:"old_affection,notnull" json:"old_affection"`
	NewAffection      int       `bun:"new_affection,notnull" json:"new_affection"`
	AffectionChange   int       `bun:"affection_change,notnull" json:"affection_change"`
	OldMood           string    `bun:"old_mood,notnull" json:"old_mood"`
	NewMood           string    `bun:"new_mood,notnull" json:"new_mood"`
	TriggerType       string    `bun:"trigger_type,notnull" json:"trigger_type"`
	TriggerContent    *string   `bun:"trigger_content" json:"trigger_content"`
	Context           map[string]interface{} `bun:"context,type:jsonb,default:'{}'" json:"context"`
	CreatedAt         time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	
	// Relations
	User      *UserDB      `bun:"rel:belongs-to,join:user_id=id"`
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

// EmotionMilestoneDB 情感里程碑資料庫模型
type EmotionMilestoneDB struct {
	bun.BaseModel `bun:"table:emotion_milestones,alias:em"`
	
	ID               string    `bun:"id,pk" json:"id"`
	UserID           string    `bun:"user_id,notnull" json:"user_id"`
	CharacterID      string    `bun:"character_id,notnull" json:"character_id"`
	MilestoneType    string    `bun:"milestone_type,notnull" json:"milestone_type"`
	Description      string    `bun:"description,notnull" json:"description"`
	AffectionLevel   int       `bun:"affection_level,notnull" json:"affection_level"`
	AchievedAt       time.Time `bun:"achieved_at,notnull,default:now()" json:"achieved_at"`
	
	// Relations
	User      *UserDB      `bun:"rel:belongs-to,join:user_id=id"`
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
}

