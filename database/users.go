package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser 創建新用戶
func CreateUser(user *models.User) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Generate user ID
	user.ID = "user_" + utils.GenerateID(12)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (
			id, username, email, password_hash, nickname, 
			gender, birth_date, avatar_url, is_verified, is_adult,
			preferences, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = DB.Exec(query,
		user.ID,
		user.Username,
		user.Email,
		string(hashedPassword),
		user.Nickname,
		user.Gender,
		user.BirthDate,
		user.AvatarURL,
		user.IsVerified,
		user.IsAdult,
		"{}",
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// GetUserByUsername 根據用戶名獲取用戶
func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	var passwordHash string
	var birthDate sql.NullTime
	var avatarURL sql.NullString
	var lastLoginAt sql.NullTime

	query := `
		SELECT id, username, email, password_hash, nickname, gender, birth_date,
		       avatar_url, is_verified, is_adult, created_at, updated_at, last_login_at
		FROM users WHERE username = $1
	`

	err := DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&passwordHash,
		&user.Nickname,
		&user.Gender,
		&birthDate,
		&avatarURL,
		&user.IsVerified,
		&user.IsAdult,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Password = passwordHash // Store hash for verification
	if birthDate.Valid {
		user.BirthDate = birthDate.Time
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetUserByEmail 根據郵箱獲取用戶
func GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	var passwordHash string
	var birthDate sql.NullTime
	var avatarURL sql.NullString
	var lastLoginAt sql.NullTime

	query := `
		SELECT id, username, email, password_hash, nickname, gender, birth_date,
		       avatar_url, is_verified, is_adult, created_at, updated_at, last_login_at
		FROM users WHERE email = $1
	`

	err := DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&passwordHash,
		&user.Nickname,
		&user.Gender,
		&birthDate,
		&avatarURL,
		&user.IsVerified,
		&user.IsAdult,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Password = passwordHash
	if birthDate.Valid {
		user.BirthDate = birthDate.Time
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetUserByID 根據ID獲取用戶
func GetUserByID(userID string) (*models.User, error) {
	user := &models.User{}
	var passwordHash string
	var birthDate sql.NullTime
	var avatarURL sql.NullString
	var lastLoginAt sql.NullTime

	query := `
		SELECT id, username, email, password_hash, nickname, gender, birth_date,
		       avatar_url, is_verified, is_adult, created_at, updated_at, last_login_at
		FROM users WHERE id = $1
	`

	err := DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&passwordHash,
		&user.Nickname,
		&user.Gender,
		&birthDate,
		&avatarURL,
		&user.IsVerified,
		&user.IsAdult,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if birthDate.Valid {
		user.BirthDate = birthDate.Time
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// VerifyPassword 驗證密碼
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// UpdateLastLogin 更新最後登入時間
func UpdateLastLogin(userID string) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`
	_, err := DB.Exec(query, time.Now(), userID)
	return err
}

// UpdateUser 更新用戶資料
func UpdateUser(user *models.User) error {
	query := `
		UPDATE users SET
			nickname = $1,
			gender = $2,
			avatar_url = $3,
			updated_at = $4
		WHERE id = $5
	`

	_, err := DB.Exec(query,
		user.Nickname,
		user.Gender,
		user.AvatarURL,
		time.Now(),
		user.ID,
	)

	return err
}

// CheckUsernameExists 檢查用戶名是否存在
func CheckUsernameExists(username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = $1`
	err := DB.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckEmailExists 檢查郵箱是否存在
func CheckEmailExists(email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1`
	err := DB.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}