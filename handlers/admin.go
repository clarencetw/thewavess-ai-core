package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"golang.org/x/crypto/bcrypt"
)

// AdminStatsResponse 系統統計響應結構
type AdminStatsResponse struct {
	Uptime          string `json:"uptime" example:"15天 8小時 32分鐘"`
	TotalRequests   string `json:"total_requests" example:"1,234,567"`
	ErrorRate       string `json:"error_rate" example:"0.02%"`
	AvgResponseTime string `json:"avg_response_time" example:"125ms"`
	ActiveUsers     string `json:"active_users" example:"42"`
	DBConnections   string `json:"db_connections" example:"8/20"`
	MemoryUsage     string `json:"memory_usage" example:"256MB"`
	GoRoutines      string `json:"go_routines" example:"45"`
}

// AdminLogsResponse 管理員日誌響應結構
type AdminLogsResponse struct {
	Logs  []models.SystemLog `json:"logs"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

var (
	startTime         = time.Now()
	totalRequests     int64
	errorCount        int64
	totalResponseTime int64
)

// GetAdminStats 獲取系統統計數據
// @Summary 獲取系統統計數據
// @Description 獲取系統運行統計信息，包括運行時間、請求數量、錯誤率等
// @Tags 管理系統
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=AdminStatsResponse} "統計數據"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 500 {object} models.APIResponse "服務器錯誤"
// @Router /admin/stats [get]
func GetAdminStats(c *gin.Context) {
	ctx := context.Background()
	dbConn := GetDB()

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

	var (
		totalUsers, todayUsers, weekUsers, activeUsers, blockedUsers int
		totalCharacters                                              int
		totalSessions, todaySessions                                 int
		totalMessages, todayMessages                                 int
		openaiRequests, grokRequests, otherRequests                  int
		openaiRequests24h, grokRequests24h, otherRequests24h         int
	)

	// 使用產品主要時區，確保「今日」等統計符合營運期望的日界線
	timezoneName := utils.GetEnvWithDefault("APP_TIMEZONE", "Asia/Taipei")
	loc, err := time.LoadLocation(timezoneName)
	if err != nil {
		loc = time.Local
	}
	nowInLoc := time.Now().In(loc)
	todayStartLocal := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), 0, 0, 0, 0, loc)
	todayStart := todayStartLocal.UTC()
	weekStart := todayStart.AddDate(0, 0, -7)
	last24h := time.Now().UTC().Add(-24 * time.Hour)

	if dbConn != nil {
		// 用戶統計（排除已軟刪除帳號）
		baseUserQuery := dbConn.NewSelect().Model((*db.UserDB)(nil)).Where("deleted_at IS NULL")
		if totalUsers, err = baseUserQuery.Clone().Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count total users")
		}
		if todayUsers, err = baseUserQuery.Clone().Where("created_at >= ?", todayStart).Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count today users")
		}
		if weekUsers, err = baseUserQuery.Clone().Where("created_at >= ?", weekStart).Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count week users")
		}
		if activeUsers, err = baseUserQuery.Clone().
			Where("last_login_at >= ?", weekStart).
			Where("status = ?", "active").
			Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count active users")
		}
		if blockedUsers, err = baseUserQuery.Clone().
			Where("status = ?", "banned").
			Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count blocked users")
		}

		// 角色、聊天、訊息
		if totalCharacters, err = dbConn.NewSelect().Model((*db.CharacterDB)(nil)).
			Where("is_active = ?", true).
			Where("deleted_at IS NULL").
			Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count characters")
		}
		if totalSessions, err = dbConn.NewSelect().Model((*db.ChatDB)(nil)).
			Where("status != ?", "deleted").
			Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count chat sessions")
		}
		if todaySessions, err = dbConn.NewSelect().Model((*db.ChatDB)(nil)).
			Where("created_at >= ?", todayStart).
			Where("status != ?", "deleted").
			Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count today chat sessions")
		}
		if totalMessages, err = dbConn.NewSelect().Model((*db.MessageDB)(nil)).Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count total messages")
		}
		if todayMessages, err = dbConn.NewSelect().Model((*db.MessageDB)(nil)).
			Where("created_at >= ?", todayStart).
			Count(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Failed to count today messages")
		}

		// AI 引擎使用統計
		openaiRequests, _ = dbConn.NewSelect().Model((*db.MessageDB)(nil)).
			Where("role = ?", "assistant").
			Where("ai_engine = ?", "openai").
			Count(ctx)

		grokRequests, _ = dbConn.NewSelect().Model((*db.MessageDB)(nil)).
			Where("role = ?", "assistant").
			Where("ai_engine = ?", "grok").
			Count(ctx)

		otherRequests, _ = dbConn.NewSelect().Model((*db.MessageDB)(nil)).
			Where("role = ?", "assistant").
			Where("ai_engine NOT IN (?)", bun.In([]string{"openai", "grok"})).
			Count(ctx)

		openaiRequests24h, _ = dbConn.NewSelect().Model((*db.MessageDB)(nil)).
			Where("role = ?", "assistant").
			Where("ai_engine = ?", "openai").
			Where("created_at >= ?", last24h).
			Count(ctx)

		grokRequests24h, _ = dbConn.NewSelect().Model((*db.MessageDB)(nil)).
			Where("role = ?", "assistant").
			Where("ai_engine = ?", "grok").
			Where("created_at >= ?", last24h).
			Count(ctx)

		otherRequests24h, _ = dbConn.NewSelect().Model((*db.MessageDB)(nil)).
			Where("role = ?", "assistant").
			Where("ai_engine NOT IN (?)", bun.In([]string{"openai", "grok"})).
			Where("created_at >= ?", last24h).
			Count(ctx)
	}

	totalAIRequests := openaiRequests + grokRequests + otherRequests
	totalAIRequests24h := openaiRequests24h + grokRequests24h + otherRequests24h

	var openaiPercentage, grokPercentage float64
	if totalAIRequests > 0 {
		openaiPercentage = math.Round((float64(openaiRequests)/float64(totalAIRequests))*100*100) / 100
		grokPercentage = math.Round((float64(grokRequests)/float64(totalAIRequests))*100*100) / 100
	}

	aiEngines := gin.H{
		"total_requests":  totalAIRequests,
		"openai_requests": openaiRequests,
		"grok_requests":   grokRequests,
		"last_24h": gin.H{
			"total":  totalAIRequests24h,
			"openai": openaiRequests24h,
			"grok":   grokRequests24h,
		},
		"breakdown": gin.H{
			"openai": gin.H{
				"requests":   openaiRequests,
				"percentage": openaiPercentage,
				"last_24h":   openaiRequests24h,
			},
			"grok": gin.H{
				"requests":   grokRequests,
				"percentage": grokPercentage,
				"last_24h":   grokRequests24h,
			},
		},
	}
	if otherRequests > 0 {
		aiEngines["other_requests"] = otherRequests
	}
	if otherRequests24h > 0 {
		aiEngines["last_24h"].(gin.H)["other"] = otherRequests24h
	}

	aiConfig := gin.H{
		"openai": gin.H{
			"model":       utils.GetEnvWithDefault("OPENAI_MODEL", "gpt-4o"),
			"temperature": utils.GetEnvFloatWithDefault("OPENAI_TEMPERATURE", 0.8),
			"max_tokens":  utils.GetEnvIntWithDefault("OPENAI_MAX_TOKENS", 1200),
			"base_url":    utils.GetEnvWithDefault("OPENAI_API_URL", "https://api.openai.com/v1"),
			"enabled":     utils.GetEnvWithDefault("OPENAI_API_KEY", "") != "",
		},
		"grok": gin.H{
			"model":       utils.GetEnvWithDefault("GROK_MODEL", "grok-3"),
			"temperature": utils.GetEnvFloatWithDefault("GROK_TEMPERATURE", 0.9),
			"max_tokens":  utils.GetEnvIntWithDefault("GROK_MAX_TOKENS", 2000),
			"base_url":    utils.GetEnvWithDefault("GROK_API_URL", "https://api.x.ai/v1"),
			"enabled":     utils.GetEnvWithDefault("GROK_API_KEY", "") != "",
		},
	}

	mistralModel := utils.GetEnvWithDefault("MISTRAL_MODEL", "")
	if mistralModel != "" || utils.GetEnvWithDefault("MISTRAL_API_KEY", "") != "" {
		aiConfig["mistral"] = gin.H{
			"model":       mistralModel,
			"temperature": utils.GetEnvFloatWithDefault("MISTRAL_TEMPERATURE", 0.8),
			"max_tokens":  utils.GetEnvIntWithDefault("MISTRAL_MAX_TOKENS", 1200),
			"base_url":    utils.GetEnvWithDefault("MISTRAL_API_URL", ""),
			"enabled":     utils.GetEnvWithDefault("MISTRAL_API_KEY", "") != "",
		}
	}

	// 構建完整的統計響應
	stats := gin.H{
		// 基本系統統計
		"uptime":            uptimeStr,
		"total_requests":    formatNumber(totalRequests),
		"error_rate":        errorRate,
		"avg_response_time": avgResponseTime,
		"memory_usage":      memoryUsage,
		"go_routines":       goroutines,

		// 業務統計
		"users": gin.H{
			"total":     totalUsers,
			"today_new": todayUsers,
			"week_new":  weekUsers,
			"active_7d": activeUsers,
			"blocked":   blockedUsers,
		},

		"characters": gin.H{
			"total":  totalCharacters,
			"active": totalCharacters,
		},

		"chats": gin.H{
			"total_sessions": totalSessions,
			"today_sessions": todaySessions,
			"total_messages": totalMessages,
			"today_messages": todayMessages,
		},

		// 系統服務狀態
		"services": gin.H{
			"database": getDatabaseStatus(ctx),
			"ai_services": gin.H{
				"openai": getOpenAIStatus(),
				"grok":   getGrokStatus(),
			},
			"memory_system": gin.H{
				"status":         "healthy",
				"total_memories": 0, // 記憶功能已整合到relationships.emotion_data
			},
		},

		// 系統性能指標
		"performance": gin.H{
			"memory_allocated": memoryUsage,
			"memory_sys":       fmt.Sprintf("%.1fMB", float64(m.Sys)/1024/1024),
			"gc_count":         m.NumGC,
			"goroutines":       goroutines,
		},
		"ai_engines": aiEngines,
		"ai_config":  aiConfig,

		// 時間戳
		"last_updated": time.Now(),
		"timezone":     timezoneName,
	}

	utils.Logger.WithFields(logrus.Fields{
		"type":        "admin_stats_request",
		"uptime":      uptimeStr,
		"memory":      memoryUsage,
		"goroutines":  goroutines,
		"total_users": totalUsers,
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
// @Accept json
// @Produce json
// @Security BearerAuth
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
		"type":  "admin_logs_request",
		"page":  page,
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

// GetAdminUserByID 獲取單個用戶詳情
// @Summary 獲取用戶詳情
// @Description 管理員查看指定用戶詳細資訊
// @Tags 管理員 - 用戶管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用戶ID"
// @Success 200 {object} models.APIResponse{data=models.UserResponse} "用戶詳情"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 404 {object} models.APIResponse "用戶不存在"
// @Router /admin/users/{id} [get]
func GetAdminUserByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "用戶ID為必填項目",
			Error: &models.APIError{
				Code:    "USER_ID_REQUIRED",
				Message: "用戶ID為必填項目",
			},
		})
		return
	}

	ctx := context.Background()
	var user db.UserDB

	err := services.GetDB().NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("獲取用戶詳情失敗")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "用戶不存在",
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "用戶不存在",
			},
		})
		return
	}

	userResponse := models.UserFromDB(&user).ToResponse()

	utils.Logger.WithFields(logrus.Fields{
		"user_id": userID,
		"admin":   c.GetString("admin_id"),
	}).Info("管理員獲取用戶詳情")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取用戶詳情成功",
		Data:    userResponse,
	})
}

// UpdateAdminUserStatus 更新用戶狀態
// @Summary 更新用戶狀態
// @Description 管理員更新用戶狀態（封鎖/解封）
// @Tags 管理員 - 用戶管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用戶ID"
// @Param request body object{status=string} true "狀態更新請求"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.APIResponse "請求參數錯誤"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 404 {object} models.APIResponse "用戶不存在"
// @Router /admin/users/{id}/status [put]
func UpdateAdminUserStatus(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "用戶ID為必填項目",
			Error: &models.APIError{
				Code:    "USER_ID_REQUIRED",
				Message: "用戶ID為必填項目",
			},
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive suspended"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "無效的輸入格式",
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "無效的輸入格式",
			},
		})
		return
	}

	ctx := context.Background()

	// 檢查用戶是否存在
	exists, err := services.GetDB().NewSelect().
		Model((*db.UserDB)(nil)).
		Where("id = ?", userID).
		Exists(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("檢查用戶存在性失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "資料庫錯誤",
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "資料庫錯誤",
			},
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "用戶不存在",
			Error: &models.APIError{
				Code:    "USER_NOT_FOUND",
				Message: "用戶不存在",
			},
		})
		return
	}

	// 更新用戶狀態
	_, err = services.GetDB().NewUpdate().
		Model((*db.UserDB)(nil)).
		Set("status = ?", req.Status).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("更新用戶狀態失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "更新用戶狀態失敗",
			Error: &models.APIError{
				Code:    "UPDATE_FAILED",
				Message: "更新用戶狀態失敗",
			},
		})
		return
	}

	utils.Logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"new_status": req.Status,
		"admin":      c.GetString("admin_id"),
	}).Info("管理員更新用戶狀態")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("用戶狀態已更新為：%s", req.Status),
		Data:    nil,
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

// getDatabaseStatus 獲取資料庫狀態
func getDatabaseStatus(ctx context.Context) gin.H {
	start := time.Now()

	// 測試資料庫連接
	var count int
	err := GetDB().NewSelect().ColumnExpr("1").Scan(ctx, &count)

	pingTime := time.Since(start)

	if err != nil {
		return gin.H{
			"status":     "error",
			"ping_time":  "N/A",
			"error":      err.Error(),
			"last_check": time.Now().Format("15:04:05"),
		}
	}

	return gin.H{
		"status":     "healthy",
		"ping_time":  fmt.Sprintf("%.2fms", float64(pingTime.Nanoseconds())/1000000),
		"last_check": time.Now().Format("15:04:05"),
	}
}

// getOpenAIStatus 獲取 OpenAI 服務狀態
func getOpenAIStatus() gin.H {
	openaiKey := utils.GetEnvWithDefault("OPENAI_API_KEY", "")

	if openaiKey == "" {
		return gin.H{
			"status":     "error",
			"error":      "API key not configured",
			"last_check": time.Now().Format("15:04:05"),
		}
	}

	return gin.H{
		"status":     "healthy",
		"api_key":    "configured",
		"last_check": time.Now().Format("15:04:05"),
	}
}

// getGrokStatus 獲取 Grok 服務狀態
func getGrokStatus() gin.H {
	grokKey := utils.GetEnvWithDefault("GROK_API_KEY", "")

	if grokKey == "" {
		return gin.H{
			"status":     "error",
			"error":      "API key not configured",
			"last_check": time.Now().Format("15:04:05"),
		}
	}

	return gin.H{
		"status":     "healthy",
		"api_key":    "configured",
		"last_check": time.Now().Format("15:04:05"),
	}
}

// GetAdminUsers godoc
// @Summary      獲取用戶列表
// @Description  管理員用：支援分頁和篩選的用戶列表
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

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")

	// 設置分頁限制
	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}

	// 構建查詢
	query := GetDB().NewSelect().Model((*db.UserDB)(nil))

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

	// 處理排序
	orderClause := "created_at DESC" // 默認排序
	if sortBy != "" {
		// 允許的排序字段
		validSortFields := map[string]string{
			"username":   "username",
			"email":      "email",
			"status":     "status",
			"created_at": "created_at",
			"updated_at": "updated_at",
		}

		if field, valid := validSortFields[sortBy]; valid {
			direction := "DESC"
			if sortOrder == "asc" {
				direction = "ASC"
			}
			orderClause = field + " " + direction
		}
	}

	// 分頁查詢
	var users []db.UserDB
	err = query.
		Order(orderClause).
		Offset((page-1)*limit).
		Limit(limit).
		Scan(ctx, &users)

	if err != nil {
		utils.Logger.WithError(err).WithField("admin_id", adminID).Error("Failed to query users")
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
	for i := range users {
		userModel := models.UserFromDB(&users[i])
		if userModel == nil {
			continue
		}
		userResponses = append(userResponses, userModel.ToResponse())
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
// @Summary      更新用戶資料
// @Description  管理員用：更新指定用戶的資料信息
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

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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
	var user db.UserDB
	err := GetDB().NewSelect().
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
		var existingUser db.UserDB
		exists, err := GetDB().NewSelect().
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
		var existingUser db.UserDB
		exists, err := GetDB().NewSelect().
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
	updateData := db.UserDB{
		UpdatedAt: time.Now(),
	}

	// 構建動態更新查詢
	updateQuery := GetDB().NewUpdate().Model(&updateData)
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
		updateData.Nickname = &req.Nickname
		updateQuery = updateQuery.Column("nickname")
		hasUpdates = true
	}
	if req.Gender != "" {
		updateData.Gender = &req.Gender
		updateQuery = updateQuery.Column("gender")
		hasUpdates = true
	}
	if req.BirthDate != nil {
		updateData.BirthDate = req.BirthDate
		updateQuery = updateQuery.Column("birth_date")
		hasUpdates = true
	}
	if req.AvatarURL != "" {
		updateData.AvatarURL = &req.AvatarURL
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
	err = GetDB().NewSelect().
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

	userResponse := models.UserFromDB(&user).ToResponse()

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "用戶資料更新成功",
		Data:    userResponse,
	})
}

// UpdateAdminUserPassword godoc
// @Summary      重置用戶密碼
// @Description  管理員用：重置指定用戶的密碼
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

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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
	var user db.UserDB
	err := GetDB().NewSelect().
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
	_, err = GetDB().NewUpdate().
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

// AdminSearchChats godoc
// @Summary      搜尋聊天記錄
// @Description  管理員用：支援全局搜尋和空查詢的聊天記錄
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        query query string false "搜尋關鍵詞（空值則返回所有記錄）"
// @Param        user_id query string false "特定用戶ID過濾"
// @Param        character_id query string false "角色ID過濾"
// @Param        date_from query string false "開始日期"
// @Param        date_to query string false "結束日期"
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(50)
// @Param        include_user_info query bool false "是否包含用戶信息" default(true)
// @Success      200 {object} models.APIResponse "搜尋成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      403 {object} models.APIResponse{error=models.APIError} "權限不足"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "服務器錯誤"
// @Router       /admin/chats/search [get]
func AdminSearchChats(c *gin.Context) {
	startTime := time.Now()
	ctx := context.Background()

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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
	query := c.Query("query")              // 管理員可以使用空查詢
	userIDFilter := c.Query("user_id")     // 特定用戶過濾
	characterID := c.Query("character_id") // 角色過濾
	dateFrom := c.Query("date_from")       // 日期範圍
	dateTo := c.Query("date_to")
	sortBy := c.Query("sort_by")                               // 排序字段
	sortOrder := c.Query("sort_order")                         // 排序方向
	includeUserInfo := c.Query("include_user_info") != "false" // 默認包含用戶信息

	// 解析分頁參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50 // 管理員默認較高的限制
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	// 執行管理員專用搜尋
	results, totalCount, err := adminSearchChatSessions(ctx, query, userIDFilter, characterID, dateFrom, dateTo, sortBy, sortOrder, page, limit, includeUserInfo)
	if err != nil {
		utils.Logger.WithError(err).WithField("admin_id", adminID).Error("管理員搜尋聊天記錄失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SEARCH_ERROR",
				Message: "搜尋失敗",
			},
		})
		return
	}

	totalPages := (totalCount + limit - 1) / limit
	searchTime := time.Since(startTime)

	utils.Logger.WithFields(map[string]interface{}{
		"admin_id":    adminID,
		"query":       query,
		"user_filter": userIDFilter,
		"total_found": totalCount,
		"search_time": searchTime.Milliseconds(),
	}).Info("管理員執行聊天記錄搜尋")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "搜尋成功",
		Data: gin.H{
			"chats":        results,
			"total_found":  totalCount,
			"current_page": page,
			"total_pages":  totalPages,
			"search_time":  fmt.Sprintf("%dms", searchTime.Milliseconds()),
			"filters": gin.H{
				"query":             query,
				"user_id":           userIDFilter,
				"character_id":      characterID,
				"date_from":         dateFrom,
				"date_to":           dateTo,
				"include_user_info": includeUserInfo,
			},
		},
	})
}

// adminSearchChatSessions 管理員專用的聊天會話搜尋函數
func adminSearchChatSessions(ctx context.Context, query, userIDFilter, characterID, dateFrom, dateTo, sortBy, sortOrder string, page, limit int, includeUserInfo bool) ([]gin.H, int, error) {
	database := GetDB()
	if database == nil {
		return nil, 0, fmt.Errorf("database connection unavailable")
	}

	// 構建基礎查詢 - 直接查詢聊天會話而不是消息
	baseQuery := database.NewSelect().
		Model((*db.ChatDB)(nil)).
		Column("c.id", "c.title", "c.character_id", "c.user_id", "c.created_at", "c.updated_at", "c.status").
		Column("char.name", "char.avatar_url").
		Join("JOIN characters char ON char.id = c.character_id").
		Where("c.status != ?", "deleted")

	// 如果有用戶過濾
	if userIDFilter != "" {
		baseQuery = baseQuery.Where("c.user_id = ?", userIDFilter)
	}

	// 如果有角色過濾
	if characterID != "" {
		baseQuery = baseQuery.Where("c.character_id = ?", characterID)
	}

	// 日期範圍過濾
	if dateFrom != "" {
		baseQuery = baseQuery.Where("c.created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		baseQuery = baseQuery.Where("c.created_at <= ?", dateTo)
	}

	// 如果有搜尋查詢，需要在相關的消息中搜尋
	if query != "" && query != "*" {
		// 子查詢：找到包含關鍵詞的消息對應的聊天會話ID
		subQuery := database.NewSelect().
			Model((*db.MessageDB)(nil)).
			Column("DISTINCT chat_id").
			Where("to_tsvector('simple', dialogue) @@ plainto_tsquery('simple', ?)", query)

		baseQuery = baseQuery.Where("c.id IN (?)", subQuery)
	}

	// 計算總數
	countQuery := baseQuery
	totalCount, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count chat sessions: %w", err)
	}

	// 分頁和排序
	offset := (page - 1) * limit
	var results []struct {
		db.ChatDB
		CharacterName   string `bun:"name"`
		CharacterAvatar string `bun:"avatar_url"`
	}

	// 處理排序
	orderClause := "c.updated_at DESC" // 默認以最近更新排序
	if sortBy != "" {
		validSortFields := map[string]string{
			"title":          "c.title",
			"username":       "", // 需要透過 JOIN 處理
			"character_name": "char.name",
			"created_at":     "c.created_at",
			"updated_at":     "c.updated_at",
		}

		if field, valid := validSortFields[sortBy]; valid && field != "" {
			direction := "DESC"
			if sortOrder == "asc" {
				direction = "ASC"
			}
			orderClause = field + " " + direction
		}
	}

	err = baseQuery.
		Order(orderClause).
		Limit(limit).
		Offset(offset).
		Scan(ctx, &results)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute admin chat search: %w", err)
	}

	// 構建響應數據
	chatResults := make([]gin.H, len(results))
	for i, result := range results {
		chatData := gin.H{
			"id":               result.ID,
			"title":            result.Title,
			"character_id":     result.CharacterID,
			"character_name":   result.CharacterName,
			"character_avatar": result.CharacterAvatar,
			"status":           result.Status,
			"created_at":       result.CreatedAt,
			"updated_at":       result.UpdatedAt,
		}

		// 根據需要添加用戶信息
		if includeUserInfo {
			// 獲取用戶信息
			var user db.UserDB
			err := database.NewSelect().
				Model(&user).
				Column("id", "username", "display_name", "email").
				Where("id = ?", result.UserID).
				Scan(ctx)

			if err == nil {
				chatData["user"] = gin.H{
					"id":           user.ID,
					"username":     user.Username,
					"display_name": user.DisplayName,
					"email":        user.Email,
				}
			} else {
				chatData["user"] = gin.H{
					"id":       result.UserID,
					"username": "未知用戶",
				}
			}
		} else {
			chatData["user_id"] = result.UserID
		}

		// 獲取消息統計
		var messageCount int
		messageCount, _ = database.NewSelect().
			Model((*db.MessageDB)(nil)).
			Where("chat_id = ?", result.ID).
			Count(ctx)

		chatData["message_count"] = messageCount

		// 獲取最後一條消息預覽
		var lastMessage db.MessageDB
		err = database.NewSelect().
			Model(&lastMessage).
			Where("chat_id = ?", result.ID).
			Order("created_at DESC").
			Limit(1).
			Scan(ctx)

		if err == nil {
			preview := lastMessage.Dialogue
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			chatData["last_message"] = gin.H{
				"content":    preview,
				"created_at": lastMessage.CreatedAt,
				"role":       lastMessage.Role,
			}
		}

		// 獲取關係狀態資料
		var relationship db.RelationshipDB
		err = database.NewSelect().
			Model(&relationship).
			Where("user_id = ? AND character_id = ? AND chat_id = ?", result.UserID, result.CharacterID, result.ID).
			Scan(ctx)

		if err == nil {
			// 直接使用 AI 設定的關係狀態，不進行轉換
			chatData["relationship"] = gin.H{
				"affection_level":    relationship.Affection,
				"relationship_stage": relationship.Relationship, // 直接使用資料庫中的關係狀態
				"mood":               relationship.Mood,
				"intimacy_level":     relationship.IntimacyLevel,
				"total_interactions": relationship.TotalInteractions,
			}
		} else {
			// 如果沒有關係記錄，使用預設值
			chatData["relationship"] = gin.H{
				"affection_level":    0,
				"relationship_stage": "stranger", // 使用預設的英文值，讓前端處理顯示
				"mood":               "neutral",
				"intimacy_level":     "distant",
				"total_interactions": 0,
			}
		}

		chatResults[i] = chatData
	}

	return chatResults, totalCount, nil
}

// AdminGetChatHistory godoc
// @Summary      獲取聊天記錄
// @Description  管理員用：獲取聊天記錄，無需用戶認證
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "聊天會話ID"
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(50)
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      403 {object} models.APIResponse{error=models.APIError} "權限不足"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "聊天會話不存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "服務器錯誤"
// @Router       /admin/chats/{chat_id}/history [get]
func AdminGetChatHistory(c *gin.Context) {
	startTime := time.Now()
	ctx := context.Background()

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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

	// 獲取聊天會話ID
	chatID := c.Param("chat_id")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CHAT_ID",
				Message: "聊天會話ID不能為空",
			},
		})
		return
	}

	// 解析分頁參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50 // 管理員默認較高的限制
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	// 執行管理員專用聊天記錄獲取
	messages, totalCount, sessionInfo, err := adminGetChatMessages(ctx, chatID, page, limit)
	if err != nil {
		if err.Error() == "chat session not found" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "CHAT_NOT_FOUND",
					Message: "聊天會話不存在",
				},
			})
			return
		}

		utils.Logger.WithError(err).WithFields(map[string]interface{}{
			"admin_id": adminID,
			"chat_id":  chatID,
		}).Error("管理員獲取聊天記錄失敗")

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SERVER_ERROR",
				Message: "獲取聊天記錄失敗",
			},
		})
		return
	}

	totalPages := (totalCount + limit - 1) / limit
	searchTime := time.Since(startTime)

	utils.Logger.WithFields(map[string]interface{}{
		"admin_id":    adminID,
		"chat_id":     chatID,
		"total_found": totalCount,
		"search_time": searchTime.Milliseconds(),
	}).Info("管理員獲取聊天記錄成功")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取聊天記錄成功",
		Data: gin.H{
			"messages":     messages,
			"session_info": sessionInfo,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_count":  totalCount,
				"limit":        limit,
			},
			"fetch_time": fmt.Sprintf("%dms", searchTime.Milliseconds()),
		},
	})
}

// adminGetChatMessages 管理員專用的聊天消息獲取函數
func adminGetChatMessages(ctx context.Context, chatID string, page, limit int) ([]gin.H, int, gin.H, error) {
	database := GetDB()
	if database == nil {
		return nil, 0, nil, fmt.Errorf("database connection unavailable")
	}

	// 驗證聊天會話是否存在
	var session db.ChatDB
	err := database.NewSelect().
		Model(&session).
		Where("id = ? AND status != ?", chatID, "deleted").
		Scan(ctx)

	if err != nil {
		return nil, 0, nil, fmt.Errorf("chat session not found")
	}

	// 獲取角色信息
	var character db.CharacterDB
	err = database.NewSelect().
		Model(&character).
		Where("id = ?", session.CharacterID).
		Scan(ctx)

	if err != nil {
		return nil, 0, nil, fmt.Errorf("character not found")
	}

	// 獲取用戶信息
	var user db.UserDB
	err = database.NewSelect().
		Model(&user).
		Column("id", "username", "display_name", "email").
		Where("id = ?", session.UserID).
		Scan(ctx)

	if err != nil {
		return nil, 0, nil, fmt.Errorf("user not found")
	}

	// 構建會話信息
	sessionInfo := gin.H{
		"id":         session.ID,
		"title":      session.Title,
		"status":     session.Status,
		"created_at": session.CreatedAt,
		"updated_at": session.UpdatedAt,
		"character": gin.H{
			"id":         character.ID,
			"name":       character.Name,
			"avatar_url": character.AvatarURL,
		},
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"email":        user.Email,
		},
	}

	// 獲取消息總數
	totalCount, err := database.NewSelect().
		Model((*db.MessageDB)(nil)).
		Where("chat_id = ?", chatID).
		Count(ctx)

	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to count messages: %w", err)
	}

	// 獲取分頁消息
	offset := (page - 1) * limit
	var messages []db.MessageDB
	err = database.NewSelect().
		Model(&messages).
		Where("chat_id = ?", chatID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	// 轉換為響應格式
	messageResults := make([]gin.H, len(messages))
	for i, msg := range messages {
		messageResults[i] = gin.H{
			"id":                msg.ID,
			"role":              msg.Role,
			"dialogue":          msg.Dialogue,
			"scene_description": msg.SceneDescription,
			"action":            msg.Action,
			"nsfw_level":        msg.NSFWLevel,
			"ai_engine":         msg.AIEngine,
			"response_time_ms":  msg.ResponseTimeMs,
			"created_at":        msg.CreatedAt,
		}
	}

	return messageResults, totalCount, sessionInfo, nil
}

// UpdateCharacterStatus 更新角色狀態
// @Summary 更新角色狀態
// @Description 管理員更新角色的啟用/停用狀態
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Param request body object{is_active=bool} true "狀態更新請求"
// @Success 200 {object} models.APIResponse{data=object{character_id=string,is_active=bool}} "更新成功"
// @Failure 400 {object} models.APIResponse "請求參數錯誤"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 404 {object} models.APIResponse "角色不存在"
// @Router /admin/character/{id}/status [put]
func UpdateCharacterStatus(c *gin.Context) {
	ctx := context.Background()
	characterID := c.Param("id")

	// 解析請求體
	var request struct {
		IsActive bool `json:"is_active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.LogServiceEvent("admin_character_status_update_failed", map[string]interface{}{
			"character_id": characterID,
			"error":        err.Error(),
		})
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "請求參數錯誤",
		})
		return
	}

	// 檢查角色是否存在
	var character db.CharacterDB
	err := GetDB().NewSelect().
		Model(&character).
		Where("id = ?", characterID).
		Scan(ctx)

	if err != nil {
		utils.LogServiceEvent("admin_character_not_found", map[string]interface{}{
			"character_id": characterID,
		})
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "角色不存在",
		})
		return
	}

	// 更新角色狀態
	_, err = GetDB().NewUpdate().
		Model(&character).
		Set("is_active = ?", request.IsActive).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", characterID).
		Exec(ctx)

	if err != nil {
		utils.LogServiceEvent("admin_character_status_update_failed", map[string]interface{}{
			"character_id": characterID,
			"error":        err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "更新角色狀態失敗",
		})
		return
	}

	// 記錄成功日誌
	action := "啟用"
	if !request.IsActive {
		action = "停用"
	}

	utils.LogServiceEvent("admin_character_status_updated", map[string]interface{}{
		"character_id": characterID,
		"is_active":    request.IsActive,
		"action":       action,
	})

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("角色%s成功", action),
		Data: gin.H{
			"character_id": characterID,
			"is_active":    request.IsActive,
		},
	})
}

