package models

import (
	"time"
)

// Tag 標籤領域模型
type Tag struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Color       string    `json:"color,omitempty"`
	Description string    `json:"description,omitempty"`
	UsageCount  int       `json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// CharacterTag 角色標籤關聯
type CharacterTag struct {
	ID          string    `json:"id"`
	CharacterID string    `json:"character_id"`
	TagID       string    `json:"tag_id"`
	CreatedAt   time.Time `json:"created_at"`
	
	// 關聯數據
	Character *Character `json:"character,omitempty"`
	Tag       *Tag       `json:"tag,omitempty"`
}

// TagWithStats 帶統計信息的標籤
type TagWithStats struct {
	Tag
	Trend           string  `json:"trend"`            // up, down, stable
	TrendPercentage float64 `json:"trend_percentage"`
}

// TagCategory 標籤分類
type TagCategory struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Count       int    `json:"count"`
}

// TagsResponse API 響應格式
type TagsResponse struct {
	Tags       []Tag         `json:"tags"`
	Categories []TagCategory `json:"categories"`
	TotalCount int           `json:"total_count"`
}

// PopularTagsResponse 熱門標籤響應格式
type PopularTagsResponse struct {
	Tags      []TagWithStats `json:"tags"`
	Period    string         `json:"period"`
	UpdatedAt time.Time      `json:"updated_at"`
}