package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// StartNovel godoc
// @Summary      開始小說模式
// @Description  開始一個新的互動小說體驗
// @Tags         Novel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        novel body object true "小說設定"
// @Success      201 {object} models.APIResponse "創建成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Router       /novel/start [post]
func StartNovel(c *gin.Context) {
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
		CharacterID string   `json:"character_id" binding:"required"`
		Genre       string   `json:"genre" binding:"required"`
		Setting     string   `json:"setting"`
		Tags        []string `json:"tags"`
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

	// 靜態數據回應
	novel := gin.H{
		"novel_id":     utils.GenerateID(16),
		"user_id":      userID,
		"character_id": req.CharacterID,
		"genre":        req.Genre,
		"setting":      req.Setting,
		"status":       "active",
		"chapter": gin.H{
			"current":     1,
			"title":       "命運的邂逅",
			"scene":       "繁華都市的咖啡廳，午後的陽光透過落地窗灑在桌面上",
			"content":     "你坐在常去的咖啡廳角落，手裡捧著一杯拿鐵，正專注地看著筆記本電腦上的文件。突然，一個熟悉又陌生的聲音在耳邊響起...",
			"atmosphere":  "溫馨而略帶緊張",
		},
		"choices": []gin.H{
			{
				"id":     "choice_001",
				"text":   "抬頭看向聲音的來源",
				"type":   "neutral",
				"hint":   "直接面對可能會有意想不到的發展",
			},
			{
				"id":     "choice_002",
				"text":   "假裝沒聽見，繼續看電腦",
				"type":   "avoid",
				"hint":   "迴避可能會錯過重要的機會",
			},
			{
				"id":     "choice_003",
				"text":   "微笑著回應一聲",
				"type":   "friendly",
				"hint":   "友善的態度總是好的開始",
			},
		},
		"stats": gin.H{
			"affection":   0,
			"tension":     5,
			"plot_points": 0,
			"choices_made": 0,
		},
		"created_at": utils.GetCurrentTimestampString(),
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "小說模式已開始",
		Data:    novel,
	})
}

