package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/handlers"
	"github.com/clarencetw/thewavess-ai-core/middleware"
)

// SetupRoutes 設置所有 API 路由
func SetupRoutes(router *gin.RouterGroup) {
	// 系統管理路由（無需認證）
	router.GET("/version", handlers.GetVersion)
	router.GET("/status", handlers.GetStatus)

	// 用戶認證路由（無需認證）
	userAuth := router.Group("/user")
	{
		userAuth.POST("/register", handlers.RegisterUser)
		userAuth.POST("/login", handlers.LoginUser)
	}

	// 需要認證的路由
	authenticated := router.Group("")
	authenticated.Use(middleware.AuthMiddleware())

	// 用戶管理路由
	user := authenticated.Group("/user")
	{
		user.POST("/logout", handlers.LogoutUser)
		user.POST("/refresh", handlers.RefreshToken)
		user.GET("/profile", handlers.GetProfile)
		user.PUT("/profile", handlers.UpdateProfile)
		user.PUT("/preferences", handlers.UpdatePreferences)
		user.GET("/character", handlers.GetCurrentCharacter)
		user.PUT("/character", handlers.SelectCharacter)
	}

	// 角色管理路由
	character := authenticated.Group("/character")
	{
		character.GET("/list", handlers.GetCharacterList)
		character.GET("/:character_id", handlers.GetCharacterDetails)
		character.GET("/:character_id/stats", handlers.GetCharacterStats)
	}

	// 對話管理路由
	chat := authenticated.Group("/chat")
	{
		chat.POST("/session", handlers.CreateChatSession)
		chat.GET("/session/:session_id", handlers.GetChatSession)
		chat.GET("/sessions", handlers.GetChatSessions)
		chat.POST("/message", handlers.SendMessage)
		chat.POST("/regenerate", handlers.RegenerateMessage)
		chat.PUT("/session/:session_id/mode", handlers.UpdateSessionMode)
		chat.GET("/session/:session_id/history", handlers.GetMessageHistory)
		chat.POST("/session/:session_id/tag", handlers.AddSessionTags)
		chat.GET("/session/:session_id/export", handlers.ExportChatHistory)
		chat.DELETE("/session/:session_id", handlers.DeleteChatSession)
	}

	// TODO: 添加其他模組路由
	// - 小說模式路由
	// - 情感系統路由
	// - TTS 語音路由
	// - 記憶系統路由
	// - 標籤系統路由
}