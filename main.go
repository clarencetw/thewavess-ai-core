package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

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
	
	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase()
	
	// Run database migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	
	// Initialize Gin router
	router := gin.Default()
	
	// Add middleware
	router.Use(utils.RequestIDMiddleware())
	router.Use(utils.RecoverMiddleware())

	// Static files for web interface
	router.Static("/public", "./public")
	router.StaticFile("/", "./public/index.html")

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
		"endpoints": map[string]string{
			"web_interface": "http://localhost:8080",
			"swagger_ui":    "http://localhost:8080/swagger/index.html",
			"health_check":  "http://localhost:8080/health",
		},
	})
	
	log.Println("Starting Thewavess AI Core API server on :8080")
	log.Println("Web interface: http://localhost:8080")
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
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
		"status":  "ok",
		"service": "thewavess-ai-core",
		"version": "1.0.0",
	}
	
	// Check database health
	if err := database.HealthCheck(); err != nil {
		response["database"] = "unhealthy"
		response["database_error"] = err.Error()
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}
	
	response["database"] = "healthy"
	c.JSON(http.StatusOK, response)
}