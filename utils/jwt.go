package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT 聲明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// AdminJWTClaims 管理員JWT 聲明
type AdminJWTClaims struct {
	AdminID     string   `json:"admin_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成訪問令牌
func GenerateAccessToken(userID, username, email string) (string, error) {
	secretKey := GetEnvWithDefault("JWT_SECRET", "your-super-secret-jwt-key-here")
	if secretKey == "" {
		secretKey = "default-secret-key-for-development"
	}

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小時有效期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "thewavess-ai-core",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID string) (string, error) {
	secretKey := GetEnvWithDefault("JWT_SECRET", "your-super-secret-jwt-key-here")
	if secretKey == "" {
		secretKey = "default-secret-key-for-development"
	}

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7天有效期
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "thewavess-ai-core",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateToken 驗證令牌
func ValidateToken(tokenString string) (*JWTClaims, error) {
	secretKey := GetEnvWithDefault("JWT_SECRET", "your-super-secret-jwt-key-here")
	if secretKey == "" {
		secretKey = "default-secret-key-for-development"
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken 驗證刷新令牌
func ValidateRefreshToken(tokenString string) (string, error) {
	secretKey := GetEnvWithDefault("JWT_SECRET", "your-super-secret-jwt-key-here")
	if secretKey == "" {
		secretKey = "default-secret-key-for-development"
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, nil
	}

	return "", errors.New("invalid refresh token")
}

// GenerateAdminAccessToken 生成管理員訪問令牌
func GenerateAdminAccessToken(adminID, username, email, role string, permissions []string) (string, error) {
	secretKey := GetEnvWithDefault("JWT_SECRET", "your-super-secret-jwt-key-here")
	if secretKey == "" {
		secretKey = "default-secret-key-for-development"
	}

	claims := AdminJWTClaims{
		AdminID:     adminID,
		Username:    username,
		Email:       email,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)), // 8小時有效期（比用戶短）
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "thewavess-ai-core-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateAdminToken 驗證管理員令牌
func ValidateAdminToken(tokenString string) (*AdminJWTClaims, error) {
	secretKey := GetEnvWithDefault("JWT_SECRET", "your-super-secret-jwt-key-here")
	if secretKey == "" {
		secretKey = "default-secret-key-for-development"
	}

	token, err := jwt.ParseWithClaims(tokenString, &AdminJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AdminJWTClaims); ok && token.Valid {
		// 檢查是否為管理員令牌（通過issuer判斷）
		if claims.Issuer != "thewavess-ai-core-admin" {
			return nil, errors.New("not an admin token")
		}
		return claims, nil
	}

	return nil, errors.New("invalid admin token")
}