// Admin Character Management APIs

// AdminGetCharacters 管理員獲取所有角色列表
// @Summary 獲取角色列表
// @Description 管理員用：包含追蹤信息和軟刪除狀態的角色列表
// @Tags 管理員 - 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(50)
// @Param status query string false "狀態篩選" Enums(active,inactive,all)
// @Param created_by query string false "創建者篩選"
// @Param type query string false "角色類型篩選" Enums(system,user,all)
// @Param include_deleted query bool false "是否包含已刪除" default(false)
// @Success 200 {object} models.APIResponse{data=object} "角色列表"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 500 {object} models.APIResponse "服務器錯誤"
// @Router /admin/characters [get]
func AdminGetCharacters(c *gin.Context) {
	ctx := context.Background()

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	status := c.Query("status")        // active, inactive, all
	createdBy := c.Query("created_by") // 特定創建者篩選
	charType := c.Query("type")        // system, user, all
	sortBy := c.Query("sort_by")       // 排序字段
	sortOrder := c.Query("sort_order") // 排序方向
	includeDeleted := c.Query("include_deleted") == "true"

	// 構建查詢
	query := GetDB().NewSelect().
		Model((*db.CharacterDB)(nil)).
		Column("id", "name", "type", "locale", "is_active", "avatar_url", "tags", "popularity",
			"user_description", "created_by", "updated_by", "is_public", "is_system",
			"created_at", "updated_at", "deleted_at")

	// 軟刪除篩選
	if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	// 狀態篩選
	if status != "" && status != "all" {
		if status == "active" {
			query = query.Where("is_active = ?", true)
		} else if status == "inactive" {
			query = query.Where("is_active = ?", false)
		}
	}

	// 創建者篩選
	if createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}

	// 類型篩選
	if charType != "" && charType != "all" {
		if charType == "system" {
			query = query.Where("is_system = ?", true)
		} else if charType == "user" {
			query = query.Where("is_system = ?", false)
		}
	}

	// 計算總數
	total, err := query.Count(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to count characters")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取角色總數失敗",
			},
		})
		return
	}

	// 處理排序
	orderClause := "updated_at DESC" // 默認排序
	if sortBy != "" {
		validSortFields := map[string]string{
			"name":       "name",
			"type":       "type",
			"status":     "is_active",
			"creator":    "created_by",
			"popularity": "popularity",
			"updated_at": "updated_at",
		}

		if field, valid := validSortFields[sortBy]; valid {
			direction := "DESC"
			if sortOrder == "asc" {
				direction = "ASC"
			}
			orderClause = field + " " + direction
		}
	}

	// 分頁查詢
	var characters []db.CharacterDB
	err = query.
		Order(orderClause).
		Offset((page-1)*limit).
		Limit(limit).
		Scan(ctx, &characters)

	if err != nil {
		utils.Logger.WithError(err).WithField("admin_id", adminID).Error("Failed to query characters")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取角色列表失敗",
			},
		})
		return
	}

	// 轉換為響應格式
	var characterResponses []gin.H
	for _, char := range characters {
		charResponse := gin.H{
			"id":               char.ID,
			"name":             char.Name,
			"type":             char.Type,
			"locale":           char.Locale,
			"is_active":        char.IsActive,
			"avatar_url":       char.AvatarURL,
			"tags":             char.Tags,
			"popularity":       char.Popularity,
			"user_description": char.UserDescription,
			"created_by":       char.CreatedBy,
			"updated_by":       char.UpdatedBy,
			"is_public":        char.IsPublic,
			"is_system":        char.IsSystem,
			"created_at":       char.CreatedAt,
			"updated_at":       char.UpdatedAt,
			"deleted_at":       char.DeletedAt,
		}

		// 獲取創建者和更新者的用戶名
		if char.CreatedBy != nil && *char.CreatedBy != "" {
			var user db.UserDB
			err := GetDB().NewSelect().
				Model(&user).
				Column("username", "display_name").
				Where("id = ?", *char.CreatedBy).
				Scan(ctx)
			if err == nil {
				charResponse["created_by_name"] = user.Username
				if user.DisplayName != nil {
					charResponse["created_by_display_name"] = *user.DisplayName
				}
			}
		}

		characterResponses = append(characterResponses, charResponse)
	}

	// 計算分頁信息
	totalPages := (total + limit - 1) / limit

	utils.Logger.WithFields(logrus.Fields{
		"admin_id": adminID,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"filters": map[string]interface{}{
			"status":          status,
			"created_by":      createdBy,
			"type":            charType,
			"include_deleted": includeDeleted,
		},
	}).Info("管理員獲取角色列表")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取角色列表成功",
		Data: gin.H{
			"characters": characterResponses,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_count":  total,
				"limit":        limit,
			},
			"filters": gin.H{
				"status":          status,
				"created_by":      createdBy,
				"type":            charType,
				"include_deleted": includeDeleted,
			},
		},
	})
}

