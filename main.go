package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	_ "github.com/clarencetw/thewavess-ai-core/docs"
	"github.com/clarencetw/thewavess-ai-core/routes"
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
	// Initialize Gin router
	router := gin.Default()

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", healthCheck)

	// Setup API routes
	api := router.Group("/api/v1")
	routes.SetupRoutes(api)

	// Start server
	log.Println("Starting Thewavess AI Core API server on :8080")
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
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "thewavess-ai-core",
		"version": "1.0.0",
	})
}