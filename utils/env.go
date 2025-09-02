package utils

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// 全域載入標記
var envLoaded = false

// LoadEnv 載入環境變數（只載入一次）
func LoadEnv() error {
	if envLoaded {
		return nil
	}

	err := godotenv.Load()
	if err != nil {
		// 在生產環境中，.env 文件可能不存在（使用系統環境變數）
		if os.Getenv("ENVIRONMENT") != "production" {
			return err
		}
	}

	envLoaded = true
	return nil
}

// GetEnvWithDefault 獲取環境變數，如果不存在則使用預設值
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvIntWithDefault 獲取整數環境變數，如果不存在或無效則使用預設值
func GetEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvFloatWithDefault 獲取浮點數環境變數，如果不存在或無效則使用預設值
func GetEnvFloatWithDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// GetEnvBoolWithDefault 獲取布林環境變數，如果不存在或無效則使用預設值
func GetEnvBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// RequiredEnv 獲取必要的環境變數，如果不存在則 panic
func RequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Required environment variable not set: " + key)
	}
	return value
}
