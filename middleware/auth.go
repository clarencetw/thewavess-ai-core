package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
)

// AuthMiddleware JWT 認證中間件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 獲取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "需要認證",
				Error: &models.APIError{
					Code:    "MISSING_AUTH_HEADER",
					Message: "缺少 Authorization header",
				},
			})
			c.Abort()
			return
		}

		// 檢查 Bearer token 格式
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "認證格式錯誤",
				Error: &models.APIError{
					Code:    "INVALID_AUTH_FORMAT",
					Message: "Authorization header 格式應為 'Bearer <token>'",
				},
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// TODO: 實作 JWT token 驗證邏輯
		// 這裡應該包含：
		// 1. JWT token 解析和驗證
		// 2. 檢查 token 是否過期
		// 3. 從 token 中提取用戶資訊
		// 4. 檢查用戶是否存在且為活躍狀態

		// 暫時的假驗證邏輯
		if token == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "無效的認證令牌",
				Error: &models.APIError{
					Code:    "INVALID_TOKEN",
					Message: "JWT token 無效或已過期",
				},
			})
			c.Abort()
			return
		}

		// 設定用戶 ID 到 context（實際實作時從 JWT 中提取）
		c.Set("user_id", "550e8400-e29b-41d4-a716-446655440000")
		c.Set("username", "demo_user")

		c.Next()
	}
}