package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
)

// RelationshipStatusResponse 提供目前對話的關係快照
type RelationshipStatusResponse struct {
	UserID            string `json:"user_id"`
	CharacterID       string `json:"character_id"`
	ChatID            string `json:"chat_id"`
	Affection         int    `json:"affection"`
	Mood              string `json:"mood"`
	MoodIntensity     int    `json:"mood_intensity"`
	MoodDescription   string `json:"mood_description"`
	Relationship      string `json:"relationship"`
	IntimacyLevel     string `json:"intimacy_level"`
	TotalInteractions int    `json:"total_interactions"`
	LastInteractionAt string `json:"last_interaction_at"`
	UpdatedAt         string `json:"updated_at"`
}

// AffectionLevelResponse 提供好感度細節
type AffectionLevelResponse struct {
	UserID             string `json:"user_id"`
	CharacterID        string `json:"character_id"`
	ChatID             string `json:"chat_id"`
	Current            int    `json:"current"`
	LevelName          string `json:"level_name"`
	LevelTier          int    `json:"level_tier"`
	Description        string `json:"description"`
	NextLevelThreshold int    `json:"next_level_threshold"`
	PointsToNext       int    `json:"points_to_next"`
	UpdatedAt          string `json:"updated_at"`
}

// RelationshipHistoryResponse 回傳情感歷史事件（若有）
type RelationshipHistoryResponse struct {
	UserID            string                  `json:"user_id"`
	CharacterID       string                  `json:"character_id"`
	ChatID            string                  `json:"chat_id"`
	CharacterName     string                  `json:"character_name,omitempty"`
	CurrentAffection  int                     `json:"current_affection"`
	TotalInteractions int                     `json:"total_interactions"`
	History           []AffectionHistoryEntry `json:"history"`
	UpdatedAt         string                  `json:"updated_at"`
}

// AffectionHistoryEntry 描述一次情感變化事件
type AffectionHistoryEntry struct {
	Timestamp       string `json:"timestamp"`
	TriggerType     string `json:"trigger_type"`
	TriggerContent  string `json:"trigger_content"`
	AffectionBefore int    `json:"affection_before"`
	AffectionAfter  int    `json:"affection_after"`
	AffectionChange int    `json:"affection_change"`
	MoodBefore      string `json:"mood_before"`
	MoodAfter       string `json:"mood_after"`
}

// ChatInfo 聊天信息结构，用于验证chat_id并获取相关信息
type ChatInfo struct {
	ChatID      string `json:"chat_id"`
	UserID      string `json:"user_id"`
	CharacterID string `json:"character_id"`
	Title       string `json:"title"`
}

// validateChatAndGetInfo 驗證 chat_id 並取得會話資料
func validateChatAndGetInfo(c *gin.Context, chatID, userID string) (*ChatInfo, *models.APIError) {
	if chatID == "" {
		return nil, &models.APIError{
			Code:    "MISSING_CHAT_ID",
			Message: "chat_id 參數是必填的",
		}
	}

	ctx := c.Request.Context()
	if ctx == nil {
		ctx = context.Background()
	}

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
		}).Warn("查詢聊天記錄失敗")

		return nil, &models.APIError{
			Code:    "CHAT_NOT_FOUND",
			Message: "指定的聊天記錄不存在",
		}
	}

	if chat.UserID != userID {
		return nil, &models.APIError{
			Code:    "CHAT_ACCESS_DENIED",
			Message: "無權訪問該聊天記錄",
		}
	}

	return &ChatInfo{
		ChatID:      chat.ID,
		UserID:      chat.UserID,
		CharacterID: chat.CharacterID,
		Title:       chat.Title,
	}, nil
}

// GetRelationshipStatus godoc
// @Summary      獲取情感狀態
// @Description  獲取指定聊天中角色的情感狀態
// @Tags         Relationships
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "聊天ID"
// @Success      200 {object} models.APIResponse{data=RelationshipStatusResponse}
// @Failure      400 {object} models.APIResponse{error=models.APIError}
// @Failure      401 {object} models.APIResponse{error=models.APIError}
// @Router       /relationships/chat/{chat_id}/status [get]
func GetRelationshipStatus(c *gin.Context) {
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
	chatID := c.Param("chat_id")

	chatInfo, apiError := validateChatAndGetInfo(c, chatID, userIDStr)
	if apiError != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   apiError,
		})
		return
	}

	relationship, err := getChatRelationship(c.Request.Context(), userIDStr, chatInfo.CharacterID, chatID)
	if err != nil {
		utils.Logger.WithError(err).Error("查詢關係狀態失敗")
		status := http.StatusInternalServerError
		apiErr := &models.APIError{
			Code:    "RELATIONSHIP_NOT_FOUND",
			Message: "未找到關係狀態",
		}

		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}

		c.JSON(status, models.APIResponse{
			Success: false,
			Message: "未找到關係狀態",
			Error:   apiErr,
		})
		return
	}

	response := buildRelationshipStatusResponse(relationship, chatInfo)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取關係狀態成功",
		Data:    response,
	})
}