// MakeNovelChoice godoc
// @Summary      做出選擇
// @Description  在小說中做出劇情選擇
// @Tags         Novel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        choice body object true "選擇信息"
// @Success      200 {object} models.APIResponse "選擇成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Router       /novel/choice [post]
func MakeNovelChoice(c *gin.Context) {
	var req struct {
		NovelID  string `json:"novel_id" binding:"required"`
		ChoiceID string `json:"choice_id" binding:"required"`
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

	// 靜態數據回應 - 根據選擇返回不同結果
	var consequence gin.H
	switch req.ChoiceID {
	case "choice_001":
		consequence = gin.H{
			"text":            "你抬起頭，視線與一雙深邃的眼眸相遇。是陸燁銘，那個在商界呼風喚雨的男人，此刻正站在你面前，嘴角掛著一絲若有似無的微笑。",
			"affection_change": "+5",
			"tension_change":   "+3",
			"unlock":          "special_dialogue_1",
		}
	case "choice_002":
		consequence = gin.H{
			"text":            "你繼續盯著螢幕，但能感覺到那道視線依然停留在你身上。幾秒後，一隻修長的手輕輕合上了你的筆記本。「裝作沒看見我，可不太禮貌。」",
			"affection_change": "-2",
			"tension_change":   "+8",
			"unlock":          "dominant_route",
		}
	default:
		consequence = gin.H{
			"text":            "你抬頭微笑著打了個招呼。對方似乎對你的反應很滿意，拉開對面的椅子坐了下來。",
			"affection_change": "+3",
			"tension_change":   "+1",
			"unlock":          "friendly_route",
		}
	}

	progress := gin.H{
		"novel_id":      req.NovelID,
		"choice_made":   req.ChoiceID,
		"consequence":   consequence,
		"chapter": gin.H{
			"current":    1,
			"progress":   "35%",
			"title":      "命運的邂逅",
		},
		"next_scene": gin.H{
			"content":    consequence["text"],
			"atmosphere": "tension building",
			"bgm":        "soft_tension.mp3",
		},
		"new_choices": []gin.H{
			{
				"id":   "choice_004",
				"text": "「好久不見，沒想到會在這裡遇到你。」",
				"type": "casual",
			},
			{
				"id":   "choice_005",
				"text": "「陸總，這麼巧？」",
				"type": "formal",
			},
			{
				"id":   "choice_006",
				"text": "不說話，等待對方先開口",
				"type": "passive",
			},
		},
		"stats": gin.H{
			"affection":    8,
			"tension":      13,
			"plot_points":  1,
			"choices_made": 1,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "選擇已記錄，故事繼續",
		Data:    progress,
	})
}

// GetNovelProgress godoc
// @Summary      獲取小說進度
// @Description  獲取當前小說的進度信息
// @Tags         Novel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        novel_id path string true "小說ID"
// @Success      200 {object} models.APIResponse "獲取成功"
// @Router       /novel/progress/{novel_id} [get]
func GetNovelProgress(c *gin.Context) {
	novelID := c.Param("novel_id")

	// 靜態數據回應
	progress := gin.H{
		"novel_id": novelID,
		"status":   "in_progress",
		"chapter_info": gin.H{
			"current":       3,
			"total":         10,
			"current_title": "意外的告白",
			"progress":      "67%",
		},
		"storyline": gin.H{
			"main_route":      "romantic",
			"sub_routes":      []string{"friendship", "mystery"},
			"critical_points": 2,
			"endings_unlocked": 0,
		},
		"character_relationships": []gin.H{
			{
				"character_id":   "char_001",
				"character_name": "陸燁銘",
				"relationship":   "戀人未滿",
				"affection":      72,
				"trust":          65,
			},
		},
		"achievements": []gin.H{
			{
				"id":          "ach_001",
				"name":        "初次心動",
				"description": "第一次讓角色心動",
				"unlocked_at": time.Now().AddDate(0, 0, -2),
			},
			{
				"id":          "ach_002",
				"name":        "勇敢告白",
				"description": "主動表達了自己的心意",
				"unlocked_at": time.Now().AddDate(0, 0, -1),
			},
		},
		"save_points": []gin.H{
			{
				"id":         "save_001",
				"chapter":    1,
				"scene":      "咖啡廳初遇",
				"created_at": time.Now().AddDate(0, 0, -5),
			},
			{
				"id":         "save_002",
				"chapter":    2,
				"scene":      "雨夜送傘",
				"created_at": time.Now().AddDate(0, 0, -3),
			},
			{
				"id":         "save_003",
				"chapter":    3,
				"scene":      "告白前夕",
				"created_at": time.Now().AddDate(0, 0, -1),
			},
		},
		"reading_time":    "2h 35m",
		"last_updated":    utils.GetCurrentTimestampString(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取小說進度成功",
		Data:    progress,
	})
}

// GetNovelList godoc
// @Summary      獲取小說列表
// @Description  獲取用戶的小說列表
// @Tags         Novel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse "獲取成功"
// @Router       /novel/list [get]
func GetNovelList(c *gin.Context) {
	// 靜態數據回應
	novels := []gin.H{
		{
			"novel_id":       "novel_001",
			"title":          "霸道總裁的溫柔",
			"character_name": "陸燁銘",
			"genre":          "現代言情",
			"status":         "in_progress",
			"progress":       "35%",
			"last_played":    time.Now().AddDate(0, 0, -1),
			"cover_image":    "https://placehold.co/300x400/purple/white?text=Novel",
		},
		{
			"novel_id":       "novel_002",
			"title":          "古風醫者傳",
			"character_name": "沈言墨",
			"genre":          "古風",
			"status":         "completed",
			"progress":       "100%",
			"last_played":    time.Now().AddDate(0, 0, -7),
			"cover_image":    "https://placehold.co/300x400/brown/white?text=Novel",
			"ending":         "完美結局",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取小說列表成功",
		Data: gin.H{
			"novels":      novels,
			"total_count": len(novels),
			"statistics": gin.H{
				"completed":      1,
				"in_progress":    1,
				"total_reading":  "15h 42m",
				"endings_unlocked": 3,
			},
		},
	})
}

// SaveNovelProgress godoc
// @Summary      保存小說進度
// @Description  手動保存小說進度到存檔點
// @Tags         Novel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        save body object true "保存請求"
// @Success      200 {object} models.APIResponse "保存成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /novel/progress/save [post]
func SaveNovelProgress(c *gin.Context) {
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
		NovelID     string `json:"novel_id" binding:"required"`
		SaveName    string `json:"save_name" binding:"required"`
		Description string `json:"description"`
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

	// 靜態回應 - 模擬進度保存
	saveData := gin.H{
		"save_id":    utils.GenerateID(16),
		"novel_id":   req.NovelID,
		"user_id":    userID,
		"save_name":  req.SaveName,
		"description": req.Description,
		"progress_snapshot": gin.H{
			"chapter":        3,
			"scene":          "告白前夕",
			"affection":      72,
			"tension":        15,
			"choices_made":   23,
			"unlocked_routes": []string{"romantic", "friendship"},
		},
		"game_state": gin.H{
			"character_relationships": gin.H{
				"陸燁銘": gin.H{"affection": 72, "trust": 68},
			},
			"unlocked_content": []string{"special_scene_003", "exclusive_dialogue_015"},
			"achievements":     []string{"first_kiss", "heart_to_heart"},
		},
		"saved_at": time.Now(),
		"file_size": "1.2KB",
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "小說進度保存成功",
		Data:    saveData,
	})
}

