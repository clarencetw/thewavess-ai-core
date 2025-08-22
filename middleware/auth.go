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

// AdminMiddleware 管理員權限中間件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先需要通過基本認證
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "未授權",
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "用戶未通過認證",
				},
			})
			c.Abort()
			return
		}

		// 檢查管理員權限 (目前簡化實現，實際應查詢用戶角色)
		userIDStr := userID.(string)
		
		// TODO: 實際項目中應該查詢用戶的角色/權限
		// 這裡暫時使用簡化邏輯：特定用戶ID或用戶名為admin的用戶
		username, usernameExists := c.Get("username")
		if usernameExists && username.(string) == "admin" {
			c.Next()
			return
		}
		
		// 或者基於用戶ID的管理員列表（生產環境應該從數據庫查詢）
		adminUsers := map[string]bool{
			"admin": true,
			// 可以添加其他管理員用戶ID
		}
		
		if adminUsers[userIDStr] {
			c.Next()
			return
		}

		// 非管理員用戶
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "需要管理員權限",
			Error: &models.APIError{
				Code:    "INSUFFICIENT_PRIVILEGES",
				Message: "此操作需要管理員權限",
			},
		})
		c.Abort()
	}
}

