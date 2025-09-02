package models

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/google/uuid"
)

// Admin 管理員領域模型
type Admin struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	DisplayName *string    `json:"display_name,omitempty"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	Permissions []string   `json:"permissions"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Notes       *string    `json:"notes,omitempty"`
}

// AdminCreateRequest 創建管理員請求
type AdminCreateRequest struct {
	Username    string   `json:"username" binding:"required,min=3,max=50"`
	Email       string   `json:"email" binding:"required,email"`
	Password    string   `json:"password" binding:"required,min=8"`
	DisplayName *string  `json:"display_name,omitempty"`
	Role        string   `json:"role" binding:"required,oneof=super_admin admin"`
	Permissions []string `json:"permissions,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	CreatedBy   string   `json:"-"` // 不在JSON中顯示，由後端設定
}

// AdminUpdateRequest 更新管理員請求
type AdminUpdateRequest struct {
	DisplayName *string  `json:"display_name,omitempty"`
	Role        *string  `json:"role,omitempty" binding:"omitempty,oneof=super_admin admin"`
	Status      *string  `json:"status,omitempty" binding:"omitempty,oneof=active inactive suspended"`
	Permissions []string `json:"permissions,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
}

// AdminLoginRequest 登入請求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse 登入響應
type AdminLoginResponse struct {
	Admin       Admin  `json:"admin"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// AdminListQuery 列表查詢參數
type AdminListQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
	Role     string `form:"role" binding:"omitempty,oneof=super_admin content_admin system_admin data_admin"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive suspended"`
	Search   string `form:"search"`
}

// AdminListResponse 列表響應
type AdminListResponse struct {
	Admins     []*Admin           `json:"admins"`
	Pagination PaginationResponse `json:"pagination"`
}

// AdminFromDB 資料庫模型轉換
func AdminFromDB(adminDB *db.AdminDB) *Admin {
	if adminDB == nil {
		return nil
	}

	return &Admin{
		ID:          adminDB.ID,
		Username:    adminDB.Username,
		Email:       adminDB.Email,
		DisplayName: adminDB.DisplayName,
		Role:        string(adminDB.Role),
		Status:      string(adminDB.Status),
		Permissions: adminDB.Permissions,
		LastLoginAt: adminDB.LastLoginAt,
		CreatedAt:   adminDB.CreatedAt,
		UpdatedAt:   adminDB.UpdatedAt,
		Notes:       adminDB.Notes,
	}
}

// ToDB 轉換資料庫模型
func (a *Admin) ToDB() *db.AdminDB {
	return &db.AdminDB{
		ID:          a.ID,
		Username:    a.Username,
		Email:       a.Email,
		DisplayName: a.DisplayName,
		Role:        db.AdminRole(a.Role),
		Status:      db.AdminStatus(a.Status),
		Permissions: a.Permissions,
		LastLoginAt: a.LastLoginAt,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		Notes:       a.Notes,
	}
}

// Validate 驗證數據
func (req *AdminCreateRequest) Validate() error {
	// 用戶名格式檢查
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return fmt.Errorf("用戶名長度必須在3-50字符之間")
	}

	// 密碼強度檢查
	if len(req.Password) < 8 {
		return fmt.Errorf("密碼長度至少8個字符")
	}

	// 角色有效性檢查
	validRoles := map[string]bool{
		"super_admin": true,
		"admin":       true,
	}
	if !validRoles[req.Role] {
		return fmt.Errorf("無效的管理員角色: %s", req.Role)
	}

	return nil
}

// HashPassword 密碼加密
func (req *AdminCreateRequest) HashPassword() (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密碼加密失敗: %w", err)
	}
	return string(hashedPassword), nil
}

// ToAdmin 轉換管理員模型
func (req *AdminCreateRequest) ToAdmin() (*Admin, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 生成管理員ID
	adminID := "admin_" + uuid.New().String()

	admin := &Admin{
		ID:          adminID,
		Username:    req.Username,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Role:        req.Role,
		Status:      "active",
		Permissions: req.Permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Notes:       req.Notes,
	}

	return admin, nil
}

// ValidateAdminPassword 驗證密碼
func ValidateAdminPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GetRolePermissions 獲取角色權限
func GetRolePermissions(role string) []string {
	switch role {
	case "super_admin":
		return []string{"*"}
	case "admin":
		return []string{"basic"}
	default:
		return []string{}
	}
}

// 角色管理模型

// AdminCharacterUpdateRequest 更新角色請求
type AdminCharacterUpdateRequest struct {
	Name            string   `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Type            string   `json:"type,omitempty" binding:"omitempty,oneof=dominant gentle playful mystery reliable"`
	Locale          string   `json:"locale,omitempty" binding:"omitempty,oneof=zh-TW en-US ja-JP"`
	IsActive        *bool    `json:"is_active,omitempty"`
	AvatarURL       *string  `json:"avatar_url,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Popularity      *int     `json:"popularity,omitempty" binding:"omitempty,min=0,max=100"`
	UserDescription *string  `json:"user_description,omitempty"`
	IsPublic        *bool    `json:"is_public,omitempty"`
}

// AdminUserUpdateRequest 更新用戶請求
type AdminUserUpdateRequest struct {
	Username    string     `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Email       string     `json:"email,omitempty" binding:"omitempty,email"`
	DisplayName *string    `json:"display_name,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	Status      string     `json:"status,omitempty" binding:"omitempty,oneof=active inactive suspended"`
	Nickname    string     `json:"nickname,omitempty"`
	Gender      string     `json:"gender,omitempty" binding:"omitempty,oneof=male female other"`
	BirthDate   *time.Time `json:"birth_date,omitempty"`
	AvatarURL   string     `json:"avatar_url,omitempty"`
	IsVerified  *bool      `json:"is_verified,omitempty"`
	IsAdult     *bool      `json:"is_adult,omitempty"`
}

// AdminPasswordUpdateRequest 重置密碼請求
type AdminPasswordUpdateRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
