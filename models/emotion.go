package models

import (
	"time"
	"github.com/uptrace/bun"
)

// EmotionState 情感狀態表
type EmotionState struct {
	bun.BaseModel `bun:"table:emotion_states,alias:es"`
	
	ID            string    `bun:"id,pk" json:"id"`
	UserID        string    `bun:"user_id,notnull" json:"user_id"`
	CharacterID   string    `bun:"character_id,notnull" json:"character_id"`
	Affection     int       `bun:"affection,notnull,default:30" json:"affection"`
	Mood          string    `bun:"mood,notnull,default:'neutral'" json:"mood"`
	Relationship  string    `bun:"relationship,notnull,default:'stranger'" json:"relationship"`
	IntimacyLevel string    `bun:"intimacy_level,notnull,default:'distant'" json:"intimacy_level"`
	TotalInteractions int   `bun:"total_interactions,default:0" json:"total_interactions"`
	LastInteraction   time.Time `bun:"last_interaction,default:now()" json:"last_interaction"`
	CreatedAt     time.Time `bun:"created_at,nullzero,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at,nullzero,default:current_timestamp" json:"updated_at"`
	
	// Relations
	User      *User      `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Character *Character `bun:"rel:belongs-to,join:character_id=id" json:"character,omitempty"`
}

// EmotionHistory 情感變化歷史表
type EmotionHistory struct {
	bun.BaseModel `bun:"table:emotion_history,alias:eh"`
	
	ID              string                 `bun:"id,pk" json:"id"`
	UserID          string                 `bun:"user_id,notnull" json:"user_id"`
	CharacterID     string                 `bun:"character_id,notnull" json:"character_id"`
	OldAffection    int                    `bun:"old_affection,notnull" json:"old_affection"`
	NewAffection    int                    `bun:"new_affection,notnull" json:"new_affection"`
	AffectionChange int                    `bun:"affection_change,notnull" json:"affection_change"`
	OldMood         string                 `bun:"old_mood,notnull" json:"old_mood"`
	NewMood         string                 `bun:"new_mood,notnull" json:"new_mood"`
	TriggerType     string                 `bun:"trigger_type,notnull" json:"trigger_type"`
	TriggerContent  string                 `bun:"trigger_content" json:"trigger_content"`
    Context         map[string]interface{} `bun:"context,type:jsonb" json:"context"` // 建議存放本次變動的解釋(explanations)、命中規則、NSFW等級、訊息長度等debug資訊
	CreatedAt       time.Time              `bun:"created_at,nullzero,default:current_timestamp" json:"created_at"`
	
	// Relations
	User      *User      `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Character *Character `bun:"rel:belongs-to,join:character_id=id" json:"character,omitempty"`
}

// EmotionMilestone 情感里程碑表
type EmotionMilestone struct {
	bun.BaseModel `bun:"table:emotion_milestones,alias:em"`
	
	ID             string    `bun:"id,pk" json:"id"`
	UserID         string    `bun:"user_id,notnull" json:"user_id"`
	CharacterID    string    `bun:"character_id,notnull" json:"character_id"`
	MilestoneType  string    `bun:"milestone_type,notnull" json:"milestone_type"`
	Description    string    `bun:"description,notnull" json:"description"`
	AffectionLevel int       `bun:"affection_level,notnull" json:"affection_level"`
	AchievedAt     time.Time `bun:"achieved_at,nullzero,default:current_timestamp" json:"achieved_at"`
	
	// Relations
	User      *User      `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Character *Character `bun:"rel:belongs-to,join:character_id=id" json:"character,omitempty"`
}
