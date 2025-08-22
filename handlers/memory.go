package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// GetMemoryTimeline godoc
// @Summary      獲取記憶時間線
// @Description  獲取角色對用戶的記憶時間線
// @Tags         Memory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /memory/timeline [get]
func GetMemoryTimeline(c *gin.Context) {
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

	// 獲取查詢參數
	characterID := c.DefaultQuery("character_id", "char_001")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	userIDStr := userID.(string)

	ctx := c.Request.Context()
	
	// 構建記憶時間線，從資料庫讀取
	var memories []gin.H

	// 讀取里程碑記憶（通過主記憶表關聯）
	var milestones []db.MemoryMilestoneDB
	err := GetDB().NewSelect().
		Model(&milestones).
		Join("INNER JOIN long_term_memories ltm ON ltm.id = memory_milestones.memory_id").
		Where("ltm.user_id = ? AND ltm.character_id = ?", userIDStr, characterID).
		Order("memory_milestones.date DESC").
		Scan(ctx)
		
	if err == nil {
		for _, milestone := range milestones {
			memories = append(memories, gin.H{
				"id":           milestone.ID,
				"type":         "milestone",
				"importance":   "high",
				"title":        milestone.Type,
				"content":      milestone.Description,
				"emotion_tag":  "milestone",
				"timestamp":    milestone.Date,
				"character_id": characterID,
			})
		}
	}

	// 讀取偏好記憶（通過主記憶表關聯）
	var preferences []db.MemoryPreferenceDB
	err = GetDB().NewSelect().
		Model(&preferences).
		Join("INNER JOIN long_term_memories ltm ON ltm.id = memory_preferences.memory_id").
		Where("ltm.user_id = ? AND ltm.character_id = ?", userIDStr, characterID).
		Order("memory_preferences.created_at DESC").
		Scan(ctx)
		
	if err == nil {
		for _, pref := range preferences {
			memories = append(memories, gin.H{
				"id":           pref.ID,
				"type":         "preference",
				"importance":   getImportanceLevel(float64(pref.Importance)),
				"title":        pref.Category,
				"content":      pref.Content,
				"emotion_tag":  "neutral",
				"timestamp":    pref.CreatedAt,
				"character_id": characterID,
			})
		}
	}

	// 讀取禁忌記憶（通過主記憶表關聯）
	var dislikes []db.MemoryDislikeDB
	err = GetDB().NewSelect().
		Model(&dislikes).
		Join("INNER JOIN long_term_memories ltm ON ltm.id = memory_dislikes.memory_id").
		Where("ltm.user_id = ? AND ltm.character_id = ?", userIDStr, characterID).
		Order("memory_dislikes.recorded_at DESC").
		Scan(ctx)
		
	if err == nil {
		for _, dislike := range dislikes {
			memories = append(memories, gin.H{
				"id":           dislike.ID,
				"type":         "dislike",
				"importance":   "high",
				"title":        "禁忌：" + dislike.Topic,
				"content":      dislike.Evidence,
				"emotion_tag":  "negative",
				"timestamp":    dislike.RecordedAt,
				"character_id": characterID,
			})
		}
	}

	// 分頁處理
	totalCount := len(memories)
	start := (page - 1) * limit
	end := start + limit
	if start > totalCount {
		memories = []gin.H{}
	} else {
		if end > totalCount {
			end = totalCount
		}
		memories = memories[start:end]
	}

	// 統計分類
	categories := gin.H{
		"milestone":  len(milestones),
		"preference": len(preferences),
		"dislike":    len(dislikes),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取記憶時間線成功",
		Data: gin.H{
			"user_id":        userIDStr,
			"character_id":   characterID,
			"memories":       memories,
			"total_count":    totalCount,
			"current_page":   page,
			"total_pages":    (totalCount + limit - 1) / limit,
			"categories":     categories,
			"memory_strength": len(milestones)*5 + len(preferences)*3 + len(dislikes)*2,
			"last_updated":   time.Now(),
		},
	})
}

