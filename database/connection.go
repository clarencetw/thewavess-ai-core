package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	
	"github.com/clarencetw/thewavess-ai-core/utils"
)

var DB *bun.DB

// InitDB 初始化數據庫連接
func InitDB() error {
	// 確保環境變數已載入
	utils.LoadEnv()
	
	// 構建數據庫連接字符串
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		utils.GetEnvWithDefault("DB_USER", "postgres"),
		utils.GetEnvWithDefault("DB_PASSWORD", "password"),
		utils.GetEnvWithDefault("DB_HOST", "localhost"),
		utils.GetEnvWithDefault("DB_PORT", "5432"),
		utils.GetEnvWithDefault("DB_NAME", "thewavess_ai_core"),
		utils.GetEnvWithDefault("DB_SSLMODE", "disable"),
	)

	// 創建 PostgreSQL 連接
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	
	// 創建 Bun DB 實例
	DB = bun.NewDB(sqldb, pgdialect.New())

	// 開發環境下啟用詳細查詢日誌
	if utils.GetEnvWithDefault("GO_ENV", "development") != "production" {
		DB.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	// 測試連接
	ctx := context.Background()
	if err := DB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	utils.Logger.Info("Database connection established successfully")
	return nil
}

// CloseDB 關閉 Bun 數據庫連接
func CloseDB() error {
	if DB != nil {
		if err := DB.Close(); err != nil {
			utils.Logger.WithError(err).Error("Failed to close Bun database connection")
			return err
		}
		utils.Logger.Info("Bun database connection closed")
	}
	return nil
}

// GetDB 獲取 Bun DB 實例
func GetDB() *bun.DB {
	return DB
}