package handlers

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// AdminStatsResponse 系統統計響應結構
type AdminStatsResponse struct {
	Uptime            string `json:"uptime" example:"15天 8小時 32分鐘"`
	TotalRequests     string `json:"total_requests" example:"1,234,567"`
	ErrorRate         string `json:"error_rate" example:"0.02%"`
	AvgResponseTime   string `json:"avg_response_time" example:"125ms"`
	ActiveUsers       string `json:"active_users" example:"42"`
	DBConnections     string `json:"db_connections" example:"8/20"`
	MemoryUsage       string `json:"memory_usage" example:"256MB"`
	GoRoutines        string `json:"go_routines" example:"45"`
}

// AdminLogsResponse 系統日誌響應結構
type AdminLogsResponse struct {
	Logs []models.SystemLog `json:"logs"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

var (
	startTime    = time.Now()
	totalRequests int64
	errorCount   int64
	totalResponseTime int64
)

// GetAdminStats 獲取系統統計數據
// @Summary 獲取系統統計數據
// @Description 獲取系統運行統計信息，包括運行時間、請求數量、錯誤率等
// @Tags 管理系統
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.APIResponse{data=AdminStatsResponse} "統計數據"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 500 {object} models.APIResponse "服務器錯誤"
// @Router /admin/stats [get]
func GetAdminStats(c *gin.Context) {
	// 計算系統運行時間
	uptime := time.Since(startTime)
	uptimeStr := formatUptime(uptime)

	// 獲取記憶體使用情況
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsage := fmt.Sprintf("%.1fMB", float64(m.Alloc)/1024/1024)

	// 獲取 Goroutine 數量
	goroutines := fmt.Sprintf("%d", runtime.NumGoroutine())

	// 計算錯誤率
	var errorRate string
	if totalRequests > 0 {
		rate := float64(errorCount) / float64(totalRequests) * 100
		errorRate = fmt.Sprintf("%.2f%%", rate)
	} else {
		errorRate = "0.00%"
	}

	// 計算平均響應時間
	var avgResponseTime string
	if totalRequests > 0 {
		avg := totalResponseTime / totalRequests
		avgResponseTime = fmt.Sprintf("%dms", avg)
	} else {
		avgResponseTime = "0ms"
	}

	// 獲取活躍用戶數（模擬）
	activeUsers := "42"

	// 獲取資料庫連接數（模擬）
	dbConnections := "8/20"

	stats := AdminStatsResponse{
		Uptime:          uptimeStr,
		TotalRequests:   formatNumber(totalRequests),
		ErrorRate:       errorRate,
		AvgResponseTime: avgResponseTime,
		ActiveUsers:     activeUsers,
		DBConnections:   dbConnections,
		MemoryUsage:     memoryUsage,
		GoRoutines:      goroutines,
	}

	utils.Logger.WithFields(logrus.Fields{
		"type": "admin_stats_request",
		"uptime": uptimeStr,
		"memory": memoryUsage,
		"goroutines": goroutines,
	}).Info("Admin stats requested")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "系統統計獲取成功",
		Data:    stats,
	})
}

// GetAdminLogs 獲取系統日誌
// @Summary 獲取系統日誌
// @Description 獲取系統運行日誌，支持分頁和級別篩選
// @Tags 管理系統
// @Security BearerAuth
// @Produce json
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(50)
// @Param level query string false "日誌級別" Enums(debug,info,warning,error,all)
// @Success 200 {object} models.APIResponse{data=AdminLogsResponse} "日誌數據"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 500 {object} models.APIResponse "服務器錯誤"
// @Router /admin/logs [get]
func GetAdminLogs(c *gin.Context) {
	// 獲取查詢參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	level := c.Query("level")

	// 從內存中獲取最近的日誌（實際項目中應該從資料庫或日誌文件讀取）
	logs := getRecentLogs(page, limit, level)
	total := getTotalLogCount(level)

	response := AdminLogsResponse{
		Logs:  logs,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	utils.Logger.WithFields(logrus.Fields{
		"type": "admin_logs_request",
		"page": page,
		"limit": limit,
		"level": level,
		"total": total,
	}).Info("Admin logs requested")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "系統日誌獲取成功",
		Data:    response,
	})
}

// 輔助函數

// formatUptime 格式化運行時間
func formatUptime(duration time.Duration) string {
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%d天 %d小時 %d分鐘", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小時 %d分鐘", hours, minutes)
	} else {
		return fmt.Sprintf("%d分鐘", minutes)
	}
}

// formatNumber 格式化數字，添加千分位分隔符
func formatNumber(n int64) string {
	str := fmt.Sprintf("%d", n)
	result := ""
	
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	
	return result
}

// getRecentLogs 獲取最近的日誌
func getRecentLogs(page, limit int, level string) []models.SystemLog {
	logService := services.GetLogService()
	logs, _ := logService.GetLogs(page, limit, level)
	return logs
}

// getTotalLogCount 獲取日誌總數
func getTotalLogCount(level string) int {
	logService := services.GetLogService()
	_, total := logService.GetLogs(1, 1, level)
	return total
}

// IncrementRequestCount 增加請求計數
func IncrementRequestCount(isError bool, responseTime int64) {
	totalRequests++
	totalResponseTime += responseTime
	
	if isError {
		errorCount++
	}
}

// GetAdminUsers godoc
// @Summary      管理員獲取用戶列表
// @Description  獲取所有用戶列表，包含分頁和篩選功能
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(10)
// @Param        status query string false "用戶狀態篩選" Enums(active,inactive,banned)
// @Param        search query string false "搜索關鍵字（用戶名、郵箱）"
// @Success      200 {object} models.APIResponse{data=object} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      403 {object} models.APIResponse{error=models.APIError} "權限不足"
// @Router       /admin/users [get]
func GetAdminUsers(c *gin.Context) {
	ctx := context.Background()

	// TODO: 檢查管理員權限
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

	// 獲取查詢參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	status := c.Query("status")
	search := c.Query("search")

	// 設置分頁限制
	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}

	// 構建查詢
	query := database.DB.NewSelect().Model((*models.User)(nil))

	// 狀態篩選
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 搜索篩選
	if search != "" {
		query = query.Where("(username ILIKE ? OR email ILIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	// 計算總數
	total, err := query.Count(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to count users")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取用戶總數失敗",
			},
		})
		return
	}

	// 分頁查詢
	var users []models.User
	err = query.
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(ctx, &users)

	if err != nil {
		utils.Logger.WithError(err).WithField("admin_id", userID).Error("Failed to query users")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取用戶列表失敗",
			},
		})
		return
	}

	// 轉換為響應格式
	var userResponses []*models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	// 計算分頁信息
	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取用戶列表成功",
		Data: gin.H{
			"users": userResponses,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_count":  total,
				"limit":        limit,
			},
		},
	})
}

// UpdateAdminUser godoc
// @Summary      管理員更新用戶資料
// @Description  管理員更新指定用戶的資料信息
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "用戶ID"
// @Param        user body models.AdminUserUpdateRequest true "更新資料"
// @Success      200 {object} models.APIResponse{data=models.UserResponse} "更新成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      403 {object} models.APIResponse{error=models.APIError} "權限不足"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "用戶不存在"
// @Router       /admin/users/{id} [put]
func UpdateAdminUser(c *gin.Context) {
	ctx := context.Background()

	// TODO: 檢查管理員權限
	adminID, exists := c.Get("user_id")
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

	// 獲取用戶ID
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_USER_ID",
				Message: "用戶ID不能為空",
			},
		})
		return
	}

	var req models.AdminUserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "輸入參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 檢查用戶是否存在
	var user models.User
	err := database.DB.NewSelect().
		Model(&user).
		Where("id = ?", userID).
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

	// 檢查郵箱和用戶名唯一性（如果有更新）
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		exists, err := database.DB.NewSelect().
			Model(&existingUser).
			Where("email = ? AND id != ?", req.Email, userID).
			Exists(ctx)

		if err != nil {
			utils.Logger.WithError(err).Error("Failed to check email uniqueness")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "檢查郵箱唯一性失敗",
				},
			})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "EMAIL_EXISTS",
					Message: "郵箱已被其他用戶使用",
				},
			})
			return
		}
	}

	if req.Username != "" && req.Username != user.Username {
		var existingUser models.User
		exists, err := database.DB.NewSelect().
			Model(&existingUser).
			Where("username = ? AND id != ?", req.Username, userID).
			Exists(ctx)

		if err != nil {
			utils.Logger.WithError(err).Error("Failed to check username uniqueness")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "檢查用戶名唯一性失敗",
				},
			})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "USERNAME_EXISTS",
					Message: "用戶名已被其他用戶使用",
				},
			})
			return
		}
	}

	// 構建更新數據
	updateData := models.User{
		UpdatedAt: time.Now(),
	}

	// 構建動態更新查詢
	updateQuery := database.DB.NewUpdate().Model(&updateData)
	hasUpdates := false

	if req.Username != "" {
		updateData.Username = req.Username
		updateQuery = updateQuery.Column("username")
		hasUpdates = true
	}
	if req.Email != "" {
		updateData.Email = req.Email
		updateQuery = updateQuery.Column("email")
		hasUpdates = true
	}
	if req.DisplayName != nil {
		updateData.DisplayName = req.DisplayName
		updateQuery = updateQuery.Column("display_name")
		hasUpdates = true
	}
	if req.Bio != nil {
		updateData.Bio = req.Bio
		updateQuery = updateQuery.Column("bio")
		hasUpdates = true
	}
	if req.Status != "" {
		updateData.Status = req.Status
		updateQuery = updateQuery.Column("status")
		hasUpdates = true
	}
	if req.Nickname != "" {
		updateData.Nickname = req.Nickname
		updateQuery = updateQuery.Column("nickname")
		hasUpdates = true
	}
	if req.Gender != "" {
		updateData.Gender = req.Gender
		updateQuery = updateQuery.Column("gender")
		hasUpdates = true
	}
	if req.BirthDate != nil {
		updateData.BirthDate = req.BirthDate
		updateQuery = updateQuery.Column("birth_date")
		hasUpdates = true
	}
	if req.AvatarURL != "" {
		updateData.AvatarURL = req.AvatarURL
		updateQuery = updateQuery.Column("avatar_url")
		hasUpdates = true
	}
	if req.IsVerified != nil {
		updateData.IsVerified = *req.IsVerified
		updateQuery = updateQuery.Column("is_verified")
		hasUpdates = true
	}
	if req.IsAdult != nil {
		updateData.IsAdult = *req.IsAdult
		updateQuery = updateQuery.Column("is_adult")
		hasUpdates = true
	}
	if req.Preferences != nil {
		updateData.Preferences = req.Preferences
		updateQuery = updateQuery.Column("preferences")
		hasUpdates = true
	}

	if !hasUpdates {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NO_UPDATES",
				Message: "沒有要更新的字段",
			},
		})
		return
	}

	// 總是更新 updated_at
	updateQuery = updateQuery.Column("updated_at").Where("id = ?", userID)

	// 執行更新
	result, err := updateQuery.Exec(ctx)
	if err != nil {
		utils.Logger.WithError(err).WithFields(map[string]interface{}{
			"admin_id": adminID,
			"user_id":  userID,
		}).Error("Failed to update user")
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
				Message: "用戶不存在或未發生更新",
			},
		})
		return
	}

	// 獲取更新後的用戶信息
	err = database.DB.NewSelect().
		Model(&user).
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

	utils.Logger.WithFields(map[string]interface{}{
		"admin_id": adminID,
		"user_id":  userID,
	}).Info("Admin updated user profile")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "用戶資料更新成功",
		Data:    user.ToResponse(),
	})
}

// UpdateAdminUserPassword godoc
// @Summary      管理員重置用戶密碼
// @Description  管理員重置指定用戶的密碼
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "用戶ID"
// @Param        password body models.AdminPasswordUpdateRequest true "新密碼"
// @Success      200 {object} models.APIResponse "更新成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      403 {object} models.APIResponse{error=models.APIError} "權限不足"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "用戶不存在"
// @Router       /admin/users/{id}/password [put]
func UpdateAdminUserPassword(c *gin.Context) {
	ctx := context.Background()

	// TODO: 檢查管理員權限
	adminID, exists := c.Get("user_id")
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

	// 獲取用戶ID
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_USER_ID",
				Message: "用戶ID不能為空",
			},
		})
		return
	}

	var req models.AdminPasswordUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "輸入參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 檢查用戶是否存在
	var user models.User
	err := database.DB.NewSelect().
		Model(&user).
		Where("id = ?", userID).
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

	// 加密新密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
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

	// 更新密碼
	_, err = database.DB.NewUpdate().
		Model(&user).
		Set("password_hash = ?", string(hashedPassword)).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithFields(map[string]interface{}{
			"admin_id": adminID,
			"user_id":  userID,
		}).Error("Failed to update user password")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "更新用戶密碼失敗",
			},
		})
		return
	}

	utils.Logger.WithFields(map[string]interface{}{
		"admin_id": adminID,
		"user_id":  userID,
	}).Info("Admin updated user password")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "用戶密碼更新成功",
	})
}