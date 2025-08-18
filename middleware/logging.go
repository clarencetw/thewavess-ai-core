package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/handlers"
	"github.com/clarencetw/thewavess-ai-core/services"
)

// LoggingMiddleware 創建日誌記錄中間件
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 記錄 API 請求統計
		isError := param.StatusCode >= 400
		responseTime := param.Latency.Nanoseconds() / int64(time.Millisecond)
		
		handlers.IncrementRequestCount(isError, responseTime)
		
		// 記錄到結構化日誌
		userID := ""
		if param.Keys != nil {
			if uid, exists := param.Keys["user_id"]; exists {
				if uidStr, ok := uid.(string); ok {
					userID = uidStr
				}
			}
		}
		
		data := map[string]interface{}{
			"ip":         param.ClientIP,
			"user_agent": param.Request.UserAgent(),
		}
		
		services.LogAPIEvent(
			param.Method,
			param.Path,
			param.StatusCode,
			responseTime,
			userID,
			data,
		)
		
		// 返回空字符串，因為我們已經通過結構化日誌記錄了
		return ""
	})
}