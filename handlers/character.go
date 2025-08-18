package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// GetCharacterList godoc
// @Summary      獲取角色列表
// @Description  獲取可用角色列表，支援分頁和篩選，無需認證
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Param        type query string false "角色類型篩選" Enums(gentle,dominant,ascetic,sunny,cunning)
// @Param        tags query string false "標籤篩選，多個用逗號分隔"
// @Success      200 {object} models.APIResponse{data=models.CharacterListResponse} "獲取成功"
// @Router       /character/list [get]
func GetCharacterList(c *gin.Context) {
	ctx := context.Background()

	// 解析查詢參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// 構建查詢
	query := database.DB.NewSelect().
		Model((*models.Character)(nil)).
		Where("is_active = ?", true)

	// 應用篩選
	if typeFilter := c.Query("type"); typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}

	if tagsFilter := c.Query("tags"); tagsFilter != "" {
		// PostgreSQL 數組查詢
		query = query.Where("tags && ?", []string{tagsFilter})
	}

	// 獲取總數
	totalCount, err := query.Count(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to count characters")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "無法查詢角色數量",
			},
		})
		return
	}

	// 分頁查詢
	var characters []*models.Character
	err = query.
		Order("popularity DESC", "name ASC").
		Limit(limit).
		Offset((page - 1) * limit).
		Scan(ctx, &characters)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to query characters")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "無法查詢角色列表",
			},
		})
		return
	}

	// 轉換為響應格式
	characterResponses := make([]*models.CharacterResponse, len(characters))
	for i, char := range characters {
		characterResponses[i] = char.ToResponse()
	}

	// 計算分頁信息
	totalPages := (totalCount + limit - 1) / limit

	response := &models.CharacterListResponse{
		Characters: characterResponses,
		Pagination: models.PaginationResponse{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  totalCount,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取角色列表成功",
		Data:    response,
	})
}

// GetCharacterByID godoc
// @Summary      獲取角色詳情
// @Description  獲取特定角色的詳細信息
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        id path string true "角色ID"
// @Success      200 {object} models.APIResponse{data=models.CharacterResponse} "獲取成功"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Router       /character/{id} [get]
func GetCharacterByID(c *gin.Context) {
	ctx := context.Background()
	characterID := c.Param("id")

	var character models.Character
	err := database.DB.NewSelect().
		Model(&character).
		Where("id = ? AND is_active = ?", characterID, true).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("character_id", characterID).Error("Failed to query character")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取角色詳情成功",
		Data:    character.ToResponse(),
	})
}

// CreateCharacter godoc
// @Summary      創建角色
// @Description  創建新角色
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        character body models.Character true "角色信息"
// @Success      201 {object} models.APIResponse{data=models.CharacterResponse} "創建成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Router       /character [post]
func CreateCharacter(c *gin.Context) {
	ctx := context.Background()
	
	var character models.Character
	if err := c.ShouldBindJSON(&character); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "輸入參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 設置基本信息
	character.ID = utils.GenerateID(16)
	character.CreatedAt = time.Now()
	character.UpdatedAt = time.Now()

	// 插入數據庫
	_, err := database.DB.NewInsert().Model(&character).Exec(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to create character")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "創建角色失敗",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "角色創建成功",
		Data:    character.ToResponse(),
	})
}

// UpdateCharacter godoc
// @Summary      更新角色
// @Description  更新角色信息
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "角色ID"
// @Param        character body models.Character true "角色信息"
// @Success      200 {object} models.APIResponse{data=models.CharacterResponse} "更新成功"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Router       /character/{id} [put]
func UpdateCharacter(c *gin.Context) {
	ctx := context.Background()
	characterID := c.Param("id")

	var updateData models.Character
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "輸入參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 更新時間
	updateData.UpdatedAt = time.Now()

	// 執行更新
	result, err := database.DB.NewUpdate().
		Model(&updateData).
		OmitZero().
		Where("id = ? AND is_active = ?", characterID, true).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("character_id", characterID).Error("Failed to update character")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "更新角色失敗",
			},
		})
		return
	}

	// 檢查是否有行被更新
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在",
			},
		})
		return
	}

	// 獲取更新後的角色信息
	var character models.Character
	err = database.DB.NewSelect().
		Model(&character).
		Where("id = ?", characterID).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to fetch updated character")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取更新後角色信息失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "角色更新成功",
		Data:    character.ToResponse(),
	})
}

