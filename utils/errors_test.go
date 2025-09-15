package utils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestAPIError_基本錯誤功能測試
func TestAPIError_基本錯誤功能測試(t *testing.T) {
	t.Run("創建新的API錯誤", func(t *testing.T) {
		err := NewAPIError("TEST_ERROR", "測試錯誤", "這是詳細信息", http.StatusBadRequest)

		assert.Equal(t, "TEST_ERROR", err.Code)
		assert.Equal(t, "測試錯誤", err.Message)
		assert.Equal(t, "這是詳細信息", err.Details)
		assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	})

	t.Run("API錯誤Error方法", func(t *testing.T) {
		err := NewAPIError("VALIDATION_ERROR", "參數錯誤", "用戶名不能為空", http.StatusBadRequest)

		expected := "[VALIDATION_ERROR] 參數錯誤: 用戶名不能為空"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("WithDetails添加詳細信息", func(t *testing.T) {
		baseErr := ErrValidation
		newErr := baseErr.WithDetails("用戶名長度不足")

		// 原錯誤不變
		assert.Empty(t, baseErr.Details)
		// 新錯誤有詳細信息
		assert.Equal(t, "用戶名長度不足", newErr.Details)
		assert.Equal(t, baseErr.Code, newErr.Code)
		assert.Equal(t, baseErr.Message, newErr.Message)
	})

	t.Run("WithContext添加上下文信息", func(t *testing.T) {
		baseErr := ErrUnauthorized
		context := map[string]interface{}{
			"user_id": "user_123",
			"ip":      "192.168.1.1",
		}
		newErr := baseErr.WithContext(context)

		assert.Nil(t, baseErr.Context)
		assert.Equal(t, context, newErr.Context)
	})
}

// TestAPIError_預定義錯誤測試
func TestAPIError_預定義錯誤測試(t *testing.T) {
	tests := []struct {
		name           string
		err            *APIError
		expectedCode   string
		expectedStatus int
	}{
		{"驗證錯誤", ErrValidation, "VALIDATION_ERROR", http.StatusBadRequest},
		{"未授權", ErrUnauthorized, "UNAUTHORIZED", http.StatusUnauthorized},
		{"權限不足", ErrForbidden, "FORBIDDEN", http.StatusForbidden},
		{"資源不存在", ErrNotFound, "RESOURCE_NOT_FOUND", http.StatusNotFound},
		{"請求頻率超限", ErrRateLimit, "RATE_LIMIT_EXCEEDED", http.StatusTooManyRequests},
		{"內部服務器錯誤", ErrInternalServer, "INTERNAL_SERVER_ERROR", http.StatusInternalServerError},
		{"服務不可用", ErrServiceUnavailable, "SERVICE_UNAVAILABLE", http.StatusServiceUnavailable},
		{"AI服務不可用", ErrAIServiceUnavailable, "AI_SERVICE_UNAVAILABLE", http.StatusServiceUnavailable},
		{"處理失敗", ErrProcessingFailed, "PROCESSING_FAILED", http.StatusInternalServerError},
		{"無效會話", ErrInvalidSession, "INVALID_SESSION", http.StatusBadRequest},
		{"NSFW內容被阻擋", ErrNSFWContentBlocked, "NSFW_CONTENT_BLOCKED", http.StatusUnavailableForLegalReasons},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedCode, tt.err.Code)
			assert.Equal(t, tt.expectedStatus, tt.err.StatusCode)
			assert.NotEmpty(t, tt.err.Message)
		})
	}
}

// TestHandleError_統一錯誤處理測試
func TestHandleError_統一錯誤處理測試(t *testing.T) {
	// 設置 Gin 為測試模式
	gin.SetMode(gin.TestMode)

	// 初始化 Logger 以避免 nil pointer
	InitLogger()

	t.Run("處理API錯誤", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("request_id", "test-request-123")

		// 設置測試請求
		req := httptest.NewRequest("POST", "/api/test", nil)
		req.Header.Set("User-Agent", "TestAgent/1.0")
		c.Request = req

		apiErr := ErrValidation.WithDetails("用戶名不能為空")
		HandleError(c, apiErr)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		// 檢查響應體包含預期的錯誤信息
		body := w.Body.String()
		assert.Contains(t, body, "VALIDATION_ERROR")
		assert.Contains(t, body, "請求參數驗證失敗")
		assert.Contains(t, body, "用戶名不能為空")
		assert.Contains(t, body, "\"success\":false")
	})

	t.Run("處理普通錯誤", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 設置測試請求
		req := httptest.NewRequest("GET", "/api/error", nil)
		c.Request = req

		normalErr := errors.New("這是普通錯誤")
		HandleError(c, normalErr)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		body := w.Body.String()
		assert.Contains(t, body, "INTERNAL_SERVER_ERROR")
		assert.Contains(t, body, "內部服務器錯誤")
		assert.Contains(t, body, "這是普通錯誤")
	})

	t.Run("帶上下文的錯誤處理", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 模擬 HTTP 請求
		req := httptest.NewRequest("POST", "/api/test", nil)
		req.Header.Set("User-Agent", "TestAgent/1.0")
		c.Request = req

		context := map[string]interface{}{
			"user_id": "user_123",
			"action":  "create_chat",
		}
		apiErr := ErrForbidden.WithContext(context)
		HandleError(c, apiErr)

		assert.Equal(t, http.StatusForbidden, w.Code)

		body := w.Body.String()
		assert.Contains(t, body, "FORBIDDEN")
		assert.Contains(t, body, "權限不足")
	})
}

// TestRecoverMiddleware_恢復panic測試
func TestRecoverMiddleware_恢復panic測試(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常請求不觸發panic", func(t *testing.T) {
		// 創建Gin引擎來測試middleware
		router := gin.New()
		router.Use(RecoverMiddleware())

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		// 不應該panic
		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("處理panic並返回錯誤", func(t *testing.T) {
		// 創建Gin引擎來測試middleware
		router := gin.New()
		router.Use(RecoverMiddleware())

		router.GET("/panic", func(c *gin.Context) {
			c.Set("request_id", "panic-test-123")
			panic("測試panic")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/panic", nil)

		// middleware應該捕獲panic而不是傳播
		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})

		// 應該返回內部服務器錯誤
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		body := w.Body.String()
		assert.Contains(t, body, "INTERNAL_SERVER_ERROR")
		assert.Contains(t, body, "服務發生異常")
	})
}

// TestGetCurrentTimestamp_時間戳功能測試
func TestGetCurrentTimestamp_時間戳功能測試(t *testing.T) {
	t.Run("生成當前時間戳", func(t *testing.T) {
		timestamp1 := GetCurrentTimestamp()
		timestamp2 := GetCurrentTimestamp()

		// 兩次調用的時間戳應該不同（除非在同一納秒內調用）
		assert.True(t, timestamp2 >= timestamp1)

		// 時間戳應該是正數且合理（大於某個基準值）
		// 2020-01-01的納秒時間戳作為最小值
		minTimestamp := int64(1577836800000000000)
		assert.True(t, timestamp1 > minTimestamp)
	})
}
