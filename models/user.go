package models

import "time"

// User 用戶模型
type User struct {
	BaseModel
	Username    string    `json:"username" gorm:"uniqueIndex" example:"alice123"`
	Email       string    `json:"email" gorm:"uniqueIndex" example:"alice@example.com"`
	Nickname    string    `json:"nickname" example:"小愛"`
	BirthDate   time.Time `json:"birth_date" example:"2000-01-01T00:00:00Z"`
	Gender      string    `json:"gender" example:"female" enums:"female,male,other"`
	AvatarURL   string    `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	IsActive    bool      `json:"is_active" example:"true"`
	LastLoginAt time.Time `json:"last_login_at,omitempty" example:"2023-01-01T00:00:00Z"`
}

// UserRegisterRequest 用戶註冊請求
type UserRegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=20" example:"alice123"`
	Email     string `json:"email" binding:"required,email" example:"alice@example.com"`
	Password  string `json:"password" binding:"required,min=8" example:"password123"`
	BirthDate string `json:"birth_date" binding:"required" example:"2000-01-01"`
	Gender    string `json:"gender" binding:"required" example:"female" enums:"female,male,other"`
	Nickname  string `json:"nickname" binding:"required,min=1,max=50" example:"小愛"`
}

// UserLoginRequest 用戶登入請求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required" example:"alice123"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// AuthResponse 認證回應
type AuthResponse struct {
	UserID       string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int    `json:"expires_in" example:"3600"`
}

// UserProfile 用戶個人資料
type UserProfile struct {
	BaseModel
	Username     string                 `json:"username" example:"alice123"`
	Email        string                 `json:"email" example:"alice@example.com"`
	Nickname     string                 `json:"nickname" example:"小愛"`
	BirthDate    time.Time              `json:"birth_date" example:"2000-01-01T00:00:00Z"`
	Gender       string                 `json:"gender" example:"female"`
	AvatarURL    string                 `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	Preferences  map[string]interface{} `json:"preferences,omitempty"`
	CharacterID  string                 `json:"character_id,omitempty" example:"char_001"`
	TotalChats   int                    `json:"total_chats" example:"150"`
	JoinedAt     time.Time              `json:"joined_at" example:"2023-01-01T00:00:00Z"`
	LastActiveAt time.Time              `json:"last_active_at" example:"2023-12-01T12:00:00Z"`
}

// UpdateProfileRequest 更新個人資料請求
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname,omitempty" example:"新暱稱"`
	AvatarURL string `json:"avatar_url,omitempty" example:"https://example.com/new-avatar.jpg"`
}

// UpdatePreferencesRequest 更新偏好設定請求
type UpdatePreferencesRequest struct {
	Preferences map[string]interface{} `json:"preferences"`
}