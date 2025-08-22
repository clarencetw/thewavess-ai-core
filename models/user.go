package models

import (
	"time"
)

// User 用戶領域模型
type User struct {
	ID           string                 `json:"id"`
	Username     string                 `json:"username"`
	Email        string                 `json:"email"`
	Password     string                 `json:"-"`
	DisplayName  *string                `json:"display_name,omitempty"`
	Bio          *string                `json:"bio,omitempty"`
	Status       string                 `json:"status"`
	Nickname     *string                `json:"nickname,omitempty"`
	Gender       *string                `json:"gender,omitempty"`
	BirthDate    *time.Time             `json:"birth_date,omitempty"`
	AvatarURL    *string                `json:"avatar_url,omitempty"`
	IsVerified   bool                   `json:"is_verified"`
	IsAdult      bool                   `json:"is_adult"`
	Preferences  map[string]interface{} `json:"preferences,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastLoginAt  *time.Time             `json:"last_login_at,omitempty"`

	// 關聯
	Sessions []*ChatSession `json:"sessions,omitempty"`
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
	Nickname    *string                `json:"nickname,omitempty"`
	Gender      *string                `json:"gender,omitempty"`
	BirthDate   *time.Time             `json:"birth_date,omitempty"`
	AvatarURL   *string                `json:"avatar_url,omitempty"`
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

// Admin 管理相關請求結構
type AdminUserUpdateRequest struct {
	Username    string                 `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Email       string                 `json:"email,omitempty" binding:"omitempty,email"`
	DisplayName *string                `json:"display_name,omitempty"`
	Bio         *string                `json:"bio,omitempty"`
	Status      string                 `json:"status,omitempty" binding:"omitempty,oneof=active inactive banned"`
	Nickname    string                 `json:"nickname,omitempty"`
	Gender      string                 `json:"gender,omitempty" binding:"omitempty,oneof=male female other"`
	BirthDate   *time.Time             `json:"birth_date,omitempty"`
	AvatarURL   string                 `json:"avatar_url,omitempty"`
	IsVerified  *bool                  `json:"is_verified,omitempty"`
	IsAdult     *bool                  `json:"is_adult,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

type AdminPasswordUpdateRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
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