// DeleteCharacter godoc
// @Summary      刪除角色
// @Description  刪除指定角色 (軟刪除，設為非活躍)
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "角色ID"
// @Success      200 {object} models.APIResponse "刪除成功"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Failure      403 {object} models.APIResponse{error=models.APIError} "無法刪除系統預設角色"
// @Router       /character/{id} [delete]
func DeleteCharacter(c *gin.Context) {
	ctx := context.Background()
	characterID := c.Param("id")

	// 檢查是否為系統預設角色，防止刪除
	systemCharacters := []string{"char_001", "char_002"}
	for _, sysChar := range systemCharacters {
		if characterID == sysChar {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "SYSTEM_CHARACTER_PROTECTED",
					Message: "系統預設角色無法刪除",
				},
			})
			return
		}
	}

	// 檢查角色是否存在
	var character models.Character
	err := database.DB.NewSelect().
		Model(&character).
		Where("id = ? AND is_active = ?", characterID, true).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("character_id", characterID).Error("Character not found for deletion")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在",
			},
		})
		return
	}

	// 軟刪除：設為非活躍狀態
	result, err := database.DB.NewUpdate().
		Model((*models.Character)(nil)).
		Set("is_active = ?", false).
		Set("updated_at = ?", time.Now()).
		Where("id = ? AND is_active = ?", characterID, true).
		Exec(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("character_id", characterID).Error("Failed to delete character")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "刪除角色失敗",
			},
		})
		return
	}

	// 檢查是否有行被更新
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在或已被刪除",
			},
		})
		return
	}

	// 記錄刪除事件
	utils.Logger.WithFields(map[string]interface{}{
		"character_id":   characterID,
		"character_name": character.Name,
		"deleted_by":     "api_user", // 可以從 JWT 中獲取實際用戶ID
	}).Info("Character deleted successfully")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "角色刪除成功",
		Data: map[string]interface{}{
			"character_id":   characterID,
			"character_name": character.Name,
			"deleted_at":     time.Now(),
			"status":         "deleted",
		},
	})
}

// GetCharacterStats godoc
// @Summary      獲取角色統計
// @Description  獲取角色的詳細統計信息
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        id path string true "角色ID"
// @Success      200 {object} models.APIResponse{data=models.CharacterStatsResponse} "獲取成功"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Router       /character/{id}/stats [get]
func GetCharacterStats(c *gin.Context) {
	ctx := context.Background()
	characterID := c.Param("id")

	// 檢查角色是否存在
	var character models.Character
	err := database.DB.NewSelect().
		Model(&character).
		Where("id = ? AND is_active = ?", characterID, true).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithField("character_id", characterID).Error("Character not found for stats")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "CHARACTER_NOT_FOUND",
				Message: "角色不存在",
			},
		})
		return
	}

	// 獲取統計數據
	stats, err := getCharacterStatistics(ctx, characterID, &character)
	if err != nil {
		utils.Logger.WithError(err).WithField("character_id", characterID).Error("Failed to get character statistics")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "STATISTICS_ERROR",
				Message: "獲取統計數據失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取角色統計成功",
		Data:    stats,
	})
}

// GetCurrentCharacter godoc
// @Summary      獲取當前選中角色
// @Description  獲取用戶當前選中的對話角色
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/character [get]
func GetCurrentCharacter(c *gin.Context) {
	// 檢查認證
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	// 靜態數據回應 - 模擬當前角色
	currentCharacter := gin.H{
		"user_id": userID,
		"character": gin.H{
			"id":           "char_001",
			"name":         "陸燁銘",
			"avatar_url":   "https://placehold.co/300x300/darkblue/white?text=陸燁銘",
			"description":  "冷峻霸道的集團總裁，外表冰冷內心火熱",
			"personality":  []string{"霸道", "溫柔", "專情", "成熟"},
			"background":   "陸氏集團年輕總裁，商界傳奇人物",
		},
		"relationship": gin.H{
			"selected_at":     time.Now().AddDate(0, 0, -15),
			"interaction_days": 15,
			"total_messages":   156,
			"affection_level":  72,
			"relationship_status": "戀人未滿",
		},
		"preferences": gin.H{
			"conversation_style": "深度交流",
			"scenario_preference": "現實向",
			"interaction_mode":   "romantic",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取當前角色成功",
		Data:    currentCharacter,
	})
}

