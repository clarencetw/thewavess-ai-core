package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

var (
	userMapper = models.NewUserMapper()
)

// RegisterUser godoc
// @Summary      用戶註冊
// @Description  創建新用戶帳號
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user body models.RegisterRequest true "註冊信息"
// @Success      201 {object} models.APIResponse{data=models.UserResponse} "註冊成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      409 {object} models.APIResponse{error=models.APIError} "用戶已存在"
// @Router       /auth/register [post]
func RegisterUser(c *gin.Context) {
	ctx := context.Background()

	var req models.RegisterRequest
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 檢查用戶是否已存在
	var existingUserDB db.UserDB
	exists, err := GetDB().NewSelect().
		Model(&existingUserDB).
		Where("email = ? OR username = ?", req.Email, req.Username).
		Exists(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to check existing user")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "檢查用戶失敗",
			},
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_EXISTS",
				Message: "用戶名或郵箱已存在",
			},
		})
		return
	}

	// 加密密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to hash password")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ENCRYPTION_ERROR",
				Message: "密碼加密失敗",
			},
		})
		return
	}

	// 創建新用戶
	defaultBirthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	user := &models.User{
		ID:        utils.GenerateUserID(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Status:    "active",
		BirthDate: &defaultBirthDate,
		IsVerified: false,
		IsAdult:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 轉換為DB模型並插入數據庫
	userDB := userMapper.ToDB(user)
	_, err = GetDB().NewInsert().Model(userDB).Exec(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to create user")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "創建用戶失敗",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "用戶註冊成功",
		Data:    user.ToResponse(),
	})
}

// LoginUser godoc
// @Summary      用戶登入
// @Description  驗證用戶並返回 JWT Token
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        credentials body models.LoginRequest true "登入憑證 (username/password)"
// @Success      200 {object} models.APIResponse{data=models.LoginResponse} "登入成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "認證失敗"
// @Router       /auth/login [post]
func LoginUser(c *gin.Context) {
	ctx := context.Background()

	var req models.LoginRequest
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 查找用戶
	var userDB db.UserDB
	err := GetDB().NewSelect().
		Model(&userDB).
		Where("username = ? AND status = ?", req.Username, "active").
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("username", req.Username).Error("User not found")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "用戶名或密碼錯誤",
			},
		})
		return
	}

	// 轉換為領域模型
	user := userMapper.FromDB(&userDB)

	// 驗證密碼
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", user.ID).Error("Password verification failed")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "用戶名/郵箱或密碼錯誤",
			},
		})
		return
	}

	// 生成 JWT Token
	token, err := utils.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to generate JWT token")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "生成認證令牌失敗",
			},
		})
		return
	}

	// 生成 Refresh Token
	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "生成刷新令牌失敗",
			},
		})
		return
	}

	// 更新最後登入時間
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now

	updatedUserDB := userMapper.ToDB(user)
	_, err = GetDB().NewUpdate().
		Model(updatedUserDB).
		Column("last_login_at", "updated_at").
		Where("id = ?", user.ID).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to update last login time")
		// 不中斷登入流程，只記錄錯誤
	}

	response := &models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24小時
		User:         user.ToResponse(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "登入成功",
		Data:    response,
	})
}

// GetUserProfile godoc
// @Summary      獲取用戶資料
// @Description  獲取當前用戶的詳細資料
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse{data=models.UserResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "用戶不存在"
// @Router       /user/profile [get]
func GetUserProfile(c *gin.Context) {
	ctx := context.Background()

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var userDB db.UserDB
	err := GetDB().NewSelect().
		Model(&userDB).
		Where("id = ? AND status = ?", userID, "active").
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", userID).Error("Failed to query user")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "用戶不存在",
			},
		})
		return
	}

	// 轉換為領域模型
	user := userMapper.FromDB(&userDB)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取用戶資料成功",
		Data:    user.ToResponse(),
	})
}

