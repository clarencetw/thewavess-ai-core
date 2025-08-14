package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
)

// LogoutUser godoc
// @Summary      用戶登出
// @Description  註銷當前用戶會話
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse "登出成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/logout [post]
func LogoutUser(c *gin.Context) {
	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	// 模擬登出成功
	// 在真實環境中，這裡會將 token 加入黑名單或從 Redis 中刪除
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "登出成功",
		Data: map[string]interface{}{
			"logout_time": time.Now(),
			"message":     "您已安全登出",
		},
	})
}

// RefreshToken godoc
// @Summary      刷新存取令牌
// @Description  使用 refresh token 獲取新的 access token
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object{refresh_token=string} true "刷新令牌請求"
// @Success      200 {object} models.APIResponse{data=models.AuthResponse} "刷新成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "令牌無效或過期"
// @Router       /user/refresh [post]
func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "請求參數驗證失敗",
				Details: err.Error(),
			},
		})
		return
	}

	// 簡單驗證 refresh token（在真實環境中會驗證 JWT）
	if len(req.RefreshToken) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REFRESH_TOKEN",
				Message: "無效的 refresh token",
			},
		})
		return
	}

	// 生成新的 tokens
	newAccessToken := "mock_new_access_token_" + time.Now().Format("20060102150405")
	newRefreshToken := "mock_new_refresh_token_" + time.Now().Format("20060102150405")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "令牌刷新成功",
		Data: models.AuthResponse{
			UserID:       "user_from_refresh_token",
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    3600,
		},
	})
}

// UpdateProfile godoc
// @Summary      更新用戶個人資料
// @Description  更新當前用戶的個人資料
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.UpdateProfileRequest true "更新資料"
// @Success      200 {object} models.APIResponse{data=models.UserProfile} "更新成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/profile [put]
func UpdateProfile(c *gin.Context) {
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "請求參數驗證失敗",
				Details: err.Error(),
			},
		})
		return
	}

	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	// 模擬更新用戶資料
	updatedProfile := models.UserProfile{
		BaseModel: models.BaseModel{
			ID:        "user_demo_001",
			CreatedAt: time.Now().AddDate(0, -6, 0),
			UpdatedAt: time.Now(),
		},
		Username:     "demo_user",
		Email:        "demo@example.com",
		Nickname:     req.Nickname,
		AvatarURL:    req.AvatarURL,
		JoinedAt:     time.Now().AddDate(0, -6, 0),
		LastActiveAt: time.Now(),
		TotalChats:   42,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "個人資料更新成功",
		Data:    updatedProfile,
	})
}

// UpdatePreferences godoc
// @Summary      更新用戶偏好設定
// @Description  更新用戶的應用偏好設定
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.UpdatePreferencesRequest true "偏好設定"
// @Success      200 {object} models.APIResponse "更新成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/preferences [put]
func UpdatePreferences(c *gin.Context) {
	var req models.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "請求參數驗證失敗",
				Details: err.Error(),
			},
		})
		return
	}

	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "偏好設定更新成功",
		Data: map[string]interface{}{
			"preferences": req.Preferences,
			"updated_at":  time.Now(),
		},
	})
}