// AdminGetCharacterByID 管理員獲取單個角色詳情
// @Summary 獲取角色詳情
// @Description 管理員用：包含創建歷史的角色詳細信息
// @Tags 管理員 - 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Success 200 {object} models.APIResponse{data=object} "角色詳情"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 404 {object} models.APIResponse "角色不存在"
// @Failure 500 {object} models.APIResponse "服務器錯誤"
// @Router /admin/characters/{id} [get]
func AdminGetCharacterByID(c *gin.Context) {
	ctx := context.Background()
	characterID := c.Param("id")

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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

	if characterID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "角色ID為必填項目",
			Error: &models.APIError{
				Code:    "CHARACTER_ID_REQUIRED",
				Message: "角色ID為必填項目",
			},
		})
		return
	}

	// 獲取角色詳情（包含已刪除）
	var character db.CharacterDB
	err := GetDB().NewSelect().
		Model(&character).
		Where("id = ?", characterID).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("character_id", characterID).Error("Character not found")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "角色不存在",
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在",
			},
		})
		return
	}

	// 構建詳細響應
	charResponse := gin.H{
		"id":               character.ID,
		"name":             character.Name,
		"type":             character.Type,
		"locale":           character.Locale,
		"is_active":        character.IsActive,
		"avatar_url":       character.AvatarURL,
		"tags":             character.Tags,
		"popularity":       character.Popularity,
		"user_description": character.UserDescription,
		"created_by":       character.CreatedBy,
		"updated_by":       character.UpdatedBy,
		"is_public":        character.IsPublic,
		"is_system":        character.IsSystem,
		"created_at":       character.CreatedAt,
		"updated_at":       character.UpdatedAt,
		"deleted_at":       character.DeletedAt,
	}

	// 獲取創建者信息（可能是用戶或管理員）
	if character.CreatedBy != nil && *character.CreatedBy != "" {
		// 先嘗試從用戶表查詢
		var creator db.UserDB
		err := GetDB().NewSelect().
			Model(&creator).
			Column("id", "username", "display_name", "email").
			Where("id = ?", *character.CreatedBy).
			Scan(ctx)
		if err == nil {
			charResponse["creator"] = gin.H{
				"id":           creator.ID,
				"username":     creator.Username,
				"display_name": creator.DisplayName,
				"email":        creator.Email,
				"type":         "user",
			}
		} else {
			// 如果用戶表沒找到，嘗試從管理員表查詢
			var admin db.AdminDB
			err := GetDB().NewSelect().
				Model(&admin).
				Column("id", "username", "display_name", "email").
				Where("id = ?", *character.CreatedBy).
				Scan(ctx)
			if err == nil {
				charResponse["creator"] = gin.H{
					"id":           admin.ID,
					"username":     admin.Username,
					"display_name": admin.DisplayName,
					"email":        admin.Email,
					"type":         "admin",
				}
			}
		}
	}

	// 獲取更新者信息（可能是用戶或管理員）
	if character.UpdatedBy != nil && *character.UpdatedBy != "" &&
		(character.CreatedBy == nil || *character.UpdatedBy != *character.CreatedBy) {
		// 先嘗試從用戶表查詢
		var updater db.UserDB
		err := GetDB().NewSelect().
			Model(&updater).
			Column("id", "username", "display_name", "email").
			Where("id = ?", *character.UpdatedBy).
			Scan(ctx)
		if err == nil {
			charResponse["updater"] = gin.H{
				"id":           updater.ID,
				"username":     updater.Username,
				"display_name": updater.DisplayName,
				"email":        updater.Email,
				"type":         "user",
			}
		} else {
			// 如果用戶表沒找到，嘗試從管理員表查詢
			var admin db.AdminDB
			err := GetDB().NewSelect().
				Model(&admin).
				Column("id", "username", "display_name", "email").
				Where("id = ?", *character.UpdatedBy).
				Scan(ctx)
			if err == nil {
				charResponse["updater"] = gin.H{
					"id":           admin.ID,
					"username":     admin.Username,
					"display_name": admin.DisplayName,
					"email":        admin.Email,
					"type":         "admin",
				}
			}
		}
	}

	// 獲取使用統計
	var chatCount int
	chatCount, _ = GetDB().NewSelect().
		Model((*db.ChatDB)(nil)).
		Where("character_id = ? AND status != ?", characterID, "deleted").
		Count(ctx)

	var messageCount int
	messageCount, _ = GetDB().NewSelect().
		Model((*db.MessageDB)(nil)).
		Join("INNER JOIN chats c ON m.chat_id = c.id").
		Where("c.character_id = ?", characterID).
		Count(ctx)

	charResponse["usage_stats"] = gin.H{
		"chat_sessions":  chatCount,
		"total_messages": messageCount,
	}

	utils.Logger.WithFields(logrus.Fields{
		"admin_id":     adminID,
		"character_id": characterID,
	}).Info("管理員獲取角色詳情")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取角色詳情成功",
		Data:    charResponse,
	})
}

