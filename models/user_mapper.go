package models

import (
	"github.com/clarencetw/thewavess-ai-core/models/db"
)

// UserMapper 用戶模型轉換器
type UserMapper struct{}

// NewUserMapper 創建新的用戶映射器
func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

// FromDB 從資料庫模型轉換為領域模型
func (m *UserMapper) FromDB(userDB *db.UserDB) *User {
	if userDB == nil {
		return nil
	}

	return &User{
		ID:           userDB.ID,
		Username:     userDB.Username,
		Email:        userDB.Email,
		Password:     userDB.Password,
		DisplayName:  userDB.DisplayName,
		Bio:          userDB.Bio,
		Status:       userDB.Status,
		Nickname:     userDB.Nickname,
		Gender:       userDB.Gender,
		BirthDate:    userDB.BirthDate,
		AvatarURL:    userDB.AvatarURL,
		IsVerified:   userDB.IsVerified,
		IsAdult:      userDB.IsAdult,
		Preferences:  userDB.Preferences,
		CreatedAt:    userDB.CreatedAt,
		UpdatedAt:    userDB.UpdatedAt,
		LastLoginAt:  userDB.LastLoginAt,
	}
}

// ToDB 從領域模型轉換為資料庫模型
func (m *UserMapper) ToDB(user *User) *db.UserDB {
	if user == nil {
		return nil
	}

	return &db.UserDB{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Password:     user.Password,
		DisplayName:  user.DisplayName,
		Bio:          user.Bio,
		Status:       user.Status,
		Nickname:     user.Nickname,
		Gender:       user.Gender,
		BirthDate:    user.BirthDate,
		AvatarURL:    user.AvatarURL,
		IsVerified:   user.IsVerified,
		IsAdult:      user.IsAdult,
		Preferences:  user.Preferences,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		LastLoginAt:  user.LastLoginAt,
	}
}

// FromDBList 從資料庫模型列表轉換為領域模型列表
func (m *UserMapper) FromDBList(userDBs []*db.UserDB) []*User {
	if userDBs == nil {
		return nil
	}

	users := make([]*User, len(userDBs))
	for i, userDB := range userDBs {
		users[i] = m.FromDB(userDB)
	}
	return users
}

// ToDBList 從領域模型列表轉換為資料庫模型列表
func (m *UserMapper) ToDBList(users []*User) []*db.UserDB {
	if users == nil {
		return nil
	}

	userDBs := make([]*db.UserDB, len(users))
	for i, user := range users {
		userDBs[i] = m.ToDB(user)
	}
	return userDBs
}