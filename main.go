package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading .env file")
	} else {
		log.Println("Successfully loaded .env file")
	}

	// Initialize logger
	utils.InitLogger()

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
	router.Use(utils.RequestIDMiddleware())
	router.Use(utils.RecoverMiddleware())

	// Root path redirect to web interface
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/public/")
	})
	
	// Static files for web interface
	router.Static("/public", "./public")

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

	log.Println("Starting Thewavess AI Core API server on :8080")
	log.Println("Web interface: http://localhost:8080")
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	if !dbInitialized {
		log.Println("⚠️  Warning: Running in database-free mode - some endpoints may not work")
	}
	log.Fatal(http.ListenAndServe(":8080", router))
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
