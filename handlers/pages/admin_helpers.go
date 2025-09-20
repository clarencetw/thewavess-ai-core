package pages

import (
	"context"
	"fmt"

	"github.com/clarencetw/thewavess-ai-core/handlers"
	dbmodel "github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/gin-gonic/gin"
)

// Helper functions for admin pages

func getSystemStats(ctx context.Context) (gin.H, error) {
	db := handlers.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection unavailable")
	}

	// 用戶統計
	totalUsers, _ := db.NewSelect().Model((*dbmodel.UserDB)(nil)).Count(ctx)
	activeUsers, _ := db.NewSelect().Model((*dbmodel.UserDB)(nil)).Where("status = ?", "active").Count(ctx)
	todayUsers, _ := db.NewSelect().Model((*dbmodel.UserDB)(nil)).Where("DATE(created_at) = CURRENT_DATE").Count(ctx)
	blockedUsers, _ := db.NewSelect().Model((*dbmodel.UserDB)(nil)).Where("status = ?", "banned").Count(ctx)

	// 聊天統計
	totalChats, _ := db.NewSelect().Model((*dbmodel.ChatDB)(nil)).Count(ctx)
	totalMessages, _ := db.NewSelect().Model((*dbmodel.MessageDB)(nil)).Count(ctx)
	todayChats, _ := db.NewSelect().Model((*dbmodel.ChatDB)(nil)).Where("DATE(created_at) = CURRENT_DATE").Count(ctx)
	todayMessages, _ := db.NewSelect().Model((*dbmodel.MessageDB)(nil)).Where("DATE(created_at) = CURRENT_DATE").Count(ctx)

	// 角色統計
	totalCharacters, _ := db.NewSelect().Model((*dbmodel.CharacterDB)(nil)).Count(ctx)
	activeCharacters := totalCharacters // 所有角色都視為活躍

	// 基本系統資訊
	systemInfo := gin.H{
		"Uptime":    "Unknown",
		"Version":   "1.0.0",
		"GoVersion": "1.23",
	}

	return gin.H{
		"Users": gin.H{
			"Total":   totalUsers,
			"Active":  activeUsers,
			"Today":   todayUsers,
			"Blocked": blockedUsers,
		},
		"Chats": gin.H{
			"TotalSessions": totalChats,
			"TotalMessages": totalMessages,
			"TodaySessions": todayChats,
			"TodayMessages": todayMessages,
		},
		"Characters": gin.H{
			"Total":  totalCharacters,
			"Active": activeCharacters,
		},
		"System": systemInfo,
	}, nil
}

func getDefaultStats() gin.H {
	return gin.H{
		"Users": gin.H{
			"Total":   0,
			"Active":  0,
			"Today":   0,
			"Blocked": 0,
		},
		"Chats": gin.H{
			"TotalSessions": 0,
			"TotalMessages": 0,
			"TodaySessions": 0,
			"TodayMessages": 0,
		},
		"Characters": gin.H{
			"Total":  0,
			"Active": 0,
		},
		"System": gin.H{
			"Uptime":    "Unknown",
			"Version":   "1.0.0",
			"GoVersion": "1.23",
		},
	}
}

func getUsersWithStats(ctx context.Context, page, limit int) ([]dbmodel.UserDB, int, gin.H, error) {
	db := handlers.GetDB()
	if db == nil {
		return nil, 0, nil, fmt.Errorf("database connection unavailable")
	}

	// 獲取用戶總數
	totalCount, err := db.NewSelect().Model((*dbmodel.UserDB)(nil)).Count(ctx)
	if err != nil {
		return nil, 0, nil, err
	}

	// 獲取分頁用戶列表
	offset := (page - 1) * limit
	var users []dbmodel.UserDB
	err = db.NewSelect().
		Model(&users).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)
	if err != nil {
		return nil, 0, nil, err
	}

	// 計算用戶統計
	activeCount, _ := db.NewSelect().Model((*dbmodel.UserDB)(nil)).Where("status = ?", "active").Count(ctx)
	todayCount, _ := db.NewSelect().Model((*dbmodel.UserDB)(nil)).Where("DATE(created_at) = CURRENT_DATE").Count(ctx)
	blockedCount, _ := db.NewSelect().Model((*dbmodel.UserDB)(nil)).Where("status = ?", "banned").Count(ctx)

	userStats := gin.H{
		"Total":   totalCount,
		"Active":  activeCount,
		"Today":   todayCount,
		"Blocked": blockedCount,
	}

	return users, totalCount, userStats, nil
}

func getUserStatsDefault() gin.H {
	return gin.H{
		"Total":   0,
		"Active":  0,
		"Today":   0,
		"Blocked": 0,
	}
}

