package utils

import (
	"net/http"
	"strings"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationHelper 提供統一的驗證錯誤處理
type ValidationHelper struct{}

// BindAndValidate 綁定並驗證請求，返回是否成功
func (v *ValidationHelper) BindAndValidate(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		v.HandleValidationError(c, err)
		return false
	}
	return true
}

// HandleValidationError 處理驗證錯誤並返回標準格式
func (v *ValidationHelper) HandleValidationError(c *gin.Context, err error) {
	var message string
	var code string = "INVALID_INPUT"

	// 檢查是否為驗證錯誤
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		message = v.formatValidationErrors(validationErrors)
		code = "VALIDATION_ERROR"
	} else {
		// JSON 解析錯誤或其他綁定錯誤
		message = "請求格式錯誤: " + err.Error()
	}

	c.JSON(http.StatusBadRequest, models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
	})
}

// formatValidationErrors 格式化驗證錯誤為友好的中文消息
func (v *ValidationHelper) formatValidationErrors(errs validator.ValidationErrors) string {
	var messages []string

	for _, err := range errs {
		fieldName := v.getChineseFieldName(err.Field())

		var message string
		switch err.Tag() {
		case "required":
			message = fieldName + "為必填項目"
		case "email":
			message = fieldName + "必須為有效的郵箱地址"
		case "min":
			if err.Kind().String() == "string" {
				message = fieldName + "長度不能少於 " + err.Param() + " 個字符"
			} else {
				message = fieldName + "數值不能小於 " + err.Param()
			}
		case "max":
			if err.Kind().String() == "string" {
				message = fieldName + "長度不能超過 " + err.Param() + " 個字符"
			} else {
				message = fieldName + "數值不能大於 " + err.Param()
			}
		case "oneof":
			message = fieldName + "必須為以下值之一: " + err.Param()
		case "url":
			message = fieldName + "必須為有效的 URL"
		default:
			message = fieldName + "格式不正確 (" + err.Tag() + ")"
		}
		messages = append(messages, message)
	}

	return strings.Join(messages, "; ")
}

// getChineseFieldName 將英文字段名轉換為中文
func (v *ValidationHelper) getChineseFieldName(field string) string {
	fieldMap := map[string]string{
		"Username":    "用戶名",
		"Email":       "郵箱",
		"Password":    "密碼",
		"Gender":      "性別",
		"AvatarURL":   "頭像",
		"Name":        "角色名稱",
		"Type":        "角色類型",
		"Locale":      "語言",
		"Description": "描述",
		"Background":  "背景",
		"Tone":        "語氣",
		"SceneType":   "場景類型",
		"TimeOfDay":   "時間",
		"StateKey":    "狀態鍵",
		"CharacterID": "角色ID",
		"SessionID":   "對話ID",
		"Message":     "訊息內容",
		"Mode":        "模式",
		"Tag":         "標籤",
		"Page":        "頁碼",
		"PageSize":    "每頁數量",
		"SortOrder":   "排序方式",
		"Age":         "年齡",
	}

	if chinese, exists := fieldMap[field]; exists {
		return chinese
	}
	return field
}

// 全局驗證助手實例
var ValidationHelperInstance = &ValidationHelper{}
