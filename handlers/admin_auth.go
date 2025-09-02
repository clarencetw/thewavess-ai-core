package handlers

import (
	"net/http"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
)

// AdminLogin godoc
// @Summary      管理員登入
// @Description  管理員身份認證登入
// @Tags         管理員認證
// @Accept       json
// @Produce      json
// @Param        request body models.AdminLoginRequest true "登入資訊"
// @Success      200 {object} models.APIResponse{data=models.AdminLoginResponse} "登入成功"
// @Failure      400 {object} models.APIResponse "請求參數錯誤"
// @Failure      401 {object} models.APIResponse "認證失敗"
// @Failure      500 {object} models.APIResponse "伺服器錯誤"
// @Router       /admin/auth/login [post]
func AdminLogin(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "請求參數錯誤",
			Error: &models.APIError{
				Code:    "INVALID_REQUEST_BODY",
				Message: err.Error(),
			},
		})
		return
	}

	adminService := services.GetAdminService()
	admin, err := adminService.AuthenticateAdmin(c.Request.Context(), &req)
	if err != nil {
		// 記錄登入失敗事件
		clientIP := utils.GetClientIP(c)
		userAgent := c.Request.Header.Get("User-Agent")
		utils.LogAdminAction(
			"",
			req.Username,
			"LOGIN_FAILED",
			"",
			clientIP,
			userAgent,
			false,
			"管理員登入失敗: "+err.Error(),
		)
		
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "登入失敗",
			Error: &models.APIError{
				Code:    "AUTHENTICATION_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	// 生成管理員JWT令牌
	accessToken, err := utils.GenerateAdminAccessToken(
		admin.ID,
		admin.Username,
		admin.Email,
		admin.Role,
		admin.Permissions,
	)
	if err != nil {
		utils.Logger.WithError(err).Error("生成管理員JWT令牌失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "令牌生成失敗",
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_FAILED",
				Message: "無法生成認證令牌",
			},
		})
		return
	}

	// 記錄管理員登入事件
	clientIP := utils.GetClientIP(c)
	userAgent := c.Request.Header.Get("User-Agent")
	utils.LogAdminAction(
		admin.ID,
		admin.Email,
		"LOGIN",
		"",
		clientIP,
		userAgent,
		true,
		"管理員成功登入系統",
	)

	response := models.AdminLoginResponse{
		Admin:       *admin,
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   8 * 3600, // 8小時 (秒)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "管理員登入成功",
		Data:    response,
	})
}

// CreateAdmin godoc
// @Summary      創建管理員
// @Description  創建新的管理員帳號（需要超級管理員權限）
// @Tags         管理員管理
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body models.AdminCreateRequest true "管理員資訊"
// @Success      201 {object} models.APIResponse{data=models.Admin} "創建成功"
// @Failure      400 {object} models.APIResponse "請求參數錯誤"
// @Failure      401 {object} models.APIResponse "未授權"
// @Failure      403 {object} models.APIResponse "權限不足"
// @Failure      409 {object} models.APIResponse "管理員已存在"
// @Failure      500 {object} models.APIResponse "伺服器錯誤"
// @Router       /admin/admins [post]
func CreateAdmin(c *gin.Context) {
	var req models.AdminCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "請求參數錯誤",
			Error: &models.APIError{
				Code:    "INVALID_REQUEST_BODY",
				Message: err.Error(),
			},
		})
		return
	}

	// 獲取創建者信息
	creatorID, _ := c.Get("admin_id")
	creatorEmail, _ := c.Get("admin_email")
	req.CreatedBy = creatorID.(string)

	adminService := services.GetAdminService()
	admin, err := adminService.CreateAdmin(c.Request.Context(), &req)
	if err != nil {
		// 判斷錯誤類型
		statusCode := http.StatusInternalServerError
		errorCode := "ADMIN_CREATION_FAILED"

		// 這裡可以根據具體錯誤類型設定不同的狀態碼
		if contains := (err.Error() == "管理員用戶名或郵箱已存在"); contains {
			statusCode = http.StatusConflict
			errorCode = "ADMIN_ALREADY_EXISTS"
		}

		// 記錄創建管理員失敗事件
		clientIP := utils.GetClientIP(c)
		userAgent := c.Request.Header.Get("User-Agent")
		utils.LogAdminAction(
			creatorID.(string),
			creatorEmail.(string),
			"CREATE_ADMIN_FAILED",
			req.Email,
			clientIP,
			userAgent,
			false,
			"創建管理員失敗: "+err.Error(),
		)

		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Message: "創建管理員失敗",
			Error: &models.APIError{
				Code:    errorCode,
				Message: err.Error(),
			},
		})
		return
	}

	// 記錄創建管理員成功事件
	clientIP := utils.GetClientIP(c)
	userAgent := c.Request.Header.Get("User-Agent")
	utils.LogAdminAction(
		creatorID.(string),
		creatorEmail.(string),
		"CREATE_ADMIN",
		admin.Email,
		clientIP,
		userAgent,
		true,
		"成功創建管理員: "+admin.Username,
	)

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "管理員創建成功",
		Data:    admin,
	})
}

// GetAdminList godoc
// @Summary      獲取管理員列表
// @Description  支援分頁和篩選（需系統管理員權限）
// @Tags         管理員管理
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        page query int false "頁數" default(1)
// @Param        page_size query int false "每頁數量" default(20)
// @Param        role query string false "角色篩選"
// @Param        status query string false "狀態篩選"
// @Param        search query string false "搜尋"
// @Success      200 {object} models.APIResponse{data=models.AdminListResponse} "獲取成功"
// @Failure      400 {object} models.APIResponse "請求參數錯誤"
// @Failure      401 {object} models.APIResponse "未授權"
// @Failure      403 {object} models.APIResponse "權限不足"
// @Failure      500 {object} models.APIResponse "伺服器錯誤"
// @Router       /admin/admins [get]
func GetAdminList(c *gin.Context) {
	var query models.AdminListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "查詢參數錯誤",
			Error: &models.APIError{
				Code:    "INVALID_QUERY_PARAMS",
				Message: err.Error(),
			},
		})
		return
	}

	adminService := services.GetAdminService()
	admins, pagination, err := adminService.ListAdmins(c.Request.Context(), &query)
	if err != nil {
		utils.Logger.WithError(err).Error("獲取管理員列表失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "獲取管理員列表失敗",
			Error: &models.APIError{
				Code:    "ADMIN_LIST_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	response := models.AdminListResponse{
		Admins:     admins,
		Pagination: *pagination,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取管理員列表成功",
		Data:    response,
	})
}

// GetAdminProfile godoc
// @Summary      獲取管理員資料
// @Description  獲取當前登入管理員的個人資料
// @Tags         管理員認證
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200 {object} models.APIResponse{data=models.Admin} "獲取成功"
// @Failure      401 {object} models.APIResponse "未授權"
// @Failure      404 {object} models.APIResponse "管理員不存在"
// @Failure      500 {object} models.APIResponse "伺服器錯誤"
// @Router       /admin/profile [get]
func GetAdminProfile(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "管理員認證失敗",
			Error: &models.APIError{
				Code:    "ADMIN_AUTH_FAILED",
				Message: "無法獲取管理員身份信息",
			},
		})
		return
	}

	adminService := services.GetAdminService()
	admin, err := adminService.GetAdmin(c.Request.Context(), adminID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "管理員不存在",
			Error: &models.APIError{
				Code:    "ADMIN_NOT_FOUND",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取管理員資料成功",
		Data:    admin,
	})
}
