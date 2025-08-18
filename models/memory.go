package models

import (
	"time"

	"github.com/uptrace/bun"
)

// LongTermMemoryModel 長期記憶主表
type LongTermMemoryModel struct {
	bun.BaseModel `bun:"table:long_term_memories"`

	ID          string    `bun:"id,pk" json:"id"`
	UserID      string    `bun:"user_id,notnull" json:"user_id"`
	CharacterID string    `bun:"character_id,notnull" json:"character_id"`
	LastUpdated time.Time `bun:"last_updated,notnull" json:"last_updated"`
	CreatedAt   time.Time `bun:"created_at,notnull" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull" json:"updated_at"`

	// 關聯
	Preferences  []MemoryPreference  `bun:"rel:has-many,join:id=memory_id" json:"preferences,omitempty"`
	Nicknames    []MemoryNickname    `bun:"rel:has-many,join:id=memory_id" json:"nicknames,omitempty"`
	Milestones   []MemoryMilestone   `bun:"rel:has-many,join:id=memory_id" json:"milestones,omitempty"`
	Dislikes     []MemoryDislike     `bun:"rel:has-many,join:id=memory_id" json:"dislikes,omitempty"`
	PersonalInfo []MemoryPersonalInfo `bun:"rel:has-many,join:id=memory_id" json:"personal_info,omitempty"`
}

// MemoryPreference 偏好記憶
type MemoryPreference struct {
	bun.BaseModel `bun:"table:memory_preferences"`

	ID         string    `bun:"id,pk" json:"id"`
	MemoryID   string    `bun:"memory_id,notnull" json:"memory_id"`
	Content    string    `bun:"content,notnull" json:"content"`
	Category   string    `bun:"category,notnull" json:"category"`
	Importance int       `bun:"importance,notnull" json:"importance"`
	CreatedAt  time.Time `bun:"created_at,notnull" json:"created_at"`
}

// MemoryNickname 稱呼記憶
type MemoryNickname struct {
	bun.BaseModel `bun:"table:memory_nicknames"`

	ID        string    `bun:"id,pk" json:"id"`
	MemoryID  string    `bun:"memory_id,notnull" json:"memory_id"`
	Nickname  string    `bun:"nickname,notnull" json:"nickname"`
	Frequency int       `bun:"frequency,notnull" json:"frequency"`
	LastUsed  time.Time `bun:"last_used,notnull" json:"last_used"`
	CreatedAt time.Time `bun:"created_at,notnull" json:"created_at"`
}

// MemoryMilestone 里程碑記憶
type MemoryMilestone struct {
	bun.BaseModel `bun:"table:memory_milestones"`

	ID          string    `bun:"id,pk" json:"id"`
	MemoryID    string    `bun:"memory_id,notnull" json:"memory_id"`
	Type        string    `bun:"type,notnull" json:"type"`
	Description string    `bun:"description,notnull" json:"description"`
	Affection   int       `bun:"affection,notnull" json:"affection"`
	Date        time.Time `bun:"date,notnull" json:"date"`
	CreatedAt   time.Time `bun:"created_at,notnull" json:"created_at"`
}

// MemoryDislike 禁忌記憶
type MemoryDislike struct {
	bun.BaseModel `bun:"table:memory_dislikes"`

	ID         string    `bun:"id,pk" json:"id"`
	MemoryID   string    `bun:"memory_id,notnull" json:"memory_id"`
	Topic      string    `bun:"topic,notnull" json:"topic"`
	Severity   int       `bun:"severity,notnull" json:"severity"`
	Evidence   string    `bun:"evidence" json:"evidence"`
	RecordedAt time.Time `bun:"recorded_at,notnull" json:"recorded_at"`
	CreatedAt  time.Time `bun:"created_at,notnull" json:"created_at"`
}

// MemoryPersonalInfo 個人信息記憶
type MemoryPersonalInfo struct {
	bun.BaseModel `bun:"table:memory_personal_info"`

	ID        string    `bun:"id,pk" json:"id"`
	MemoryID  string    `bun:"memory_id,notnull" json:"memory_id"`
	InfoType  string    `bun:"info_type,notnull" json:"info_type"`
	Content   string    `bun:"content,notnull" json:"content"`
	CreatedAt time.Time `bun:"created_at,notnull" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull" json:"updated_at"`
}