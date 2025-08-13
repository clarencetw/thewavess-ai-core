package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
)

// RegisterUser godoc
// @Summary      用戶註冊
// @Description  創建新用戶帳號
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body models.UserRegisterRequest true "用戶註冊資訊"
// @Success      201 {object} models.APIResponse{data=models.AuthResponse} "註冊成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      409 {object} models.APIResponse{error=models.APIError} "用戶名或信箱已存在"
// @Router       /user/register [post]
func RegisterUser(c *gin.Context) {
	var req models.UserRegisterRequest
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

	// 檢查年齡是否滿 18 歲
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_BIRTH_DATE",
				Message: "生日格式錯誤，請使用 YYYY-MM-DD 格式",
				Details: err.Error(),
			},
		})
		return
	}

	age := time.Now().Year() - birthDate.Year()
	if age < 18 {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "AGE_RESTRICTION",
				Message: "本服務僅限 18 歲以上成年用戶使用",
			},
		})
		return
	}

	// 模擬檢查用戶名是否已存在
	if req.Username == "admin" || req.Username == "test" {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USERNAME_EXISTS",
				Message: "用戶名已存在",
			},
		})
		return
	}

	// 模擬檢查信箱是否已存在
	if req.Email == "admin@example.com" {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "EMAIL_EXISTS",
				Message: "信箱已被使用",
			},
		})
		return
	}

	// 生成模擬用戶 ID 和 Token
	userID := "user_" + req.Username + "_" + time.Now().Format("20060102150405")
	accessToken := "mock_access_token_" + userID
	refreshToken := "mock_refresh_token_" + userID

	// 返回註冊成功回應
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "用戶註冊成功",
		Data: models.AuthResponse{
			UserID:       userID,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
		},
	})
}

// LoginUser godoc
// @Summary      用戶登入
// @Description  用戶身份驗證
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body models.UserLoginRequest true "登入憑證"
// @Success      200 {object} models.APIResponse{data=models.AuthResponse} "登入成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "認證失敗"
// @Router       /user/login [post]
func LoginUser(c *gin.Context) {
	var req models.UserLoginRequest
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

	// 模擬驗證用戶憑證
	// 在真實環境中，這裡會檢查數據庫中的用戶和密碼哈希
	validCredentials := map[string]string{
		"alice123":  "password123",
		"bob456":    "password456",
		"charlie":   "mypassword",
		"demo_user": "demo123",
	}

	expectedPassword, userExists := validCredentials[req.Username]
	if !userExists || expectedPassword != req.Password {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "用戶名或密碼錯誤",
			},
		})
		return
	}

	// 生成模擬 Token
	userID := "user_" + req.Username + "_" + time.Now().Format("20060102150405")
	accessToken := "mock_access_token_" + userID
	refreshToken := "mock_refresh_token_" + userID

	// 返回登入成功回應
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "登入成功",
		Data: models.AuthResponse{
			UserID:       userID,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
		},
	})
}

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
// @Summary      刷新 JWT Token
// @Description  使用 refresh token 獲取新的 access token
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse{data=models.AuthResponse} "Token 刷新成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "Token 無效"
// @Router       /user/refresh [post]
func RefreshToken(c *gin.Context) {
	// 獲取當前 token 或 refresh token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_TOKEN",
				Message: "缺少認證 Token",
			},
		})
		return
	}

	// 模擬 token 驗證和刷新
	// 在真實環境中，這裡會驗證 refresh token 的有效性
	if !strings.Contains(authHeader, "refresh") && !strings.Contains(authHeader, "mock") {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REFRESH_TOKEN",
				Message: "無效的刷新 Token",
			},
		})
		return
	}

	// 生成新的 token
	userID := "user_refreshed_" + time.Now().Format("20060102150405")
	newAccessToken := "mock_access_token_" + userID
	newRefreshToken := "mock_refresh_token_" + userID

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Token 刷新成功",
		Data: models.AuthResponse{
			UserID:       userID,
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    3600,
		},
	})
}

