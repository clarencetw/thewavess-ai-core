package database

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"sync"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/clarencetw/thewavess-ai-core/utils"
)

// App 代表應用程式上下文，管理資料庫連接和配置
type App struct {
	db          *bun.DB
	dbOnce      sync.Once
	cfg         *Config
	hookManager *HookManager
}

// Config 資料庫配置結構
type Config struct {
	DSN        string
	Production bool
	Debug      bool
	// 連接池配置
	MaxOpenConns int
	MaxIdleConns int
}

// NewApp 創建新的應用程式實例
func NewApp() *App {
	return &App{
		cfg:         loadConfig(),
		hookManager: NewHookManager(),
	}
}

// loadConfig 載入資料庫配置
func loadConfig() *Config {
	// 確保環境變數已載入
	utils.LoadEnv()

	// 構建資料庫連接字符串
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		utils.GetEnvWithDefault("DB_USER", "postgres"),
		utils.GetEnvWithDefault("DB_PASSWORD", "password"),
		utils.GetEnvWithDefault("DB_HOST", "localhost"),
		utils.GetEnvWithDefault("DB_PORT", "5432"),
		utils.GetEnvWithDefault("DB_NAME", "thewavess_ai_core"),
		utils.GetEnvWithDefault("DB_SSLMODE", "disable"),
	)

	isProduction := utils.GetEnvWithDefault("GO_ENV", "development") == "production"
	debug := utils.GetEnvWithDefault("DB_DEBUG", "true") == "true"

	// 計算最佳連接池大小
	maxConns := 4 * runtime.GOMAXPROCS(0)
	if maxConns < 4 {
		maxConns = 4
	}
	if maxConns > 32 {
		maxConns = 32
	}

	return &Config{
		DSN:          dsn,
		Production:   isProduction,
		Debug:        debug && !isProduction,
		MaxOpenConns: maxConns,
		MaxIdleConns: maxConns,
	}
}

// DB 獲取資料庫實例（線程安全的懶加載）
func (app *App) DB() *bun.DB {
	app.dbOnce.Do(func() {
		// 創建 PostgreSQL 連接
		sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(app.cfg.DSN)))

		// 配置連接池
		sqldb.SetMaxOpenConns(app.cfg.MaxOpenConns)
		sqldb.SetMaxIdleConns(app.cfg.MaxIdleConns)

		// 創建 Bun DB 實例
		db := bun.NewDB(sqldb, pgdialect.New())

		// 生產環境優化：設置忽略未知欄位以增強遷移過程的彈性
		if app.cfg.Production {
			// 注意：WithDiscardUnknownColumns 是查詢級別的選項，不是 DB 級別
			// 這裡我們僅做標記，實際使用時需要在查詢中設置
			utils.Logger.Info("Production mode: will use DiscardUnknownColumns on queries")
		}

		// 開發環境下啟用詳細查詢日誌
		if app.cfg.Debug {
			db.AddQueryHook(bundebug.NewQueryHook(
				bundebug.WithVerbose(true),
				bundebug.FromEnv("BUNDEBUG"),
			))
		}

		// 測試連接
		ctx := context.Background()
		if err := db.PingContext(ctx); err != nil {
			utils.Logger.WithError(err).Fatal("Failed to ping database")
		}

		app.db = db
		utils.Logger.Info("Database connection established successfully")
	})
	return app.db
}

// Close 關閉資料庫連接
func (app *App) Close() error {
	if app.db != nil {
		if err := app.db.Close(); err != nil {
			utils.Logger.WithError(err).Error("Failed to close database connection")
			return err
		}
		utils.Logger.Info("Database connection closed")
	}
	return nil
}

// Config 獲取配置
func (app *App) Config() *Config {
	return app.cfg
}

// OnStart 註冊啟動鉤子
func (app *App) OnStart(name string, hook Hook) {
	app.hookManager.OnStart(name, hook)
}

// RunStartHooks 執行所有啟動鉤子
func (app *App) RunStartHooks(ctx context.Context) error {
	return app.hookManager.RunHooks(ctx, app)
}

// GetRegisteredHooks 獲取已註冊的鉤子名稱列表
func (app *App) GetRegisteredHooks() []string {
	return app.hookManager.GetRegisteredHooks()
}
