package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnv_載入與預設值行為測試
func TestEnv_載入與預設值行為測試(t *testing.T) {
	t.Run("非生產環境缺少dotenv返回錯誤", func(t *testing.T) {
		t.Cleanup(func() { envLoaded = false })
		envLoaded = false

		tempDir := t.TempDir()
		oldWD, err := os.Getwd()
		assert.NoError(t, err)
		assert.NoError(t, os.Chdir(tempDir))
		t.Cleanup(func() {
			os.Chdir(oldWD) //nolint:errcheck
		})

		t.Setenv("ENVIRONMENT", "development")

		err = LoadEnv()
		assert.Error(t, err)
		assert.False(t, envLoaded, "失敗時不應標記為已載入")
	})

	t.Run("生產環境缺少dotenv忽略錯誤", func(t *testing.T) {
		t.Cleanup(func() { envLoaded = false })
		envLoaded = false

		tempDir := t.TempDir()
		oldWD, err := os.Getwd()
		assert.NoError(t, err)
		assert.NoError(t, os.Chdir(tempDir))
		t.Cleanup(func() {
			os.Chdir(oldWD) //nolint:errcheck
		})

		t.Setenv("ENVIRONMENT", "production")

		err = LoadEnv()
		assert.NoError(t, err)
		assert.True(t, envLoaded, "成功載入後應標記為已載入")

		// 第二次調用應直接返回
		err = LoadEnv()
		assert.NoError(t, err)
	})

	t.Run("成功讀取dotenv並載入環境變數", func(t *testing.T) {
		t.Cleanup(func() { envLoaded = false })
		envLoaded = false

		tempDir := t.TempDir()
		oldWD, err := os.Getwd()
		assert.NoError(t, err)
		assert.NoError(t, os.Chdir(tempDir))
		t.Cleanup(func() {
			os.Chdir(oldWD) //nolint:errcheck
		})

		content := "TEST_KEY=測試值\n"
		assert.NoError(t, os.WriteFile(filepath.Join(tempDir, ".env"), []byte(content), 0o600))

		t.Setenv("ENVIRONMENT", "development")

		err = LoadEnv()
		assert.NoError(t, err)
		assert.Equal(t, "測試值", os.Getenv("TEST_KEY"))
	})
}

// TestEnv_各類型讀取測試
func TestEnv_各類型讀取測試(t *testing.T) {
	t.Run("字串環境變數", func(t *testing.T) {
		t.Setenv("STRING_KEY", "自訂")
		assert.Equal(t, "自訂", GetEnvWithDefault("STRING_KEY", "預設"))
		assert.Equal(t, "預設", GetEnvWithDefault("MISSING_KEY", "預設"))
	})

	t.Run("整數環境變數", func(t *testing.T) {
		t.Setenv("INT_KEY", "42")
		assert.Equal(t, 42, GetEnvIntWithDefault("INT_KEY", 7))

		t.Setenv("INT_KEY_INVALID", "not-number")
		assert.Equal(t, -1, GetEnvIntWithDefault("INT_KEY_INVALID", -1))
	})

	t.Run("浮點數環境變數", func(t *testing.T) {
		t.Setenv("FLOAT_KEY", "3.1415")
		assert.InDelta(t, 3.1415, GetEnvFloatWithDefault("FLOAT_KEY", 1.0), 0.0001)

		t.Setenv("FLOAT_KEY_INVALID", "not-float")
		assert.Equal(t, 2.5, GetEnvFloatWithDefault("FLOAT_KEY_INVALID", 2.5))
	})

	t.Run("布林環境變數", func(t *testing.T) {
		cases := []struct {
			value    string
			expected bool
		}{
			{"true", true},
			{"1", true},
			{"yes", true},
			{"false", false},
			{"0", false},
			{"", false},
		}

		for _, c := range cases {
			c := c
			t.Run("值為"+c.value, func(t *testing.T) {
				key := "BOOL_KEY"
				t.Setenv(key, c.value)
				assert.Equal(t, c.expected, GetEnvBoolWithDefault(key, false))
			})
		}

		assert.Equal(t, true, GetEnvBoolWithDefault("BOOL_MISSING", true))
	})

	t.Run("必要環境變數", func(t *testing.T) {
		t.Setenv("REQUIRED_KEY", "存在的值")
		assert.Equal(t, "存在的值", RequiredEnv("REQUIRED_KEY"))

		assert.Panics(t, func() {
			RequiredEnv("MISSING_REQUIRED_KEY")
		})
	})
}
