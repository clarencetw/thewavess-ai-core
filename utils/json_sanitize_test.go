package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSanitizeLooseJSONForNewlines_字串換行處理測試
func TestSanitizeLooseJSONForNewlines_字串換行處理測試(t *testing.T) {
	t.Run("空字串保持不變", func(t *testing.T) {
		assert.Equal(t, "", SanitizeLooseJSONForNewlines(""))
	})

	t.Run("字串中的裸換行被轉義", func(t *testing.T) {
		input := "{\"text\":\"第一行\n第二行\"}"
		output := SanitizeLooseJSONForNewlines(input)
		assert.Equal(t, "{\"text\":\"第一行\\n第二行\"}", output)
	})

	t.Run("字串中的CRLF換行被統一", func(t *testing.T) {
		input := "{\"text\":\"第一行\r\n第二行\"}"
		output := SanitizeLooseJSONForNewlines(input)
		assert.Equal(t, "{\"text\":\"第一行\\n第二行\"}", output)
	})

	t.Run("字串外的換行保持原樣", func(t *testing.T) {
		input := "{\n  \"value\": 123\n}"
		output := SanitizeLooseJSONForNewlines(input)
		assert.Equal(t, input, output)
	})

	t.Run("已轉義的換行不重複處理", func(t *testing.T) {
		input := "{\"text\":\"第一行\\n第二行\"}"
		output := SanitizeLooseJSONForNewlines(input)
		assert.Equal(t, input, output)
	})

	t.Run("處理包含反斜線與引號的字串", func(t *testing.T) {
		input := "{\"text\":\"路徑\\\\file\\\"name\n下一行\"}"
		output := SanitizeLooseJSONForNewlines(input)
		assert.Equal(t, "{\"text\":\"路徑\\\\file\\\"name\\n下一行\"}", output)
	})

	t.Run("多個字串片段都正確處理", func(t *testing.T) {
		input := "{\"a\":\"甲\n乙\",\"b\":\"丙\n丁\"}"
		output := SanitizeLooseJSONForNewlines(input)
		assert.Equal(t, "{\"a\":\"甲\\n乙\",\"b\":\"丙\\n丁\"}", output)
	})
}
