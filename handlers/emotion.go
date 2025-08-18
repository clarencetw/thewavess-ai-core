package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// GetEmotionStatus godoc
// @Summary      獲取情感狀態
// @Description  獲取當前角色的情感狀態
// @Tags         Emotion
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /emotion/status [get]
func GetEmotionStatus(c *gin.Context) {
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
	userIDStr := userID.(string)

	// 使用實際的情感管理器
	emotionManager := services.GetEmotionManager()
	currentEmotion := emotionManager.GetEmotionState(userIDStr, characterID)

	// 獲取情感統計數據
	stats := emotionManager.GetSimpleEmotionStats(userIDStr, characterID)

    // 構建真實的情感狀態響應
    // TODO(擴充建議): 若後端保存了「本輪規則命中明細 explanations」，
    // 可在此一併返回，方便前端顯示「為什麼加/減分」。
    // 例如新增欄位: "explanations": ["命中正向詞 '喜歡' +2", "NSFW 等級 3 +1", "長訊息 +1"]
    emotionStatus := gin.H{
		"user_id":      userIDStr,
		"character_id": characterID,
		"current_emotion": gin.H{
			"type":        currentEmotion.Mood,
			"intensity":   currentEmotion.Affection,
			"description": getEmotionDescription(currentEmotion.Mood, currentEmotion.Affection),
		},
		"relationship": gin.H{
			"status":      currentEmotion.Relationship,
			"intimacy":    currentEmotion.IntimacyLevel,
			"affection":   currentEmotion.Affection,
		},
		"statistics": stats,
		"updated_at": utils.GetCurrentTimestampString(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取情感狀態成功",
		Data:    emotionStatus,
	})
}

// GetAffectionLevel godoc
// @Summary      獲取好感度
// @Description  獲取當前角色對用戶的好感度數據
// @Tags         Emotion
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /emotion/affection [get]
func GetAffectionLevel(c *gin.Context) {
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
	userIDStr := userID.(string)

	// 使用實際的情感管理器
	emotionManager := services.GetEmotionManager()
	currentEmotion := emotionManager.GetEmotionState(userIDStr, characterID)
	stats := emotionManager.GetSimpleEmotionStats(userIDStr, characterID)

	// 計算等級和進度
	levelName, levelTier := getAffectionLevelInfo(currentEmotion.Affection)
	nextLevelThreshold := getNextLevelThreshold(levelTier)
	pointsNeeded := nextLevelThreshold - currentEmotion.Affection

    // 構建真實的好感度響應
    // TODO(擴充建議): 回傳「下一步建議」或「加速升級提示」，
    // 例如根據常命中/未命中規則建議使用者互動方式。
    affectionData := gin.H{
		"user_id":      userIDStr,
		"character_id": characterID,
		"affection_level": gin.H{
			"current":     currentEmotion.Affection,
			"max":         100,
			"level_name":  levelName,
			"level_tier":  levelTier,
			"description": getAffectionDescription(levelTier),
		},
		"progress": gin.H{
			"to_next_level":  nextLevelThreshold,
			"points_needed":  max(0, pointsNeeded),
			"estimated_days": max(1, pointsNeeded/2), // 假設每天平均+2好感度
		},
		"relationship": gin.H{
			"status":    currentEmotion.Relationship,
			"intimacy":  currentEmotion.IntimacyLevel,
			"mood":      currentEmotion.Mood,
		},
		"statistics": stats,
		"updated_at": utils.GetCurrentTimestampString(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取好感度數據成功",
		Data:    affectionData,
	})
}

// TriggerEmotionEvent godoc
// @Summary      觸發情感事件
// @Description  觸發特定的情感事件，影響角色情緒
// @Tags         Emotion
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        event body object true "事件信息"
// @Success      200 {object} models.APIResponse "觸發成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Router       /emotion/event [post]
func TriggerEmotionEvent(c *gin.Context) {
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
		EventType string                 `json:"event_type" binding:"required"`
		Intensity float64                `json:"intensity"`
		Context   map[string]interface{} `json:"context"`
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
    // TODO(擴充建議): 可將這個事件流接入 EmotionManager，
    // 依據 event_type -> 權重表，實際更新情感並返回真實的 before/after。
    eventResponse := gin.H{
		"event_id":   utils.GenerateID(16),
		"user_id":    userID,
		"event_type": req.EventType,
		"result": gin.H{
			"emotion_change": gin.H{
				"before":     "neutral",
				"after":      "happy",
				"delta":      "+15",
			},
			"affection_change": gin.H{
				"before":     68,
				"after":      71,
				"delta":      "+3",
			},
			"unlock_content": []gin.H{
				{
					"type":        "dialogue",
					"id":          "special_001",
					"description": "解鎖了特殊對話選項",
				},
			},
			"character_response": "哇，真的嗎？你這麼說讓我很開心呢～",
		},
		"timestamp": utils.GetCurrentTimestampString(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "情感事件觸發成功",
		Data:    eventResponse,
	})
}

