package services

import (
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/uptrace/bun"
)

// BaseService 提供統一的資料庫存取
type BaseService struct {
	app *database.App
}

// NewBaseService 創建基礎服務
func NewBaseService() *BaseService {
	return &BaseService{
		app: database.GetApp(),
	}
}

// DB 獲取資料庫實例
func (s *BaseService) DB() *bun.DB {
	return s.app.DB()
}

// App 獲取應用程式實例
func (s *BaseService) App() *database.App {
	return s.app
}

// GetDB 全局輔助函數（用於遷移期間）
func GetDB() *bun.DB {
	return database.GetApp().DB()
}