// SelectCharacter godoc
// @Summary      選擇對話角色
// @Description  選擇用於對話的角色
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        selection body object true "角色選擇"
// @Success      200 {object} models.APIResponse "選擇成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /user/character [put]
func SelectCharacter(c *gin.Context) {
	// 檢查認證
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var req struct {
		CharacterID string `json:"character_id" binding:"required"`
		Preferences gin.H  `json:"preferences"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "請求參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 靜態數據回應 - 模擬角色選擇
	selection := gin.H{
		"user_id":      userID,
		"character_id": req.CharacterID,
		"character": gin.H{
			"id":   req.CharacterID,
			"name": map[string]string{
				"char_001": "陸燁銘",
				"char_002": "沈言墨",
				"char_003": "顧清歡",
			}[req.CharacterID],
			"type": "romance",
		},
		"previous_character": "char_001",
		"selected_at":       time.Now(),
		"preferences":       req.Preferences,
		"welcome_message":   "很高興再次見到你，我們繼續之前的對話吧。",
		"relationship": gin.H{
			"level":             1,
			"affection":         0,
			"interaction_count": 0,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "角色選擇成功",
		Data:    selection,
	})
}

// getCharacterStatistics 獲取角色統計數據
func getCharacterStatistics(ctx context.Context, characterID string, character *models.Character) (*models.CharacterStatsResponse, error) {
	// 基本信息
	basicInfo := models.CharacterBasicInfo{
		Name:        character.Name,
		Type:        character.Type,
		Description: character.Description,
		Tags:        character.Tags,
		Popularity:  character.Popularity,
		IsActive:    character.IsActive,
		CreatedAt:   character.CreatedAt,
	}

	// 獲取互動統計
	interactionStats, err := getInteractionStats(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// 獲取關係統計
	relationshipStats, err := getRelationshipStats(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// 獲取內容統計
	contentStats, err := getContentStats(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// 獲取用戶偏好統計
	userPreferences, err := getUserPreferencesStats(ctx, characterID)
	if err != nil {
		return nil, err
	}

	return &models.CharacterStatsResponse{
		CharacterID:       characterID,
		BasicInfo:         basicInfo,
		InteractionStats:  interactionStats,
		RelationshipStats: relationshipStats,
		ContentStats:      contentStats,
		UserPreferences:   userPreferences,
		GeneratedAt:       time.Now(),
	}, nil
}

// getInteractionStats 獲取互動統計
func getInteractionStats(ctx context.Context, characterID string) (models.CharacterInteractionStats, error) {
	stats := models.CharacterInteractionStats{
		MessagesByRole: make(map[string]int),
		EngineUsage:    make(map[string]int),
	}

	// 查詢總會話數
	totalSessions, err := database.DB.NewSelect().
		Model((*models.ChatSession)(nil)).
		Where("character_id = ?", characterID).
		Count(ctx)
	if err != nil {
		return stats, err
	}
	stats.TotalConversations = totalSessions

	// 查詢總消息數和角色分布
	var messageStats []struct {
		Role  string `bun:"role"`
		Count int    `bun:"count"`
	}
	err = database.DB.NewSelect().
		Model((*models.Message)(nil)).
		Column("role").
		ColumnExpr("COUNT(*) as count").
		Join("JOIN chat_sessions cs ON cs.id = m.session_id").
		Where("cs.character_id = ?", characterID).
		Group("role").
		Scan(ctx, &messageStats)
	if err != nil {
		return stats, err
	}

	totalMessages := 0
	for _, stat := range messageStats {
		stats.MessagesByRole[stat.Role] = stat.Count
		totalMessages += stat.Count
	}
	stats.TotalMessages = totalMessages

	// 查詢AI引擎使用分布
	var engineStats []struct {
		Engine string `bun:"ai_engine"`
		Count  int    `bun:"count"`
	}
	err = database.DB.NewSelect().
		Model((*models.Message)(nil)).
		Column("ai_engine").
		ColumnExpr("COUNT(*) as count").
		Join("JOIN chat_sessions cs ON cs.id = m.session_id").
		Where("cs.character_id = ? AND ai_engine IS NOT NULL", characterID).
		Group("ai_engine").
		Scan(ctx, &engineStats)
	if err != nil {
		return stats, err
	}

	for _, stat := range engineStats {
		if stat.Engine != "" {
			stats.EngineUsage[stat.Engine] = stat.Count
		}
	}

	// 查詢總用戶數
	err = database.DB.NewSelect().
		Model((*models.ChatSession)(nil)).
		ColumnExpr("COUNT(DISTINCT user_id)").
		Where("character_id = ?", characterID).
		Limit(1).
		Scan(ctx, &stats.TotalUsers)
	if err != nil {
		return stats, err
	}

	// 查詢最後互動時間
	var lastInteraction time.Time
	err = database.DB.NewSelect().
		Model((*models.ChatSession)(nil)).
		Column("last_message_at").
		Where("character_id = ? AND last_message_at IS NOT NULL", characterID).
		Order("last_message_at DESC").
		Limit(1).
		Scan(ctx, &lastInteraction)
	if err == nil {
		stats.LastInteraction = &lastInteraction
	}

	// 計算活躍天數（簡化計算）
	if stats.LastInteraction != nil {
		activeDays := int(time.Since(*stats.LastInteraction).Hours() / 24)
		if activeDays < 1 {
			activeDays = 1
		}
		stats.ActiveDays = activeDays
	}

	// 計算平均會話長度（基於消息數估算，假設每分鐘1條消息）
	if totalSessions > 0 {
		avgMessages := totalMessages / totalSessions
		stats.AvgSessionLength = int64(avgMessages * 60) // Convert to seconds
	}

	return stats, nil
}

// getRelationshipStats 獲取關係統計
func getRelationshipStats(ctx context.Context, characterID string) (models.CharacterRelationshipStats, error) {
	stats := models.CharacterRelationshipStats{
		RelationshipStages:   make(map[string]int),
		MoodDistribution:     make(map[string]int),
		IntimacyLevels:       make(map[string]int),
		EmotionalProgression: []models.EmotionalMilestone{},
	}

	// 查詢情感狀態分布（從消息的情感狀態字段中統計）
	var emotionStats []struct {
		Mood         string `bun:"mood"`
		Relationship string `bun:"relationship"`
		Intimacy     string `bun:"intimacy_level"`
		Count        int    `bun:"count"`
	}

	// 由於這需要從 JSONB 字段中提取數據，我們使用原生SQL查詢
	err := database.DB.NewRaw(`
		SELECT 
			emotional_state->>'mood' as mood,
			emotional_state->>'relationship' as relationship,
			emotional_state->>'intimacy_level' as intimacy_level,
			COUNT(*) as count
		FROM messages m
		JOIN chat_sessions cs ON cs.id = m.session_id
		WHERE cs.character_id = ? AND emotional_state IS NOT NULL
		GROUP BY emotional_state->>'mood', emotional_state->>'relationship', emotional_state->>'intimacy_level'
	`, characterID).Scan(ctx, &emotionStats)

	if err != nil {
		// 如果查詢失敗，使用默認統計
		stats.RelationshipStages["stranger"] = 100
		stats.MoodDistribution["neutral"] = 100
		stats.IntimacyLevels["distant"] = 100
		stats.AvgAffectionLevel = 50.0
	} else {
		// 統計關係階段、心情和親密度分布
		for _, stat := range emotionStats {
			if stat.Relationship != "" {
				stats.RelationshipStages[stat.Relationship] += stat.Count
			}
			if stat.Mood != "" {
				stats.MoodDistribution[stat.Mood] += stat.Count
			}
			if stat.Intimacy != "" {
				stats.IntimacyLevels[stat.Intimacy] += stat.Count
			}
		}

		// 計算平均好感度（簡化計算）
		var avgAffection float64
		err = database.DB.NewRaw(`
			SELECT AVG(CAST(emotional_state->>'affection' AS INTEGER)) as avg_affection
			FROM messages m
			JOIN chat_sessions cs ON cs.id = m.session_id
			WHERE cs.character_id = ? AND emotional_state->>'affection' IS NOT NULL
		`, characterID).Scan(ctx, &avgAffection)
		if err == nil {
			stats.AvgAffectionLevel = avgAffection
		} else {
			stats.AvgAffectionLevel = 50.0
		}
	}

	// 估算關鍵時刻和特殊事件（基於消息數）
	stats.KeyMoments = stats.RelationshipStages["lover"] + stats.RelationshipStages["romantic"]
	stats.SpecialEvents = len(stats.RelationshipStages)

	return stats, nil
}

// getContentStats 獲取內容統計
func getContentStats(ctx context.Context, characterID string) (models.CharacterContentStats, error) {
	stats := models.CharacterContentStats{
		NSFWLevelDistribution: make(map[string]int),
		SceneTypes:            make(map[string]int),
	}

	// 查詢NSFW級別分布
	var nsfwStats []struct {
		Level int `bun:"nsfw_level"`
		Count int `bun:"count"`
	}
	err := database.DB.NewSelect().
		Model((*models.Message)(nil)).
		Column("nsfw_level").
		ColumnExpr("COUNT(*) as count").
		Join("JOIN chat_sessions cs ON cs.id = m.session_id").
		Where("cs.character_id = ?", characterID).
		Group("nsfw_level").
		Scan(ctx, &nsfwStats)

	if err != nil {
		// 默認分布
		stats.NSFWLevelDistribution["level_1"] = 100
	} else {
		for _, stat := range nsfwStats {
			levelKey := "level_" + strconv.Itoa(stat.Level)
			stats.NSFWLevelDistribution[levelKey] = stat.Count
			
			// 統計浪漫場景（NSFW級別2-3）
			if stat.Level >= 2 && stat.Level <= 3 {
				stats.RomanticScenes += stat.Count
			}
		}
	}

	// 查詢重新生成的消息數
	regeneratedCount, err := database.DB.NewSelect().
		Model((*models.Message)(nil)).
		Join("JOIN chat_sessions cs ON cs.id = m.session_id").
		Where("cs.character_id = ? AND is_regenerated = ?", characterID, true).
		Count(ctx)
	if err == nil {
		stats.RegeneratedMessages = regeneratedCount
	}

	// 統計日常對話（NSFW級別1）
	if levelCount, exists := stats.NSFWLevelDistribution["level_1"]; exists {
		stats.DailyConversations = levelCount
	}

	// 估算記憶深刻的引言（基於消息長度）
	longMessagesCount, err := database.DB.NewSelect().
		Model((*models.Message)(nil)).
		Join("JOIN chat_sessions cs ON cs.id = m.session_id").
		Where("cs.character_id = ? AND LENGTH(content) > ?", characterID, 100).
		Count(ctx)
	if err == nil {
		stats.MemorableQuotes = longMessagesCount
	}

	return stats, nil
}

// getUserPreferencesStats 獲取用戶偏好統計
func getUserPreferencesStats(ctx context.Context, characterID string) (models.CharacterUserPreferences, error) {
	stats := models.CharacterUserPreferences{
		SessionModes: make(map[string]int),
	}

	// 查詢會話模式分布
	var modeStats []struct {
		Status string `bun:"status"`
		Count  int    `bun:"count"`
	}
	err := database.DB.NewSelect().
		Model((*models.ChatSession)(nil)).
		Column("status").
		ColumnExpr("COUNT(*) as count").
		Where("character_id = ?", characterID).
		Group("status").
		Scan(ctx, &modeStats)

	if err == nil {
		for _, stat := range modeStats {
			stats.SessionModes[stat.Status] = stat.Count
		}
	}

	// 查詢角色的標籤（從角色表獲取）
	var character models.Character
	err = database.DB.NewSelect().
		Model(&character).
		Where("id = ?", characterID).
		Scan(ctx)

	var tagStats []struct {
		Tag   string `bun:"tag"`
		Count int    `bun:"count"`
	}

	if err == nil && len(character.Tags) > 0 {
		// 將角色標籤轉換為統計格式
		for i, tag := range character.Tags {
			tagStats = append(tagStats, struct {
				Tag   string `bun:"tag"`
				Count int    `bun:"count"`
			}{
				Tag:   tag,
				Count: len(character.Tags) - i, // 簡單的權重計算
			})
		}
	}

	if len(tagStats) > 0 {
		for _, stat := range tagStats {
			stats.PopularTags = append(stats.PopularTags, stat.Tag)
		}
	}

	// 設置默認偏好（基於角色類型）
	stats.FavoriteScenarios = []string{"辦公室", "咖啡廳", "家中"}
	stats.PreferredMoods = []string{"溫柔", "浪漫", "關懷"}
	stats.InteractionStyles = []string{"深度對話", "情感交流", "日常陪伴"}

	return stats, nil
}