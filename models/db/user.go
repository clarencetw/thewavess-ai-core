package db

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// UserDB 用戶資料庫模型
type UserDB struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID          string     `bun:"id,pk" json:"id"`
	Username    string     `bun:"username,unique,notnull" json:"username"`
	Email       string     `bun:"email,unique,notnull" json:"email"`
	Password    string     `bun:"password_hash,notnull" json:"-"`
	DisplayName *string    `bun:"display_name" json:"display_name"`
	Bio         *string    `bun:"bio" json:"bio"`
	Status      string     `bun:"status,default:'active'" json:"status"`
	Nickname    *string    `bun:"nickname" json:"nickname"`
	Gender      *string    `bun:"gender" json:"gender"`
	BirthDate   *time.Time `bun:"birth_date" json:"birth_date"`
	AvatarURL   *string    `bun:"avatar_url" json:"avatar_url"`
	IsVerified  bool       `bun:"is_verified,default:false" json:"is_verified"`
	IsAdult     bool       `bun:"is_adult,default:false" json:"is_adult"`
	// 時間戳 (包含軟刪除)
	CreatedAt   time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt   time.Time  `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	DeletedAt   *time.Time `bun:"deleted_at,soft_delete,nullzero" json:"deleted_at,omitempty"`
	LastLoginAt *time.Time `bun:"last_login_at" json:"last_login_at"`

	// IP 記錄功能
	RegistrationIP *string `bun:"registration_ip" json:"registration_ip,omitempty"`
	LastLoginIP    *string `bun:"last_login_ip" json:"last_login_ip,omitempty"`

	// Relations
	Chats         []ChatDB         `bun:"rel:has-many,join:id=user_id"`
	Relationships []RelationshipDB `bun:"rel:has-many,join:id=user_id"`
}

// BeforeAppendModel 在模型操作前執行 (使用 Hook 的最佳實踐)
var _ bun.BeforeAppendModelHook = (*UserDB)(nil)

func (u *UserDB) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		now := time.Now()
		u.CreatedAt = now
		u.UpdatedAt = now
	case *bun.UpdateQuery:
		u.UpdatedAt = time.Now()
	}
	return nil
}
