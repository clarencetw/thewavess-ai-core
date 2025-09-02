package pages

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthPageHandler 顯示健康檢查頁面
func HealthPageHandler(c *gin.Context) {
	data := gin.H{
		"Title": "系統健康狀態",
	}
	c.HTML(http.StatusOK, "health.html", data)
}
