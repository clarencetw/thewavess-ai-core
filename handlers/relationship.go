package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
)

// RelationshipStatusResponse 關係狀態響應結構
type RelationshipStatusResponse struct {
	UserID              string                  `json:"user_id"`
	CharacterID         string                  `json:"character_id"`
	ChatID              string                  `json:"chat_id"`
	CurrentRelationship CurrentRelationshipInfo `json:"current_relationship"`
	RelationshipContext RelationshipContextInfo `json:"relationship_context"`
	UpdatedAt           string                  `json:"updated_at"`
}

// CurrentRelationshipInfo 當前關係信息
type CurrentRelationshipInfo struct {
	Type        string `json:"type"`
	Intensity   int    `json:"intensity"`
	Description string `json:"description"`
}

// RelationshipContextInfo 關係上下文信息
type RelationshipContextInfo struct {
	Stability       string   `json:"stability"`
	LastInteraction int      `json:"last_interaction"`
	RecentTriggers  []string `json:"recent_triggers"`
}

// AffectionLevelResponse 好感度響應結構
type AffectionLevelResponse struct {
	UserID             string                 `json:"user_id"`
	CharacterID        string                 `json:"character_id"`
	ChatID             string                 `json:"chat_id"`
	AffectionLevel     AffectionLevelInfo     `json:"affection_level"`
	Relationship       RelationshipInfo       `json:"relationship"`
	Progress           ProgressInfo           `json:"progress"`
	InteractionSummary InteractionSummaryInfo `json:"interaction_summary"`
	UpdatedAt          string                 `json:"updated_at"`
}

// AffectionLevelInfo 好感度等級信息
type AffectionLevelInfo struct {
	Current     int    `json:"current"`
	Max         int    `json:"max"`
	LevelName   string `json:"level_name"`
	LevelTier   int    `json:"level_tier"`
	Description string `json:"description"`
}

// RelationshipInfo 關係信息
type RelationshipInfo struct {
	Status   string `json:"status"`
	Intimacy string `json:"intimacy"`
}

// ProgressInfo 進度信息
type ProgressInfo struct {
	ToNextLevel   int      `json:"to_next_level"`
	PointsNeeded  int      `json:"points_needed"`
	EstimatedDays int      `json:"estimated_days"`
	GrowthTips    []string `json:"growth_tips"`
}

// InteractionSummaryInfo 互動摘要信息
type InteractionSummaryInfo struct {
	TotalInteractions int `json:"total_interactions"`
}

// ChatInfo 聊天信息结构，用于验证chat_id并获取相关信息
type ChatInfo struct {
	ChatID      string `json:"chat_id"`
	UserID      string `json:"user_id"`
	CharacterID string `json:"character_id"`
	Title       string `json:"title"`
}

// validateChatAndGetInfo 验证chat_id并获取相关信息
func validateChatAndGetInfo(c *gin.Context, chatID, userID string) (*ChatInfo, *models.APIError) {
	if chatID == "" {
		return nil, &models.APIError{
			Code:    "MISSING_CHAT_ID",
			Message: "chat_id 参数是必填的",
		}
	}

	// 查询聊天记录
	ctx := context.Background()
	var chat db.ChatDB
	err := database.GetApp().DB().NewSelect().
		Model(&chat).
		Column("id", "user_id", "character_id", "title").
		Where("id = ?", chatID).
		Scan(ctx)

	if err != nil {
		utils.Logger.WithError(err).WithFields(map[string]interface{}{
			"chat_id": chatID,
			"user_id": userID,
		}).Warn("查询聊天记录失败")

		return nil, &models.APIError{
			Code:    "CHAT_NOT_FOUND",
			Message: "指定的聊天记录不存在",
		}
	}

	// 验证用户权限
	if chat.UserID != userID {
		return nil, &models.APIError{
			Code:    "CHAT_ACCESS_DENIED",
			Message: "无权访问该聊天记录",
		}
	}

	return &ChatInfo{
		ChatID:      chat.ID,
		UserID:      chat.UserID,
		CharacterID: chat.CharacterID,
		Title:       chat.Title,
	}, nil
}

