package models

import "time"

// APIResponse 標準 API 回應格式
type APIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"操作成功"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError 錯誤回應格式
type APIError struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"請求參數無效"`
	Details string `json:"details,omitempty" example:"用戶名稱不能為空"`
}

// PaginationResponse 分頁回應格式
type PaginationResponse struct {
	CurrentPage int `json:"current_page" example:"1"`
	TotalPages  int `json:"total_pages" example:"10"`
	TotalCount  int `json:"total_count" example:"100"`
	HasNext     bool `json:"has_next" example:"true"`
	HasPrev     bool `json:"has_prev" example:"false"`
}

// PaginationRequest 分頁請求參數
type PaginationRequest struct {
	Page  int `json:"page" form:"page" example:"1" default:"1"`
	Limit int `json:"limit" form:"limit" example:"20" default:"20"`
}

// BaseModel API層基礎模型 (Data Transfer Object Layer)
// 用於 JSON 序列化、API 響應等場景，使用 UUID 字符串作為主鍵
type BaseModel struct {
	ID        string    `json:"id" db:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt time.Time `json:"created_at" db:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" example:"2023-01-01T00:00:00Z"`
}