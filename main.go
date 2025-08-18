package main

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/clarencetw/thewavess-ai-core/database"
	_ "github.com/clarencetw/thewavess-ai-core/docs"
	"github.com/clarencetw/thewavess-ai-core/routes"
	"github.com/clarencetw/thewavess-ai-core/utils"
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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// configureStaticFiles 配置靜態檔案中間件
func configureStaticFiles() gin.HandlerFunc {
	return static.Serve("/public", static.LocalFile("./public", false))
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
	// Initialize logger first
	utils.InitLogger()

	// Load environment variables from .env file
	if err := utils.LoadEnv(); err != nil {
		utils.Logger.Warn("No .env file found or error loading .env file")
	} else {
		utils.Logger.Info("Successfully loaded .env file")
	}

	// Initialize database (non-fatal for development/testing)
	dbInitialized := false
	if err := database.InitDB(); err != nil {
		utils.Logger.WithError(err).Warn("Failed to initialize database, running in database-free mode")
	} else {
		dbInitialized = true
		defer database.CloseDB()
		
		// Initialize migrations if database is available
		if err := database.InitMigrations(); err != nil {
			utils.Logger.WithError(err).Warn("Failed to initialize migrations")
		} else {
			utils.Logger.Info("Database and migrations initialized successfully")
		}
	}

	// Initialize Gin router
	router := gin.Default()

	// Add middleware
	router.Use(cors.New(configureCORS()))
	router.Use(utils.RequestIDMiddleware())
	router.Use(utils.RecoverMiddleware())

	// Root path redirect to web interface
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/public/")
	})
	
	// Static files for web interface
	router.Use(configureStaticFiles())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", healthCheck)

	// Setup API routes
	api := router.Group("/api/v1")
	routes.SetupRoutes(api)

	// Start server
	utils.LogServiceEvent("server_starting", map[string]interface{}{
		"port": 8080,
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
	if database.DB == nil {
		response["database"] = "unavailable"
		response["database_message"] = "running in database-free mode"
		response["note"] = "some endpoints may not work without database"
	} else {
		// Try to ping database
		if err := database.DB.Ping(); err != nil {
			response["database"] = "unhealthy"
			response["database_error"] = err.Error()
		} else {
			response["database"] = "healthy"
		}
	}

	// Always return 200 for health check - service is running
	c.JSON(http.StatusOK, response)
}