func getChatSessions(ctx context.Context, query, userIDFilter, characterID, dateFrom, dateTo string, page, limit int) ([]gin.H, int, error) {
	db := handlers.GetDB()
	if db == nil {
		return nil, 0, fmt.Errorf("database connection unavailable")
	}

	// Build the base query with filters
	baseQuery := db.NewSelect().Model((*dbmodel.ChatDB)(nil))

	// Apply filters
	if query != "" {
		baseQuery = baseQuery.Where("title ILIKE ?", "%"+query+"%")
	}
	if userIDFilter != "" {
		baseQuery = baseQuery.Where("user_id = ?", userIDFilter)
	}
	if characterID != "" {
		baseQuery = baseQuery.Where("character_id = ?", characterID)
	}
	if dateFrom != "" {
		baseQuery = baseQuery.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		baseQuery = baseQuery.Where("created_at <= ?", dateTo)
	}
	baseQuery = baseQuery.Where("status != ?", "deleted")

	// Get total count
	totalCount, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated chats
	var chats []dbmodel.ChatDB
	err = baseQuery.
		Order("created_at DESC").
		Limit(limit).
		Offset((page-1)*limit).
		Scan(ctx, &chats)

	if err != nil {
		return nil, 0, err
	}

	// Convert to response format with separate queries for user and character data
	result := make([]gin.H, len(chats))
	for i, chat := range chats {
		// Get user data
		var user dbmodel.UserDB
		err = db.NewSelect().Model(&user).Where("id = ?", chat.UserID).Scan(ctx)
		if err != nil {
			// Handle user not found gracefully
			user.Username = "Unknown User"
		}

		// Get character data
		var character dbmodel.CharacterDB
		err = db.NewSelect().Model(&character).Where("id = ?", chat.CharacterID).Scan(ctx)
		if err != nil {
			// Handle character not found gracefully
			character.Name = "Unknown Character"
		}

		result[i] = gin.H{
			"ID":        chat.ID,
			"Title":     chat.Title,
			"Status":    chat.Status,
			"CreatedAt": chat.CreatedAt,
			"UpdatedAt": chat.UpdatedAt,
			"User": gin.H{
				"ID":          user.ID,
				"Username":    user.Username,
				"Email":       user.Email,
				"DisplayName": user.DisplayName,
			},
			"Character": gin.H{
				"ID":        character.ID,
				"Name":      character.Name,
				"AvatarURL": character.AvatarURL,
			},
			"Relationship": gin.H{
				"AffectionLevel":    nil,
				"RelationshipStage": nil,
			},
		}
	}

	return result, totalCount, nil
}

func getCharacterList(ctx context.Context, query, characterType, locale, isActiveFilter string, page, limit int) ([]gin.H, int, error) {
	db := handlers.GetDB()
	if db == nil {
		return nil, 0, fmt.Errorf("database connection unavailable")
	}

	baseQuery := db.NewSelect().
		Model((*dbmodel.CharacterDB)(nil)).
		Where("is_active = ?", true) // 預設只顯示活躍角色

	// Apply filters
	if query != "" {
		baseQuery = baseQuery.Where("name ILIKE ?", "%"+query+"%")
	}

	if characterType != "" {
		baseQuery = baseQuery.Where("type = ?", characterType)
	}

	if locale != "" {
		baseQuery = baseQuery.Where("locale = ?", locale)
	}

	if isActiveFilter == "false" {
		baseQuery = baseQuery.Where("is_active = ?", false)
	} else if isActiveFilter == "all" {
		// 移除預設的 is_active = true 條件
		baseQuery = db.NewSelect().Model((*dbmodel.CharacterDB)(nil))
		if query != "" {
			baseQuery = baseQuery.Where("name ILIKE ?", "%"+query+"%")
		}
		if characterType != "" {
			baseQuery = baseQuery.Where("type = ?", characterType)
		}
		if locale != "" {
			baseQuery = baseQuery.Where("locale = ?", locale)
		}
	}

	// Get total count
	totalCount, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	var characters []dbmodel.CharacterDB

	err = baseQuery.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx, &characters)

	if err != nil {
		return nil, 0, err
	}

	// Convert to response format
	result := make([]gin.H, len(characters))
	for i, char := range characters {
		result[i] = gin.H{
			"ID":              char.ID,
			"Name":            char.Name,
			"Type":            char.Type,
			"Locale":          char.Locale,
			"IsActive":        char.IsActive,
			"AvatarURL":       char.AvatarURL,
			"Popularity":      char.Popularity,
			"Tags":            char.Tags,
			"UserDescription": char.UserDescription,
			"CreatedAt":       char.CreatedAt,
			"UpdatedAt":       char.UpdatedAt,
		}
	}

	return result, totalCount, nil
}

func getCharacterStats(ctx context.Context) (gin.H, error) {
	db := handlers.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection unavailable")
	}

	// 角色統計
	totalCharacters, _ := db.NewSelect().Model((*dbmodel.CharacterDB)(nil)).Count(ctx)
	activeCharacters, _ := db.NewSelect().Model((*dbmodel.CharacterDB)(nil)).Where("is_active = ?", true).Count(ctx)
	inactiveCharacters, _ := db.NewSelect().Model((*dbmodel.CharacterDB)(nil)).Where("is_active = ?", false).Count(ctx)
	todayCharacters, _ := db.NewSelect().Model((*dbmodel.CharacterDB)(nil)).Where("DATE(created_at) = CURRENT_DATE").Count(ctx)

	// 按類型統計
	romanticCount, _ := db.NewSelect().Model((*dbmodel.CharacterDB)(nil)).Where("type = ?", "romantic").Count(ctx)
	friendCount, _ := db.NewSelect().Model((*dbmodel.CharacterDB)(nil)).Where("type = ?", "friend").Count(ctx)
	mentorCount, _ := db.NewSelect().Model((*dbmodel.CharacterDB)(nil)).Where("type = ?", "mentor").Count(ctx)
	fantasyCount, _ := db.NewSelect().Model((*dbmodel.CharacterDB)(nil)).Where("type = ?", "fantasy").Count(ctx)

	return gin.H{
		"Total":    totalCharacters,
		"Active":   activeCharacters,
		"Inactive": inactiveCharacters,
		"Today":    todayCharacters,
		"Types": gin.H{
			"Romantic": romanticCount,
			"Friend":   friendCount,
			"Mentor":   mentorCount,
			"Fantasy":  fantasyCount,
		},
	}, nil
}

func getCharacterStatsDefault() gin.H {
	return gin.H{
		"Total":    0,
		"Active":   0,
		"Inactive": 0,
		"Today":    0,
		"Types": gin.H{
			"Romantic": 0,
			"Friend":   0,
			"Mentor":   0,
			"Fantasy":  0,
		},
	}
}