// GetEmotionStatus godoc
// @Summary      獲取情感狀態
// @Description  獲取指定聊天中角色的情感狀態，專注於情感和情緒相關信息
// @Tags         Relationships
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "聊天ID"
// @Success      200 {object} models.APIResponse{data=RelationshipStatusResponse} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /relationships/chat/{chat_id}/status [get]
func GetRelationshipStatus(c *gin.Context) {
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

	// 獲取並驗證chat_id參數（從URL路徑獲取）
	chatID := c.Param("chat_id")
	chatInfo, apiError := validateChatAndGetInfo(c, chatID, userIDStr)
	if apiError != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   apiError,
		})
		return
	}

	// 直接查詢 relationships 表獲取對話特定的關係狀態
	// 重要：關係記錄應該在創建聊天會話時自動初始化
	// 如果這裡返回404，檢查 CreateChatSession 是否正確創建了關係記錄
	var relationship db.RelationshipDB
	err := database.GetApp().DB().NewSelect().
		Model(&relationship).
		Where("user_id = ? AND character_id = ? AND chat_id = ?", userIDStr, chatInfo.CharacterID, chatID).
		Scan(c.Request.Context())

	if err != nil {
		utils.Logger.WithError(err).Error("查詢關係狀態失敗")
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "未找到關係狀態",
		})
		return
	}

	// 專注於當前情感狀態的響應 - 移除與affection API重複的資料
	relationshipStatus := gin.H{
		"user_id":      userIDStr,
		"character_id": chatInfo.CharacterID,
		"chat_id":      chatID,
		"current_relationship": gin.H{
			"type":        relationship.Relationship,
			"intensity":   getEmotionIntensity(relationship.Mood),
			"description": getEmotionDescription(relationship.Mood, relationship.Affection),
		},
		"relationship_context": gin.H{
			"stability":        getMoodStability(relationship.Mood),
			"last_interaction": relationship.TotalInteractions,
			"recent_triggers":  []string{},
		},
		"updated_at": utils.GetCurrentTimestampString(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取關係狀態成功",
		Data:    relationshipStatus,
	})
}

// GetAffectionLevel godoc
// @Summary      獲取好感度
// @Description  獲取指定聊天中角色對用戶的好感度數據，專注於關係進展和好感度系統
// @Tags         Relationships
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "聊天ID"
// @Success      200 {object} models.APIResponse{data=AffectionLevelResponse} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /relationships/chat/{chat_id}/affection [get]
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

	userIDStr := userID.(string)

	// 獲取並驗證chat_id參數（從URL路徑獲取）
	chatID := c.Param("chat_id")
	chatInfo, apiError := validateChatAndGetInfo(c, chatID, userIDStr)
	if apiError != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   apiError,
		})
		return
	}

	// 直接查詢 relationships 表獲取對話特定的關係狀態
	var relationship db.RelationshipDB
	err := database.GetApp().DB().NewSelect().
		Model(&relationship).
		Where("user_id = ? AND character_id = ? AND chat_id = ?", userIDStr, chatInfo.CharacterID, chatID).
		Scan(c.Request.Context())

	if err != nil {
		utils.Logger.WithError(err).Error("查詢關係狀態失敗")
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "未找到關係狀態",
		})
		return
	}

	// 計算等級和進度
	levelName, levelTier := getAffectionLevelInfo(relationship.Affection)
	nextLevelThreshold := getNextLevelThreshold(levelTier)
	pointsNeeded := nextLevelThreshold - relationship.Affection

	affectionData := gin.H{
		"user_id":      userIDStr,
		"character_id": chatInfo.CharacterID,
		"chat_id":      chatID,
		"affection_level": gin.H{
			"current":     relationship.Affection,
			"max":         100,
			"level_name":  levelName,
			"level_tier":  levelTier,
			"description": getAffectionDescription(levelTier),
		},
		"relationship": gin.H{
			"status":   relationship.Relationship,
			"intimacy": relationship.IntimacyLevel,
		},
		"progress": gin.H{
			"to_next_level":  nextLevelThreshold,
			"points_needed":  max(0, pointsNeeded),
			"estimated_days": max(1, pointsNeeded/2),
			"growth_tips":    getGrowthTips(levelTier), // 新增成長建議
		},
		"interaction_summary": gin.H{
			"total_interactions": relationship.TotalInteractions,
		},
		"updated_at": utils.GetCurrentTimestampString(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取好感度數據成功",
		Data:    affectionData,
	})
}