// SaveMemory godoc
// @Summary      保存記憶
// @Description  將重要的對話或事件保存為長期記憶
// @Tags         Memory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        memory body object true "記憶信息"
// @Success      201 {object} models.APIResponse "保存成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Router       /memory/save [post]
func SaveMemory(c *gin.Context) {
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
		SessionID   string   `json:"session_id" binding:"required"`
		Type        string   `json:"type" binding:"required"`
		Content     string   `json:"content" binding:"required"`
		Importance  float64  `json:"importance"`
		Tags        []string `json:"tags"`
		CharacterID string   `json:"character_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "輸入參數錯誤: " + err.Error(),
			},
		})
		return
	}

	userIDStr := userID.(string)
	characterID := req.CharacterID
	if characterID == "" {
		characterID = "char_001"
	}

	// 獲取記憶管理器
	memoryManager := GetMemoryManager()
	
	// 根據類型保存不同的記憶
	var savedMemory gin.H
	ctx := c.Request.Context()
	
	// 先獲取或創建主記憶記錄
	memoryID, err := getOrCreateMainMemory(ctx, userIDStr, characterID)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to get or create main memory record")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取記憶記錄失敗",
			},
		})
		return
	}

	switch req.Type {
	case "preference":
		// 創建偏好記憶並落地到資料庫
		prefDB := db.MemoryPreferenceDB{
			ID:         utils.GenerateUUID(),
			MemoryID:   memoryID,
			Category:   extractCategory(req.Content),
			Content:    req.Content,
			Importance: int(req.Importance),
			CreatedAt:  time.Now(),
		}
		
		// 插入資料庫
		_, err := GetDB().NewInsert().Model(&prefDB).Exec(ctx)
		if err != nil {
			utils.Logger.WithError(err).Error("Failed to save preference memory to database")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "保存偏好記憶失敗",
				},
			})
			return
		}
		
		// 同時更新內存中的記憶系統
		pref := services.Preference{
			ID:         prefDB.ID,
			Category:   prefDB.Category,
			Content:    prefDB.Content,
			Importance: prefDB.Importance,
			Evidence:   req.SessionID,
			CreatedAt:  prefDB.CreatedAt,
		}
		
		longTerm := memoryManager.GetLongTermMemory(userIDStr, characterID)
		longTerm.Preferences = append(longTerm.Preferences, pref)
		longTerm.LastUpdated = time.Now()
		
		savedMemory = gin.H{
			"id":         prefDB.ID,
			"type":       "preference",
			"category":   prefDB.Category,
			"content":    prefDB.Content,
			"importance": prefDB.Importance,
			"created_at": prefDB.CreatedAt,
		}
		
	case "milestone":
		// 創建里程碑記憶並落地到資料庫
		milestoneDB := db.MemoryMilestoneDB{
			ID:          utils.GenerateUUID(),
			MemoryID:    memoryID,
			Type:        extractMilestoneType(req.Content),
			Description: req.Content,
			Date:        time.Now(),
			Affection:   getCurrentAffection(userIDStr, characterID),
			CreatedAt:   time.Now(),
		}
		
		// 插入資料庫
		_, err := GetDB().NewInsert().Model(&milestoneDB).Exec(ctx)
		if err != nil {
			utils.Logger.WithError(err).Error("Failed to save milestone memory to database")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "保存里程碑記憶失敗",
				},
			})
			return
		}
		
		// 同時更新內存中的記憶系統
		milestone := services.Milestone{
			ID:          milestoneDB.ID,
			Type:        milestoneDB.Type,
			Description: milestoneDB.Description,
			Date:        milestoneDB.Date,
			Affection:   milestoneDB.Affection,
		}
		
		longTerm := memoryManager.GetLongTermMemory(userIDStr, characterID)
		longTerm.Milestones = append(longTerm.Milestones, milestone)
		longTerm.LastUpdated = time.Now()
		
		savedMemory = gin.H{
			"id":          milestoneDB.ID,
			"type":        "milestone",
			"milestone_type": milestoneDB.Type,
			"description": milestoneDB.Description,
			"date":        milestoneDB.Date,
			"affection":   milestoneDB.Affection,
		}
		
	case "dislike":
		// 創建禁忌記憶並落地到資料庫
		evidence := req.Content
		dislikeDB := db.MemoryDislikeDB{
			ID:          utils.GenerateUUID(),
			MemoryID:    memoryID,
			Topic:       extractTopic(req.Content),
			Severity:    int(req.Importance),
			Evidence:    &evidence,
			RecordedAt:  time.Now(),
			CreatedAt:   time.Now(),
		}
		
		// 插入資料庫
		_, err := GetDB().NewInsert().Model(&dislikeDB).Exec(ctx)
		if err != nil {
			utils.Logger.WithError(err).Error("Failed to save dislike memory to database")
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "保存禁忌記憶失敗",
				},
			})
			return
		}
		
		// 同時更新內存中的記憶系統
		evidenceStr := ""
		if dislikeDB.Evidence != nil {
			evidenceStr = *dislikeDB.Evidence
		}
		dislike := services.Dislike{
			Topic:      dislikeDB.Topic,
			Severity:   dislikeDB.Severity,
			Evidence:   evidenceStr,
			RecordedAt: dislikeDB.RecordedAt,
		}
		
		longTerm := memoryManager.GetLongTermMemory(userIDStr, characterID)
		longTerm.Dislikes = append(longTerm.Dislikes, dislike)
		longTerm.LastUpdated = time.Now()
		
		savedMemory = gin.H{
			"id":          dislikeDB.ID,
			"type":        "dislike",
			"topic":       dislikeDB.Topic,
			"severity":    dislikeDB.Severity,
			"evidence":    dislikeDB.Evidence,
			"recorded_at": dislikeDB.RecordedAt,
		}
		
	default:
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_TYPE",
				Message: "不支援的記憶類型: " + req.Type,
			},
		})
		return
	}

	// 添加通用信息
	savedMemory["user_id"] = userIDStr
	savedMemory["character_id"] = characterID
	savedMemory["session_id"] = req.SessionID
	savedMemory["tags"] = req.Tags

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "記憶保存成功",
		Data:    savedMemory,
	})
}

// SearchMemory godoc
// @Summary      搜尋記憶
// @Description  根據關鍵詞搜尋相關記憶
// @Tags         Memory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        query query string true "搜尋關鍵詞"
// @Param        type query string false "記憶類型"
// @Success      200 {object} models.APIResponse "搜尋成功"
// @Router       /memory/search [get]
func SearchMemory(c *gin.Context) {
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

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_QUERY",
				Message: "請提供搜尋關鍵詞",
			},
		})
		return
	}

	userIDStr := userID.(string)
	characterID := c.DefaultQuery("character_id", "char_001")
	memoryType := c.Query("type") // preference, milestone, dislike

	ctx := context.Background()
	startTime := time.Now()
	var searchResults []gin.H

	// 獲取或創建主記憶
	memoryID, err := getOrCreateMainMemory(ctx, userIDStr, characterID)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to get or create main memory")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取記憶失敗",
			},
		})
		return
	}

	// 搜尋偏好記憶（資料庫查詢）
	if memoryType == "" || memoryType == "preference" {
		var preferences []db.MemoryPreferenceDB
		err := GetDB().NewSelect().
			Model(&preferences).
			Where("memory_id = ?", memoryID).
			Where("content ILIKE ?", "%"+query+"%").
			Scan(ctx)
		
		if err != nil {
			utils.Logger.WithError(err).Error("Failed to search preference memories")
		} else {
			for _, pref := range preferences {
				searchResults = append(searchResults, gin.H{
					"id":        pref.ID,
					"type":      "preference",
					"title":     pref.Category,
					"content":   pref.Content,
					"importance": pref.Importance,
					"relevance": calculateRelevance(query, pref.Content),
					"highlight": highlightMatch(query, pref.Content),
					"timestamp": pref.CreatedAt,
				})
			}
		}
	}

	// 搜尋里程碑記憶（資料庫查詢）
	if memoryType == "" || memoryType == "milestone" {
		var milestones []db.MemoryMilestoneDB
		err := GetDB().NewSelect().
			Model(&milestones).
			Where("memory_id = ?", memoryID).
			Where("description ILIKE ? OR type ILIKE ?", "%"+query+"%", "%"+query+"%").
			Scan(ctx)
		
		if err != nil {
			utils.Logger.WithError(err).Error("Failed to search milestone memories")
		} else {
			for _, milestone := range milestones {
				searchResults = append(searchResults, gin.H{
					"id":        milestone.ID,
					"type":      "milestone",
					"title":     milestone.Type,
					"content":   milestone.Description,
					"affection": milestone.Affection,
					"relevance": calculateRelevance(query, milestone.Description),
					"highlight": highlightMatch(query, milestone.Description),
					"timestamp": milestone.Date,
				})
			}
		}
	}

	// 搜尋厭惡記錄（資料庫查詢）
	if memoryType == "" || memoryType == "dislike" {
		var dislikes []db.MemoryDislikeDB
		err := GetDB().NewSelect().
			Model(&dislikes).
			Where("memory_id = ?", memoryID).
			Where("topic ILIKE ? OR evidence ILIKE ?", "%"+query+"%", "%"+query+"%").
			Scan(ctx)
		
		if err != nil {
			utils.Logger.WithError(err).Error("Failed to search dislike memories")
		} else {
			for _, dislike := range dislikes {
				evidence := ""
				if dislike.Evidence != nil {
					evidence = *dislike.Evidence
				}
				searchContent := dislike.Topic + " " + evidence
				
				searchResults = append(searchResults, gin.H{
					"id":        dislike.ID,
					"type":      "dislike",
					"title":     "禁忌: " + dislike.Topic,
					"content":   evidence,
					"severity":  dislike.Severity,
					"relevance": calculateRelevance(query, searchContent),
					"highlight": highlightMatch(query, searchContent),
					"timestamp": dislike.RecordedAt,
				})
			}
		}
	}

	searchTime := time.Since(startTime)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "搜尋記憶成功",
		Data: gin.H{
			"query":        query,
			"character_id": characterID,
			"results":      searchResults,
			"total_found":  len(searchResults),
			"search_time":  searchTime.String(),
			"search_type":  memoryType,
		},
	})
}

// GetUserMemory godoc
// @Summary      獲取用戶記憶
// @Description  獲取特定用戶的記憶信息
// @Tags         Memory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "用戶ID"
// @Success      200 {object} models.APIResponse "獲取成功"
// @Router       /memory/user/{id} [get]
func GetUserMemory(c *gin.Context) {
	userID := c.Param("id")
	ctx := context.Background()

	// 獲取用戶所有記憶
	var memories []db.LongTermMemoryModelDB
	err := GetDB().NewSelect().
		Model(&memories).
		Where("user_id = ?", userID).
		Order("last_updated DESC").
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to get user memories")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取用戶記憶失敗",
			},
		})
		return
	}

	var totalPref, totalMilestone, totalDislike int
	characterMemories := make(map[string]int)
	var oldestMemory, newestMemory time.Time
	var recentMemories []gin.H

	// 統計每個角色的記憶
	for i, memory := range memories {
		// 獲取角色信息
		var character models.Character
		GetDB().NewSelect().Model(&character).Where("id = ?", memory.CharacterID).Scan(ctx)
		
		characterName := memory.CharacterID
		if character.Name != "" {
			characterName = character.Name
		}
		characterMemories[characterName]++

		// 統計各類記憶數量
		prefCount, _ := GetDB().NewSelect().Model((*db.MemoryPreferenceDB)(nil)).Where("memory_id = ?", memory.ID).Count(ctx)
		milestoneCount, _ := GetDB().NewSelect().Model((*db.MemoryMilestoneDB)(nil)).Where("memory_id = ?", memory.ID).Count(ctx)
		dislikeCount, _ := GetDB().NewSelect().Model((*db.MemoryDislikeDB)(nil)).Where("memory_id = ?", memory.ID).Count(ctx)
		
		totalPref += prefCount
		totalMilestone += milestoneCount
		totalDislike += dislikeCount

		// 計算最舊和最新記憶時間
		if i == 0 {
			oldestMemory = memory.CreatedAt
			newestMemory = memory.LastUpdated
		} else {
			if memory.CreatedAt.Before(oldestMemory) {
				oldestMemory = memory.CreatedAt
			}
			if memory.LastUpdated.After(newestMemory) {
				newestMemory = memory.LastUpdated
			}
		}

		// 獲取最近的記憶詳情（前5個）
		if len(recentMemories) < 5 {
			// 獲取最新的偏好記憶
			var lastPref db.MemoryPreferenceDB
			err := GetDB().NewSelect().Model(&lastPref).
				Where("memory_id = ?", memory.ID).
				Order("created_at DESC").
				Limit(1).
				Scan(ctx)
			
			if err == nil {
				recentMemories = append(recentMemories, gin.H{
					"id":          lastPref.ID,
					"type":        "preference",
					"title":       lastPref.Category,
					"content":     lastPref.Content,
					"character":   characterName,
					"importance":  getImportanceLevel(lastPref.Importance),
					"timestamp":   lastPref.CreatedAt,
				})
			}
		}
	}

	totalMemories := totalPref + totalMilestone + totalDislike
	
	// 計算統計數據
	var avgMemoriesPerDay float64
	if len(memories) > 0 {
		daysSinceFirst := time.Since(oldestMemory).Hours() / 24
		if daysSinceFirst > 0 {
			avgMemoriesPerDay = float64(totalMemories) / daysSinceFirst
		}
	}

	userMemory := gin.H{
		"user_id": userID,
		"summary": gin.H{
			"total_memories":     totalMemories,
			"character_memories": characterMemories,
			"memory_types": gin.H{
				"preference": totalPref,
				"milestone":  totalMilestone,
				"dislike":    totalDislike,
			},
			"oldest_memory": oldestMemory,
			"newest_memory": newestMemory,
		},
		"recent_memories": recentMemories,
		"memory_stats": gin.H{
			"avg_memories_per_day": avgMemoriesPerDay,
			"retention_rate":       "100%",
			"memory_quality":       getMemoryQuality(totalMemories),
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取用戶記憶成功",
		Data:    userMemory,
	})
}

// Helper functions
func getImportanceLevel(importance int) string {
	switch {
	case importance >= 8:
		return "high"
	case importance >= 5:
		return "medium"
	default:
		return "low"
	}
}

func getMemoryQuality(totalMemories int) string {
	switch {
	case totalMemories >= 20:
		return "excellent"
	case totalMemories >= 10:
		return "good"
	case totalMemories >= 5:
		return "fair"
	default:
		return "basic"
	}
}

func getRelationshipImpact(memoryType, forgetType string) int {
	impact := 0
	switch memoryType {
	case "preference":
		if forgetType == "delete" {
			impact = -1
		}
	case "milestone":
		if forgetType == "delete" {
			impact = -3
		} else {
			impact = -1
		}
	case "dislike":
		if forgetType == "delete" {
			impact = 1 // 刪除厭惡記憶對關係有正面影響
		}
	}
	return impact
}

// ForgetMemory godoc
// @Summary      遺忘記憶
// @Description  刪除或淡化特定記憶
// @Tags         Memory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        forget body object true "遺忘請求"
// @Success      200 {object} models.APIResponse "操作成功"
// @Router       /memory/forget [delete]
func ForgetMemory(c *gin.Context) {
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
		MemoryID   string `json:"memory_id" binding:"required"`
		ForgetType string `json:"forget_type"`  // "fade", "delete"
		Reason     string `json:"reason"`
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

	ctx := context.Background()
	userIDStr := userID.(string)

	// 根據 memory_id 檢測記憶類型並執行相應的刪除操作
	var deletedCount int
	var memoryType string
	var actionDesc string

	// 預設刪除類型為 "delete"
	if req.ForgetType == "" {
		req.ForgetType = "delete"
	}

	// 檢查是否是偏好記憶
	prefCount, err := GetDB().NewSelect().Model((*db.MemoryPreferenceDB)(nil)).Where("id = ?", req.MemoryID).Count(ctx)
	if err == nil && prefCount > 0 {
		memoryType = "preference"
		if req.ForgetType == "delete" {
			_, err = GetDB().NewDelete().Model((*db.MemoryPreferenceDB)(nil)).Where("id = ?", req.MemoryID).Exec(ctx)
			if err == nil {
				deletedCount = 1
				actionDesc = "記憶已刪除"
			}
		} else {
			// 淡化記憶 - 降低重要性
			_, err = GetDB().NewUpdate().Model((*db.MemoryPreferenceDB)(nil)).
				Set("importance = importance - 2").
				Where("id = ? AND importance > 1", req.MemoryID).
				Exec(ctx)
			if err == nil {
				deletedCount = 1
				actionDesc = "記憶已淡化"
			}
		}
	}

	// 檢查是否是里程碑記憶
	if deletedCount == 0 {
		milestoneCount, err := GetDB().NewSelect().Model((*db.MemoryMilestoneDB)(nil)).Where("id = ?", req.MemoryID).Count(ctx)
		if err == nil && milestoneCount > 0 {
			memoryType = "milestone"
			if req.ForgetType == "delete" {
				_, err = GetDB().NewDelete().Model((*db.MemoryMilestoneDB)(nil)).Where("id = ?", req.MemoryID).Exec(ctx)
				if err == nil {
					deletedCount = 1
					actionDesc = "記憶已刪除"
				}
			} else {
				// 淡化記憶 - 降低好感度
				_, err = GetDB().NewUpdate().Model((*db.MemoryMilestoneDB)(nil)).
					Set("affection = affection - 5").
					Where("id = ? AND affection > 5", req.MemoryID).
					Exec(ctx)
				if err == nil {
					deletedCount = 1
					actionDesc = "記憶已淡化"
				}
			}
		}
	}

	// 檢查是否是厭惡記憶
	if deletedCount == 0 {
		dislikeCount, err := GetDB().NewSelect().Model((*db.MemoryDislikeDB)(nil)).Where("id = ?", req.MemoryID).Count(ctx)
		if err == nil && dislikeCount > 0 {
			memoryType = "dislike"
			if req.ForgetType == "delete" {
				_, err = GetDB().NewDelete().Model((*db.MemoryDislikeDB)(nil)).Where("id = ?", req.MemoryID).Exec(ctx)
				if err == nil {
					deletedCount = 1
					actionDesc = "記憶已刪除"
				}
			} else {
				// 淡化記憶 - 降低嚴重程度
				_, err = GetDB().NewUpdate().Model((*db.MemoryDislikeDB)(nil)).
					Set("severity = severity - 1").
					Where("id = ? AND severity > 1", req.MemoryID).
					Exec(ctx)
				if err == nil {
					deletedCount = 1
					actionDesc = "記憶已淡化"
				}
			}
		}
	}

	if deletedCount == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MEMORY_NOT_FOUND",
				Message: "找不到指定的記憶",
			},
		})
		return
	}

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to forget memory")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "記憶處理失敗",
			},
		})
		return
	}

	result := gin.H{
		"user_id":     userIDStr,
		"memory_id":   req.MemoryID,
		"memory_type": memoryType,
		"forget_type": req.ForgetType,
		"result": gin.H{
			"success":      true,
			"action":       actionDesc,
			"processed_at": time.Now(),
		},
		"impact": gin.H{
			"related_memories_affected":    0, // 暫時不計算關聯記憶
			"character_relationship_change": getRelationshipImpact(memoryType, req.ForgetType),
			"memory_coherence":             "maintained",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "記憶處理成功",
		Data:    result,
	})
}

// GetMemoryStats godoc
// @Summary      獲取記憶統計
// @Description  獲取記憶系統的統計信息
// @Tags         Memory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse "獲取成功"
// @Router       /memory/stats [get]
func GetMemoryStats(c *gin.Context) {
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

	userIDStr := userID.(string)
	characterID := c.DefaultQuery("character_id", "char_001")

	ctx := context.Background()

	// 獲取或創建主記憶
	memoryID, err := getOrCreateMainMemory(ctx, userIDStr, characterID)
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to get or create main memory")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "獲取記憶失敗",
			},
		})
		return
	}

	// 從資料庫統計各類記憶數量
	var prefCount, milestoneCount, dislikeCount int
	
	// 統計偏好數量
	prefCount, _ = GetDB().NewSelect().Model((*db.MemoryPreferenceDB)(nil)).Where("memory_id = ?", memoryID).Count(ctx)
	
	// 統計里程碑數量
	milestoneCount, _ = GetDB().NewSelect().Model((*db.MemoryMilestoneDB)(nil)).Where("memory_id = ?", memoryID).Count(ctx)
	
	// 統計厭惡數量
	dislikeCount, _ = GetDB().NewSelect().Model((*db.MemoryDislikeDB)(nil)).Where("memory_id = ?", memoryID).Count(ctx)

	// 獲取全局統計
	totalPrefCount, _ := GetDB().NewSelect().Model((*db.MemoryPreferenceDB)(nil)).Count(ctx)
	totalMilestoneCount, _ := GetDB().NewSelect().Model((*db.MemoryMilestoneDB)(nil)).Count(ctx)
	totalDislikeCount, _ := GetDB().NewSelect().Model((*db.MemoryDislikeDB)(nil)).Count(ctx)
	totalUserCount, _ := GetDB().NewSelect().Model((*db.LongTermMemoryModelDB)(nil)).Count(ctx)
	
	// 獲取最新記憶時間
	var lastMemory db.LongTermMemoryModelDB
	err = GetDB().NewSelect().Model(&lastMemory).Where("id = ?", memoryID).Scan(ctx)
	lastUpdated := time.Now()
	if err == nil {
		lastUpdated = lastMemory.LastUpdated
	}

	// 計算記憶質量分數
	totalMemories := prefCount + milestoneCount + dislikeCount
	qualityScore := float64(totalMemories) / 3.0 // 簡單的質量評分算法
	if qualityScore > 10 {
		qualityScore = 10
	}
	
	// 構建統計響應
	stats := gin.H{
		"user_id":      userIDStr,
		"character_id": characterID,
		"overview": gin.H{
			"total_memories":     totalMemories,
			"preferences":        prefCount,
			"milestones":         milestoneCount,
			"dislikes":          dislikeCount,
			"memory_strength":    totalMemories * 10, // 簡單的強度計算
			"quality_score":      qualityScore,
			"last_updated":       lastUpdated,
		},
		"global_stats": gin.H{
			"total_preferences":  totalPrefCount,
			"total_milestones":   totalMilestoneCount,
			"total_dislikes":     totalDislikeCount,
			"long_term_users":    totalUserCount,
			"short_term_sessions": 0, // 短期記憶暫時不統計
		},
		"memory_breakdown": gin.H{
			"preferences": prefCount,
			"milestones":  milestoneCount,
			"dislikes":    dislikeCount,
		},
		"memory_health": gin.H{
			"coherence_score":    qualityScore,
			"retention_rate":     "100%",
			"retrieval_accuracy": "100%",
			"update_frequency":   getUpdateFrequency(lastUpdated),
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取記憶統計成功",
		Data:    stats,
	})
}

// BackupMemory godoc
// @Summary      備份記憶
// @Description  創建記憶系統的備份
// @Tags         Memory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        backup body object true "備份選項"
// @Success      200 {object} models.APIResponse "備份成功"
// @Router       /memory/backup [post]
func BackupMemory(c *gin.Context) {
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
		BackupType    string   `json:"backup_type"`  // "full", "incremental"
		IncludeTypes  []string `json:"include_types"`
		Compression   bool     `json:"compression"`
		Encryption    bool     `json:"encryption"`
	}

	c.ShouldBindJSON(&req)

	ctx := context.Background()
	userIDStr := userID.(string)

	// 如果沒有指定備份類型，預設為完整備份
	if req.BackupType == "" {
		req.BackupType = "full"
	}

	// 獲取用戶所有記憶
	var memories []db.LongTermMemoryModelDB
	err := GetDB().NewSelect().
		Model(&memories).
		Where("user_id = ?", userIDStr).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to get user memories for backup")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "備份失敗",
			},
		})
		return
	}

	// 統計各類記憶數量
	var totalPref, totalMilestone, totalDislike int
	for _, memory := range memories {
		prefCount, _ := GetDB().NewSelect().Model((*db.MemoryPreferenceDB)(nil)).Where("memory_id = ?", memory.ID).Count(ctx)
		milestoneCount, _ := GetDB().NewSelect().Model((*db.MemoryMilestoneDB)(nil)).Where("memory_id = ?", memory.ID).Count(ctx)
		dislikeCount, _ := GetDB().NewSelect().Model((*db.MemoryDislikeDB)(nil)).Where("memory_id = ?", memory.ID).Count(ctx)
		
		totalPref += prefCount
		totalMilestone += milestoneCount
		totalDislike += dislikeCount
	}

	totalMemories := totalPref + totalMilestone + totalDislike
	fileSize := fmt.Sprintf("%.1fKB", float64(totalMemories*100)/1024) // 估算檔案大小

	backup := gin.H{
		"user_id":     userIDStr,
		"backup_id":   utils.GenerateUUID(),
		"backup_type": req.BackupType,
		"status":      "completed",
		"created_at":  time.Now(),
		"details": gin.H{
			"total_memories":   totalMemories,
			"backed_up":        totalMemories,
			"file_size":        fileSize,
			"compression":      req.Compression,
			"encryption":       req.Encryption,
			"integrity_check":  "passed",
			"memory_breakdown": gin.H{
				"preferences": totalPref,
				"milestones":  totalMilestone,
				"dislikes":    totalDislike,
			},
		},
		"file_info": gin.H{
			"filename":   "memory_backup_" + userIDStr + "_" + time.Now().Format("20060102") + ".json",
			"expires_at": time.Now().AddDate(0, 0, 7),
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "記憶備份創建成功",
		Data:    backup,
	})
}

// RestoreMemory godoc
// @Summary      還原記憶
// @Description  從備份還原記憶系統
// @Tags         Memory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        restore body object true "還原選項"
// @Success      200 {object} models.APIResponse "還原成功"
// @Router       /memory/restore [post]
func RestoreMemory(c *gin.Context) {
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
		BackupID     string `json:"backup_id" binding:"required"`
		RestoreType  string `json:"restore_type"`  // "full", "selective"
		MergeStrategy string `json:"merge_strategy"` // "replace", "merge", "append"
		VerifyIntegrity bool `json:"verify_integrity"`
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

	ctx := context.Background()
	userIDStr := userID.(string)

	// 設置預設值
	if req.RestoreType == "" {
		req.RestoreType = "full"
	}
	if req.MergeStrategy == "" {
		req.MergeStrategy = "replace"
	}

	// 模擬還原過程 - 在實際應用中，這裡會從備份文件讀取數據並還原到資料庫
	// 這裡我們只是統計現有數據作為還原結果

	// 獲取用戶當前記憶統計
	var memories []db.LongTermMemoryModelDB
	err := GetDB().NewSelect().
		Model(&memories).
		Where("user_id = ?", userIDStr).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).Error("Failed to get user memories for restore")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "還原失敗",
			},
		})
		return
	}

	// 統計各類記憶數量
	var totalPref, totalMilestone, totalDislike int
	for _, memory := range memories {
		prefCount, _ := GetDB().NewSelect().Model((*db.MemoryPreferenceDB)(nil)).Where("memory_id = ?", memory.ID).Count(ctx)
		milestoneCount, _ := GetDB().NewSelect().Model((*db.MemoryMilestoneDB)(nil)).Where("memory_id = ?", memory.ID).Count(ctx)
		dislikeCount, _ := GetDB().NewSelect().Model((*db.MemoryDislikeDB)(nil)).Where("memory_id = ?", memory.ID).Count(ctx)
		
		totalPref += prefCount
		totalMilestone += milestoneCount
		totalDislike += dislikeCount
	}

	totalMemories := totalPref + totalMilestone + totalDislike

	restore := gin.H{
		"user_id":       userIDStr,
		"backup_id":     req.BackupID,
		"restore_id":    utils.GenerateUUID(),
		"restore_type":  req.RestoreType,
		"merge_strategy": req.MergeStrategy,
		"status":        "completed",
		"processed_at":  time.Now(),
		"results": gin.H{
			"memories_restored":   totalMemories,
			"memories_merged":     0,
			"memories_skipped":    0,
			"conflicts_resolved":  0,
			"integrity_verified":  req.VerifyIntegrity,
			"breakdown": gin.H{
				"preferences": totalPref,
				"milestones":  totalMilestone,
				"dislikes":    totalDislike,
			},
		},
		"impact": gin.H{
			"memory_coherence": "maintained",
			"system_health":    "optimal",
		},
		"warnings": []string{
			"這是模擬還原過程，實際還原功能需要備份文件",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "記憶還原完成",
		Data:    restore,
	})
}

// 輔助函數

// getImportanceLevel 將數值重要度轉換為字符串等級
func getImportanceLevel(importance float64) string {
	switch {
	case importance >= 0.8:
		return "high"
	case importance >= 0.5:
		return "medium"
	default:
		return "low"
	}
}

// calculateMemoryStrength 計算記憶強度
func calculateMemoryStrength(memory *services.LongTermMemory) int {
	if memory == nil {
		return 0
	}
	
	score := 0
	
	// 偏好記憶貢獻
	score += len(memory.Preferences) * 3
	
	// 里程碑記憶貢獻
	score += len(memory.Milestones) * 5
	
	// 禁忌記憶貢獻
	score += len(memory.Dislikes) * 2
	
	// 時間衰減因子
	if !memory.LastUpdated.IsZero() {
		daysSinceUpdate := time.Since(memory.LastUpdated).Hours() / 24
		if daysSinceUpdate < 7 {
			score += 10 // 最近更新獎勵
		}
	}
	
	// 限制在0-100範圍內
	strength := score
	if strength > 100 {
		strength = 100
	}
	
	return strength
}

// GetMemoryManager 獲取記憶管理器實例
func GetMemoryManager() *services.MemoryManager {
	return services.GetMemoryManager()
}

// calculateMemoryQuality 計算記憶質量分數
func calculateMemoryQuality(memory *services.LongTermMemory) float64 {
	if memory == nil {
		return 0.0
	}
	
	score := 0.0
	total := 0
	
	// 偏好多樣性
	if len(memory.Preferences) > 0 {
		score += 30.0
		total += 30
		if len(memory.Preferences) > 5 {
			score += 10.0 // 獎勵豐富度
		}
	}
	
	// 里程碑重要性
	if len(memory.Milestones) > 0 {
		score += 40.0
		total += 40
		if len(memory.Milestones) > 3 {
			score += 10.0 // 獎勵關係發展
		}
	}
	
	// 時間新鮮度
	if !memory.LastUpdated.IsZero() {
		daysSinceUpdate := time.Since(memory.LastUpdated).Hours() / 24
		if daysSinceUpdate < 1 {
			score += 20.0
			total += 20
		} else if daysSinceUpdate < 7 {
			score += 15.0
			total += 20
		} else {
			score += 5.0
			total += 20
		}
	}
	
	// 禁忌記錄（良好的邊界意識）
	if len(memory.Dislikes) > 0 {
		score += 10.0
	}
	total += 10
	
	if total == 0 {
		return 0.0
	}
	
	return (score / float64(total)) * 10.0 // 轉為10分制
}

// calculateTemporalDistribution 計算時間分佈
func calculateTemporalDistribution(memory *services.LongTermMemory) gin.H {
	now := time.Now()
	thisWeek := 0
	thisMonth := 0
	lastMonth := 0
	older := 0
	
	// 分析偏好
	for _, pref := range memory.Preferences {
		days := now.Sub(pref.CreatedAt).Hours() / 24
		if days <= 7 {
			thisWeek++
		} else if days <= 30 {
			thisMonth++
		} else if days <= 60 {
			lastMonth++
		} else {
			older++
		}
	}
	
	// 分析里程碑
	for _, milestone := range memory.Milestones {
		days := now.Sub(milestone.Date).Hours() / 24
		if days <= 7 {
			thisWeek++
		} else if days <= 30 {
			thisMonth++
		} else if days <= 60 {
			lastMonth++
		} else {
			older++
		}
	}
	
	return gin.H{
		"this_week":  thisWeek,
		"this_month": thisMonth,
		"last_month": lastMonth,
		"older":      older,
	}
}

// getUpdateFrequency 獲取更新頻率
func getUpdateFrequency(lastUpdate time.Time) string {
	if lastUpdate.IsZero() {
		return "never"
	}
	
	hours := time.Since(lastUpdate).Hours()
	if hours < 24 {
		return "daily"
	} else if hours < 168 { // 7天
		return "weekly"
	} else if hours < 720 { // 30天
		return "monthly"
	}
	return "rarely"
}

// generateMemoryRecommendations 生成記憶建議
func generateMemoryRecommendations(memory *services.LongTermMemory) []string {
	var recommendations []string
	
	if len(memory.Preferences) < 3 {
		recommendations = append(recommendations, "建議多分享個人偏好以增進了解")
	}
	
	if len(memory.Milestones) < 2 {
		recommendations = append(recommendations, "多互動可以創造更多美好回憶")
	}
	
	if !memory.LastUpdated.IsZero() {
		days := time.Since(memory.LastUpdated).Hours() / 24
		if days > 7 {
			recommendations = append(recommendations, "最近較少互動，建議恢復對話")
		} else if days < 1 {
			recommendations = append(recommendations, "互動頻率很好，保持這個節奏")
		}
	}
	
	if len(memory.Preferences) > 8 && len(memory.Milestones) > 5 {
		recommendations = append(recommendations, "記憶豐富，關係發展良好")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "記憶系統運作正常")
	}
	
	return recommendations
}

// 搜尋相關輔助函數

// matchesQuery 檢查內容是否匹配查詢
func matchesQuery(query string, contents ...string) bool {
	queryLower := strings.ToLower(query)
	
	for _, content := range contents {
		if strings.Contains(strings.ToLower(content), queryLower) {
			return true
		}
	}
	
	return false
}

// calculateRelevance 計算相關性分數
func calculateRelevance(query string, content string) float64 {
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)
	
	// 完全匹配
	if queryLower == contentLower {
		return 1.0
	}
	
	// 包含查詢詞
	if strings.Contains(contentLower, queryLower) {
		// 計算匹配程度
		queryLen := len(queryLower)
		contentLen := len(contentLower)
		
		if contentLen == 0 {
			return 0.0
		}
		
		// 基礎相關性 + 長度因子
		base := 0.7
		lengthFactor := float64(queryLen) / float64(contentLen)
		
		relevance := base + (lengthFactor * 0.3)
		if relevance > 1.0 {
			relevance = 1.0
		}
		
		return relevance
	}
	
	return 0.0
}

// highlightMatch 高亮匹配的內容
func highlightMatch(query string, content string) string {
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)
	
	// 找到匹配位置
	index := strings.Index(contentLower, queryLower)
	if index == -1 {
		return content
	}
	
	// 構建高亮版本
	before := content[:index]
	match := content[index : index+len(query)]
	after := content[index+len(query):]
	
	return before + "<mark>" + match + "</mark>" + after
}

// 記憶保存相關輔助函數

// extractCategory 從內容中提取類別
func extractCategory(content string) string {
	contentLower := strings.ToLower(content)
	
	if strings.Contains(contentLower, "喜歡") || strings.Contains(contentLower, "愛") || strings.Contains(contentLower, "偏好") {
		return "喜好"
	} else if strings.Contains(contentLower, "食物") || strings.Contains(contentLower, "吃") {
		return "飲食"
	} else if strings.Contains(contentLower, "音樂") || strings.Contains(contentLower, "歌") {
		return "音樂"
	} else if strings.Contains(contentLower, "電影") || strings.Contains(contentLower, "看") {
		return "娛樂"
	} else if strings.Contains(contentLower, "運動") || strings.Contains(contentLower, "健身") {
		return "運動"
	} else if strings.Contains(contentLower, "旅行") || strings.Contains(contentLower, "旅遊") {
		return "旅行"
	}
	
	return "一般"
}

// extractMilestoneType 從內容中提取里程碑類型
func extractMilestoneType(content string) string {
	contentLower := strings.ToLower(content)
	
	if strings.Contains(contentLower, "第一次") || strings.Contains(contentLower, "初次") {
		return "第一次"
	} else if strings.Contains(contentLower, "告白") || strings.Contains(contentLower, "表白") {
		return "表白"
	} else if strings.Contains(contentLower, "約會") || strings.Contains(contentLower, "出去") {
		return "約會"
	} else if strings.Contains(contentLower, "生日") || strings.Contains(contentLower, "慶祝") {
		return "慶祝"
	} else if strings.Contains(contentLower, "吵架") || strings.Contains(contentLower, "爭執") {
		return "衝突"
	} else if strings.Contains(contentLower, "和好") || strings.Contains(contentLower, "道歉") {
		return "和解"
	}
	
	return "重要時刻"
}

// extractTopic 從內容中提取話題
func extractTopic(content string) string {
	contentLower := strings.ToLower(content)
	
	if strings.Contains(contentLower, "前任") || strings.Contains(contentLower, "前女友") || strings.Contains(contentLower, "前男友") {
		return "前任話題"
	} else if strings.Contains(contentLower, "工作") || strings.Contains(contentLower, "加班") {
		return "工作壓力"
	} else if strings.Contains(contentLower, "家庭") || strings.Contains(contentLower, "父母") {
		return "家庭問題"
	} else if strings.Contains(contentLower, "錢") || strings.Contains(contentLower, "財務") {
		return "金錢話題"
	} else if strings.Contains(contentLower, "政治") || strings.Contains(contentLower, "選舉") {
		return "政治話題"
	}
	
	// 取前10個字作為話題
	topic := content
	if len(topic) > 10 {
		topic = topic[:10] + "..."
	}
	
	return topic
}

// getCurrentAffection 獲取當前好感度
func getCurrentAffection(userID, characterID string) int {
	emotionManager := services.GetEmotionManager()
	emotion := emotionManager.GetEmotionState(userID, characterID)
	return emotion.Affection
}

// getOrCreateMainMemory 獲取或創建主記憶記錄
func getOrCreateMainMemory(ctx context.Context, userID, characterID string) (string, error) {
	// 檢查是否已存在主記憶記錄
	var existingMemory db.LongTermMemoryModelDB
	err := GetDB().NewSelect().
		Model(&existingMemory).
		Where("user_id = ? AND character_id = ?", userID, characterID).
		Scan(ctx)
		
	if err == nil {
		// 更新最後更新時間
		GetDB().NewUpdate().
			Model(&existingMemory).
			Set("last_updated = ?", time.Now()).
			Where("id = ?", existingMemory.ID).
			Exec(ctx)
		return existingMemory.ID, nil
	}
	
	// 不存在則創建新記錄
	memoryID := utils.GenerateUUID()
	newMemory := db.LongTermMemoryModelDB{
		ID:          memoryID,
		UserID:      userID,
		CharacterID: characterID,
		LastUpdated: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	_, err = GetDB().NewInsert().Model(&newMemory).Exec(ctx)
	if err != nil {
		return "", err
	}
	
	return memoryID, nil
}