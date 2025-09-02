package db

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// AdminRole 管理員角色枚舉
type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "super_admin" // 超級管理員 (所有權限)
	AdminRoleAdmin      AdminRole = "admin"       // 一般管理員 (管理權限)
)

// AdminStatus 管理員狀態枚舉
type AdminStatus string

const (
	AdminStatusActive    AdminStatus = "active"    // 啟用
	AdminStatusInactive  AdminStatus = "inactive"  // 停用
	AdminStatusSuspended AdminStatus = "suspended" // 暫停
)

// AdminDB 管理員資料庫模型
type AdminDB struct {
	bun.BaseModel `bun:"table:admins,alias:a"`

	// 基本資訊
	ID          string  `bun:"id,pk" json:"id"`
	Username    string  `bun:"username,unique,notnull" json:"username"`
	Email       string  `bun:"email,unique,notnull" json:"email"`
	Password    string  `bun:"password_hash,notnull" json:"-"`
	DisplayName *string `bun:"display_name" json:"display_name"`

	// 權限與狀態
	Role        AdminRole   `bun:"role,notnull,default:'admin'" json:"role"`
	Status      AdminStatus `bun:"status,notnull,default:'active'" json:"status"`
	Permissions []string    `bun:"permissions,array,default:'{}'" json:"permissions"`

	// 安全相關
	LastLoginAt    *time.Time `bun:"last_login_at" json:"last_login_at"`
	LastLoginIP    *string    `bun:"last_login_ip" json:"last_login_ip"`
	FailedAttempts int        `bun:"failed_attempts,default:0" json:"failed_attempts"`
	LockedUntil    *time.Time `bun:"locked_until" json:"locked_until"`

	// 時間戳
	CreatedAt time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	CreatedBy *string   `bun:"created_by" json:"created_by"` // 創建者ID
	Notes     *string   `bun:"notes" json:"notes"`           // 管理員備注
}

// BeforeAppendModel 在模型操作前執行
var _ bun.BeforeAppendModelHook = (*AdminDB)(nil)

func (a *AdminDB) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		now := time.Now()
		a.CreatedAt = now
		a.UpdatedAt = now
	case *bun.UpdateQuery:
		a.UpdatedAt = time.Now()
	}
	return nil
}

// IsActive 檢查管理員是否處於活躍狀態
func (a *AdminDB) IsActive() bool {
	if a.Status != AdminStatusActive {
		return false
	}

	// 檢查是否被鎖定
	if a.LockedUntil != nil && time.Now().Before(*a.LockedUntil) {
		return false
	}

	return true
}

// HasPermission 檢查管理員是否擁有特定權限
func (a *AdminDB) HasPermission(permission string) bool {
	// 超級管理員擁有所有權限
	if a.Role == AdminRoleSuperAdmin {
		return true
	}

	// 檢查具體權限
	for _, perm := range a.Permissions {
		if perm == permission {
			return true
		}
	}

	return false
}

// IncrementFailedAttempts 增加失敗嘗試次數
func (a *AdminDB) IncrementFailedAttempts() {
	a.FailedAttempts++

	// 如果失敗次數達到5次，鎖定帳號30分鐘
	if a.FailedAttempts >= 5 {
		lockUntil := time.Now().Add(30 * time.Minute)
		a.LockedUntil = &lockUntil
		a.Status = AdminStatusSuspended
	}
}

// ResetFailedAttempts 重置失敗嘗試次數
func (a *AdminDB) ResetFailedAttempts() {
	a.FailedAttempts = 0
	a.LockedUntil = nil
	if a.Status == AdminStatusSuspended {
		a.Status = AdminStatusActive
	}
}
