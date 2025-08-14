package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
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

		// 驗證 JWT token
		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "無效的認證令牌",
				Error: &models.APIError{
					Code:    "INVALID_TOKEN",
					Message: "JWT token 無效或已過期: " + err.Error(),
				},
			})
			c.Abort()
			return
		}

		// 設定用戶資訊到 context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		c.Next()
	}
}