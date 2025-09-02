package handlers

import (
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/uptrace/bun"
)

// GetDB 統一的資料庫存取函數
func GetDB() *bun.DB {
	return database.GetApp().DB()
}

// GetApp 統一的應用程式實例存取函數
func GetApp() *database.App {
	return database.GetApp()
}
