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
// @Description  創建新角色（管理員功能）
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
// @Description  更新角色信息（管理員功能）
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

// GetCharacterStats godoc
// @Summary      獲取角色統計
// @Description  獲取角色的詳細統計信息
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        id path string true "角色ID"
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Router       /character/{id}/stats [get]
func GetCharacterStats(c *gin.Context) {
	characterID := c.Param("id")

	// 靜態數據回應 - 模擬角色統計
	stats := gin.H{
		"character_id": characterID,
		"basic_info": gin.H{
			"name":         "陸燁銘",
			"age":          28,
			"profession":   "集團總裁",
			"personality":  "外冷內熱、霸道溫柔",
		},
		"interaction_stats": gin.H{
			"total_conversations": 156,
			"total_messages":      2847,
			"avg_session_length":  "45分鐘",
			"last_interaction":    time.Now().AddDate(0, 0, -1),
			"active_days":         23,
		},
		"relationship_stats": gin.H{
			"affection_level":     72,
			"trust_level":         68,
			"intimacy_level":      45,
			"relationship_stage":  "戀人未滿",
			"key_moments":         8,
		},
		"content_stats": gin.H{
			"romantic_scenes":     34,
			"daily_conversations": 89,
			"special_events":      12,
			"memorable_quotes":    23,
		},
		"user_preferences": gin.H{
			"favorite_scenarios":  []string{"辦公室", "咖啡廳", "家中"},
			"preferred_mood":      "溫柔霸道",
			"interaction_style":   "深度對話",
		},
		"generated_at": time.Now(),
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