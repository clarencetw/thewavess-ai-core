package db

import (
	"time"
	"github.com/uptrace/bun"
)

// TagDB 標籤資料庫模型
type TagDB struct {
	bun.BaseModel `bun:"table:tags,alias:t"`

	ID          string    `bun:"id,pk" json:"id"`
	Name        string    `bun:"name,notnull" json:"name"`
	Category    string    `bun:"category,notnull" json:"category"` // genre, personality, role, style
	Color       string    `bun:"color" json:"color"`
	Description string    `bun:"description" json:"description"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// CharacterTagDB 角色標籤關聯資料庫模型
type CharacterTagDB struct {
	bun.BaseModel `bun:"table:character_tags,alias:ct"`

	ID          string    `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	CharacterID string    `bun:"character_id,notnull" json:"character_id"`
	TagID       string    `bun:"tag_id,notnull" json:"tag_id"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`

	// 關聯
	Character *CharacterDB `bun:"rel:belongs-to,join:character_id=id"`
	Tag       *TagDB       `bun:"rel:belongs-to,join:tag_id=id"`
}