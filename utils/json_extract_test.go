package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestJSONExtract_JSON提取功能測試
func TestJSONExtract_JSON提取功能測試(t *testing.T) {
	t.Run("純JSON對象提取", func(t *testing.T) {
		input := `{"a":1,"b":[2,3],"c":{"d":"e"}}`
		output, err := ExtractJSONFromText(input)

		assert.NoError(t, err, "提取純JSON不應該出錯")
		assert.Equal(t, input, output, "純JSON提取後應該保持不變")
	})

	t.Run("帶前後綴文字的JSON提取", func(t *testing.T) {
		input := "這是結果：\n```json\n{\n  \"x\": [1, 2, 3],\n  \"y\": \"ok\"\n}\n```\n謝謝！"
		output, err := ExtractJSONFromText(input)

		assert.NoError(t, err, "提取帶前後綴的JSON不應該出錯")

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err, "提取的JSON應該能正常解析")
		assert.Equal(t, "ok", result["y"].(string), "解析的JSON內容應該正確")

		// 檢查陣列內容
		x, ok := result["x"].([]any)
		assert.True(t, ok, "x欄位應該是陣列")
		assert.Len(t, x, 3, "陣列應該有3個元素")
	})

	t.Run("JSON陣列提取", func(t *testing.T) {
		input := "前綴文字 [ {\n\"a\":1}, {\"b\":2} ] 後綴文字"
		output, err := ExtractJSONFromText(input)

		assert.NoError(t, err, "提取JSON陣列不應該出錯")

		var result []map[string]int
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err, "提取的JSON陣列應該能正常解析")
		assert.Len(t, result, 2, "陣列應該有2個元素")
		assert.Equal(t, 1, result[0]["a"], "第一個元素的a值應該是1")
		assert.Equal(t, 2, result[1]["b"], "第二個元素的b值應該是2")
	})

	t.Run("字符串內包含大括號的JSON提取", func(t *testing.T) {
		input := "前綴 {\n  \"text\": \"包含 } 大括號\",\n  \"ok\":true\n} 後綴"
		output, err := ExtractJSONFromText(input)

		assert.NoError(t, err, "提取包含特殊字符的JSON不應該出錯")

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err, "提取的JSON應該能正常解析")
		assert.True(t, result["ok"].(bool), "ok欄位應該是true")
		assert.Equal(t, "包含 } 大括號", result["text"].(string), "text欄位應該包含正確的內容")
	})

	t.Run("無JSON內容的文字", func(t *testing.T) {
		input := "這裡沒有JSON，抱歉。只有普通的文字內容。"
		_, err := ExtractJSONFromText(input)

		assert.Error(t, err, "沒有JSON的文字應該返回錯誤")
	})

	t.Run("複雜嵌套JSON提取", func(t *testing.T) {
		input := `
		以下是複雜的配置：
		{
			"config": {
				"database": {
					"host": "localhost",
					"port": 5432,
					"users": ["admin", "user"]
				},
				"features": {
					"enabled": true,
					"options": {"debug": false, "timeout": 30}
				}
			}
		}
		配置結束。
		`
		output, err := ExtractJSONFromText(input)

		assert.NoError(t, err, "提取複雜嵌套JSON不應該出錯")

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err, "複雜JSON應該能正常解析")

		// 檢查嵌套結構
		config, ok := result["config"].(map[string]any)
		assert.True(t, ok, "config欄位應該是對象")

		database, ok := config["database"].(map[string]any)
		assert.True(t, ok, "database欄位應該是對象")
		assert.Equal(t, "localhost", database["host"].(string), "host應該是localhost")

		users, ok := database["users"].([]any)
		assert.True(t, ok, "users欄位應該是陣列")
		assert.Len(t, users, 2, "users陣列應該有2個元素")
	})

	t.Run("結構不平衡的JSON處理", func(t *testing.T) {
		input := "這是不平衡的JSON：{\"key\": \"value\""
		_, err := ExtractJSONFromText(input)

		assert.Error(t, err, "結構不平衡的JSON應該返回錯誤")
	})

	t.Run("JSON語法有效性檢驗", func(t *testing.T) {
		// 注意：ExtractJSONFromText 只檢查結構平衡，不檢查JSON語法有效性
		// 這個測試驗證該行為是正確的
		input := "結構平衡但語法無效：{\"key\": invalid_value}"
		output, err := ExtractJSONFromText(input)

		assert.NoError(t, err, "結構平衡的內容應該被成功提取")
		assert.Contains(t, output, "invalid_value", "提取的內容應該包含原始文字")

		// 但是用json.Unmarshal解析時會失敗
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		assert.Error(t, err, "語法無效的JSON在解析時應該失敗")
	})
}
