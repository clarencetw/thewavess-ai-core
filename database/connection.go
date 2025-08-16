package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	
	"github.com/clarencetw/thewavess-ai-core/utils"
)

var DB *bun.DB

// InitDB 初始化數據庫連接
func InitDB() error {
	// 構建數據庫連接字符串
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// 創建 PostgreSQL 連接
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	
	// 創建 Bun DB 實例
	DB = bun.NewDB(sqldb, pgdialect.New())

	// 開發環境下啟用詳細查詢日誌
	if os.Getenv("GO_ENV") != "production" {
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