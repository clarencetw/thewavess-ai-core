package utils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// 測試用的結構體
type TestUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Gender   string `json:"gender" binding:"omitempty,oneof=male female other"`
	Age      int    `json:"age" binding:"omitempty,min=13,max=120"`
}

// TestValidationHelper_參數驗證測試
func TestValidationHelper_參數驗證測試(t *testing.T) {
	gin.SetMode(gin.TestMode)
	helper := &ValidationHelper{}

	t.Run("有效參數驗證成功", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 設置有效的JSON請求
		validJSON := `{
			"username": "testuser",
			"email": "test@example.com",
			"password": "password123",
			"gender": "male",
			"age": 25
		}`
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(validJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		var req TestUserRequest
		result := helper.BindAndValidate(c, &req)

		assert.True(t, result)
		assert.Equal(t, "testuser", req.Username)
		assert.Equal(t, "test@example.com", req.Email)
		assert.Equal(t, "password123", req.Password)
		assert.Equal(t, "male", req.Gender)
		assert.Equal(t, 25, req.Age)
	})

	t.Run("必填欄位驗證失敗", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 缺少必填欄位的JSON
		invalidJSON := `{
			"username": "",
			"password": "123"
		}`
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		var req TestUserRequest
		result := helper.BindAndValidate(c, &req)

		assert.False(t, result)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		body := w.Body.String()
		assert.Contains(t, body, "VALIDATION_ERROR")
		assert.Contains(t, body, "用戶名為必填項目")
		assert.Contains(t, body, "郵箱為必填項目")
		assert.Contains(t, body, "密碼長度不能少於 8 個字符")
	})

	t.Run("郵箱格式驗證失敗", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		invalidJSON := `{
			"username": "testuser",
			"email": "invalid-email",
			"password": "password123"
		}`
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		var req TestUserRequest
		result := helper.BindAndValidate(c, &req)

		assert.False(t, result)
		body := w.Body.String()
		assert.Contains(t, body, "郵箱必須為有效的郵箱地址")
	})

	t.Run("枚舉值驗證失敗", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		invalidJSON := `{
			"username": "testuser",
			"email": "test@example.com",
			"password": "password123",
			"gender": "invalid_gender"
		}`
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		var req TestUserRequest
		result := helper.BindAndValidate(c, &req)

		assert.False(t, result)
		body := w.Body.String()
		assert.Contains(t, body, "性別必須為以下值之一: male female other")
	})

	t.Run("數值範圍驗證失敗", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		invalidJSON := `{
			"username": "testuser",
			"email": "test@example.com",
			"password": "password123",
			"age": 150
		}`
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		var req TestUserRequest
		result := helper.BindAndValidate(c, &req)

		assert.False(t, result)
		body := w.Body.String()
		assert.Contains(t, body, "年齡數值不能大於 120")
	})

	t.Run("JSON格式錯誤", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		invalidJSON := `{invalid json`
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		var req TestUserRequest
		result := helper.BindAndValidate(c, &req)

		assert.False(t, result)
		body := w.Body.String()
		assert.Contains(t, body, "INVALID_INPUT")
		assert.Contains(t, body, "請求格式錯誤")
	})
}

// TestValidationHelper_中文欄位名稱映射測試
func TestValidationHelper_中文欄位名稱映射測試(t *testing.T) {
	helper := &ValidationHelper{}

	tests := []struct {
		englishField string
		chineseField string
	}{
		{"Username", "用戶名"},
		{"Email", "郵箱"},
		{"Password", "密碼"},
		{"Gender", "性別"},
		{"AvatarURL", "頭像"},
		{"Name", "角色名稱"},
		{"Description", "描述"},
		{"CharacterID", "角色ID"},
		{"Message", "訊息內容"},
		{"UnknownField", "UnknownField"}, // 未映射的欄位保持原樣
	}

	for _, tt := range tests {
		t.Run(tt.englishField+"映射為"+tt.chineseField, func(t *testing.T) {
			result := helper.getChineseFieldName(tt.englishField)
			assert.Equal(t, tt.chineseField, result)
		})
	}
}

// TestValidationHelper_錯誤訊息格式化測試
func TestValidationHelper_錯誤訊息格式化測試(t *testing.T) {
	helper := &ValidationHelper{}

	t.Run("格式化多個驗證錯誤", func(t *testing.T) {
		// 創建模擬的驗證錯誤
		// 注意：這裡我們主要測試格式化邏輯，實際的validator.ValidationErrors需要通過實際驗證產生
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 使用無效數據觸發多個驗證錯誤
		invalidJSON := `{
			"username": "a",
			"email": "invalid",
			"password": "123",
			"age": -5
		}`
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		var req TestUserRequest
		err := c.ShouldBindJSON(&req)
		assert.Error(t, err)

		// 檢查是否為驗證錯誤
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			message := helper.formatValidationErrors(validationErrors)

			// 檢查錯誤訊息包含中文描述
			assert.Contains(t, message, "用戶名長度不能少於")
			assert.Contains(t, message, "郵箱必須為有效的郵箱地址")
			assert.Contains(t, message, "密碼長度不能少於")

			// 檢查多個錯誤用分號分隔
			assert.Contains(t, message, ";")
		}
	})
}

// TestValidationHelper_全域實例測試
func TestValidationHelper_全域實例測試(t *testing.T) {
	t.Run("全域驗證助手實例存在", func(t *testing.T) {
		assert.NotNil(t, ValidationHelperInstance)
		assert.IsType(t, &ValidationHelper{}, ValidationHelperInstance)
	})

	t.Run("全域實例功能正常", func(t *testing.T) {
		result := ValidationHelperInstance.getChineseFieldName("Username")
		assert.Equal(t, "用戶名", result)
	})
}
