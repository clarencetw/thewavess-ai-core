package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
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
		CharacterID string                 `json:"character_id" binding:"required"`
		EventType   string                 `json:"event_type" binding:"required"`
		Intensity   float64                `json:"intensity"`
		Context     map[string]interface{} `json:"context"`
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

	// 獲取情感管理器
	emotionManager := services.GetEmotionManager()
	
	// 獲取觸發前的情感狀態
	beforeState := emotionManager.GetEmotionState(userIDStr, req.CharacterID)
	
	// 根據事件類型計算情感變化
	var affectionChange int
	var moodChange string
	var characterResponse string
	
	switch req.EventType {
	case "gift":
		affectionChange = 5
		moodChange = "happy"
		characterResponse = "謝謝你的禮物，我很喜歡！"
	case "compliment":
		affectionChange = 3
		moodChange = "pleased"
		characterResponse = "你這樣誇我，我好開心呢～"
	case "romantic_gesture":
		affectionChange = 8
		moodChange = "romantic"
		characterResponse = "你這樣做讓我心跳好快..."
	case "deep_conversation":
		affectionChange = 6
		moodChange = "loving"
		characterResponse = "和你聊天總是讓我覺得很有意思"
	case "physical_touch":
		affectionChange = 4
		moodChange = "shy"
		characterResponse = "你...你在做什麼呀，好害羞..."
	case "special_moment":
		affectionChange = 10
		moodChange = "passionate"
		characterResponse = "這一刻我想永遠記住..."
	default:
		affectionChange = 2
		moodChange = "neutral"
		characterResponse = "嗯，我知道了"
	}
	
	// 應用強度調整
	if req.Intensity > 0 {
		affectionChange = int(float64(affectionChange) * req.Intensity)
	}
	
	// 更新情感狀態
	newAffection := beforeState.Affection + affectionChange
	if newAffection > 100 {
		newAffection = 100
	}
	if newAffection < 0 {
		newAffection = 0
	}
	
	// 保存到資料庫
	eventID := utils.GenerateUUID()
	emotionEvent := &db.EmotionStateDB{
		ID:           eventID,
		UserID:       userIDStr,
		CharacterID:  req.CharacterID,
		Affection:    newAffection,
		Mood:         moodChange,
		Relationship: beforeState.Relationship,
		IntimacyLevel: beforeState.IntimacyLevel,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	_, err := database.GetApp().DB().NewInsert().Model(emotionEvent).Exec(c)
	if err != nil {
		utils.Logger.WithError(err).Error("保存情感事件失敗")
		// 繼續執行，不返回錯誤
	}
	
	// 更新情感管理器的狀態 (暫時註釋，方法不存在)
	// emotionManager.UpdateEmotionStateByEvent(userIDStr, req.CharacterID, req.EventType, affectionChange)
	
	// 獲取觸發後的狀態
	afterState := emotionManager.GetEmotionState(userIDStr, req.CharacterID)
	
	// 檢查是否解鎖新內容
	var unlockedContent []gin.H
	if afterState.Affection >= 25 && beforeState.Affection < 25 {
		unlockedContent = append(unlockedContent, gin.H{
			"type":        "dialogue",
			"id":          "friendly_talk",
			"description": "解鎖友好對話選項",
		})
	}
	if afterState.Affection >= 50 && beforeState.Affection < 50 {
		unlockedContent = append(unlockedContent, gin.H{
			"type":        "scene",
			"id":          "intimate_scene",
			"description": "解鎖親密場景",
		})
	}
	if afterState.Affection >= 70 && beforeState.Affection < 70 {
		unlockedContent = append(unlockedContent, gin.H{
			"type":        "special",
			"id":          "lover_mode",
			"description": "解鎖戀人模式",
		})
	}
	
	eventResponse := gin.H{
		"event_id":   eventID,
		"user_id":    userIDStr,
		"character_id": req.CharacterID,
		"event_type": req.EventType,
		"result": gin.H{
			"emotion_change": gin.H{
				"before": beforeState.Mood,
				"after":  afterState.Mood,
				"delta":  fmt.Sprintf("%s -> %s", beforeState.Mood, afterState.Mood),
			},
			"affection_change": gin.H{
				"before": beforeState.Affection,
				"after":  afterState.Affection,
				"delta":  fmt.Sprintf("%+d", affectionChange),
			},
			"unlock_content": unlockedContent,
			"character_response": characterResponse,
		},
		"timestamp": time.Now(),
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

	userIDStr := userID.(string)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	
	// 查詢角色名稱
	var character models.Character
	err := database.GetApp().DB().NewSelect().
		Model(&character).
		Where("id = ?", characterID).
		Scan(c)
	
	characterName := "未知角色"
	if err == nil {
		characterName = character.Name
	}
	
	// 查詢情感狀態歷史
	var emotionStates []db.EmotionStateDB
	err = database.GetApp().DB().NewSelect().
		Model(&emotionStates).
		Where("user_id = ? AND character_id = ? AND created_at >= ?", 
			userIDStr, characterID, time.Now().AddDate(0, 0, -days)).
		Order("created_at ASC").
		Scan(c)
	
	if err != nil {
		utils.Logger.WithError(err).Error("查詢情感歷史失敗")
		// 返回空歷史而非錯誤
		emotionStates = []db.EmotionStateDB{}
	}
	
	// 獲取當前情感狀態
	emotionManager := services.GetEmotionManager()
	currentEmotion := emotionManager.GetEmotionState(userIDStr, characterID)
	
	// 轉換歷史記錄
	var historyEntries []gin.H
	var positiveChanges, negativeChanges int
	var highestAffection int = currentEmotion.Affection
	
	prevAffection := 0
	for i, state := range emotionStates {
		change := state.Affection - prevAffection
		if i > 0 { // 跳過第一筆，因為沒有前一筆比較
			if change > 0 {
				positiveChanges++
			} else if change < 0 {
				negativeChanges++
			}
		}
		
		if state.Affection > highestAffection {
			highestAffection = state.Affection
		}
		
		// 根據好感度變化推斷事件類型
		eventName := getEventNameFromAffectionChange(change, state.Mood)
		
		historyEntries = append(historyEntries, gin.H{
			"date":      state.CreatedAt,
			"affection": state.Affection,
			"event":     eventName,
			"change":    change,
			"trigger":   state.Mood,
		})
		
		prevAffection = state.Affection
	}
	
	// 計算統計數據
	growthRate := "0/天"
	if len(emotionStates) > 1 {
		totalDays := emotionStates[len(emotionStates)-1].CreatedAt.Sub(emotionStates[0].CreatedAt).Hours() / 24
		if totalDays > 0 {
			totalGrowth := emotionStates[len(emotionStates)-1].Affection - emotionStates[0].Affection
			rate := float64(totalGrowth) / totalDays
			growthRate = fmt.Sprintf("%.1f/天", rate)
		}
	}
	
	// 生成里程碑
	var milestones []gin.H
	milestoneThresholds := []int{25, 50, 70, 90}
	milestoneNames := []string{"初步信任", "心動時刻", "深度依戀", "完美愛情"}
	milestoneDescs := []string{"開始對你產生信任感", "對你產生了特殊的感情", "已經深深愛上了你", "達到了完美的愛情狀態"}
	
	for i, threshold := range milestoneThresholds {
		if currentEmotion.Affection >= threshold {
			// 找到達到該閾值的時間點
			var achievedAt time.Time
			for _, state := range emotionStates {
				if state.Affection >= threshold {
					achievedAt = state.CreatedAt
					break
				}
			}
			if achievedAt.IsZero() {
				achievedAt = time.Now() // 如果找不到具體時間，使用當前時間
			}
			
			milestones = append(milestones, gin.H{
				"level":       threshold,
				"name":        milestoneNames[i],
				"achieved_at": achievedAt,
				"description": milestoneDescs[i],
			})
		}
	}
	
	history := gin.H{
		"user_id":           userIDStr,
		"character_id":      characterID,
		"character_name":    characterName,
		"current_affection": currentEmotion.Affection,
		"history":           historyEntries,
		"statistics": gin.H{
			"total_interactions": len(emotionStates),
			"positive_changes":   positiveChanges,
			"negative_changes":   negativeChanges,
			"highest_affection":  highestAffection,
			"growth_rate":        growthRate,
		},
		"milestones": milestones,
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

	userIDStr := userID.(string)
	
	// 查詢角色名稱
	var character models.Character
	err := database.GetApp().DB().NewSelect().
		Model(&character).
		Where("id = ?", characterID).
		Scan(c)
	
	characterName := "未知角色"
	if err == nil {
		characterName = character.Name
	}
	
	// 獲取當前情感狀態
	emotionManager := services.GetEmotionManager()
	currentEmotion := emotionManager.GetEmotionState(userIDStr, characterID)
	
	// 定義里程碑數據
	milestoneData := []struct {
		id              string
		name            string
		description     string
		requiredAffection int
		unlockContent   string
		specialScene    string
	}{
		{"milestone_001", "初次見面", "第一次相遇的特殊時刻", 0, "解鎖基礎對話模式", "初次邂逅"},
		{"milestone_002", "破冰時刻", "開始產生好感的瞬間", 20, "解鎖溫柔對話選項", "溫暖交流"},
		{"milestone_003", "心動瞬間", "第一次感受到特殊情感", 40, "解鎖浪漫場景模式", "心跳加速"},
		{"milestone_004", "情感共鳴", "心靈深度契合的時刻", 60, "解鎖深度情感對話", "心靈相通"},
		{"milestone_005", "深度依戀", "彼此不可分割的深度情感", 70, "解鎖專屬稱呼和親密動作", "深情告白"},
		{"milestone_006", "心意相通", "完全理解彼此的心意", 80, "解鎖專屬結局路線", "心有靈犀"},
		{"milestone_007", "完美結合", "達到最完美的關係狀態", 95, "解鎖所有特殊內容", "完美愛情"},
	}
	
	// 查詢情感歷史來確定里程碑達成時間
	var emotionStates []db.EmotionStateDB
	err = database.GetApp().DB().NewSelect().
		Model(&emotionStates).
		Where("user_id = ? AND character_id = ?", userIDStr, characterID).
		Order("created_at ASC").
		Scan(c)
	
	if err != nil {
		utils.Logger.WithError(err).Error("查詢情感歷史失敗")
		emotionStates = []db.EmotionStateDB{}
	}
	
	// 分類已達成和未達成的里程碑
	var achievedMilestones []gin.H
	var upcomingMilestones []gin.H
	
	for _, milestone := range milestoneData {
		if currentEmotion.Affection >= milestone.requiredAffection {
			// 已達成的里程碑
			var achievedAt time.Time
			
			// 從歷史記錄中找到第一次達到此好感度的時間
			for _, state := range emotionStates {
				if state.Affection >= milestone.requiredAffection {
					achievedAt = state.CreatedAt
					break
				}
			}
			
			// 如果沒有找到歷史記錄，使用預設時間
			if achievedAt.IsZero() {
				daysAgo := (100 - milestone.requiredAffection) / 3 // 簡單的時間估算
				achievedAt = time.Now().AddDate(0, 0, -daysAgo)
			}
			
			achievedMilestones = append(achievedMilestones, gin.H{
				"id":                 milestone.id,
				"name":               milestone.name,
				"description":        milestone.description,
				"required_affection": milestone.requiredAffection,
				"achieved_at":        achievedAt,
				"unlock_content":     milestone.unlockContent,
				"special_scene":      milestone.specialScene,
			})
		} else {
			// 未達成的里程碑
			progress := float64(currentEmotion.Affection) / float64(milestone.requiredAffection) * 100
			if progress > 100 {
				progress = 100
			}
			
			pointsNeeded := milestone.requiredAffection - currentEmotion.Affection
			eta := "未知"
			if pointsNeeded > 0 {
				// 假設平均每天增長2點好感度
				daysNeeded := pointsNeeded / 2
				if daysNeeded < 1 {
					eta = "即將達成"
				} else {
					eta = fmt.Sprintf("%d-%d天", daysNeeded, daysNeeded+2)
				}
			}
			
			upcomingMilestones = append(upcomingMilestones, gin.H{
				"id":                 milestone.id,
				"name":               milestone.name,
				"description":        milestone.description,
				"required_affection": milestone.requiredAffection,
				"progress":           fmt.Sprintf("%.0f%%", progress),
				"unlock_content":     milestone.unlockContent,
				"hint":               getMilestoneHint(milestone.requiredAffection),
				"eta":                eta,
			})
		}
	}
	
	// 計算統計數據
	totalMilestones := len(milestoneData)
	achievedCount := len(achievedMilestones)
	completionRate := float64(achievedCount) / float64(totalMilestones) * 100
	
	// 確定當前階段
	currentStage := getCurrentRelationshipStage(currentEmotion.Affection)
	
	// 計算下一個里程碑的ETA
	nextMilestoneETA := "已完成所有里程碑"
	if len(upcomingMilestones) > 0 {
		milestone := upcomingMilestones[0]
		if eta, exists := milestone["eta"]; exists {
			if etaStr, ok := eta.(string); ok {
				nextMilestoneETA = etaStr
			}
		}
	}
	
	milestones := gin.H{
		"user_id":              userIDStr,
		"character_id":         characterID,
		"character_name":       characterName,
		"current_stage":        currentStage,
		"relationship_level":   currentEmotion.Affection,
		"achieved_milestones":  achievedMilestones,
		"upcoming_milestones":  upcomingMilestones,
		"statistics": gin.H{
			"total_milestones":   totalMilestones,
			"achieved_count":     achievedCount,
			"completion_rate":    fmt.Sprintf("%.0f%%", completionRate),
			"next_milestone_eta": nextMilestoneETA,
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

// getEventNameFromAffectionChange 根據好感度變化推斷事件名稱
func getEventNameFromAffectionChange(change int, mood string) string {
	if change == 0 {
		return "初次見面"
	}
	
	switch {
	case change >= 10:
		return "特殊時刻"
	case change >= 7:
		return "深度交流"
	case change >= 5:
		return "溫馨互動"
	case change >= 3:
		return "愉快對話"
	case change >= 1:
		return "日常互動"
	case change < 0:
		return "情緒波動"
	default:
		return "平常交流"
	}
}

// getCurrentRelationshipStage 根據好感度獲取當前關係階段
func getCurrentRelationshipStage(affection int) string {
	switch {
	case affection >= 95:
		return "完美愛情期"
	case affection >= 80:
		return "心意相通期"
	case affection >= 70:
		return "深度依戀期"
	case affection >= 60:
		return "情感共鳴期"
	case affection >= 40:
		return "心動時期"
	case affection >= 20:
		return "破冰期"
	default:
		return "初識期"
	}
}

// getMilestoneHint 根據里程碑要求獲取提示
func getMilestoneHint(requiredAffection int) string {
	switch requiredAffection {
	case 20:
		return "多進行友好的對話，表現關心"
	case 40:
		return "分享更多個人想法，增加互動頻率"
	case 60:
		return "進行深度交流，表達真摯情感"
	case 70:
		return "表現專一和承諾，建立深度信任"
	case 80:
		return "分享內心秘密，展現完全的理解"
	case 95:
		return "經歷重要決定，證明彼此的重要性"
	default:
		return "繼續保持良好的互動"
	}
}
