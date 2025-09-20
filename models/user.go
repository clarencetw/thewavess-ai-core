package models

import (
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/pkg/gravatar"
)

// User 用戶領域模型
type User struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Password    string     `json:"-"`
	DisplayName *string    `json:"display_name,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	Status      string     `json:"status"`
	Nickname    *string    `json:"nickname,omitempty"`
	Gender      *string    `json:"gender,omitempty"`
	BirthDate   *time.Time `json:"birth_date,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	IsVerified  bool       `json:"is_verified"`
	IsAdult     bool       `json:"is_adult"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// 關聯
	Chats []*Chat `json:"chats,omitempty"`
}



// UserResponse 用戶響應（隱藏敏感信息）
type UserResponse struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	DisplayName *string    `json:"display_name,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	Status      string     `json:"status"`
	Nickname    *string    `json:"nickname,omitempty"`
	Gender      *string    `json:"gender,omitempty"`
	BirthDate   *time.Time `json:"birth_date,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	IsVerified  bool       `json:"is_verified"`
	IsAdult     bool       `json:"is_adult"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// 用戶相關請求結構
type RegisterRequest struct {
	Username    string     `json:"username" binding:"required,min=3,max=50"`
	Email       string     `json:"email" binding:"required,email"`
	Password    string     `json:"password" binding:"required,min=8"`
	DisplayName *string    `json:"display_name,omitempty"`
	Nickname    *string    `json:"nickname,omitempty"`
	Gender      *string    `json:"gender,omitempty" binding:"omitempty,oneof=male female other"`
	BirthDate   *time.Time `json:"birth_date,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用戶名
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	DisplayName *string    `json:"display_name,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	BirthDate   *time.Time `json:"birth_date,omitempty"`
	Gender      *string    `json:"gender,omitempty"`
	Nickname    *string    `json:"nickname,omitempty"`
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

// UserFromDB 從資料庫模型轉換為領域模型
func UserFromDB(userDB *db.UserDB) *User {
	if userDB == nil {
		return nil
	}

	return &User{
		ID:          userDB.ID,
		Username:    userDB.Username,
		Email:       userDB.Email,
		Password:    userDB.Password,
		DisplayName: userDB.DisplayName,
		Bio:         userDB.Bio,
		Status:      userDB.Status,
		Nickname:    userDB.Nickname,
		Gender:      userDB.Gender,
		BirthDate:   userDB.BirthDate,
		AvatarURL:   userDB.AvatarURL,
		IsVerified:  userDB.IsVerified,
		IsAdult:     userDB.IsAdult,
		CreatedAt:   userDB.CreatedAt,
		UpdatedAt:   userDB.UpdatedAt,
		LastLoginAt: userDB.LastLoginAt,
	}
}

// ToResponse 轉換為響應格式
func (u *User) ToResponse() *UserResponse {
	if u == nil {
		return nil
	}

	resp := &UserResponse{
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
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: u.LastLoginAt,
	}

	if (resp.AvatarURL == nil || strings.TrimSpace(*resp.AvatarURL) == "") && strings.TrimSpace(u.Email) != "" {
		defaultURL := gravatar.URL(u.Email, 256)
		resp.AvatarURL = &defaultURL
	}

	return resp
}