// GetAffectionLevel godoc
// @Summary      獲取好感度
// @Description  獲取指定聊天中角色對用戶的好感度數據
// @Tags         Relationships
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "聊天ID"
// @Success      200 {object} models.APIResponse{data=AffectionLevelResponse}
// @Failure      400 {object} models.APIResponse{error=models.APIError}
// @Failure      401 {object} models.APIResponse{error=models.APIError}
// @Router       /relationships/chat/{chat_id}/affection [get]
func GetAffectionLevel(c *gin.Context) {
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
	chatID := c.Param("chat_id")

	chatInfo, apiError := validateChatAndGetInfo(c, chatID, userIDStr)
	if apiError != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   apiError,
		})
		return
	}

	relationship, err := getChatRelationship(c.Request.Context(), userIDStr, chatInfo.CharacterID, chatID)
	if err != nil {
		utils.Logger.WithError(err).Error("查詢關係狀態失敗")
		status := http.StatusInternalServerError
		apiErr := &models.APIError{
			Code:    "RELATIONSHIP_NOT_FOUND",
			Message: "未找到關係狀態",
		}

		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}

		c.JSON(status, models.APIResponse{
			Success: false,
			Message: "未找到關係狀態",
			Error:   apiErr,
		})
		return
	}

	response := buildAffectionLevelResponse(relationship, chatInfo)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取好感度數據成功",
		Data:    response,
	})
}

// GetRelationshipHistory godoc
// @Summary      獲取好感度歷史
// @Description  獲取指定聊天中角色好感度變化歷史記錄
// @Tags         Relationships
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id path string true "聊天ID"
// @Success      200 {object} models.APIResponse{data=RelationshipHistoryResponse}
// @Failure      400 {object} models.APIResponse{error=models.APIError}
// @Failure      401 {object} models.APIResponse{error=models.APIError}
// @Router       /relationships/chat/{chat_id}/history [get]
func GetRelationshipHistory(c *gin.Context) {
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
	chatID := c.Param("chat_id")

	chatInfo, apiError := validateChatAndGetInfo(c, chatID, userIDStr)
	if apiError != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   apiError,
		})
		return
	}

	relationship, err := getChatRelationship(c.Request.Context(), userIDStr, chatInfo.CharacterID, chatID)
	if err != nil {
		utils.Logger.WithError(err).Error("查詢關係狀態失敗")
		status := http.StatusInternalServerError
		apiErr := &models.APIError{
			Code:    "RELATIONSHIP_NOT_FOUND",
			Message: "未找到關係狀態",
		}

		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}

		c.JSON(status, models.APIResponse{
			Success: false,
			Message: "未找到關係狀態",
			Error:   apiErr,
		})
		return
	}

	characterName := ""
	ctx := c.Request.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	var character db.CharacterDB
	err = database.GetApp().DB().NewSelect().
		Model(&character).
		Column("name").
		Where("id = ?", chatInfo.CharacterID).
		Scan(ctx)
	if err == nil && character.Name != "" {
		characterName = character.Name
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.Logger.WithError(err).WithField("character_id", chatInfo.CharacterID).Warn("查詢角色名稱失敗")
	}

	response := buildRelationshipHistoryResponse(relationship, chatInfo, characterName)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取好感度歷史成功",
		Data:    response,
	})
}

