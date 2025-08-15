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