// AdminUpdateCharacter 管理員更新角色
// @Summary 更新角色
// @Description 管理員用：更新指定角色的信息
// @Tags 管理員 - 角色管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "角色ID"
// @Param request body models.AdminCharacterUpdateRequest true "更新請求"
// @Success 200 {object} models.APIResponse{data=object} "更新成功"
// @Failure 400 {object} models.APIResponse "請求參數錯誤"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 404 {object} models.APIResponse "角色不存在"
// @Failure 500 {object} models.APIResponse "服務器錯誤"
// @Router /admin/characters/{id} [put]
func AdminUpdateCharacter(c *gin.Context) {
	ctx := context.Background()
	characterID := c.Param("id")

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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

	if characterID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "角色ID為必填項目",
			Error: &models.APIError{
				Code:    "CHARACTER_ID_REQUIRED",
				Message: "角色ID為必填項目",
			},
		})
		return
	}

	var req models.AdminCharacterUpdateRequest
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

	// 檢查角色是否存在
	var character db.CharacterDB
	err := GetDB().NewSelect().
		Model(&character).
		Where("id = ? AND deleted_at IS NULL", characterID).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("character_id", characterID).Error("Character not found")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在",
			},
		})
		return
	}

	// 構建更新數據
	updateQuery := GetDB().NewUpdate().
		Model((*db.CharacterDB)(nil)).
		Set("updated_at = ?", time.Now()).
		Set("updated_by = ?", adminID)

	hasUpdates := false

	if req.Name != "" {
		updateQuery = updateQuery.Set("name = ?", req.Name)
		hasUpdates = true
	}
	if req.Type != "" {
		updateQuery = updateQuery.Set("type = ?", req.Type)
		hasUpdates = true
	}
	if req.Locale != "" {
		updateQuery = updateQuery.Set("locale = ?", req.Locale)
		hasUpdates = true
	}
	if req.IsActive != nil {
		updateQuery = updateQuery.Set("is_active = ?", *req.IsActive)
		hasUpdates = true
	}
	if req.AvatarURL != nil {
		updateQuery = updateQuery.Set("avatar_url = ?", req.AvatarURL)
		hasUpdates = true
	}
	if req.Tags != nil {
		updateQuery = updateQuery.Set("tags = ?", pgdialect.Array(req.Tags))
		hasUpdates = true
	}
	if req.Popularity != nil {
		updateQuery = updateQuery.Set("popularity = ?", *req.Popularity)
		hasUpdates = true
	}
	if req.UserDescription != nil {
		updateQuery = updateQuery.Set("user_description = ?", req.UserDescription)
		hasUpdates = true
	}
	if req.IsPublic != nil {
		updateQuery = updateQuery.Set("is_public = ?", *req.IsPublic)
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

	// 執行更新
	result, err := updateQuery.
		Where("id = ? AND deleted_at IS NULL", characterID).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithFields(map[string]interface{}{
			"admin_id":     adminID,
			"character_id": characterID,
		}).Error("Failed to update character")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "更新角色失敗",
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
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在或未發生更新",
			},
		})
		return
	}

	// 獲取更新後的角色信息
	err = GetDB().NewSelect().
		Model(&character).
		Where("id = ?", characterID).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to fetch updated character")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取更新後角色信息失敗",
			},
		})
		return
	}

	utils.Logger.WithFields(map[string]interface{}{
		"admin_id":     adminID,
		"character_id": characterID,
	}).Info("Admin updated character")

	// 構建響應數據
	charResponse := gin.H{
		"id":               character.ID,
		"name":             character.Name,
		"type":             character.Type,
		"locale":           character.Locale,
		"is_active":        character.IsActive,
		"avatar_url":       character.AvatarURL,
		"tags":             character.Tags,
		"popularity":       character.Popularity,
		"user_description": character.UserDescription,
		"created_by":       character.CreatedBy,
		"updated_by":       character.UpdatedBy,
		"is_public":        character.IsPublic,
		"is_system":        character.IsSystem,
		"created_at":       character.CreatedAt,
		"updated_at":       character.UpdatedAt,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "角色更新成功",
		Data:    charResponse,
	})
}