// GetAffectionHistory godoc
// @Summary      獲取好感度歷史
// @Description  獲取角色好感度變化歷史記錄
// @Tags         Emotion
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        character_id query string true "角色ID"
// @Param        days query int false "查詢天數" default(30)
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /emotion/affection/history [get]
func GetAffectionHistory(c *gin.Context) {
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

	characterID := c.Query("character_id")
	if characterID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_CHARACTER_ID",
				Message: "請提供角色ID",
			},
		})
		return
	}

	// 靜態數據回應 - 模擬好感度歷史
	history := gin.H{
		"user_id":      userID,
		"character_id": characterID,
		"character_name": "陸燁銘",
		"current_affection": 72,
		"history": []gin.H{
			{
				"date":       time.Now().AddDate(0, 0, -30),
				"affection":  0,
				"event":      "初次見面",
				"change":     0,
				"trigger":    "character_select",
			},
			{
				"date":       time.Now().AddDate(0, 0, -28),
				"affection":  15,
				"event":      "第一次深度對話",
				"change":     15,
				"trigger":    "meaningful_conversation",
			},
			{
				"date":       time.Now().AddDate(0, 0, -25),
				"affection":  28,
				"event":      "分享個人秘密",
				"change":     13,
				"trigger":    "personal_sharing",
			},
			{
				"date":       time.Now().AddDate(0, 0, -20),
				"affection":  45,
				"event":      "雨夜相伴",
				"change":     17,
				"trigger":    "romantic_moment",
			},
			{
				"date":       time.Now().AddDate(0, 0, -15),
				"affection":  58,
				"event":      "第一次約會",
				"change":     13,
				"trigger":    "special_event",
			},
			{
				"date":       time.Now().AddDate(0, 0, -10),
				"affection":  65,
				"event":      "情感共鳴",
				"change":     7,
				"trigger":    "emotional_connection",
			},
			{
				"date":       time.Now().AddDate(0, 0, -5),
				"affection":  72,
				"event":      "心意相通",
				"change":     7,
				"trigger":    "mutual_understanding",
			},
		},
		"statistics": gin.H{
			"total_interactions": 156,
			"positive_changes":   23,
			"negative_changes":   2,
			"highest_affection":  72,
			"growth_rate":        "2.4/天",
		},
		"milestones": []gin.H{
			{
				"level":        25,
				"name":         "初步信任",
				"achieved_at":  time.Now().AddDate(0, 0, -25),
				"description":  "開始對你產生信任感",
			},
			{
				"level":        50,
				"name":         "心動時刻",
				"achieved_at":  time.Now().AddDate(0, 0, -18),
				"description":  "對你產生了特殊的感情",
			},
			{
				"level":        70,
				"name":         "深度依戀",
				"achieved_at":  time.Now().AddDate(0, 0, -5),
				"description":  "已經深深愛上了你",
			},
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取好感度歷史成功",
		Data:    history,
	})
}