func getChatRelationship(ctx context.Context, userID, characterID, chatID string) (*db.RelationshipDB, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var relationship db.RelationshipDB
	err := database.GetApp().DB().NewSelect().
		Model(&relationship).
		Where("user_id = ? AND character_id = ? AND chat_id = ?", userID, characterID, chatID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &relationship, nil
}

func buildRelationshipStatusResponse(relationship *db.RelationshipDB, chatInfo *ChatInfo) *RelationshipStatusResponse {
	chatID := chatInfo.ChatID
	if relationship.ChatID != nil && *relationship.ChatID != "" {
		chatID = *relationship.ChatID
	}

	return &RelationshipStatusResponse{
		UserID:            relationship.UserID,
		CharacterID:       relationship.CharacterID,
		ChatID:            chatID,
		Affection:         relationship.Affection,
		Mood:              relationship.Mood,
		MoodIntensity:     getEmotionIntensity(relationship.Mood),
		MoodDescription:   getEmotionDescription(relationship.Mood, relationship.Affection),
		Relationship:      relationship.Relationship,
		IntimacyLevel:     relationship.IntimacyLevel,
		TotalInteractions: relationship.TotalInteractions,
		LastInteractionAt: formatTime(relationship.LastInteraction),
		UpdatedAt:         formatTime(relationship.UpdatedAt),
	}
}

func buildAffectionLevelResponse(relationship *db.RelationshipDB, chatInfo *ChatInfo) *AffectionLevelResponse {
	chatID := chatInfo.ChatID
	if relationship.ChatID != nil && *relationship.ChatID != "" {
		chatID = *relationship.ChatID
	}

	levelName, levelTier := getAffectionLevelInfo(relationship.Affection)
	nextThreshold := getNextLevelThreshold(levelTier)
	pointsToNext := max(0, nextThreshold-relationship.Affection)

	return &AffectionLevelResponse{
		UserID:             relationship.UserID,
		CharacterID:        relationship.CharacterID,
		ChatID:             chatID,
		Current:            relationship.Affection,
		LevelName:          levelName,
		LevelTier:          levelTier,
		Description:        getAffectionDescription(levelTier),
		NextLevelThreshold: nextThreshold,
		PointsToNext:       pointsToNext,
		UpdatedAt:          formatTime(relationship.UpdatedAt),
	}
}

func buildRelationshipHistoryResponse(relationship *db.RelationshipDB, chatInfo *ChatInfo, characterName string) *RelationshipHistoryResponse {
	chatID := chatInfo.ChatID
	if relationship.ChatID != nil && *relationship.ChatID != "" {
		chatID = *relationship.ChatID
	}

	historyEntries := extractHistoryEntries(relationship.EmotionData)

	return &RelationshipHistoryResponse{
		UserID:            relationship.UserID,
		CharacterID:       relationship.CharacterID,
		ChatID:            chatID,
		CharacterName:     characterName,
		CurrentAffection:  relationship.Affection,
		TotalInteractions: relationship.TotalInteractions,
		History:           historyEntries,
		UpdatedAt:         formatTime(relationship.UpdatedAt),
	}
}

func extractHistoryEntries(emotionData map[string]interface{}) []AffectionHistoryEntry {
	if len(emotionData) == 0 {
		return []AffectionHistoryEntry{}
	}

	rawHistory, ok := emotionData["history"].([]interface{})
	if !ok || len(rawHistory) == 0 {
		return []AffectionHistoryEntry{}
	}

	entries := make([]AffectionHistoryEntry, 0, len(rawHistory))
	for _, item := range rawHistory {
		historyMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		entries = append(entries, AffectionHistoryEntry{
			Timestamp:       parseTimestamp(historyMap["timestamp"]),
			TriggerType:     toString(historyMap["trigger_type"]),
			TriggerContent:  toString(historyMap["trigger_content"]),
			AffectionBefore: toInt(historyMap["old_affection"]),
			AffectionAfter:  toInt(historyMap["new_affection"]),
			AffectionChange: toInt(historyMap["affection_change"]),
			MoodBefore:      toString(historyMap["old_mood"]),
			MoodAfter:       toString(historyMap["new_mood"]),
		})
	}

	return entries
}

func parseTimestamp(value interface{}) string {
	switch v := value.(type) {
	case time.Time:
		return v.UTC().Format(time.RFC3339)
	case string:
		if v == "" {
			return ""
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
		return v
	default:
		return ""
	}
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func toInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
		if f, err := v.Float64(); err == nil {
			return int(f)
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

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
		"calm":       "角色保持平靜，專注傾聽",
		"content":    "角色很滿足當前的對話",
		"worried":    "角色擔心你的狀態",
		"angry":      "角色感到憤怒，需要冷靜",
		"sad":        "角色有點難過，需要關心",
	}

	baseDesc := descriptions[mood]
	if baseDesc == "" {
		baseDesc = "角色心情平靜"
	}

	switch {
	case affection >= 80:
		return baseDesc + "，對你的愛意溢於言表"
	case affection >= 60:
		return baseDesc + "，對你有很深的感情"
	case affection >= 40:
		return baseDesc + "，對你很有好感"
	case affection >= 20:
		return baseDesc + "，開始對你產生興趣"
	default:
		return baseDesc
	}
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
		1: 25,
		2: 50,
		3: 70,
		4: 90,
		5: 100,
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
		return 5
	}
}
