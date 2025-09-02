package middleware

import (
	"time"

	"github.com/clarencetw/thewavess-ai-core/handlers"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware 添加請求追蹤ID
func RequestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 檢查是否已有 Request ID（例如從客戶端傳入）
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 生成新的 Request ID
			requestID = uuid.New().String()
		}

		// 設置到 context 和 response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// 繼續處理請求
		c.Next()
	})
}

// LoggingMiddleware 創建日誌記錄中間件
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 記錄 API 請求統計
		isError := param.StatusCode >= 400
		responseTime := param.Latency.Nanoseconds() / int64(time.Millisecond)

		handlers.IncrementRequestCount(isError, responseTime)

		// 記錄到結構化日誌
		userID := ""
		requestID := ""
		if param.Keys != nil {
			if uid, exists := param.Keys["user_id"]; exists {
				if uidStr, ok := uid.(string); ok {
					userID = uidStr
				}
			}
			if rid, exists := param.Keys["request_id"]; exists {
				if ridStr, ok := rid.(string); ok {
					requestID = ridStr
				}
			}
		}

		data := map[string]interface{}{
			"ip":         param.ClientIP,
			"user_agent": param.Request.UserAgent(),
			"request_id": requestID,
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
