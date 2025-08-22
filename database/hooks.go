package database

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
)

// RegisterDefaultHooks 註冊預設的應用程式鉤子
func RegisterDefaultHooks(app *App) {
	// 資料庫健康檢查鉤子
	app.OnStart("database_health_check", DatabaseHealthCheckHook)
	
	// 系統資源檢查鉤子
	app.OnStart("system_resource_check", SystemResourceCheckHook)
	
	// 配置驗證鉤子
	app.OnStart("config_validation", ConfigValidationHook)
	
	// 資料庫連接池驗證鉤子
	app.OnStart("connection_pool_validation", ConnectionPoolValidationHook)
}

// DatabaseHealthCheckHook 資料庫健康檢查鉤子
func DatabaseHealthCheckHook(ctx context.Context, app *App) error {
	db := app.DB()
	
	start := time.Now()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	pingDuration := time.Since(start)
	utils.Logger.WithField("ping_duration", pingDuration).Info("Database health check passed")
	
	// 檢查是否響應時間過長
	if pingDuration > 5*time.Second {
		utils.Logger.WithField("ping_duration", pingDuration).Warn("Database ping response time is high")
	}
	
	return nil
}

// SystemResourceCheckHook 系統資源檢查鉤子
func SystemResourceCheckHook(ctx context.Context, app *App) error {
	// 檢查 CPU 核心數
	cpuCores := runtime.GOMAXPROCS(0)
	utils.Logger.WithField("cpu_cores", cpuCores).Info("System CPU cores detected")
	
	// 檢查記憶體統計
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	utils.Logger.WithFields(map[string]interface{}{
		"heap_alloc_mb":     bToMB(m.HeapAlloc),
		"heap_sys_mb":       bToMB(m.HeapSys),
		"heap_idle_mb":      bToMB(m.HeapIdle),
		"heap_released_mb":  bToMB(m.HeapReleased),
		"gc_runs":          m.NumGC,
	}).Info("System memory statistics")
	
	// 警告如果記憶體使用過高
	if bToMB(m.HeapAlloc) > 1024 { // 1GB
		utils.Logger.WithField("heap_alloc_mb", bToMB(m.HeapAlloc)).Warn("High memory usage detected")
	}
	
	return nil
}

// ConfigValidationHook 配置驗證鉤子
func ConfigValidationHook(ctx context.Context, app *App) error {
	cfg := app.Config()
	
	// 驗證 DSN 格式
	if cfg.DSN == "" {
		return fmt.Errorf("database DSN is empty")
	}
	
	// 驗證連接池配置
	if cfg.MaxOpenConns <= 0 {
		return fmt.Errorf("invalid MaxOpenConns: %d", cfg.MaxOpenConns)
	}
	
	if cfg.MaxIdleConns <= 0 {
		return fmt.Errorf("invalid MaxIdleConns: %d", cfg.MaxIdleConns)
	}
	
	utils.Logger.WithFields(map[string]interface{}{
		"production":     cfg.Production,
		"debug":         cfg.Debug,
		"max_open_conns": cfg.MaxOpenConns,
		"max_idle_conns": cfg.MaxIdleConns,
	}).Info("Database configuration validated")
	
	return nil
}

// ConnectionPoolValidationHook 連接池驗證鉤子
func ConnectionPoolValidationHook(ctx context.Context, app *App) error {
	db := app.DB()
	
	// 獲取底層的 sql.DB 以檢查連接池統計
	sqlDB := db.DB
	
	stats := sqlDB.Stats()
	utils.Logger.WithFields(map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
	}).Info("Connection pool statistics")
	
	// 檢查是否有連接洩漏的跡象
	if stats.OpenConnections > stats.MaxOpenConnections/2 {
		utils.Logger.WithFields(map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"max_connections":  stats.MaxOpenConnections,
		}).Warn("High number of open connections detected")
	}
	
	return nil
}

// bToMB 將位元組轉換為 MB
func bToMB(b uint64) uint64 {
	return b / 1024 / 1024
}

// DatabaseMigrationStatusHook 資料庫遷移狀態檢查鉤子
func DatabaseMigrationStatusHook(ctx context.Context, app *App) error {
	// 這個鉤子需要遷移系統的支援，這裡提供一個基本的檢查
	db := app.DB()
	
	// 檢查是否存在 bun_migrations 表
	var exists bool
	err := db.NewSelect().
		ColumnExpr("EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'bun_migrations')").
		Scan(ctx, &exists)
	
	if err != nil {
		return fmt.Errorf("failed to check migration table: %w", err)
	}
	
	if !exists {
		utils.Logger.Warn("Migration table not found - database may not be initialized")
		return nil
	}
	
	// 獲取遷移統計
	var migrationCount int
	err = db.NewSelect().
		Table("bun_migrations").
		ColumnExpr("COUNT(*)").
		Scan(ctx, &migrationCount)
	
	if err != nil {
		return fmt.Errorf("failed to count migrations: %w", err)
	}
	
	utils.Logger.WithField("migration_count", migrationCount).Info("Database migration status checked")
	
	return nil
}

// RegisterMigrationStatusHook 註冊遷移狀態檢查鉤子（可選）
func RegisterMigrationStatusHook(app *App) {
	app.OnStart("migration_status_check", DatabaseMigrationStatusHook)
}