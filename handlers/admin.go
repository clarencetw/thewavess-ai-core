package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
// @Router /api/v1/admin/stats [get]
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
// @Router /api/v1/admin/logs [get]
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