package db

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// RelationshipDB 用戶與角色關係資料庫模型（重命名自 EmotionStateDB）
type RelationshipDB struct {
	bun.BaseModel `bun:"table:relationships,alias:r"`

	ID          string  `bun:"id,pk" json:"id"`
	UserID      string  `bun:"user_id,notnull" json:"user_id"`
	CharacterID string  `bun:"character_id,notnull" json:"character_id"`
	ChatID      *string `bun:"chat_id" json:"chat_id"` // 重要：多會話架構必需字段，每個對話獨立關係狀態

	// 核心關係數據
	Affection         int       `bun:"affection,notnull,default:30" json:"affection"`
	Mood              string    `bun:"mood,notnull,default:'neutral'" json:"mood"`
	Relationship      string    `bun:"relationship,notnull,default:'stranger'" json:"relationship"`
	IntimacyLevel     string    `bun:"intimacy_level,notnull,default:'distant'" json:"intimacy_level"`
	TotalInteractions int       `bun:"total_interactions,default:0" json:"total_interactions"`
	LastInteraction   time.Time `bun:"last_interaction,notnull,default:now()" json:"last_interaction"`

	// JSONB 儲存情感歷史記錄
	EmotionData map[string]interface{} `bun:"emotion_data,type:jsonb,default:'{}'" json:"emotion_data"` // 儲存情感變化歷史

	// 時間戳
	CreatedAt time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`

	// Relations
	User      *UserDB      `bun:"rel:belongs-to,join:user_id=id"`
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
	Chat      *ChatDB      `bun:"rel:belongs-to,join:chat_id=id"`
}

// BeforeAppendModel 在模型操作前執行（Bun ORM Hook）
var _ bun.BeforeAppendModelHook = (*RelationshipDB)(nil)

func (r *RelationshipDB) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		now := time.Now()
		r.CreatedAt = now
		r.UpdatedAt = now
		r.LastInteraction = now
	case *bun.UpdateQuery:
		r.UpdatedAt = time.Now()
		r.LastInteraction = time.Now()
	}
	return nil
}

// AddEmotionHistory 添加情感變化歷史到 emotion_data
func (r *RelationshipDB) AddEmotionHistory(triggerType, triggerContent string, oldAffection, newAffection int, oldMood, newMood string) {
	if r.EmotionData == nil {
		r.EmotionData = make(map[string]interface{})
	}

	historyEntry := map[string]interface{}{
		"timestamp":        time.Now(),
		"trigger_type":     triggerType,
		"trigger_content":  triggerContent,
		"old_affection":    oldAffection,
		"new_affection":    newAffection,
		"affection_change": newAffection - oldAffection,
		"old_mood":         oldMood,
		"new_mood":         newMood,
	}

	// 保持最近的50條歷史記錄
	if histories, ok := r.EmotionData["history"].([]interface{}); ok {
		if len(histories) >= 50 {
			histories = histories[1:] // 移除最舊的記錄
		}
		histories = append(histories, historyEntry)
		r.EmotionData["history"] = histories
	} else {
		r.EmotionData["history"] = []interface{}{historyEntry}
	}
}