// GetRelationshipMilestones godoc
// @Summary      獲取關係里程碑
// @Description  獲取與角色的關係發展里程碑
// @Tags         Emotion
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        character_id query string true "角色ID"
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /emotion/milestones [get]
func GetRelationshipMilestones(c *gin.Context) {
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

	characterID := c.Query("character_id")
	if characterID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_CHARACTER_ID",
				Message: "請提供角色ID",
			},
		})
		return
	}

	// 靜態數據回應 - 模擬關係里程碑
	milestones := gin.H{
		"user_id":         userID,
		"character_id":    characterID,
		"character_name":  "陸燁銘",
		"current_stage":   "深度依戀期",
		"relationship_level": 72,
		"achieved_milestones": []gin.H{
			{
				"id":             "milestone_001",
				"name":           "初次見面",
				"description":    "第一次與陸燁銘相遇",
				"required_affection": 0,
				"achieved_at":    time.Now().AddDate(0, 0, -30),
				"unlock_content": "解鎖基礎對話模式",
				"special_scene":  "咖啡廳邂逅",
			},
			{
				"id":             "milestone_002", 
				"name":           "破冰時刻",
				"description":    "第一次看到他溫柔的一面",
				"required_affection": 20,
				"achieved_at":    time.Now().AddDate(0, 0, -26),
				"unlock_content": "解鎖溫柔對話選項",
				"special_scene":  "雨夜送傘",
			},
			{
				"id":             "milestone_003",
				"name":           "心動瞬間",
				"description":    "第一次感受到他的在意",
				"required_affection": 40,
				"achieved_at":    time.Now().AddDate(0, 0, -20),
				"unlock_content": "解鎖浪漫場景模式",
				"special_scene":  "辦公室加班",
			},
			{
				"id":             "milestone_004",
				"name":           "情感共鳴",
				"description":    "心靈深度契合的時刻",
				"required_affection": 60,
				"achieved_at":    time.Now().AddDate(0, 0, -12),
				"unlock_content": "解鎖深度情感對話",
				"special_scene":  "海邊漫步",
			},
			{
				"id":             "milestone_005",
				"name":           "深度依戀",
				"description":    "彼此不可分割的深度情感",
				"required_affection": 70,
				"achieved_at":    time.Now().AddDate(0, 0, -5),
				"unlock_content": "解鎖專屬稱呼和親密動作",
				"special_scene":  "星空下的告白",
			},
		},
		"upcoming_milestones": []gin.H{
			{
				"id":             "milestone_006",
				"name":           "心意相通",
				"description":    "完全理解彼此的心意",
				"required_affection": 80,
				"progress":       "90%",
				"unlock_content": "解鎖專屬結局路線",
				"hint":           "繼續深度交流，分享更多內心想法",
			},
			{
				"id":             "milestone_007",
				"name":           "完美結合",
				"description":    "達到最完美的關係狀態",
				"required_affection": 95,
				"progress":       "76%",
				"unlock_content": "解鎖所有特殊內容",
				"hint":           "需要經歷重要的人生抉擇時刻",
			},
		},
		"statistics": gin.H{
			"total_milestones":    7,
			"achieved_count":      5,
			"completion_rate":     "71%",
			"next_milestone_eta":  "3-5天",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取關係里程碑成功",
		Data:    milestones,
	})
}

// 輔助函數

// getEmotionDescription 獲取情感描述
func getEmotionDescription(mood string, affection int) string {
	descriptions := map[string]string{
		"happy":      "角色現在心情很好，對你的回應會更加積極",
		"excited":    "角色感到興奮，期待和你的互動",
		"shy":        "角色有些害羞，但對你很有好感",
		"romantic":   "角色陷入了浪漫的情緒中",
		"passionate": "角色充滿激情，渴望更深入的交流",
		"pleased":    "角色對你很滿意，心情愉悅",
		"loving":     "角色深深愛著你，全心全意",
		"friendly":   "角色對你很友好，樂於交談",
		"polite":     "角色保持著禮貌的態度",
		"neutral":    "角色心情平靜，態度中性",
		"concerned":  "角色對你有些擔心",
		"annoyed":    "角色有些煩躁，需要安撫",
	}
	
	baseDesc := descriptions[mood]
	if baseDesc == "" {
		baseDesc = "角色心情平靜"
	}
	
	if affection >= 80 {
		return baseDesc + "，對你的愛意溢於言表"
	} else if affection >= 60 {
		return baseDesc + "，對你有很深的感情"
	} else if affection >= 40 {
		return baseDesc + "，對你很有好感"
	} else if affection >= 20 {
		return baseDesc + "，開始對你產生興趣"
	}
	
	return baseDesc
}

// getAffectionLevelInfo 獲取好感度等級資訊
func getAffectionLevelInfo(affection int) (string, int) {
	switch {
	case affection >= 90:
		return "摯愛", 5
	case affection >= 70:
		return "戀人", 4
	case affection >= 50:
		return "親密", 3
	case affection >= 25:
		return "友好", 2
	default:
		return "陌生", 1
	}
}

// getNextLevelThreshold 獲取下一等級閾值
func getNextLevelThreshold(currentTier int) int {
	thresholds := map[int]int{
		1: 25,  // 陌生 -> 友好
		2: 50,  // 友好 -> 親密
		3: 70,  // 親密 -> 戀人
		4: 90,  // 戀人 -> 摯愛
		5: 100, // 摯愛 -> 完美
	}
	
	if threshold, exists := thresholds[currentTier]; exists {
		return threshold
	}
	return 100
}

// getAffectionDescription 獲取好感度描述
func getAffectionDescription(tier int) string {
	descriptions := map[int]string{
		1: "角色對你還不太熟悉，保持著基本的禮貌",
		2: "角色開始對你有好感，願意進行友好的交流",
		3: "角色對你有很深的好感，願意分享更多私人話題",
		4: "角色深深愛著你，渴望更親密的關係",
		5: "角色完全愛上了你，你們是完美的伴侶",
	}
	
	if desc, exists := descriptions[tier]; exists {
		return desc
	}
	return descriptions[1]
}

// max 輔助函數
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