// AdminRestoreCharacter 管理員恢復已刪除角色
// @Summary 恢復角色
// @Description 管理員用：恢復已軟刪除的角色
// @Tags 管理員 - 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Success 200 {object} models.APIResponse "恢復成功"
// @Failure 400 {object} models.APIResponse "請求參數錯誤"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 404 {object} models.APIResponse "角色不存在"
// @Failure 500 {object} models.APIResponse "服務器錯誤"
// @Router /admin/characters/{id}/restore [post]
func AdminRestoreCharacter(c *gin.Context) {
	ctx := context.Background()
	characterID := c.Param("id")

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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

	if characterID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "角色ID為必填項目",
			Error: &models.APIError{
				Code:    "CHARACTER_ID_REQUIRED",
				Message: "角色ID為必填項目",
			},
		})
		return
	}

	// 檢查角色是否存在且被軟刪除
	var count int
	count, err := GetDB().NewSelect().
		Model((*db.CharacterDB)(nil)).
		WhereDeleted().
		Where("id = ?", characterID).
		Count(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to check character existence")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "檢查角色狀態失敗",
			},
		})
		return
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在或未被刪除",
			},
		})
		return
	}

	// 恢復角色 - 手動恢復軟刪除
	_, err = GetDB().NewUpdate().
		Model((*db.CharacterDB)(nil)).
		WhereAllWithDeleted().
		Set("deleted_at = NULL").
		Set("updated_by = ?", adminID).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", characterID).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithFields(map[string]interface{}{
			"admin_id":     adminID,
			"character_id": characterID,
		}).Error("Failed to restore character")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "恢復角色失敗",
			},
		})
		return
	}

	utils.Logger.WithFields(map[string]interface{}{
		"admin_id":     adminID,
		"character_id": characterID,
	}).Info("Admin restored character")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "角色恢復成功",
		Data: gin.H{
			"character_id": characterID,
			"restored_at":  time.Now(),
			"restored_by":  adminID,
		},
	})
}

