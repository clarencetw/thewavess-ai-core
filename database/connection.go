package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDatabase 初始化數據庫連接
func InitDatabase() error {
	// 從環境變數讀取數據庫配置
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "thewavess_ai_core"
	}
	
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	// 構建連接字符串
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// 連接數據庫
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// 設置連接池參數
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// 測試連接
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	utils.Logger.WithFields(map[string]interface{}{
		"host":   host,
		"port":   port,
		"dbname": dbname,
	}).Info("Database connected successfully")

	return nil
}

// CloseDatabase 關閉數據庫連接
func CloseDatabase() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// HealthCheck 數據庫健康檢查
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return DB.Ping()
}