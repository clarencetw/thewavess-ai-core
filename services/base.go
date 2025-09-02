package services

import (
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/uptrace/bun"
)

// GetDB 全局輔助函數（用於遷移期間）
func GetDB() *bun.DB {
	return database.GetApp().DB()
}
