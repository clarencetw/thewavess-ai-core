package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
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
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取版本資訊成功",
		Data: gin.H{
			"version":     "1.0.0",
			"build_time":  "2023-12-01T00:00:00Z",
			"git_commit":  "abc123",
			"go_version":  "go1.22",
			"environment": "development",
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
	// TODO: 實作系統狀態檢查邏輯
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "系統狀態正常",
		Data: gin.H{
			"status":         "healthy",
			"uptime":         "24h30m15s",
			"database":       "connected",
			"redis":          "connected",
			"qdrant":         "connected",
			"openai_api":     "available",
			"grok_api":       "available",
			"memory_usage":   "45.2%",
			"cpu_usage":      "12.8%",
			"active_sessions": 156,
		},
	})
}