// AdminPermanentDeleteCharacter 管理員永久刪除角色
// @Summary 永久刪除角色
// @Description 管理員用：永久刪除角色（不可恢復）
// @Tags 管理員 - 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Success 200 {object} models.APIResponse "刪除成功"
// @Failure 400 {object} models.APIResponse "請求參數錯誤"
// @Failure 401 {object} models.APIResponse "未授權"
// @Failure 404 {object} models.APIResponse "角色不存在"
// @Failure 500 {object} models.APIResponse "服務器錯誤"
// @Router /admin/characters/{id}/permanent [delete]
func AdminPermanentDeleteCharacter(c *gin.Context) {
	ctx := context.Background()
	characterID := c.Param("id")

	// 檢查管理員權限
	adminID, exists := c.Get("admin_id")
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

	if characterID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "角色ID為必填項目",
			Error: &models.APIError{
				Code:    "CHARACTER_ID_REQUIRED",
				Message: "角色ID為必填項目",
			},
		})
		return
	}

	// 檢查角色是否存在（包括已軟刪除的）
	var count int
	count, err := GetDB().NewSelect().
		Model((*db.CharacterDB)(nil)).
		WhereAllWithDeleted().
		Where("id = ?", characterID).
		Count(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to check character existence")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "檢查角色狀態失敗",
			},
		})
		return
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在",
			},
		})
		return
	}

	// 檢查是否有相關聊天記錄
	var chatCount int
	chatCount, _ = GetDB().NewSelect().
		Model((*db.ChatDB)(nil)).
		Where("character_id = ? AND status != ?", characterID, "deleted").
		Count(ctx)

	if chatCount > 0 {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_IN_USE",
				Message: fmt.Sprintf("角色仍有 %d 個聊天會話，無法永久刪除", chatCount),
			},
		})
		return
	}

	// 永久刪除角色 - 使用 Bun ForceDelete
	_, err = GetDB().NewDelete().
		Model((*db.CharacterDB)(nil)).
		Where("id = ?", characterID).
		ForceDelete().
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithFields(map[string]interface{}{
			"admin_id":     adminID,
			"character_id": characterID,
		}).Error("Failed to permanently delete character")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "永久刪除角色失敗",
			},
		})
		return
	}

	utils.Logger.WithFields(map[string]interface{}{
		"admin_id":     adminID,
		"character_id": characterID,
	}).Warn("Admin permanently deleted character")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "角色永久刪除成功",
		Data: gin.H{
			"character_id": characterID,
			"deleted_at":   time.Now(),
			"deleted_by":   adminID,
		},
	})
}