// GetAffectionHistory godoc
// @Summary      獲取好感度歷史
// @Description  獲取指定聊天中角色好感度變化歷史記錄
// @Tags         Relationships
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "聊天ID"
// @Param        days query int false "查詢天數" default(30)
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /relationships/chat/{chat_id}/history [get]
func GetRelationshipHistory(c *gin.Context) {
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

	// 獲取並驗證chat_id參數（從URL路徑獲取）
	chatID := c.Param("chat_id")
	chatInfo, apiError := validateChatAndGetInfo(c, chatID, userIDStr)
	if apiError != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   apiError,
		})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	// 查詢角色名稱
	var character db.CharacterDB
	err := database.GetApp().DB().NewSelect().
		Model(&character).
		Column("name").
		Where("id = ?", chatInfo.CharacterID).
		Scan(c.Request.Context(), &character)

	characterName := "未知角色"
	if err == nil && character.Name != "" {
		characterName = character.Name
	} else {
		utils.Logger.WithError(err).WithField("character_id", chatInfo.CharacterID).Warn("查詢角色名稱失敗")
	}

	// 查詢基於chat_id的情感狀態歷史（優先）或基於character_id的歷史（兼容）
	var emotionStates []db.RelationshipDB
	err = database.GetApp().DB().NewSelect().
		Model(&emotionStates).
		Where("user_id = ? AND character_id = ? AND (chat_id = ? OR chat_id IS NULL) AND created_at >= ?",
			userIDStr, chatInfo.CharacterID, chatID, time.Now().AddDate(0, 0, -days)).
		Order("created_at ASC").
		Scan(c)

	if err != nil {
		utils.Logger.WithError(err).Error("查詢情感歷史失敗")
		// 返回空歷史而非錯誤
		emotionStates = []db.RelationshipDB{}
	}

	// 直接查詢當前關係狀態
	var currentRelationship db.RelationshipDB
	err = database.GetApp().DB().NewSelect().
		Model(&currentRelationship).
		Where("user_id = ? AND character_id = ? AND chat_id = ?", userIDStr, chatInfo.CharacterID, chatID).
		Scan(c)

	if err != nil {
		utils.Logger.WithError(err).Error("查詢當前關係狀態失敗")
		// 返回默認值
		currentRelationship.Affection = 50
	}

	// 轉換歷史記錄
	var historyEntries []gin.H
	var positiveChanges, negativeChanges int
	var highestAffection int = currentRelationship.Affection

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

	history := gin.H{
		"user_id":           userIDStr,
		"character_id":      chatInfo.CharacterID,
		"chat_id":           chatID,
		"character_name":    characterName,
		"current_affection": currentRelationship.Affection,
		"history":           historyEntries,
		"statistics": gin.H{
			"total_interactions": len(emotionStates),
			"positive_changes":   positiveChanges,
			"negative_changes":   negativeChanges,
			"highest_affection":  highestAffection,
			"growth_rate":        growthRate,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取好感度歷史成功",
		Data:    history,
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

// getEmotionIntensity 根據情緒類型計算強度值 (1-10)
func getEmotionIntensity(mood string) int {
	switch mood {
	case "happy", "excited", "ecstatic":
		return 8
	case "pleased", "content":
		return 6
	case "neutral", "calm":
		return 5
	case "shy", "nervous":
		return 4
	case "disappointed", "sad":
		return 3
	case "upset", "angry":
		return 7
	case "concerned", "worried":
		return 4
	default:
		return 5 // 默認中性強度
	}
}

// getMoodStability 獲取情緒穩定性描述
func getMoodStability(mood string) string {
	switch mood {
	case "neutral", "calm", "content":
		return "stable"
	case "excited", "ecstatic", "upset", "angry":
		return "volatile"
	case "happy", "pleased":
		return "positive_stable"
	case "sad", "disappointed":
		return "negative_stable"
	default:
		return "moderate"
	}
}

// getGrowthTips 根據好感度等級提供成長建議
func getGrowthTips(levelTier int) []string {
	switch levelTier {
	case 1: // 陌生
		return []string{"多進行日常對話", "表現友好態度", "避免冒犯性言論"}
	case 2: // 友好
		return []string{"分享個人興趣", "表達關心", "增加互動頻率"}
	case 3: // 親密
		return []string{"進行深度交流", "表達真摯情感", "建立共同回憶"}
	case 4: // 戀人
		return []string{"表現專一承諾", "分享內心秘密", "創造浪漫時刻"}
	case 5: // 摯愛
		return []string{"維持深度連結", "經歷重要決定", "證明彼此重要性"}
	default:
		return []string{"保持良好互動"}
	}
}
