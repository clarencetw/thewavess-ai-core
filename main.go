package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/clarencetw/thewavess-ai-core/database"
	_ "github.com/clarencetw/thewavess-ai-core/docs"
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
//
// Note: This API supports multiple environments:
// - Local development: http://localhost:8080
// - Development server: https://thewavess-ai-core.clarence.ltd
// Change the host in Swagger UI as needed.

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

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Add middleware
	router.Use(cors.New(configureCORS()))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(utils.RecoverMiddleware())
	router.Use(middleware.LoggingMiddleware())

	// Root path redirect to health page
	router.GET("/", pages.HealthPageHandler)

	// Static files (minimal CSS only)
	router.Static("/public", "./public")

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", healthCheck)

	// Web page routes (outside API)
	setupWebRoutes(router)

	// Setup API routes
	api := router.Group("/api/v1")
	routes.SetupRoutes(api)

	// Start server
	utils.LogServiceEvent("server_starting", map[string]interface{}{
		"port":               8080,
		"database_available": dbInitialized,
		"endpoints": map[string]string{
			"web_interface": "http://localhost:8080",
			"swagger_ui":    "http://localhost:8080/swagger/index.html",
			"health_check":  "http://localhost:8080/health",
		},
	})

	utils.Logger.Info("Starting Thewavess AI Core API server on :8080")
	utils.Logger.Info("Web interface: http://localhost:8080")
	utils.Logger.Info("Swagger UI: http://localhost:8080/swagger/index.html")
	if !dbInitialized {
		utils.Logger.Warn("⚠️  Warning: Running in database-free mode - some endpoints may not work")
	}

	if err := http.ListenAndServe(":8080", router); err != nil {
		utils.Logger.WithError(err).Fatal("Failed to start server")
	}
}

// healthCheck godoc
// @Summary      Health check
// @Description  Check if the API server is running
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
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
