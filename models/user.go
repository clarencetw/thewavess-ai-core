package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// User 用戶模型
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           string                 `bun:"id,pk" json:"id"`
	Username     string                 `bun:"username,unique,notnull" json:"username"`
	Email        string                 `bun:"email,unique,notnull" json:"email"`
	Password     string                 `bun:"password_hash,notnull" json:"-"`
	DisplayName  *string                `bun:"display_name" json:"display_name,omitempty"`
	Bio          *string                `bun:"bio" json:"bio,omitempty"`
	Status       string                 `bun:"status,default:'active'" json:"status"`
	Nickname     string                 `bun:"nickname" json:"nickname,omitempty"`
	Gender       string                 `bun:"gender" json:"gender,omitempty"`
	BirthDate    *time.Time             `bun:"birth_date" json:"birth_date,omitempty"`
	AvatarURL    string                 `bun:"avatar_url" json:"avatar_url,omitempty"`
	IsVerified   bool                   `bun:"is_verified,default:false" json:"is_verified"`
	IsAdult      bool                   `bun:"is_adult,default:false" json:"is_adult"`
	Preferences  map[string]interface{} `bun:"preferences,type:jsonb" json:"preferences,omitempty"`
	CreatedAt    time.Time              `bun:"created_at,nullzero,default:now()" json:"created_at"`
	UpdatedAt    time.Time              `bun:"updated_at,nullzero,default:now()" json:"updated_at"`
	LastLoginAt  *time.Time             `bun:"last_login_at" json:"last_login_at,omitempty"`

	// 關聯
	Sessions []*ChatSession `bun:"rel:has-many,join:id=user_id" json:"sessions,omitempty"`
}

// TableName 返回數據庫表名
func (u *User) TableName() string {
	return "users"
}

// BeforeAppendModel 在模型操作前執行
func (u *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		u.UpdatedAt = time.Now()
	}
	return nil
}

// UserCreateRequest 用戶創建請求
type UserCreateRequest struct {
	Username  string    `json:"username" binding:"required,min=3,max=50"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password" binding:"required,min=8"`
	Nickname  string    `json:"nickname,omitempty"`
	Gender    string    `json:"gender,omitempty" binding:"omitempty,oneof=male female other"`
	BirthDate time.Time `json:"birth_date" binding:"required"`
}

// UserUpdateRequest 用戶更新請求
type UserUpdateRequest struct {
	Nickname    string                 `json:"nickname,omitempty"`
	Gender      string                 `json:"gender,omitempty" binding:"omitempty,oneof=male female other"`
	BirthDate   *time.Time             `json:"birth_date,omitempty"`
	AvatarURL   string                 `json:"avatar_url,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// UserResponse 用戶響應（隱藏敏感信息）
type UserResponse struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	DisplayName *string                `json:"display_name,omitempty"`
	Bio         *string                `json:"bio,omitempty"`
	Status      string                 `json:"status"`
	Nickname    string                 `json:"nickname,omitempty"`
	Gender      string                 `json:"gender,omitempty"`
	BirthDate   *time.Time             `json:"birth_date,omitempty"`
	AvatarURL   string                 `json:"avatar_url,omitempty"`
	IsVerified  bool                   `json:"is_verified"`
	IsAdult     bool                   `json:"is_adult"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastLoginAt *time.Time             `json:"last_login_at,omitempty"`
}

// 用戶相關請求結構
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用戶名
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type UpdatePreferencesRequest struct {
	Preferences map[string]interface{} `json:"preferences" binding:"required"`
}


type LoginResponse struct {
	Token        string        `json:"token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int           `json:"expires_in"`
	User         *UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type LogoutRequest struct {
	Token string `json:"token,omitempty"`
}

// ToResponse 轉換為響應格式
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Bio:         u.Bio,
		Status:      u.Status,
		AvatarURL:   u.AvatarURL,
		IsVerified:  u.IsVerified,
		IsAdult:     u.IsAdult,
		Preferences: u.Preferences,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}

// ToLegacyResponse 轉換為舊版響應格式（向後兼容）
func (u *User) ToLegacyResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Bio:         u.Bio,
		Status:      u.Status,
		Nickname:    u.Nickname,
		Gender:      u.Gender,
		BirthDate:   u.BirthDate,
		AvatarURL:   u.AvatarURL,
		IsVerified:  u.IsVerified,
		IsAdult:     u.IsAdult,
		Preferences: u.Preferences,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}