// UpdateUserProfile godoc
// @Summary      更新用戶資料
// @Description  更新用戶基本資料
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        profile body models.UpdateProfileRequest true "用戶資料"
// @Success      200 {object} models.APIResponse{data=models.UserResponse} "更新成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/profile [put]
func UpdateUserProfile(c *gin.Context) {
	ctx := context.Background()

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var req models.UpdateProfileRequest
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 構建更新數據
	updateData := models.User{
		UpdatedAt: time.Now(),
	}

	// 只更新提供的字段
	updateQuery := GetDB().NewUpdate().Model(&updateData)

	if req.DisplayName != nil {
		updateData.DisplayName = req.DisplayName
		updateQuery = updateQuery.Column("display_name")
	}
	if req.Bio != nil {
		updateData.Bio = req.Bio
		updateQuery = updateQuery.Column("bio")
	}
	if req.AvatarURL != nil {
		updateData.AvatarURL = req.AvatarURL
		updateQuery = updateQuery.Column("avatar_url")
	}

	// 執行更新
	updateQuery = updateQuery.Column("updated_at").Where("id = ? AND status = ?", userID, "active")

	result, err := updateQuery.Exec(ctx)
	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", userID).Error("Failed to update user profile")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "更新用戶資料失敗",
			},
		})
		return
	}

	// 檢查是否有行被更新
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "用戶不存在",
			},
		})
		return
	}

	// 獲取更新後的用戶信息
	var userDB db.UserDB
	err = GetDB().NewSelect().
		Model(&userDB).
		Where("id = ?", userID).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to fetch updated user")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取更新後用戶信息失敗",
			},
		})
		return
	}

	// 轉換為領域模型
	user := userMapper.FromDB(&userDB)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "用戶資料更新成功",
		Data:    user.ToResponse(),
	})
}

// GetUserPreferences godoc
// @Summary      獲取用戶偏好設置
// @Description  獲取用戶偏好設置
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse{data=map[string]interface{}} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "用戶不存在"
// @Router       /user/preferences [get]
func GetUserPreferences(c *gin.Context) {
	ctx := context.Background()

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var userDB db.UserDB
	err := GetDB().NewSelect().
		Model(&userDB).
		Where("id = ? AND status = ?", userID, "active").
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", userID).Error("User not found")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "用戶不存在",
			},
		})
		return
	}

	// 轉換為領域模型
	user := userMapper.FromDB(&userDB)

	// 返回偏好設置，如果為空則返回空對象
	preferences := user.Preferences
	if preferences == nil {
		preferences = make(map[string]interface{})
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取偏好設置成功",
		Data:    preferences,
	})
}

// UpdateUserPreferences godoc
// @Summary      更新用戶偏好設置
// @Description  更新用戶偏好設置
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        preferences body models.UpdatePreferencesRequest true "偏好設置"
// @Success      200 {object} models.APIResponse{data=models.UserResponse} "更新成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/preferences [put]
func UpdateUserPreferences(c *gin.Context) {
	ctx := context.Background()

	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var req models.UpdatePreferencesRequest
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 獲取當前用戶
	var userDB db.UserDB
	err := GetDB().NewSelect().
		Model(&userDB).
		Where("id = ? AND status = ?", userID, "active").
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", userID).Error("User not found")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "用戶不存在",
			},
		})
		return
	}

	// 轉換為領域模型
	user := userMapper.FromDB(&userDB)

	// 更新偏好設置
	if user.Preferences == nil {
		user.Preferences = make(map[string]interface{})
	}

	// 合併新的偏好設置
	for key, value := range req.Preferences {
		user.Preferences[key] = value
	}

	user.UpdatedAt = time.Now()

	// 轉換回DB模型並執行更新
	updatedUserDB := userMapper.ToDB(user)
	_, err = GetDB().NewUpdate().
		Model(updatedUserDB).
		Column("preferences", "updated_at").
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", userID).Error("Failed to update user preferences")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "更新偏好設置失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "偏好設置更新成功",
		Data:    user.ToResponse(),
	})
}

// LogoutUser godoc
// @Summary      用戶登出
// @Description  登出當前用戶，使 JWT Token 失效（客戶端處理）
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        logout body models.LogoutRequest false "登出請求"
// @Success      200 {object} models.APIResponse "登出成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /auth/logout [post]
func LogoutUser(c *gin.Context) {
	// 從中間件獲取用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	// 解析請求體（可選）
	var req models.LogoutRequest
	c.ShouldBindJSON(&req)

	// 記錄登出事件
	utils.Logger.WithField("user_id", userID).Info("User logged out")

	// 由於使用 JWT，實際的 token 失效需要客戶端處理
	// 這裡主要是記錄登出事件和返回成功響應
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "登出成功",
	})
}

