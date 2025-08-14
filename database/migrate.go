package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/clarencetw/thewavess-ai-core/utils"
)

// RunMigrations 運行數據庫遷移
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	migrationsDir := "database/migrations"
	
	// 檢查遷移目錄是否存在
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		utils.Logger.Info("No migrations directory found, skipping migrations")
		return nil
	}

	// 創建遷移記錄表
	if err := createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// 讀取遷移文件
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %v", err)
	}

	// 排序文件（按文件名）
	sort.Strings(files)

	// 執行遷移
	for _, file := range files {
		filename := filepath.Base(file)
		
		// 檢查是否已執行
		if executed, err := isMigrationExecuted(filename); err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		} else if executed {
			utils.Logger.WithField("file", filename).Debug("Migration already executed, skipping")
			continue
		}

		// 執行遷移
		if err := executeMigrationFile(file); err != nil {
			return fmt.Errorf("failed to execute migration %s: %v", filename, err)
		}

		// 記錄遷移
		if err := recordMigration(filename); err != nil {
			return fmt.Errorf("failed to record migration %s: %v", filename, err)
		}

		utils.Logger.WithField("file", filename).Info("Migration executed successfully")
	}

	utils.Logger.Info("All migrations completed")
	return nil
}

// createMigrationsTable 創建遷移記錄表
func createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	_, err := DB.Exec(query)
	return err
}

// isMigrationExecuted 檢查遷移是否已執行
func isMigrationExecuted(filename string) (bool, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", filename).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// executeMigrationFile 執行遷移文件
func executeMigrationFile(filePath string) error {
	// 讀取文件內容
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// 執行整個SQL文件（不分割，讓PostgreSQL處理）
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %v", err)
	}

	return tx.Commit()
}

// recordMigration 記錄已執行的遷移
func recordMigration(filename string) error {
	_, err := DB.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", filename)
	return err
}