// GetProfile godoc
// @Summary      獲取用戶個人資料
// @Description  獲取當前用戶的詳細資料
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse{data=models.UserProfile} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/profile [get]
func GetProfile(c *gin.Context) {
	// 模擬從 JWT token 中獲取用戶 ID
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_TOKEN",
				Message: "缺少認證 Token",
			},
		})
		return
	}

	// 簡單驗證 token 格式 (在真實環境中會解析 JWT)
	if len(authHeader) < 20 || authHeader[:6] != "Bearer" {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_TOKEN",
				Message: "無效的認證 Token",
			},
		})
		return
	}

	// 模擬用戶資料
	mockUsers := map[string]models.UserProfile{
		"alice123": {
			BaseModel: models.BaseModel{
				ID:        "user_alice123_001",
				CreatedAt: time.Now().AddDate(0, -6, 0),
				UpdatedAt: time.Now(),
			},
			Username:     "alice123",
			Email:        "alice@example.com",
			Nickname:     "小愛",
			BirthDate:    time.Date(2000, 3, 15, 0, 0, 0, 0, time.UTC),
			Gender:       "female",
			AvatarURL:    "https://example.com/avatars/alice.jpg",
			CharacterID:  "char_001",
			TotalChats:   156,
			JoinedAt:     time.Now().AddDate(0, -6, 0),
			LastActiveAt: time.Now(),
			Preferences: map[string]interface{}{
				"nsfw_enabled":         true,
				"voice_enabled":        true,
				"notification_enabled": true,
				"preferred_voice":      "voice_001",
				"theme":               "dark",
			},
		},
		"demo_user": {
			BaseModel: models.BaseModel{
				ID:        "user_demo_001",
				CreatedAt: time.Now().AddDate(0, -1, 0),
				UpdatedAt: time.Now(),
			},
			Username:     "demo_user",
			Email:        "demo@example.com",
			Nickname:     "測試用戶",
			BirthDate:    time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:       "female",
			AvatarURL:    "",
			CharacterID:  "char_002",
			TotalChats:   23,
			JoinedAt:     time.Now().AddDate(0, -1, 0),
			LastActiveAt: time.Now(),
			Preferences: map[string]interface{}{
				"nsfw_enabled":         false,
				"voice_enabled":        false,
				"notification_enabled": false,
			},
		},
	}

	// 從 token 中模擬獲取用戶名 (真實環境中會解析 JWT)
	// 這裡簡單使用 alice123 作為默認用戶
	username := "alice123"
	if authHeader == "Bearer demo_token" {
		username = "demo_user"
	}

	userProfile, exists := mockUsers[username]
	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "用戶不存在",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取用戶資料成功",
		Data:    userProfile,
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

	// 模擬用戶名獲取
	username := "alice123"
	if authHeader == "Bearer demo_token" {
		username = "demo_user"
	}

	// 模擬更新後的用戶資料
	updatedProfile := models.UserProfile{
		BaseModel: models.BaseModel{
			ID:        "user_" + username + "_001",
			CreatedAt: time.Now().AddDate(0, -6, 0),
			UpdatedAt: time.Now(),
		},
		Username:     username,
		Email:        username + "@example.com",
		Nickname:     req.Nickname,
		BirthDate:    time.Date(2000, 3, 15, 0, 0, 0, 0, time.UTC),
		Gender:       "female",
		AvatarURL:    req.AvatarURL,
		CharacterID:  "char_001",
		TotalChats:   156,
		JoinedAt:     time.Now().AddDate(0, -6, 0),
		LastActiveAt: time.Now(),
		Preferences: map[string]interface{}{
			"nsfw_enabled":         true,
			"voice_enabled":        true,
			"notification_enabled": true,
		},
	}

	// 處理空值 - 如果沒有提供新值，保持原值
	if req.Nickname == "" {
		if username == "alice123" {
			updatedProfile.Nickname = "小愛"
		} else {
			updatedProfile.Nickname = "測試用戶"
		}
	}

	if req.AvatarURL == "" {
		if username == "alice123" {
			updatedProfile.AvatarURL = "https://example.com/avatars/alice.jpg"
		}
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

	// 驗證偏好設定的有效性
	allowedPreferences := map[string]bool{
		"nsfw_enabled":         true,
		"voice_enabled":        true,
		"notification_enabled": true,
		"preferred_voice":      true,
		"theme":               true,
		"default_character":    true,
		"language":            true,
		"auto_scene_update":   true,
	}

	for key := range req.Preferences {
		if !allowedPreferences[key] {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_PREFERENCE",
					Message: "不支援的偏好設定: " + key,
				},
			})
			return
		}
	}

	// 特別驗證 NSFW 設定 - 確保用戶理解成人內容
	if nsfwEnabled, exists := req.Preferences["nsfw_enabled"]; exists {
		if enabled, ok := nsfwEnabled.(bool); ok && enabled {
			// 這裡可以記錄用戶同意使用 NSFW 功能
			// 在真實環境中，可能需要額外的年齡驗證
		}
	}

	// 模擬保存偏好設定成功
	username := "alice123"
	if authHeader == "Bearer demo_token" {
		username = "demo_user"
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "偏好設定更新成功",
		Data: map[string]interface{}{
			"user_id":     "user_" + username + "_001",
			"preferences": req.Preferences,
			"updated_at":  time.Now(),
		},
	})
}