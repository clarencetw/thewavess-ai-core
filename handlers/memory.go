package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
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

	// 獲取記憶管理器
	memoryManager := services.GetMemoryManager()
	longTermMemory := memoryManager.GetLongTermMemory(userIDStr, characterID)

	// 構建記憶時間線
	var memories []gin.H

	// 添加里程碑記憶
	for _, milestone := range longTermMemory.Milestones {
		memories = append(memories, gin.H{
			"id":           utils.GenerateID(12),
			"type":         "milestone",
			"importance":   "high",
			"title":        milestone.Type,
			"content":      milestone.Description,
			"emotion_tag":  "milestone",
			"timestamp":    milestone.Date,
			"character_id": characterID,
		})
	}

	// 添加偏好記憶
	for _, pref := range longTermMemory.Preferences {
		memories = append(memories, gin.H{
			"id":           utils.GenerateID(12),
			"type":         "preference",
			"importance":   getImportanceLevel(float64(pref.Importance)),
			"title":        pref.Category,
			"content":      pref.Content,
			"emotion_tag":  "neutral",
			"timestamp":    pref.CreatedAt,
			"character_id": characterID,
		})
	}

	// 添加禁忌記憶
	for _, dislike := range longTermMemory.Dislikes {
		memories = append(memories, gin.H{
			"id":           utils.GenerateID(12),
			"type":         "dislike",
			"importance":   "high",
			"title":        "禁忌：" + dislike.Topic,
			"content":      dislike.Evidence,
			"emotion_tag":  "negative",
			"timestamp":    dislike.RecordedAt,
			"character_id": characterID,
		})
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
		"milestone":  len(longTermMemory.Milestones),
		"preference": len(longTermMemory.Preferences),
		"dislike":    len(longTermMemory.Dislikes),
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
			"memory_strength": calculateMemoryStrength(longTermMemory),
			"last_updated":   longTermMemory.LastUpdated,
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
	
	switch req.Type {
	case "preference":
		// 創建偏好記憶
		pref := services.Preference{
			ID:         utils.GenerateID(12),
			Category:   extractCategory(req.Content),
			Content:    req.Content,
			Importance: int(req.Importance),
			Evidence:   req.SessionID,
			CreatedAt:  time.Now(),
		}
		
		// 手動添加到長期記憶
		longTerm := memoryManager.GetLongTermMemory(userIDStr, characterID)
		longTerm.Preferences = append(longTerm.Preferences, pref)
		longTerm.LastUpdated = time.Now()
		
		savedMemory = gin.H{
			"id":         pref.ID,
			"type":       "preference",
			"category":   pref.Category,
			"content":    pref.Content,
			"importance": pref.Importance,
			"created_at": pref.CreatedAt,
		}
		
	case "milestone":
		// 創建里程碑記憶
		milestone := services.Milestone{
			ID:          utils.GenerateID(12),
			Type:        extractMilestoneType(req.Content),
			Description: req.Content,
			Date:        time.Now(),
			Affection:   getCurrentAffection(userIDStr, characterID),
		}
		
		// 手動添加到長期記憶
		longTerm := memoryManager.GetLongTermMemory(userIDStr, characterID)
		longTerm.Milestones = append(longTerm.Milestones, milestone)
		longTerm.LastUpdated = time.Now()
		
		savedMemory = gin.H{
			"id":          milestone.ID,
			"type":        "milestone",
			"milestone_type": milestone.Type,
			"description": milestone.Description,
			"date":        milestone.Date,
			"affection":   milestone.Affection,
		}
		
	case "dislike":
		// 創建禁忌記憶
		dislike := services.Dislike{
			Topic:      extractTopic(req.Content),
			Severity:   int(req.Importance),
			Evidence:   req.Content,
			RecordedAt: time.Now(),
		}
		
		// 手動添加到長期記憶
		longTerm := memoryManager.GetLongTermMemory(userIDStr, characterID)
		longTerm.Dislikes = append(longTerm.Dislikes, dislike)
		longTerm.LastUpdated = time.Now()
		
		savedMemory = gin.H{
			"id":          utils.GenerateID(12),
			"type":        "dislike",
			"topic":       dislike.Topic,
			"severity":    dislike.Severity,
			"evidence":    dislike.Evidence,
			"recorded_at": dislike.RecordedAt,
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

	// 獲取記憶管理器
	memoryManager := GetMemoryManager()
	longTermMemory := memoryManager.GetLongTermMemory(userIDStr, characterID)

	startTime := time.Now()
	var searchResults []gin.H

	// 搜尋偏好
	if memoryType == "" || memoryType == "preference" {
		for _, pref := range longTermMemory.Preferences {
			if matchesQuery(query, pref.Content, pref.Category) {
				searchResults = append(searchResults, gin.H{
					"id":        utils.GenerateID(12),
					"type":      "preference",
					"title":     pref.Category,
					"content":   pref.Content,
					"relevance": calculateRelevance(query, pref.Content),
					"highlight": highlightMatch(query, pref.Content),
					"timestamp": pref.CreatedAt,
				})
			}
		}
	}

	// 搜尋里程碑
	if memoryType == "" || memoryType == "milestone" {
		for _, milestone := range longTermMemory.Milestones {
			if matchesQuery(query, milestone.Description, milestone.Type) {
				searchResults = append(searchResults, gin.H{
					"id":        utils.GenerateID(12),
					"type":      "milestone",
					"title":     milestone.Type,
					"content":   milestone.Description,
					"relevance": calculateRelevance(query, milestone.Description),
					"highlight": highlightMatch(query, milestone.Description),
					"timestamp": milestone.Date,
				})
			}
		}
	}

	// 搜尋禁忌
	if memoryType == "" || memoryType == "dislike" {
		for _, dislike := range longTermMemory.Dislikes {
			if matchesQuery(query, dislike.Evidence, dislike.Topic) {
				searchResults = append(searchResults, gin.H{
					"id":        utils.GenerateID(12),
					"type":      "dislike",
					"title":     "禁忌: " + dislike.Topic,
					"content":   dislike.Evidence,
					"relevance": calculateRelevance(query, dislike.Evidence),
					"highlight": highlightMatch(query, dislike.Evidence),
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

	// 靜態數據回應 - 模擬用戶記憶
	userMemory := gin.H{
		"user_id": userID,
		"summary": gin.H{
			"total_memories":    45,
			"character_memories": map[string]int{
				"陸燁銘": 28,
				"沈言墨": 17,
			},
			"memory_types": gin.H{
				"conversation": 23,
				"preference":   12,
				"emotion":      8,
				"special_event": 2,
			},
			"oldest_memory": time.Now().AddDate(0, -2, 0),
			"newest_memory": time.Now().AddDate(0, 0, -1),
		},
		"recent_memories": []gin.H{
			{
				"id":          "mem_045",
				"type":        "conversation",
				"title":       "關於未來的討論",
				"character":   "陸燁銘",
				"importance":  "high",
				"timestamp":   time.Now().AddDate(0, 0, -1),
			},
			{
				"id":          "mem_044",
				"type":        "preference",
				"title":       "喜歡的咖啡類型",
				"character":   "陸燁銘",
				"importance":  "medium",
				"timestamp":   time.Now().AddDate(0, 0, -3),
			},
		},
		"memory_stats": gin.H{
			"avg_memories_per_day": 1.5,
			"retention_rate":       "95%",
			"memory_quality":       "excellent",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取用戶記憶成功",
		Data:    userMemory,
	})
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

	// 靜態回應 - 模擬記憶遺忘
	result := gin.H{
		"user_id":     userID,
		"memory_id":   req.MemoryID,
		"forget_type": req.ForgetType,
		"result": gin.H{
			"success":     true,
			"action":      "記憶已" + map[string]string{"fade": "淡化", "delete": "刪除"}[req.ForgetType],
			"processed_at": time.Now(),
		},
		"impact": gin.H{
			"related_memories_affected": 3,
			"character_relationship_change": -2,
			"memory_coherence": "maintained",
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

	// 獲取記憶管理器
	memoryManager := GetMemoryManager()
	
	// 獲取全局統計
	globalStats := memoryManager.GetMemoryStatistics()
	
	// 獲取用戶特定記憶
	longTermMemory := memoryManager.GetLongTermMemory(userIDStr, characterID)
	
	// 計算記憶質量分數
	qualityScore := calculateMemoryQuality(longTermMemory)
	
	// 構建統計響應
	stats := gin.H{
		"user_id":      userIDStr,
		"character_id": characterID,
		"overview": gin.H{
			"total_memories":     globalStats["total_preferences"].(int) + globalStats["total_milestones"].(int) + globalStats["total_dislikes"].(int),
			"preferences":        len(longTermMemory.Preferences),
			"milestones":         len(longTermMemory.Milestones),
			"dislikes":          len(longTermMemory.Dislikes),
			"memory_strength":    calculateMemoryStrength(longTermMemory),
			"quality_score":      qualityScore,
			"last_updated":       longTermMemory.LastUpdated,
		},
		"global_stats": globalStats,
		"memory_breakdown": gin.H{
			"preferences": len(longTermMemory.Preferences),
			"milestones":  len(longTermMemory.Milestones),
			"dislikes":    len(longTermMemory.Dislikes),
		},
		"temporal_distribution": calculateTemporalDistribution(longTermMemory),
		"memory_health": gin.H{
			"coherence_score":    qualityScore,
			"retention_rate":     "100%", // 內存中保持100%
			"retrieval_accuracy": "100%", // 直接訪問準確率100%
			"update_frequency":   getUpdateFrequency(longTermMemory.LastUpdated),
		},
		"recommendations": generateMemoryRecommendations(longTermMemory),
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

	// 靜態回應 - 模擬記憶備份
	backup := gin.H{
		"user_id":     userID,
		"backup_id":   utils.GenerateID(16),
		"backup_type": req.BackupType,
		"status":      "completed",
		"created_at":  time.Now(),
		"details": gin.H{
			"total_memories":   156,
			"backed_up":        156,
			"file_size":        "2.8MB",
			"compression":      req.Compression,
			"encryption":       req.Encryption,
			"integrity_check":  "passed",
		},
		"file_info": gin.H{
			"filename":     "memory_backup_" + userID.(string) + "_" + time.Now().Format("20060102"),
			"download_url": "https://example.com/backups/" + utils.GenerateID(32),
			"expires_at":   time.Now().AddDate(0, 0, 7),
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

	// 靜態回應 - 模擬記憶還原
	restore := gin.H{
		"user_id":       userID,
		"backup_id":     req.BackupID,
		"restore_id":    utils.GenerateID(16),
		"status":        "completed",
		"processed_at":  time.Now(),
		"results": gin.H{
			"memories_restored":   142,
			"memories_merged":     8,
			"memories_skipped":    6,
			"conflicts_resolved":  3,
			"integrity_verified":  true,
		},
		"impact": gin.H{
			"relationship_changes": gin.H{
				"陸燁銘": gin.H{"before": 72, "after": 68, "change": -4},
				"沈言墨": gin.H{"before": 58, "after": 61, "change": +3},
			},
			"memory_coherence": "maintained",
			"system_health":    "optimal",
		},
		"warnings": []string{
			"部分記憶時間戳已更新以維持一致性",
			"建議檢查角色關係狀態是否符合預期",
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