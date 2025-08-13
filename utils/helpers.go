package utils

import (
	"strconv"
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