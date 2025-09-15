package utils

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// TestJWT_用戶令牌完整流程測試
func TestJWT_用戶令牌完整流程測試(t *testing.T) {
	// 設置測試用的 JWT 密鑰
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 測試用戶資料
	userID := "user_123"
	username := "testuser"
	email := "test@example.com"

	// 測試生成訪問令牌
	t.Run("生成用戶訪問令牌", func(t *testing.T) {
		token, err := GenerateAccessToken(userID, username, email)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	// 測試驗證有效令牌
	t.Run("驗證有效用戶令牌", func(t *testing.T) {
		token, err := GenerateAccessToken(userID, username, email)
		assert.NoError(t, err)

		claims, err := ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, "thewavess-ai-core", claims.Issuer)
	})

	// 測試無效令牌
	t.Run("驗證無效用戶令牌", func(t *testing.T) {
		claims, err := ValidateToken("invalid.token.here")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	// 測試生成和驗證刷新令牌
	t.Run("刷新令牌生成與驗證", func(t *testing.T) {
		refreshToken, err := GenerateRefreshToken(userID)
		assert.NoError(t, err)
		assert.NotEmpty(t, refreshToken)

		extractedUserID, err := ValidateRefreshToken(refreshToken)
		assert.NoError(t, err)
		assert.Equal(t, userID, extractedUserID)
	})
}

// TestJWT_管理員令牌完整流程測試
func TestJWT_管理員令牌完整流程測試(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 測試管理員資料
	adminID := "admin_456"
	username := "admin"
	email := "admin@example.com"
	role := "super_admin"
	permissions := []string{"read", "write", "delete"}

	// 測試生成管理員訪問令牌
	t.Run("生成管理員訪問令牌", func(t *testing.T) {
		token, err := GenerateAdminAccessToken(adminID, username, email, role, permissions)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	// 測試驗證有效管理員令牌
	t.Run("驗證有效管理員令牌", func(t *testing.T) {
		token, err := GenerateAdminAccessToken(adminID, username, email, role, permissions)
		assert.NoError(t, err)

		claims, err := ValidateAdminToken(token)
		assert.NoError(t, err)
		assert.Equal(t, adminID, claims.AdminID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, permissions, claims.Permissions)
		assert.Equal(t, "thewavess-ai-core-admin", claims.Issuer)
	})

	// 測試用戶令牌不能通過管理員驗證
	t.Run("用戶令牌無法通過管理員驗證", func(t *testing.T) {
		userToken, err := GenerateAccessToken("user_123", "user", "user@test.com")
		assert.NoError(t, err)

		claims, err := ValidateAdminToken(userToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "not an admin token")
	})
}

// TestJWT_令牌過期時間測試
func TestJWT_令牌過期時間測試(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	t.Run("用戶令牌24小時有效期", func(t *testing.T) {
		token, err := GenerateAccessToken("user_123", "testuser", "test@example.com")
		assert.NoError(t, err)

		claims, err := ValidateToken(token)
		assert.NoError(t, err)

		// 檢查過期時間約為24小時後
		expectedExpiry := time.Now().Add(24 * time.Hour)
		actualExpiry := claims.ExpiresAt.Time

		// 允許5分鐘的誤差
		diff := actualExpiry.Sub(expectedExpiry).Abs()
		assert.True(t, diff < 5*time.Minute, "令牌過期時間應該是24小時後")
	})

	t.Run("管理員令牌8小時有效期", func(t *testing.T) {
		token, err := GenerateAdminAccessToken("admin_123", "admin", "admin@test.com", "admin", []string{})
		assert.NoError(t, err)

		claims, err := ValidateAdminToken(token)
		assert.NoError(t, err)

		// 檢查過期時間約為8小時後
		expectedExpiry := time.Now().Add(8 * time.Hour)
		actualExpiry := claims.ExpiresAt.Time

		// 允許5分鐘的誤差
		diff := actualExpiry.Sub(expectedExpiry).Abs()
		assert.True(t, diff < 5*time.Minute, "管理員令牌過期時間應該是8小時後")
	})

	t.Run("刷新令牌7天有效期", func(t *testing.T) {
		refreshToken, err := GenerateRefreshToken("user_123")
		assert.NoError(t, err)

		// 解析令牌以檢查過期時間
		token, err := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})
		assert.NoError(t, err)

		claims := token.Claims.(*jwt.RegisteredClaims)
		expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
		actualExpiry := claims.ExpiresAt.Time

		// 允許5分鐘的誤差
		diff := actualExpiry.Sub(expectedExpiry).Abs()
		assert.True(t, diff < 5*time.Minute, "刷新令牌過期時間應該是7天後")
	})
}

// TestJWT_環境變數測試
func TestJWT_環境變數測試(t *testing.T) {
	// 清除環境變數使用預設值
	os.Unsetenv("JWT_SECRET")

	t.Run("無環境變數時使用預設密鑰", func(t *testing.T) {
		token, err := GenerateAccessToken("user_123", "test", "test@test.com")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// 應該能夠正常驗證（因為生成和驗證都使用相同的預設密鑰）
		claims, err := ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, "user_123", claims.UserID)
	})

	// 恢復環境變數
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")
}
