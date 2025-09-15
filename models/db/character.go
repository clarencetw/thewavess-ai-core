package db

import (
	"time"

	"github.com/uptrace/bun"
)

// CharacterDB 角色核心表 (整合了多個原表的功能)
type CharacterDB struct {
	bun.BaseModel `bun:"table:characters,alias:c"`

	// 基本資訊
	ID         string   `bun:"id,pk" json:"id"`
	Name       string   `bun:"name,notnull" json:"name"`
	Type       string   `bun:"type,notnull" json:"type"`
	Locale     string   `bun:"locale,notnull,default:'zh-TW'" json:"locale"`
	IsActive   bool     `bun:"is_active,default:true" json:"is_active"`
	AvatarURL  *string  `bun:"avatar_url" json:"avatar_url"`
	Popularity int      `bun:"popularity,default:0" json:"popularity"`
	Tags       []string `bun:"tags,array,default:'{}'" json:"tags"`

	// 用戶描述支援
	UserDescription *string `bun:"user_description" json:"user_description"`

	// 用戶追蹤字段
	CreatedBy *string `bun:"created_by" json:"created_by"`
	UpdatedBy *string `bun:"updated_by" json:"updated_by"`

	// 角色狀態
	IsPublic bool `bun:"is_public,default:true" json:"is_public"`
	IsSystem bool `bun:"is_system,default:false" json:"is_system"`

	// 時間戳 (包含軟刪除)
	CreatedAt time.Time  `bun:"created_at,default:now()" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,default:now()" json:"updated_at"`
	DeletedAt *time.Time `bun:"deleted_at,soft_delete,nullzero" json:"deleted_at,omitempty"`
}
