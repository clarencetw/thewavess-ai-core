package handlers

import (
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// GetVersion godoc
// @Summary      獲取 API 版本資訊
// @Description  獲取當前 API 的版本和構建資訊
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200 {object} models.APIResponse
// @Router       /version [get]
func GetVersion(c *gin.Context) {
	// 獲取環境變數或使用默認值
	version := "1.0.0"
	buildTime := "2025-08-16T00:00:00Z"
	gitCommit := "latest"
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

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
// @Success      200 {object} models.APIResponse
// @Router       /status [get]
func GetStatus(c *gin.Context) {
	// 系統內存統計
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 檢查數據庫狀態
	dbStatus := "disconnected"
	if database.DB != nil {
		if err := database.DB.Ping(); err == nil {
			dbStatus = "connected"
		} else {
			dbStatus = "error"
		}
	}

	// 檢查環境變數
	openaiStatus := "not_configured"
	if os.Getenv("OPENAI_API_KEY") != "" {
		openaiStatus = "configured"
	}

	grokStatus := "not_configured"
	if os.Getenv("GROK_API_KEY") != "" {
		grokStatus = "configured"
	}

	// 計算內存使用率 (簡化)
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
			"environment":     os.Getenv("ENVIRONMENT"),
		},
	})
}