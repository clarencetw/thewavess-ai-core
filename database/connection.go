package database

import (
	"context"
	"sync"
)

var (
	globalApp     *App
	globalAppOnce sync.Once
)

// InitDB 初始化應用程式和資料庫連接
func InitDB() *App {
	globalAppOnce.Do(func() {
		globalApp = NewApp()

		// 註冊預設鉤子
		RegisterDefaultHooks(globalApp)
		RegisterMigrationStatusHook(globalApp)
	})
	return globalApp
}

// GetApp 獲取應用程式實例
func GetApp() *App {
	if globalApp == nil {
		return InitDB()
	}
	return globalApp
}

// RunStartupHooks 執行所有啟動鉤子
func RunStartupHooks(ctx context.Context) error {
	app := GetApp()
	return app.RunStartHooks(ctx)
}
