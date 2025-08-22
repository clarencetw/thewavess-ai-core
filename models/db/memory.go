package db

import (
	"time"
	"github.com/uptrace/bun"
)

// LongTermMemoryModelDB 長期記憶資料庫模型
type LongTermMemoryModelDB struct {
	bun.BaseModel `bun:"table:long_term_memories,alias:ltm"`

	ID          string    `bun:"id,pk" json:"id"`
	UserID      string    `bun:"user_id,notnull" json:"user_id"`
	CharacterID string    `bun:"character_id,notnull" json:"character_id"`
	LastUpdated time.Time `bun:"last_updated,notnull,default:now()" json:"last_updated"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	
	// Relations
	User            *UserDB               `bun:"rel:belongs-to,join:user_id=id"`
	Character       *CharacterDB          `bun:"rel:belongs-to,join:character_id=id"`
	Preferences     []MemoryPreferenceDB  `bun:"rel:has-many,join:id=memory_id"`
	Nicknames       []MemoryNicknameDB    `bun:"rel:has-many,join:id=memory_id"`
	Milestones      []MemoryMilestoneDB   `bun:"rel:has-many,join:id=memory_id"`
	Dislikes        []MemoryDislikeDB     `bun:"rel:has-many,join:id=memory_id"`
	PersonalInfo    []MemoryPersonalInfoDB `bun:"rel:has-many,join:id=memory_id"`
}

// MemoryPreferenceDB 記憶偏好資料庫模型
type MemoryPreferenceDB struct {
	bun.BaseModel `bun:"table:memory_preferences,alias:mp"`

	ID         string    `bun:"id,pk" json:"id"`
	MemoryID   string    `bun:"memory_id,notnull" json:"memory_id"`
	Content    string    `bun:"content,notnull" json:"content"`
	Category   string    `bun:"category,notnull" json:"category"`
	Importance int       `bun:"importance,notnull,default:5" json:"importance"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	
	// Relations
	Memory *LongTermMemoryModelDB `bun:"rel:belongs-to,join:memory_id=id"`
}

// MemoryNicknameDB 記憶暱稱資料庫模型
type MemoryNicknameDB struct {
	bun.BaseModel `bun:"table:memory_nicknames,alias:mn"`

	ID         string    `bun:"id,pk" json:"id"`
	MemoryID   string    `bun:"memory_id,notnull" json:"memory_id"`
	Nickname   string    `bun:"nickname,notnull" json:"nickname"`
	Frequency  int       `bun:"frequency,notnull,default:1" json:"frequency"`
	LastUsed   time.Time `bun:"last_used,notnull,default:now()" json:"last_used"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	
	// Relations
	Memory *LongTermMemoryModelDB `bun:"rel:belongs-to,join:memory_id=id"`
}

// MemoryMilestoneDB 記憶里程碑資料庫模型
type MemoryMilestoneDB struct {
	bun.BaseModel `bun:"table:memory_milestones,alias:mm"`

	ID          string    `bun:"id,pk" json:"id"`
	MemoryID    string    `bun:"memory_id,notnull" json:"memory_id"`
	Type        string    `bun:"type,notnull" json:"type"`
	Description string    `bun:"description,notnull" json:"description"`
	Affection   int       `bun:"affection,notnull" json:"affection"`
	Date        time.Time `bun:"date,notnull" json:"date"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	
	// Relations
	Memory *LongTermMemoryModelDB `bun:"rel:belongs-to,join:memory_id=id"`
}

// MemoryDislikeDB 記憶厭惡資料庫模型
type MemoryDislikeDB struct {
	bun.BaseModel `bun:"table:memory_dislikes,alias:md"`

	ID          string    `bun:"id,pk" json:"id"`
	MemoryID    string    `bun:"memory_id,notnull" json:"memory_id"`
	Topic       string    `bun:"topic,notnull" json:"topic"`
	Severity    int       `bun:"severity,notnull,default:3" json:"severity"`
	Evidence    *string   `bun:"evidence" json:"evidence"`
	RecordedAt  time.Time `bun:"recorded_at,notnull" json:"recorded_at"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	
	// Relations
	Memory *LongTermMemoryModelDB `bun:"rel:belongs-to,join:memory_id=id"`
}

// MemoryPersonalInfoDB 記憶個人信息資料庫模型
type MemoryPersonalInfoDB struct {
	bun.BaseModel `bun:"table:memory_personal_info,alias:mpi"`

	ID        string    `bun:"id,pk" json:"id"`
	MemoryID  string    `bun:"memory_id,notnull" json:"memory_id"`
	InfoType  string    `bun:"info_type,notnull" json:"info_type"`
	Content   string    `bun:"content,notnull" json:"content"`
	CreatedAt time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	
	// Relations
	Memory *LongTermMemoryModelDB `bun:"rel:belongs-to,join:memory_id=id"`
}

