package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/docs"
	"github.com/clarencetw/thewavess-ai-core/handlers/pages"
	"github.com/clarencetw/thewavess-ai-core/middleware"
	"github.com/clarencetw/thewavess-ai-core/routes"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// 構建時變數
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// @title           Thewavess AI Core API
// @version         1.0
// @description     女性向 AI 互動應用後端服務，提供智能對話、互動小說、情感陪伴等功能
// @termsOfService  https://thewavess.ai/terms

// @contact.name   Thewavess AI Core Team
// @contact.url    https://thewavess.ai
// @contact.email  api@thewavess.ai

// @license.name  Proprietary
// @license.url   https://thewavess.ai/license

// @host      localhost:8080
// @BasePath  /api/v1
// @schemes   http https
//
// Note: This API supports multiple environments:
// - Local development: http://localhost:8080
// - Development server: https://thewavess-ai-core.clarence.ltd
// The host can be dynamically set based on environment.

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 請輸入 'Bearer ' + JWT token，例如: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

// configureStaticFiles 已移除，改用 router.Static 直接配置

// setupWebRoutes 設置網頁路由（非 API）
func setupWebRoutes(router *gin.Engine) {
	// 管理員頁面路由（純HTML結構，無需後端認證）
	// AJAX架構：認證檢查由前端JavaScript + AJAX API完成
	adminPages := router.Group("/admin")
	{
		// 登入頁面
		adminPages.GET("/login", pages.AdminLoginPageHandler)

		// 管理頁面（純HTML結構，數據通過AJAX載入）
		adminPages.GET("/dashboard", pages.AdminDashboardPageHandler)
		adminPages.GET("/users", pages.AdminUsersPageHandler)
		adminPages.GET("/chats", pages.AdminChatHistoryPageHandler)
		adminPages.GET("/characters", pages.AdminCharactersPageHandler)
	}
}

