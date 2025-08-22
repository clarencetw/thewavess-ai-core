package db

import (
	"context"
	"time"
	
	"github.com/uptrace/bun"
)

// ChatSessionDB 聊天會話資料庫模型
type ChatSessionDB struct {
	bun.BaseModel `bun:"table:chat_sessions,alias:cs"`

	ID              string     `bun:"id,pk" json:"id"`
	UserID          string     `bun:"user_id,notnull" json:"user_id"`
	CharacterID     string     `bun:"character_id,notnull" json:"character_id"`
	Title           string     `bun:"title" json:"title"`
	Status          string     `bun:"status,default:'active'" json:"status"`
	MessageCount    int        `bun:"message_count,default:0" json:"message_count"`
	TotalCharacters int        `bun:"total_characters,default:0" json:"total_characters"`
	LastMessageAt   *time.Time `bun:"last_message_at" json:"last_message_at"`
	CreatedAt       time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	
	// Relations
	User      *UserDB      `bun:"rel:belongs-to,join:user_id=id"`
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
	Messages  []MessageDB  `bun:"rel:has-many,join:id=session_id"`
}

// MessageDB 消息資料庫模型
type MessageDB struct {
	bun.BaseModel `bun:"table:messages,alias:m"`

	ID                  string                 `bun:"id,pk" json:"id"`
	SessionID           string                 `bun:"session_id,notnull" json:"session_id"`
	Role                string                 `bun:"role,notnull" json:"role"`
	Content             string                 `bun:"content,notnull" json:"content"`
	SceneDescription    *string                `bun:"scene_description" json:"scene_description"`
	CharacterAction     *string                `bun:"character_action" json:"character_action"`
	EmotionalState      map[string]interface{} `bun:"emotional_state,type:jsonb,default:'{}'" json:"emotional_state"`
	AIEngine            *string                `bun:"ai_engine" json:"ai_engine"`
	ResponseTimeMs      *int                   `bun:"response_time_ms" json:"response_time_ms"`
	NSFWLevel           int                    `bun:"nsfw_level,default:0" json:"nsfw_level"`
	IsRegenerated       bool                   `bun:"is_regenerated,default:false" json:"is_regenerated"`
	RegenerationReason  *string                `bun:"regeneration_reason" json:"regeneration_reason"`
	CreatedAt           time.Time              `bun:"created_at,notnull,default:now()" json:"created_at"`
	
	// Relations
	Session *ChatSessionDB `bun:"rel:belongs-to,join:session_id=id"`
}


// BeforeAppendModel 在模型操作前執行 (使用 Hook 的最佳實踐)
var _ bun.BeforeAppendModelHook = (*ChatSessionDB)(nil)

func (cs *ChatSessionDB) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		now := time.Now()
		cs.CreatedAt = now
		cs.UpdatedAt = now
	case *bun.UpdateQuery:
		cs.UpdatedAt = time.Now()
	}
	return nil
}