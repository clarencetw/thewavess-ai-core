package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// RegisterUserReal godoc
// @Summary      用戶註冊
// @Description  創建新用戶帳號（真實資料庫版本）
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body models.UserRegisterRequest true "用戶註冊資訊"
// @Success      201 {object} models.APIResponse{data=models.AuthResponse} "註冊成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      409 {object} models.APIResponse{error=models.APIError} "用戶名或信箱已存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "資料庫錯誤"
// @Router       /user/register [post]
func RegisterUserReal(c *gin.Context) {
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
			},
		})
		return
	}

	age := time.Now().Year() - birthDate.Year()
	isAdult := age >= 18
	if !isAdult {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "AGE_RESTRICTION",
				Message: "本服務僅限 18 歲以上成年用戶使用",
			},
		})
		return
	}

	// 檢查用戶名是否已存在
	exists, err := database.CheckUsernameExists(req.Username)
	if err != nil {
		utils.LogError(err, "check username exists", logrus.Fields{"username": req.Username})
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "資料庫錯誤",
			},
		})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USERNAME_EXISTS",
				Message: "用戶名已被使用",
			},
		})
		return
	}

	// 檢查信箱是否已存在
	exists, err = database.CheckEmailExists(req.Email)
	if err != nil {
		utils.LogError(err, "check email exists", logrus.Fields{"email": req.Email})
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "資料庫錯誤",
			},
		})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "EMAIL_EXISTS",
				Message: "信箱已被註冊",
			},
		})
		return
	}

	// 創建用戶對象
	user := &models.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   req.Password,
		Nickname:   req.Nickname,
		Gender:     req.Gender,
		BirthDate:  birthDate,
		IsVerified: false,
		IsAdult:    isAdult,
	}

	// 保存到資料庫
	if err := database.CreateUser(user); err != nil {
		utils.LogError(err, "create user", logrus.Fields{"username": req.Username})
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "創建用戶失敗",
			},
		})
		return
	}

	// 生成 JWT Token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		utils.LogError(err, "generate access token", nil)
		accessToken = "error_generating_token"
	}
	
	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.LogError(err, "generate refresh token", nil)
		refreshToken = "error_generating_token"
	}

	// 記錄註冊事件
	utils.Logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": req.Username,
		"email":    req.Email,
		"is_adult": isAdult,
	}).Info("User registered successfully")

	// 返回註冊成功回應
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "用戶註冊成功",
		Data: models.AuthResponse{
			UserID:       user.ID,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    86400, // 24小時
		},
	})
}

// LoginUserReal godoc
// @Summary      用戶登入
// @Description  用戶身份驗證（真實資料庫版本）
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body models.UserLoginRequest true "登入憑證"
// @Success      200 {object} models.APIResponse{data=models.AuthResponse} "登入成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "認證失敗"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "Token 生成失敗"
// @Router       /user/login [post]
func LoginUserReal(c *gin.Context) {
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

	// 從資料庫獲取用戶
	user, err := database.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "用戶名或密碼錯誤",
			},
		})
		return
	}

	// 驗證密碼
	if err := database.VerifyPassword(user.Password, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "用戶名或密碼錯誤",
			},
		})
		return
	}

	// 更新最後登入時間
	if err := database.UpdateLastLogin(user.ID); err != nil {
		utils.LogError(err, "update last login", logrus.Fields{"user_id": user.ID})
	}

	// 生成 JWT Token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		utils.LogError(err, "generate access token", nil)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_ERROR",
				Message: "生成 Token 失敗",
			},
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.LogError(err, "generate refresh token", nil)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_ERROR",
				Message: "生成 Token 失敗",
			},
		})
		return
	}

	// 記錄登入事件
	utils.Logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	}).Info("User logged in successfully")

	// 返回登入成功回應
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "登入成功",
		Data: models.AuthResponse{
			UserID:       user.ID,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    86400, // 24小時
		},
	})
}

// GetProfileReal godoc
// @Summary      獲取用戶資料
// @Description  獲取當前登入用戶的個人資料（真實資料庫版本）
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse{data=models.UserProfile} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "用戶不存在"
// @Router       /user/profile [get]
func GetProfileReal(c *gin.Context) {
	// 從認證中間件獲取用戶ID
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_TOKEN",
				Message: "無效的 Token",
			},
		})
		return
	}

	// 從資料庫獲取用戶資料
	user, err := database.GetUserByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "用戶不存在",
			},
		})
		return
	}

	// 返回用戶資料（不包含密碼）
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取用戶資料成功",
		Data: map[string]interface{}{
			"user_id":     user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"nickname":    user.Nickname,
			"gender":      user.Gender,
			"birth_date":  user.BirthDate.Format("2006-01-02"),
			"avatar_url":  user.AvatarURL,
			"is_verified": user.IsVerified,
			"is_adult":    user.IsAdult,
			"created_at":  user.CreatedAt,
		},
	})
}