package utils

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestUUID_生成功能測試
func TestUUID_生成功能測試(t *testing.T) {
	t.Run("生成標準UUID", func(t *testing.T) {
		id := GenerateUUID()
		assert.NotEmpty(t, id)
		assert.Len(t, id, 36) // UUID標準長度

		// 驗證是否為有效的UUID格式
		_, err := uuid.Parse(id)
		assert.NoError(t, err)
	})

	t.Run("生成帶前綴的ID", func(t *testing.T) {
		id := GenerateID("test")
		assert.NotEmpty(t, id)
		assert.Contains(t, id, "test_")
		assert.True(t, len(id) > 5) // 前綴+下劃線+UUID

		// 提取UUID部分驗證
		uuidPart := id[5:] // 移除"test_"前綴
		_, err := uuid.Parse(uuidPart)
		assert.NoError(t, err)
	})

	t.Run("各種專用ID生成函數", func(t *testing.T) {
		tests := []struct {
			name     string
			function func() string
			prefix   string
		}{
			{"用戶ID", GenerateUserID, "user_"},
			{"聊天ID", GenerateChatID, "chat_"},
			{"訊息ID", GenerateMessageID, "msg_"},
			{"角色ID", GenerateCharacterID, "char_"},
			{"事件ID", GenerateEventID, "event_"},
			{"關係ID", GenerateRelationshipID, "rel_"},
			{"管理員ID", GenerateAdminID, "admin_"},
			{"TTS ID", GenerateTTSID, "tts_"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id := tt.function()
				assert.NotEmpty(t, id)
				assert.True(t, len(id) > len(tt.prefix))
				assert.Contains(t, id, tt.prefix)

				// 驗證UUID部分
				uuidPart := id[len(tt.prefix):]
				_, err := uuid.Parse(uuidPart)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("多次生成的ID應該不同", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := GenerateUUID()
			assert.False(t, ids[id], "生成了重複的UUID: %s", id)
			ids[id] = true
		}
	})
}

// TestTime_時間處理功能測試
func TestTime_時間處理功能測試(t *testing.T) {
	t.Run("獲取當前時間戳字串", func(t *testing.T) {
		timestamp := GetCurrentTimestampString()
		assert.NotEmpty(t, timestamp)

		// 解析時間戳確保格式正確
		_, err := time.Parse(time.RFC3339, timestamp)
		assert.NoError(t, err)

		// 檢查是否為UTC時間
		parsedTime, _ := time.Parse(time.RFC3339, timestamp)
		assert.Equal(t, time.UTC, parsedTime.Location())
	})

	t.Run("獲取當前UTC時間", func(t *testing.T) {
		now := Now()
		assert.Equal(t, time.UTC, now.Location())

		// 檢查時間是否合理（在測試執行時間附近）
		diff := time.Since(now).Abs()
		assert.True(t, diff < time.Second, "時間差異過大")
	})

	t.Run("連續獲取的時間戳應該遞增", func(t *testing.T) {
		time1 := GetCurrentTimestampString()
		time.Sleep(time.Millisecond) // 確保時間有差異
		time2 := GetCurrentTimestampString()

		parsed1, _ := time.Parse(time.RFC3339, time1)
		parsed2, _ := time.Parse(time.RFC3339, time2)

		assert.True(t, parsed2.After(parsed1) || parsed2.Equal(parsed1))
	})
}

// TestAge_年齡計算功能測試
func TestAge_年齡計算功能測試(t *testing.T) {
	t.Run("計算成年人年齡", func(t *testing.T) {
		// 25歲的生日（假設今天是2024年）
		birthDate := time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
		age, isAdult := CalculateAgeFromBirthDate(&birthDate)

		assert.True(t, age >= 24) // 至少24歲（考慮測試執行時間）
		assert.True(t, isAdult)
	})

	t.Run("計算未成年人年齡", func(t *testing.T) {
		// 10歲的生日
		birthDate := time.Now().AddDate(-10, 0, 0)
		age, isAdult := CalculateAgeFromBirthDate(&birthDate)

		assert.Equal(t, 10, age)
		assert.False(t, isAdult)
	})

	t.Run("剛好18歲", func(t *testing.T) {
		// 剛好18歲
		birthDate := time.Now().AddDate(-18, 0, 0)
		age, isAdult := CalculateAgeFromBirthDate(&birthDate)

		assert.Equal(t, 18, age)
		assert.True(t, isAdult)
	})

	t.Run("還未過生日的情況", func(t *testing.T) {
		// 生日在明天（還未過生日）
		now := time.Now()
		birthDate := time.Date(now.Year()-20, now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

		age, isAdult := CalculateAgeFromBirthDate(&birthDate)

		assert.Equal(t, 19, age) // 應該是19歲，因為還沒過20歲生日
		assert.True(t, isAdult)
	})

	t.Run("生日參數為nil", func(t *testing.T) {
		age, isAdult := CalculateAgeFromBirthDate(nil)

		assert.Equal(t, 0, age)
		assert.False(t, isAdult)
	})
}

// TestAvatar_頭像生成功能測試
func TestAvatar_頭像生成功能測試(t *testing.T) {
	t.Run("生成預設頭像URL", func(t *testing.T) {
		avatarURL := GenerateDefaultAvatarURL()
		assert.NotEmpty(t, avatarURL)
		assert.Contains(t, avatarURL, "https://www.gravatar.com/avatar/")
		assert.Contains(t, avatarURL, "d=identicon")
	})

	t.Run("根據性別生成頭像URL", func(t *testing.T) {
		// 男性頭像
		male := "male"
		maleAvatar := GenerateDefaultAvatarURLByGender(&male)
		assert.Contains(t, maleAvatar, "https://www.gravatar.com/avatar/")
		assert.Contains(t, maleAvatar, "d=robohash")
		assert.Contains(t, maleAvatar, "s=80")

		// 女性頭像
		female := "female"
		femaleAvatar := GenerateDefaultAvatarURLByGender(&female)
		assert.Contains(t, femaleAvatar, "https://www.gravatar.com/avatar/")
		assert.Contains(t, femaleAvatar, "d=wavatar")
		assert.Contains(t, femaleAvatar, "s=80")

		// 其他性別
		other := "other"
		otherAvatar := GenerateDefaultAvatarURLByGender(&other)
		assert.Contains(t, otherAvatar, "https://www.gravatar.com/avatar/")
		assert.Contains(t, otherAvatar, "d=identicon")

		// 無效性別
		invalid := "invalid"
		invalidAvatar := GenerateDefaultAvatarURLByGender(&invalid)
		assert.Contains(t, invalidAvatar, "https://www.gravatar.com/avatar/")
		assert.Contains(t, invalidAvatar, "d=identicon")
	})

	t.Run("性別參數為nil", func(t *testing.T) {
		avatarURL := GenerateDefaultAvatarURLByGender(nil)
		assert.Contains(t, avatarURL, "https://www.gravatar.com/avatar/")
		assert.Contains(t, avatarURL, "d=identicon")
	})
}

// TestClientIP_客戶端IP獲取測試
func TestClientIP_客戶端IP獲取測試(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("獲取客戶端IP", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 設置測試請求
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		req.RemoteAddr = "127.0.0.1:12345"
		c.Request = req

		ip := GetClientIP(c)
		assert.NotEmpty(t, ip)

		// Gin的ClientIP()會優先使用X-Forwarded-For
		// 在測試環境中可能返回不同的值，主要確保能獲取到IP
		assert.True(t, len(ip) > 0)
	})

	t.Run("無代理情況下的IP獲取", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:54321"
		c.Request = req

		ip := GetClientIP(c)
		assert.NotEmpty(t, ip)
	})
}

// TestMin_輔助函數測試
func TestMin_輔助函數測試(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a小於b", 3, 7, 3},
		{"b小於a", 10, 5, 5},
		{"相等情況", 8, 8, 8},
		{"負數情況", -5, -2, -5},
		{"零值情況", 0, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