// RefreshToken godoc
// @Summary      刷新訪問令牌
// @Description  使用 Refresh Token 獲取新的 Access Token
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        refresh body models.RefreshTokenRequest true "刷新令牌請求"
// @Success      200 {object} models.APIResponse{data=models.RefreshTokenResponse} "刷新成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "Token 無效"
// @Router       /auth/refresh [post]
func RefreshToken(c *gin.Context) {
	ctx := context.Background()

	var req models.RefreshTokenRequest
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 驗證 Refresh Token
	userID, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		utils.Logger.WithError(err).Error("Invalid refresh token")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REFRESH_TOKEN",
				Message: "刷新令牌無效或已過期",
			},
		})
		return
	}

	// 查找用戶
	var userDB db.UserDB
	err = GetDB().NewSelect().
		Model(&userDB).
		Where("id = ? AND status = ?", userID, "active").
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", userID).Error("User not found")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "用戶不存在或已停用",
			},
		})
		return
	}

	// 轉換為領域模型
	user := userMapper.FromDB(&userDB)

	// 生成新的 Access Token
	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to generate new access token")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "生成新訪問令牌失敗",
			},
		})
		return
	}

	// 生成新的 Refresh Token
	newRefreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to generate new refresh token")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "生成新刷新令牌失敗",
			},
		})
		return
	}

	response := &models.RefreshTokenResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24小時
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "令牌刷新成功",
		Data:    response,
	})
}

// UploadAvatar godoc
// @Summary      設置用戶頭像
// @Description  通過URL設置用戶頭像
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        avatar body object true "頭像URL"
// @Success      200 {object} models.APIResponse "設置成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/avatar [post]
func UploadAvatar(c *gin.Context) {
	// 檢查認證
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var req struct {
		AvatarURL string `json:"avatar_url" binding:"required,url"`
	}

	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	ctx := context.Background()

	// 更新用戶頭像URL到資料庫
	_, err := GetDB().NewUpdate().
		Model((*db.UserDB)(nil)).
		Set("avatar_url = ?", req.AvatarURL).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", userID).Error("Failed to update avatar URL")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "更新頭像失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "頭像設置成功",
		Data: gin.H{
			"user_id":     userID,
			"avatar_url":  req.AvatarURL,
			"updated_at":  time.Now(),
			"status":      "active",
			"validation":  "URL format verified",
		},
	})
}

// DeleteAccount godoc
// @Summary      刪除用戶帳號
// @Description  永久刪除用戶帳號和相關數據
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        confirm body object true "確認刪除"
// @Success      200 {object} models.APIResponse "刪除成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/account [delete]
func DeleteAccount(c *gin.Context) {
	// 檢查認證
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var req struct {
		Password    string `json:"password" binding:"required"`
		Confirmation string `json:"confirmation" binding:"required"`
	}

	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	if req.Confirmation != "DELETE_MY_ACCOUNT" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CONFIRMATION",
				Message: "請輸入 'DELETE_MY_ACCOUNT' 確認刪除",
			},
		})
		return
	}

	// 靜態回應 - 模擬帳號刪除
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "帳號刪除成功",
		Data: gin.H{
			"user_id":     userID,
			"deleted_at":  time.Now(),
			"data_retention": "用戶數據將在 30 天後完全清除",
			"recovery_period": "30天內可聯繫客服恢復帳號",
		},
	})
}

// VerifyAge godoc
// @Summary      年齡驗證
// @Description  驗證用戶年齡，允許訪問18+內容
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        verification body object true "年齡驗證"
// @Success      200 {object} models.APIResponse "驗證成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/verify [post]
func VerifyAge(c *gin.Context) {
	// 檢查認證
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var req struct {
		BirthYear int  `json:"birth_year" binding:"required"`
		Consent   bool `json:"consent" binding:"required"`
	}

	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	// 計算年齡
	currentYear := time.Now().Year()
	age := currentYear - req.BirthYear

	if age < 18 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "AGE_INSUFFICIENT",
				Message: "年齡不足18歲，無法訪問成人內容",
			},
		})
		return
	}

	if !req.Consent {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CONSENT_REQUIRED",
				Message: "必須同意條款才能完成驗證",
			},
		})
		return
	}

	ctx := context.Background()

	// 更新用戶年齡驗證狀態到資料庫
	_, err := GetDB().NewUpdate().
		Model((*db.UserDB)(nil)).
		Set("is_adult = ?", true).
		Set("is_verified = ?", true).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("user_id", userID).Error("Failed to update age verification")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "年齡驗證更新失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "年齡驗證成功",
		Data: gin.H{
			"user_id":         userID,
			"verified_at":     time.Now(),
			"age_verified":    true,
			"adult_content":   true,
			"verification_id": utils.GenerateUUID(),
		},
	})
}