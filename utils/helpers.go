package utils

import (
	"time"

	"github.com/Automattic/go-gravatar"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GenerateUUID 生成標準 UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateID 生成帶前綴的ID
func GenerateID(prefix string) string {
	return prefix + "_" + uuid.New().String()
}

// 常用的帶前綴ID生成函數
func GenerateUserID() string         { return GenerateID("user") }
func GenerateChatID() string         { return GenerateID("chat") }
func GenerateMessageID() string      { return GenerateID("msg") }
func GenerateCharacterID() string    { return GenerateID("char") }
func GenerateEventID() string        { return GenerateID("event") }
func GenerateRelationshipID() string { return GenerateID("rel") }
func GenerateAdminID() string        { return GenerateID("admin") } // 注意：由於循環依賴，Admin模型中直接使用uuid.New()
func GenerateTTSID() string          { return GenerateID("tts") }

// GetCurrentTimestampString 獲取當前時間戳字串
func GetCurrentTimestampString() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Now 獲取當前時間
func Now() time.Time {
	return time.Now().UTC()
}

// GetClientIP 獲取客戶端IP地址（使用Gin內建功能）
func GetClientIP(c *gin.Context) string {
	return c.ClientIP()
}

// CalculateAgeFromBirthDate 根據生日計算年齡並判斷是否為成人
func CalculateAgeFromBirthDate(birthDate *time.Time) (age int, isAdult bool) {
	if birthDate == nil {
		return 0, false
	}

	now := time.Now()
	age = now.Year() - birthDate.Year()

	// 檢查是否還沒過生日
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	isAdult = age >= 18
	return age, isAdult
}

// GenerateDefaultAvatarURL 生成預設頭像URL（使用 Gravatar）
func GenerateDefaultAvatarURL() string {
	// 使用預設 email 生成 identicon 風格頭像
	g := gravatar.NewGravatarFromEmail("default@example.com")
	g.Default = "identicon"
	g.Size = 80
	g.Rating = "pg"
	return g.GetURL()
}

// GenerateDefaultAvatarURLByGender 根據性別生成預設頭像URL（使用 Gravatar）
func GenerateDefaultAvatarURLByGender(gender *string) string {
	// 使用不同的預設 email 為不同性別生成不同風格的頭像
	baseEmail := "default@example.com"

	if gender != nil {
		switch *gender {
		case "male":
			baseEmail = "male@example.com"
		case "female":
			baseEmail = "female@example.com"
		}
	}

	// 根據性別設定不同的預設頭像風格
	g := gravatar.NewGravatarFromEmail(baseEmail)
	g.Size = 80
	g.Rating = "pg"

	if gender != nil {
		switch *gender {
		case "male":
			g.Default = "robohash" // 機器人風格適合男性
		case "female":
			g.Default = "wavatar"  // 面部風格適合女性
		default:
			g.Default = "identicon" // 幾何圖案適合其他
		}
	} else {
		g.Default = "identicon"
	}

	return g.GetURL()
}

// Min 返回兩個整數中的較小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
