package handlers

import (
	"net/http"
	"runtime"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
)

// GetVersion godoc
// @Summary      獲取 API 版本資訊
// @Description  獲取當前 API 的版本和構建資訊
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200 {object} models.APIResponse{data=object} "版本資訊"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /version [get]
func GetVersion(c *gin.Context) {
	// 從環境變數或構建時變數獲取真實資訊
	version := utils.GetEnvWithDefault("APP_VERSION", "1.0.0")
	buildTime := utils.GetEnvWithDefault("BUILD_TIME", utils.GetCurrentTimestampString())
	gitCommit := utils.GetEnvWithDefault("GIT_COMMIT", "unknown")
	environment := utils.GetEnvWithDefault("ENVIRONMENT", "development")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取版本資訊成功",
		Data: gin.H{
			"version":     version,
			"build_time":  buildTime,
			"git_commit":  gitCommit,
			"go_version":  runtime.Version(),
			"environment": environment,
		},
	})
}

// GetStatus godoc
// @Summary      獲取系統狀態
// @Description  獲取系統運行狀態和健康檢查資訊
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200 {object} models.APIResponse{data=object} "系統狀態"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /status [get]
func GetStatus(c *gin.Context) {
	// 系統內存統計
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 檢查數據庫狀態
	dbStatus := "disconnected"
	if GetDB() != nil {
		if err := GetDB().Ping(); err == nil {
			dbStatus = "connected"
		} else {
			dbStatus = "error"
		}
	}

	// 檢查環境變數
	openaiStatus := "not_configured"
	if utils.GetEnvWithDefault("OPENAI_API_KEY", "") != "" {
		openaiStatus = "configured"
	}

	grokStatus := "not_configured"
	if utils.GetEnvWithDefault("GROK_API_KEY", "") != "" {
		grokStatus = "configured"
	}

	// 計算內存使用率
	memUsageMB := float64(m.Alloc) / 1024 / 1024

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "系統狀態正常",
		Data: gin.H{
			"status":          "healthy",
			"timestamp":       utils.GetCurrentTimestampString(),
			"database":        dbStatus,
			"openai_api":      openaiStatus,
			"grok_api":        grokStatus,
			"memory_usage_mb": memUsageMB,
			"goroutines":      runtime.NumGoroutine(),
			"go_version":      runtime.Version(),
			"environment":     utils.GetEnvWithDefault("ENVIRONMENT", "development"),
		},
	})
}
