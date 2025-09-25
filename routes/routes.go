package routes

import (
	"github.com/clarencetw/thewavess-ai-core/handlers"
	"github.com/clarencetw/thewavess-ai-core/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 設置所有 API 路由
func SetupRoutes(router *gin.RouterGroup) {
	// 系統管理路由（無需認證）
	router.GET("/version", handlers.GetVersion)
	router.GET("/status", handlers.GetStatus)

	// 監控系統路由（無需認證）
	monitor := router.Group("/monitor")
	{
		monitor.GET("/health", handlers.HealthCheck)
		monitor.GET("/ready", handlers.Ready)
		monitor.GET("/live", handlers.Live)
		monitor.GET("/stats", handlers.GetSystemStats)
		monitor.GET("/metrics", handlers.GetMetrics)
		monitor.GET("/baseline", handlers.GetBaseline)
		monitor.POST("/baseline", handlers.UpdateBaseline)
	}

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
		user.POST("/avatar", handlers.UploadAvatar)
		user.DELETE("/account", handlers.DeleteAccount)
	}

	// 公開的角色端點（無需認證）
	publicCharacter := router.Group("/character")
	{
		publicCharacter.GET("/list", handlers.GetCharacterList)
		publicCharacter.GET("/search", handlers.SearchCharacters)
		publicCharacter.GET("/:id", handlers.GetCharacterByID)
		publicCharacter.GET("/:id/stats", handlers.GetCharacterStats)
	}

	// 需要認證的角色端點
	character := authenticated.Group("/character")
	{
		// 基礎角色管理（一般用戶可以創建和管理自己的角色）
		character.POST("", handlers.CreateCharacter)
		character.PUT("/:id", handlers.UpdateCharacter)
		character.DELETE("/:id", handlers.DeleteCharacter)

		// 角色配置管理（一般用戶可查看，管理員可修改）
		character.GET("/:id/profile", handlers.GetCharacterProfile)
	}

	// 對話管理路由
	chats := authenticated.Group("/chats")
	{
		chats.POST("", handlers.CreateChatSession)
		chats.GET("/:chat_id", handlers.GetChatSession)
		chats.GET("", handlers.GetChatSessions)
		chats.POST("/:chat_id/messages", handlers.SendMessage)
		chats.GET("/:chat_id/history", handlers.GetMessageHistory)
		chats.DELETE("/:chat_id", handlers.DeleteChatSession)
		chats.PUT("/:chat_id/mode", handlers.UpdateSessionMode)
		chats.GET("/:chat_id/export", handlers.ExportChatSession)
		chats.POST("/:chat_id/messages/:message_id/regenerate", handlers.RegenerateResponse)
	}

	// 關係系統路由 - 統一使用chat_id在URL路徑中
	relationships := authenticated.Group("/relationships/chat/:chat_id")
	{
		relationships.GET("/status", handlers.GetRelationshipStatus)
		relationships.GET("/affection", handlers.GetAffectionLevel)
		relationships.GET("/history", handlers.GetRelationshipHistory)
	}

	// 小說模式已移除

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
	}

	// TTS 公開路由（語音列表）
	publicTTS := router.Group("/tts")
	{
		publicTTS.GET("/voices", handlers.GetVoiceList)
	}

	// 管理員認證路由（無需認證）
	adminAuth := router.Group("/admin/auth")
	{
		adminAuth.POST("/login", handlers.AdminLogin)
	}

	// 管理員API路由（需要管理員認證）- 保留必要的AJAX API
	adminAPI := router.Group("/admin")
	adminAPI.Use(middleware.AdminMiddleware())
	{
		// 統計資料API（AJAX用）
		adminAPI.GET("/stats", handlers.GetAdminStats)

		// 系統日誌API（AJAX用）
		adminAPI.GET("/logs", handlers.GetAdminLogs)

		// 用戶管理API（AJAX用）
		adminAPI.GET("/users", handlers.GetAdminUsers)
		adminAPI.GET("/users/:id", handlers.GetAdminUserByID)
		adminAPI.PUT("/users/:id", handlers.UpdateAdminUser)
		adminAPI.PUT("/users/:id/password", handlers.UpdateAdminUserPassword)
		adminAPI.PUT("/users/:id/status", handlers.UpdateAdminUserStatus)

		// 聊天記錄API（AJAX用）
		adminAPI.GET("/chats", handlers.AdminSearchChats)
		adminAPI.GET("/chats/:chat_id/history", handlers.AdminGetChatHistory)

		// 角色管理API（AJAX用）
		adminAPI.PUT("/character/:id/status", handlers.UpdateCharacterStatus)
		adminAPI.GET("/characters", handlers.AdminGetCharacters)
		adminAPI.GET("/characters/:id", handlers.AdminGetCharacterByID)
		adminAPI.PUT("/characters/:id", handlers.AdminUpdateCharacter)
		adminAPI.POST("/characters/:id/restore", handlers.AdminRestoreCharacter)
		adminAPI.DELETE("/characters/:id/permanent", handlers.AdminPermanentDeleteCharacter)

		// 管理員管理API（僅超級管理員可訪問）
		adminManagement := adminAPI.Group("/admins")
		adminManagement.Use(middleware.RequireSuperAdmin())
		{
			adminManagement.GET("", handlers.GetAdminList)
			adminManagement.POST("", handlers.CreateAdmin)
		}
	}
}