// GetNovelSaveList godoc
// @Summary      獲取存檔列表
// @Description  獲取用戶的小說存檔列表
// @Tags         Novel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        novel_id query string false "小說ID過濾"
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /novel/progress/list [get]
func GetNovelSaveList(c *gin.Context) {
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

	novelID := c.Query("novel_id")

	// 靜態數據回應 - 模擬存檔列表
	saves := []gin.H{
		{
			"save_id":     "save_001",
			"novel_id":    "novel_001",
			"novel_title": "霸道總裁的溫柔",
			"save_name":   "第一次心動",
			"description": "咖啡廳初遇後的心跳瞬間",
			"chapter":     1,
			"scene":       "命運的邂逅",
			"progress":    "25%",
			"saved_at":    time.Now().AddDate(0, 0, -10),
			"auto_save":   false,
		},
		{
			"save_id":     "save_002",
			"novel_id":    "novel_001", 
			"novel_title": "霸道總裁的溫柔",
			"save_name":   "雨夜相伴",
			"description": "深夜加班時的溫暖時光",
			"chapter":     2,
			"scene":       "雨夜送傘",
			"progress":    "55%",
			"saved_at":    time.Now().AddDate(0, 0, -5),
			"auto_save":   false,
		},
		{
			"save_id":     "save_003",
			"novel_id":    "novel_001",
			"novel_title": "霸道總裁的溫柔",
			"save_name":   "自動存檔",
			"description": "系統自動保存",
			"chapter":     3,
			"scene":       "告白前夕",
			"progress":    "78%",
			"saved_at":    time.Now().AddDate(0, 0, -1),
			"auto_save":   true,
		},
	}

	// 如果提供了 novel_id，進行過濾
	if novelID != "" {
		filtered := []gin.H{}
		for _, save := range saves {
			if save["novel_id"] == novelID {
				filtered = append(filtered, save)
			}
		}
		saves = filtered
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取存檔列表成功",
		Data: gin.H{
			"user_id":     userID,
			"novel_id":    novelID,
			"saves":       saves,
			"total_count": len(saves),
			"statistics": gin.H{
				"manual_saves": 2,
				"auto_saves":   1,
				"storage_used": "3.6KB",
				"oldest_save":  time.Now().AddDate(0, 0, -10),
			},
		},
	})
}

// GetNovelStats godoc
// @Summary      獲取小說統計
// @Description  獲取特定小說的詳細統計信息
// @Tags         Novel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "小說ID"
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /novel/{id}/stats [get]
func GetNovelStats(c *gin.Context) {
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

	novelID := c.Param("id")

	// 靜態數據回應 - 模擬小說統計
	stats := gin.H{
		"novel_id": novelID,
		"user_id":  userID,
		"basic_info": gin.H{
			"title":       "霸道總裁的溫柔",
			"genre":       "現代言情",
			"character":   "陸燁銘",
			"total_chapters": 10,
			"word_count":  "約 50,000 字",
		},
		"progress_stats": gin.H{
			"current_chapter":    3,
			"progress_percentage": 35,
			"choices_made":      23,
			"scenes_unlocked":   15,
			"total_reading_time": "5h 23m",
			"sessions_played":   8,
		},
		"relationship_stats": gin.H{
			"affection_level":     72,
			"relationship_stage":  "深度依戀",
			"key_moments":         5,
			"romantic_scenes":     8,
			"special_events":      3,
		},
		"achievement_stats": gin.H{
			"total_achievements": 15,
			"unlocked":          8,
			"completion_rate":   "53%",
			"rare_achievements": 2,
			"recent_unlocks": []gin.H{
				{
					"id":          "ach_008",
					"name":        "深度依戀",
					"unlocked_at": time.Now().AddDate(0, 0, -2),
				},
			},
		},
		"ending_progress": gin.H{
			"available_endings": 5,
			"unlocked_endings":  0,
			"closest_ending": gin.H{
				"name":           "完美戀人結局",
				"requirements":   "好感度達到 85",
				"current_progress": "85%",
			},
		},
		"choices_analysis": gin.H{
			"romantic_choices":    15,
			"friendly_choices":    6,
			"assertive_choices":   2,
			"dominant_route":      false,
			"submissive_route":    false,
			"balanced_route":      true,
		},
		"recommendations": []string{
			"繼續與陸燁銘深度交流以提升好感度",
			"嘗試做出更多浪漫選擇解鎖特殊劇情",
			"即將解鎖完美戀人結局，請繼續努力",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取小說統計成功",
		Data:    stats,
	})
}

// DeleteNovelSave godoc
// @Summary      刪除存檔
// @Description  刪除特定的小說存檔
// @Tags         Novel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "存檔ID"
// @Success      200 {object} models.APIResponse "刪除成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /novel/progress/{id} [delete]
func DeleteNovelSave(c *gin.Context) {
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

	saveID := c.Param("id")

	// 靜態回應 - 模擬存檔刪除
	result := gin.H{
		"save_id":     saveID,
		"user_id":     userID,
		"deleted_at":  time.Now(),
		"save_info": gin.H{
			"save_name":   "雨夜相伴",
			"novel_title": "霸道總裁的溫柔",
			"chapter":     2,
			"progress":    "55%",
		},
		"impact": gin.H{
			"storage_freed": "1.2KB",
			"remaining_saves": 2,
			"backup_available": false,
		},
		"warning": "存檔一旦刪除無法恢復，請確認操作",
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "存檔刪除成功",
		Data:    result,
	})
}