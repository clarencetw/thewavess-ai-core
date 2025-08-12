package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// APIError 定義 API 錯誤類型
type APIError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Context    map[string]interface{} `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
}

// 預定義錯誤
var (
	ErrValidation = &APIError{
		Code:       "VALIDATION_ERROR",
		Message:    "請求參數驗證失敗",
		StatusCode: http.StatusBadRequest,
	}
	
	ErrUnauthorized = &APIError{
		Code:       "UNAUTHORIZED",
		Message:    "未授權訪問",
		StatusCode: http.StatusUnauthorized,
	}
	
	ErrForbidden = &APIError{
		Code:       "FORBIDDEN",
		Message:    "權限不足",
		StatusCode: http.StatusForbidden,
	}
	
	ErrNotFound = &APIError{
		Code:       "RESOURCE_NOT_FOUND",
		Message:    "資源不存在",
		StatusCode: http.StatusNotFound,
	}
	
	ErrRateLimit = &APIError{
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    "請求頻率超限",
		StatusCode: http.StatusTooManyRequests,
	}
	
	ErrInternalServer = &APIError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "內部服務器錯誤",
		StatusCode: http.StatusInternalServerError,
	}
	
	ErrServiceUnavailable = &APIError{
		Code:       "SERVICE_UNAVAILABLE",
		Message:    "服務暫時不可用",
		StatusCode: http.StatusServiceUnavailable,
	}
	
	ErrAIServiceUnavailable = &APIError{
		Code:       "AI_SERVICE_UNAVAILABLE",
		Message:    "AI 服務不可用",
		StatusCode: http.StatusServiceUnavailable,
	}
	
	ErrProcessingFailed = &APIError{
		Code:       "PROCESSING_FAILED",
		Message:    "處理失敗",
		StatusCode: http.StatusInternalServerError,
	}
	
	ErrInvalidSession = &APIError{
		Code:       "INVALID_SESSION",
		Message:    "無效的會話",
		StatusCode: http.StatusBadRequest,
	}
	
	ErrNSFWContentBlocked = &APIError{
		Code:       "NSFW_CONTENT_BLOCKED",
		Message:    "NSFW 內容被阻擋",
		StatusCode: http.StatusUnavailableForLegalReasons,
	}
)

// NewAPIError 創建新的 API 錯誤
func NewAPIError(code, message, details string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
	}
}

// WithDetails 為錯誤添加詳細信息
func (e *APIError) WithDetails(details string) *APIError {
	newErr := *e
	newErr.Details = details
	return &newErr
}

// WithContext 為錯誤添加上下文信息
func (e *APIError) WithContext(context map[string]interface{}) *APIError {
	newErr := *e
	newErr.Context = context
	return &newErr
}

// HandleError 統一錯誤處理函數
func HandleError(c *gin.Context, err error) {
	var apiErr *APIError
	
	switch e := err.(type) {
	case *APIError:
		apiErr = e
	default:
		// 對於未知錯誤，返回內部服務器錯誤
		apiErr = ErrInternalServer.WithDetails(err.Error())
	}
	
	// 記錄錯誤到日誌
	fields := logrus.Fields{
		"error_code":   apiErr.Code,
		"status_code":  apiErr.StatusCode,
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"user_agent":   c.Request.UserAgent(),
		"request_id":   c.GetString("request_id"),
	}
	
	// 添加上下文信息
	if apiErr.Context != nil {
		for k, v := range apiErr.Context {
			fields[k] = v
		}
	}
	
	LogError(apiErr, "API error occurred", fields)
	
	// 構建錯誤回應
	response := models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    apiErr.Code,
			Message: apiErr.Message,
			Details: apiErr.Details,
		},
	}
	
	c.JSON(apiErr.StatusCode, response)
}

// ValidateRequired 驗證必填字段
func ValidateRequired(fields map[string]interface{}) error {
	for fieldName, value := range fields {
		if value == nil || value == "" {
			return ErrValidation.WithDetails(fmt.Sprintf("字段 '%s' 為必填", fieldName))
		}
	}
	return nil
}

// RecoverMiddleware 恢復 panic 的中間件
func RecoverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fields := logrus.Fields{
					"panic":      fmt.Sprintf("%v", err),
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"request_id": c.GetString("request_id"),
				}
				
				Logger.WithFields(fields).Error("Panic recovered")
				
				HandleError(c, ErrInternalServer.WithDetails("服務發生異常"))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RequestIDMiddleware 為每個請求生成唯一 ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID 生成請求 ID
func generateRequestID() string {
	// 使用簡單的時間戳生成 ID，實際項目中可以使用 UUID
	return fmt.Sprintf("req_%d", GetCurrentTimestamp())
}

// GetCurrentTimestamp 獲取當前時間戳
func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}