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
	
	// 測試路由（無需認證） - 用於開發測試
	router.POST("/test/message", handlers.SendMessage)

	// 認證路由（無需認證）
	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.RegisterUser)
		auth.POST("/login", handlers.LoginUser)
		auth.POST("/refresh", handlers.RefreshToken)
	}

	// 需要認證的路由
	authenticated := router.Group("")
	authenticated.Use(middleware.AuthMiddleware())

	// 認證路由（需要認證）
	authAuthenticated := authenticated.Group("/auth")
	{
		authAuthenticated.POST("/logout", handlers.LogoutUser)
	}

	// 用戶管理路由
	user := authenticated.Group("/user")
	{
		user.GET("/profile", handlers.GetUserProfile)
		user.PUT("/profile", handlers.UpdateUserProfile)
		user.PUT("/preferences", handlers.UpdateUserPreferences)
		user.POST("/avatar", handlers.UploadAvatar)
		user.DELETE("/account", handlers.DeleteAccount)
		user.POST("/verify", handlers.VerifyAge)
		user.GET("/character", handlers.GetCurrentCharacter)
		user.PUT("/character", handlers.SelectCharacter)
	}

	// 角色管理路由
	character := authenticated.Group("/character")
	{
		character.POST("", handlers.CreateCharacter)
		character.PUT("/:id", handlers.UpdateCharacter)
	}

	// 公開的角色端點
	publicCharacter := router.Group("/character")
	{
		publicCharacter.GET("/list", handlers.GetCharacterList)
		publicCharacter.GET("/:id", handlers.GetCharacterByID)
		publicCharacter.GET("/:id/stats", handlers.GetCharacterStats)
	}

	// 對話管理路由
	chat := authenticated.Group("/chat")
	{
		chat.POST("/session", handlers.CreateChatSession)
		chat.GET("/session/:session_id", handlers.GetChatSession)
		chat.GET("/sessions", handlers.GetChatSessions)
		chat.POST("/message", handlers.SendMessage)
		chat.GET("/session/:session_id/history", handlers.GetMessageHistory)
		chat.DELETE("/session/:session_id", handlers.DeleteChatSession)
		chat.PUT("/session/:session_id/mode", handlers.UpdateSessionMode)
		chat.POST("/session/:session_id/tag", handlers.AddSessionTag)
		chat.GET("/session/:session_id/export", handlers.ExportChatSession)
		chat.POST("/regenerate", handlers.RegenerateResponse)
	}

	// 標籤系統路由（公開）
	tags := router.Group("/tags")
	{
		tags.GET("", handlers.GetAllTags)
		tags.GET("/popular", handlers.GetPopularTags)
	}

	// 情感系統路由
	emotion := authenticated.Group("/emotion")
	{
		emotion.GET("/status", handlers.GetEmotionStatus)
		emotion.GET("/affection", handlers.GetAffectionLevel)
		emotion.POST("/event", handlers.TriggerEmotionEvent)
		emotion.GET("/affection/history", handlers.GetAffectionHistory)
		emotion.GET("/milestones", handlers.GetRelationshipMilestones)
	}

	// 記憶系統路由
	memory := authenticated.Group("/memory")
	{
		memory.GET("/timeline", handlers.GetMemoryTimeline)
		memory.POST("/save", handlers.SaveMemory)
		memory.GET("/search", handlers.SearchMemory)
		memory.GET("/user/:id", handlers.GetUserMemory)
		memory.DELETE("/forget", handlers.ForgetMemory)
		memory.GET("/stats", handlers.GetMemoryStats)
		memory.POST("/backup", handlers.BackupMemory)
		memory.POST("/restore", handlers.RestoreMemory)
	}

	// 小說模式路由
	novel := authenticated.Group("/novel")
	{
		novel.POST("/start", handlers.StartNovel)
		novel.POST("/choice", handlers.MakeNovelChoice)
		novel.GET("/progress/:novel_id", handlers.GetNovelProgress)
		novel.GET("/list", handlers.GetNovelList)
		novel.POST("/progress/save", handlers.SaveNovelProgress)
		novel.GET("/progress/list", handlers.GetNovelSaveList)
		novel.GET("/:id/stats", handlers.GetNovelStats)
		novel.DELETE("/progress/:id", handlers.DeleteNovelSave)
	}

	// 搜尋功能路由
	search := authenticated.Group("/search")
	{
		search.GET("/chats", handlers.SearchChats)
		search.GET("/global", handlers.GlobalSearch)
	}

	// TTS 語音系統路由
	tts := authenticated.Group("/tts")
	{
		tts.POST("/generate", handlers.GenerateTTS)
		tts.POST("/batch", handlers.BatchGenerateTTS)
		tts.POST("/preview", handlers.PreviewTTS)
		tts.GET("/history", handlers.GetTTSHistory)
		tts.GET("/config", handlers.GetTTSConfig)
	}

	// TTS 公開路由（語音列表）
	publicTTS := router.Group("/tts")
	{
		publicTTS.GET("/voices", handlers.GetVoiceList)
	}

	// TODO: 添加其他模組路由
	// - 通知系統路由
	// - 資料分析路由
}