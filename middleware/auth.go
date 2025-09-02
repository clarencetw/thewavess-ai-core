package middleware

import (
	"net/http"
	"strings"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
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
		// 獲取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "需要管理員認證",
				Error: &models.APIError{
					Code:    "MISSING_ADMIN_AUTH_HEADER",
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
				Message: "管理員認證格式錯誤",
				Error: &models.APIError{
					Code:    "INVALID_ADMIN_AUTH_FORMAT",
					Message: "Authorization header 格式應為 'Bearer <admin_token>'",
				},
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// 驗證管理員 JWT token
		claims, err := utils.ValidateAdminToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "無效的管理員認證令牌",
				Error: &models.APIError{
					Code:    "INVALID_ADMIN_TOKEN",
					Message: "管理員 JWT token 無效或已過期: " + err.Error(),
				},
			})
			c.Abort()
			return
		}

		// 設定管理員資訊到 context
		c.Set("admin_id", claims.AdminID)
		c.Set("admin_username", claims.Username)
		c.Set("admin_email", claims.Email)
		c.Set("admin_role", claims.Role)
		c.Set("admin_permissions", claims.Permissions)
		c.Set("admin_claims", claims)

		c.Next()
	}
}

// RequireSuperAdmin 要求超級管理員權限的中間件
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先必須通過 AdminMiddleware
		adminRole, roleExists := c.Get("admin_role")

		if !roleExists {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "權限檢查失敗",
				Error: &models.APIError{
					Code:    "PERMISSION_CHECK_FAILED",
					Message: "無法驗證管理員權限",
				},
			})
			c.Abort()
			return
		}

		// 只有超級管理員可以訪問
		if adminRole.(string) != "super_admin" {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "需要超級管理員權限",
				Error: &models.APIError{
					Code:    "INSUFFICIENT_ADMIN_PRIVILEGES",
					Message: "此操作需要超級管理員權限",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
