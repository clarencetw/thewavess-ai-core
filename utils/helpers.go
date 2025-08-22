package utils

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

// ParseInt 安全地將字符串轉換為整數
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// IntToString 將整數轉換為字符串
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// StringInSlice 檢查字符串是否在切片中
func StringInSlice(str string, slice []string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// Max 返回兩個整數中的較大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min 返回兩個整數中的較小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateUUID 生成標準 UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateID 生成帶前綴的ID
func GenerateID(prefix string) string {
	return prefix + "_" + uuid.New().String()
}

// 常用的帶前綴ID生成函數
func GenerateUserID() string     { return GenerateID("user") }
func GenerateSessionID() string  { return GenerateID("session") }
func GenerateMessageID() string  { return GenerateID("msg") }
func GenerateCharacterID() string { return GenerateID("char") }
func GenerateEventID() string    { return GenerateID("event") }
func GenerateTagID() string      { return GenerateID("tag") }
func GenerateMemoryID() string   { return GenerateID("mem") }
func GenerateTTSID() string      { return GenerateID("tts") }
func GenerateNovelID() string    { return GenerateID("novel") }
func GenerateBatchID() string    { return GenerateID("batch") }

// GetCurrentTimestampString 獲取當前時間戳字串
func GetCurrentTimestampString() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Now 獲取當前時間
func Now() time.Time {
	return time.Now().UTC()
}