// configureCORS 配置 CORS 中間件
func configureCORS() cors.Config {
	config := cors.DefaultConfig()

	// 從環境變數讀取允許的來源，預設為全開
	allowedOrigins := utils.GetEnvWithDefault("CORS_ALLOWED_ORIGINS", "*")

	if allowedOrigins == "*" {
		config.AllowAllOrigins = true
	} else {
		config.AllowOrigins = strings.Split(allowedOrigins, ",")
	}

	// 從環境變數讀取允許的方法，預設為常用方法
	allowedMethods := utils.GetEnvWithDefault("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
	config.AllowMethods = strings.Split(allowedMethods, ",")

	// 從環境變數讀取允許的標頭，預設為常用標頭
	allowedHeaders := utils.GetEnvWithDefault("CORS_ALLOWED_HEADERS", "Origin,Content-Length,Content-Type,Authorization,X-Requested-With,Accept,Accept-Encoding,Accept-Language,Connection,Host,User-Agent")
	config.AllowHeaders = strings.Split(allowedHeaders, ",")

	// 允許認證
	config.AllowCredentials = true

	// 從環境變數讀取暴露的標頭
	exposedHeaders := utils.GetEnvWithDefault("CORS_EXPOSED_HEADERS", "")
	if exposedHeaders != "" {
		config.ExposeHeaders = strings.Split(exposedHeaders, ",")
	}

	return config
}

func main() {
	// 設置構建資訊為環境變數，供handlers使用
	os.Setenv("APP_VERSION", Version)
	os.Setenv("BUILD_TIME", BuildTime)
	os.Setenv("GIT_COMMIT", GitCommit)

	// Initialize logger first
	utils.InitLogger()

	// Initialize log hook after logger
	logHook := services.NewLogHook()
	utils.Logger.AddHook(logHook)

	// Load environment variables from .env file
	if err := utils.LoadEnv(); err != nil {
		utils.Logger.Warn("No .env file found or error loading .env file")
	} else {
		utils.Logger.Info("Successfully loaded .env file")
	}

	// Initialize database (non-fatal for development/testing)
	dbInitialized := false
	app := database.InitDB()
	if app == nil {
		utils.Logger.Warn("Failed to initialize database, running in database-free mode")
	} else {
		dbInitialized = true
		defer app.Close()

		// Run startup hooks
		ctx := context.Background()
		if err := app.RunStartHooks(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Startup hooks failed")
		}

		utils.Logger.Info("Database and hooks initialized successfully")
	}

	// Initialize Gin router
	router := gin.Default()

	// Load HTML templates (if templates directory exists)
	if templates, err := filepath.Glob("templates/*"); err == nil && len(templates) > 0 {
		router.LoadHTMLGlob("templates/*")
		utils.Logger.WithField("template_count", len(templates)).Info("HTML templates loaded successfully")
	} else {
		utils.Logger.Warn("No HTML templates found or templates directory does not exist")
	}

	// Add middleware
	router.Use(cors.New(configureCORS()))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(utils.RecoverMiddleware())
	router.Use(middleware.LoggingMiddleware())

	// Root path redirect to health page
	router.GET("/", pages.HealthPageHandler)

	// Static files (minimal CSS only)
	router.Static("/public", "./public")

	// Configure Swagger host dynamically based on environment
	configureSwaggerHost()

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", healthCheck)

	// Web page routes (outside API)
	setupWebRoutes(router)

	// Setup API routes
	api := router.Group("/api/v1")
	routes.SetupRoutes(api)

	// Display startup banner
	displayStartupBanner(dbInitialized)

	// Start server
	utils.LogServiceEvent("server_starting", map[string]interface{}{
		"port":               8080,
		"database_available": dbInitialized,
	})

	if err := http.ListenAndServe(":8080", router); err != nil {
		utils.Logger.WithError(err).Fatal("Failed to start server")
	}
}

// configureSwaggerHost configures Swagger host dynamically based on environment
func configureSwaggerHost() {
	// Get API host from environment variable, default to localhost:8080
	apiHost := utils.GetEnvWithDefault("API_HOST", "localhost:8080")

	// Detect if running in production/development based on host
	var schemes []string
	if strings.Contains(apiHost, "localhost") || strings.Contains(apiHost, "127.0.0.1") {
		schemes = []string{"http"}
	} else {
		schemes = []string{"https", "http"}
	}

	// Update Swagger info dynamically
	docs.SwaggerInfo.Host = apiHost
	docs.SwaggerInfo.Schemes = schemes

	utils.Logger.WithFields(map[string]interface{}{
		"swagger_host": apiHost,
		"schemes":      schemes,
	}).Info("Swagger configuration updated")
}

// healthCheck provides a simple health check endpoint outside the API
// This endpoint is not included in Swagger documentation as it's for operational use
func healthCheck(c *gin.Context) {
	response := gin.H{
		"status":    "ok",
		"service":   "thewavess-ai-core",
		"version":   "1.0.0",
		"timestamp": utils.GetCurrentTimestampString(),
	}

	// Check database health
	app := database.GetApp()
	if app == nil {
		response["database"] = "unavailable"
		response["database_message"] = "running in database-free mode"
		response["note"] = "some endpoints may not work without database"
	} else {
		// Try to ping database
		db := app.DB()
		if err := db.Ping(); err != nil {
			response["database"] = "unhealthy"
			response["database_error"] = err.Error()
		} else {
			response["database"] = "healthy"
		}
	}

	// Always return 200 for health check - service is running
	c.JSON(http.StatusOK, response)
}

// displayStartupBanner shows essential startup information
func displayStartupBanner(dbInitialized bool) {
	banner := `
████████╗██╗  ██╗███████╗██╗    ██╗ █████╗ ██╗   ██╗███████╗███████╗███████╗
╚══██╔══╝██║  ██║██╔════╝██║    ██║██╔══██╗██║   ██║██╔════╝██╔════╝██╔════╝
   ██║   ███████║█████╗  ██║ █╗ ██║███████║██║   ██║█████╗  ███████╗███████╗
   ██║   ██╔══██║██╔══╝  ██║███╗██║██╔══██║╚██╗ ██╔╝██╔══╝  ╚════██║╚════██║
   ██║   ██║  ██║███████╗╚███╔███╔╝██║  ██║ ╚████╔╝ ███████╗███████║███████║
   ╚═╝   ╚═╝  ╚═╝╚══════╝ ╚══╝╚══╝ ╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚══════╝╚══════╝
                                                    AI Core • 女性向智能對話系統`

	// Print banner
	utils.Logger.Info(banner)

	// System information
	shortCommit := GitCommit
	if len(GitCommit) > 7 {
		shortCommit = GitCommit[:7]
	}
	utils.Logger.WithFields(map[string]interface{}{
		"version":    Version,
		"build_time": BuildTime,
		"git_commit": shortCommit,
	}).Info("🚀 System Version")

	// Database status
	dbStatus := "❌ Unavailable"
	if dbInitialized {
		dbStatus = "✅ Connected"
	}
	utils.Logger.WithField("database", dbStatus).Info("💾 Database Status")

	// AI Engines status
	openaiKey := os.Getenv("OPENAI_API_KEY")
	grokKey := os.Getenv("GROK_API_KEY")

	aiEngines := []string{}
	if openaiKey != "" {
		aiEngines = append(aiEngines, "OpenAI")
	}
	if grokKey != "" {
		aiEngines = append(aiEngines, "Grok")
	}

	if len(aiEngines) == 0 {
		utils.Logger.Info("🤖 AI Engines: ❌ None configured")
	} else {
		utils.Logger.WithField("engines", strings.Join(aiEngines, ", ")).Info("🤖 AI Engines: ✅ Ready")
	}

	// Important URLs
	utils.Logger.Info("🌐 Server: http://localhost:8080")
	utils.Logger.Info("📖 API Docs: http://localhost:8080/swagger/index.html")
	utils.Logger.Info("❤️ Health: http://localhost:8080/health")

	if !dbInitialized {
		utils.Logger.Warn("⚠️  Database unavailable - running in limited mode")
	}

	utils.Logger.Info("═══════════════════════════════════════════════════════════════")
}
