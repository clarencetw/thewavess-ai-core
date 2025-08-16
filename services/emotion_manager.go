package services

import (
	"fmt"
	"strings"
	"time"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// EmotionManager 情感管理器
type EmotionManager struct {
	emotionHistory map[string]*EmotionHistory
}

// EmotionHistory 情感歷史記錄
type EmotionHistory struct {
	UserID           string                   `json:"user_id"`
	CharacterID      string                   `json:"character_id"`
	CurrentEmotion   *EmotionState           `json:"current_emotion"`
	EmotionTimeline  []EmotionSnapshot       `json:"emotion_timeline"`
	Milestones       []RelationshipMilestone `json:"milestones"`
	LastInteraction  time.Time               `json:"last_interaction"`
	TotalInteractions int                    `json:"total_interactions"`
}

// EmotionSnapshot 情感快照
type EmotionSnapshot struct {
	Timestamp   time.Time     `json:"timestamp"`
	Emotion     *EmotionState `json:"emotion"`
	Trigger     string        `json:"trigger"`
	Change      int           `json:"affection_change"`
	Context     string        `json:"context"`
}

// RelationshipMilestone 關係里程碑
type RelationshipMilestone struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	AchievedAt  time.Time `json:"achieved_at"`
	Affection   int       `json:"affection_level"`
}

// NewEmotionManager 創建情感管理器
func NewEmotionManager() *EmotionManager {
	return &EmotionManager{
		emotionHistory: make(map[string]*EmotionHistory),
	}
}

// GetEmotionState 獲取當前情感狀態
func (em *EmotionManager) GetEmotionState(userID, characterID string) *EmotionState {
	key := fmt.Sprintf("%s_%s", userID, characterID)
	
	if history, exists := em.emotionHistory[key]; exists {
		return history.CurrentEmotion
	}
	
	// 創建新的情感狀態
	initialEmotion := &EmotionState{
		Affection:     em.getInitialAffection(userID, characterID),
		Mood:          "neutral",
		Relationship:  "stranger",
		IntimacyLevel: "distant",
	}
	
	em.emotionHistory[key] = &EmotionHistory{
		UserID:            userID,
		CharacterID:       characterID,
		CurrentEmotion:    initialEmotion,
		EmotionTimeline:   []EmotionSnapshot{},
		Milestones:        []RelationshipMilestone{},
		LastInteraction:   time.Now(),
		TotalInteractions: 0,
	}
	
	return initialEmotion
}

// UpdateEmotion 更新情感狀態
func (em *EmotionManager) UpdateEmotion(currentEmotion *EmotionState, userMessage string, contentAnalysis *ContentAnalysis) *EmotionState {
	if currentEmotion == nil {
		return &EmotionState{
			Affection:     30,
			Mood:          "neutral",
			Relationship:  "stranger",
			IntimacyLevel: "distant",
		}
	}
	
	// 複製當前狀態
	newEmotion := *currentEmotion
	oldAffection := currentEmotion.Affection
	
	// 計算好感度變化
	affectionChange := em.calculateAffectionChange(userMessage, contentAnalysis)
	newEmotion.Affection += affectionChange
	
	// 限制好感度範圍
	if newEmotion.Affection > 100 {
		newEmotion.Affection = 100
	} else if newEmotion.Affection < 0 {
		newEmotion.Affection = 0
	}
	
	// 更新關係狀態
	newEmotion.Relationship = em.determineRelationship(newEmotion.Affection)
	newEmotion.IntimacyLevel = em.determineIntimacyLevel(newEmotion.Affection)
	
	// 更新心情
	newEmotion.Mood = em.determineMood(userMessage, contentAnalysis, newEmotion.Affection)
	
	utils.Logger.WithFields(logrus.Fields{
		"affection_change": affectionChange,
		"old_affection":    oldAffection,
		"new_affection":    newEmotion.Affection,
		"mood":             newEmotion.Mood,
		"relationship":     newEmotion.Relationship,
		"intimacy":         newEmotion.IntimacyLevel,
	}).Info("情感狀態更新完成")
	
	return &newEmotion
}

// getInitialAffection 獲取初始好感度
func (em *EmotionManager) getInitialAffection(userID, characterID string) int {
	// 基於用戶ID和角色ID生成一致的初始好感度
	hash := 0
	for _, c := range userID + characterID {
		hash += int(c)
	}
	return 25 + (hash % 25) // 25-50之間的值
}

// calculateAffectionChange 計算好感度變化
func (em *EmotionManager) calculateAffectionChange(userMessage string, analysis *ContentAnalysis) int {
	change := 1 // 基礎增長（每次互動）
	
	// 正面詞彙影響
	positiveWords := []string{
		"喜歡", "愛", "謝謝", "感謝", "開心", "高興", "快樂", "想念", "關心", "在意",
		"美好", "溫暖", "舒服", "安心", "放心", "信任", "依賴", "需要", "重要",
		"love", "like", "thank", "happy", "miss", "care", "warm", "trust", "need",
	}
	
	for _, word := range positiveWords {
		if strings.Contains(strings.ToLower(userMessage), word) {
			change += 2
			break
		}
	}
	
	// NSFW內容的信任加成
	if analysis.IsNSFW {
		switch analysis.Intensity {
		case 2, 3: // 浪漫和親密內容
			change += 1
		case 4, 5: // 明確成人內容
			change += 2 // 表示高度信任
		}
	}
	
	// 負面詞彙影響
	negativeWords := []string{
		"討厭", "煩", "不喜歡", "離開", "再見", "結束", "分手", "不要", "停止",
		"無聊", "失望", "生氣", "憤怒", "傷心", "難過", "痛苦", "後悔",
		"hate", "annoying", "dislike", "leave", "bye", "stop", "boring", "angry",
	}
	
	for _, word := range negativeWords {
		if strings.Contains(strings.ToLower(userMessage), word) {
			change -= 3
			break
		}
	}
	
	// 長度獎勵（表示投入度）
	if len(userMessage) > 50 {
		change += 1
	}
	
	return change
}

// determineRelationship 確定關係狀態
func (em *EmotionManager) determineRelationship(affection int) string {
	switch {
	case affection >= 90:
		return "deep_love"
	case affection >= 80:
		return "lover"
	case affection >= 70:
		return "romantic"
	case affection >= 60:
		return "close_friend"
	case affection >= 40:
		return "friend"
	case affection >= 20:
		return "acquaintance"
	default:
		return "stranger"
	}
}

// determineIntimacyLevel 確定親密程度
func (em *EmotionManager) determineIntimacyLevel(affection int) string {
	switch {
	case affection >= 85:
		return "deeply_intimate"
	case affection >= 70:
		return "intimate"
	case affection >= 55:
		return "close"
	case affection >= 35:
		return "friendly"
	case affection >= 15:
		return "polite"
	default:
		return "distant"
	}
}

// determineMood 確定心情
func (em *EmotionManager) determineMood(userMessage string, analysis *ContentAnalysis, affection int) string {
	message := strings.ToLower(userMessage)
	
	// 基於關鍵詞的心情判斷
	if strings.Contains(message, "開心") || strings.Contains(message, "高興") || strings.Contains(message, "快樂") {
		return "happy"
	}
	
	if strings.Contains(message, "難過") || strings.Contains(message, "傷心") || strings.Contains(message, "痛苦") {
		return "concerned"
	}
	
	if strings.Contains(message, "生氣") || strings.Contains(message, "憤怒") || strings.Contains(message, "煩") {
		return "annoyed"
	}
	
	if strings.Contains(message, "害羞") || strings.Contains(message, "臉紅") || strings.Contains(message, "不好意思") {
		return "shy"
	}
	
	if strings.Contains(message, "興奮") || strings.Contains(message, "激動") || strings.Contains(message, "期待") {
		return "excited"
	}
	
	// 基於NSFW內容的心情
	if analysis.IsNSFW {
		switch analysis.Intensity {
		case 2, 3:
			return "romantic"
		case 4, 5:
			return "passionate"
		}
	}
	
	// 基於好感度的默認心情
	switch {
	case affection >= 80:
		return "loving"
	case affection >= 60:
		return "pleased"
	case affection >= 40:
		return "friendly"
	case affection >= 20:
		return "polite"
	default:
		return "neutral"
	}
}

// CheckMilestones 檢查是否達到新的里程碑
func (em *EmotionManager) CheckMilestones(userID, characterID string, oldEmotion, newEmotion *EmotionState) []RelationshipMilestone {
	var newMilestones []RelationshipMilestone
	
	// 好感度里程碑
	milestonePoints := []int{20, 40, 60, 80, 100}
	for _, point := range milestonePoints {
		if oldEmotion.Affection < point && newEmotion.Affection >= point {
			milestone := RelationshipMilestone{
				ID:          fmt.Sprintf("affection_%d", point),
				Type:        "affection_milestone",
				Description: em.getAffectionMilestoneDescription(point),
				AchievedAt:  time.Now(),
				Affection:   point,
			}
			newMilestones = append(newMilestones, milestone)
		}
	}
	
	// 關係狀態里程碑
	if oldEmotion.Relationship != newEmotion.Relationship {
		milestone := RelationshipMilestone{
			ID:          fmt.Sprintf("relationship_%s", newEmotion.Relationship),
			Type:        "relationship_change",
			Description: fmt.Sprintf("關係發展到：%s", em.getRelationshipDisplayName(newEmotion.Relationship)),
			AchievedAt:  time.Now(),
			Affection:   newEmotion.Affection,
		}
		newMilestones = append(newMilestones, milestone)
	}
	
	return newMilestones
}

// getAffectionMilestoneDescription 獲取好感度里程碑描述
func (em *EmotionManager) getAffectionMilestoneDescription(point int) string {
	descriptions := map[int]string{
		20:  "初步認識，開始有些好感",
		40:  "成為朋友，彼此信任",
		60:  "關係親密，特別在意",
		80:  "深深愛戀，無法分離",
		100: "完美結合，靈魂伴侶",
	}
	
	if desc, exists := descriptions[point]; exists {
		return desc
	}
	return fmt.Sprintf("好感度達到%d", point)
}

// getRelationshipDisplayName 獲取關係顯示名稱
func (em *EmotionManager) getRelationshipDisplayName(relationship string) string {
	names := map[string]string{
		"stranger":     "陌生人",
		"acquaintance": "認識",
		"friend":       "朋友",
		"close_friend": "好友",
		"romantic":     "戀人",
		"lover":        "愛人",
		"deep_love":    "摯愛",
	}
	
	if name, exists := names[relationship]; exists {
		return name
	}
	return relationship
}

// GetEmotionAnalytics 獲取情感分析統計
func (em *EmotionManager) GetEmotionAnalytics(userID, characterID string) map[string]interface{} {
	key := fmt.Sprintf("%s_%s", userID, characterID)
	
	if history, exists := em.emotionHistory[key]; exists {
		return map[string]interface{}{
			"current_affection":    history.CurrentEmotion.Affection,
			"current_relationship": history.CurrentEmotion.Relationship,
			"total_interactions":   history.TotalInteractions,
			"milestones_count":     len(history.Milestones),
			"timeline_length":      len(history.EmotionTimeline),
			"last_interaction":     history.LastInteraction,
			"days_since_first":     time.Since(history.LastInteraction).Hours() / 24,
		}
	}
	
	return map[string]interface{}{
		"status": "no_history_found",
	}
}

// SaveEmotionSnapshot 保存情感快照
func (em *EmotionManager) SaveEmotionSnapshot(userID, characterID, trigger, context string, oldEmotion, newEmotion *EmotionState) {
	key := fmt.Sprintf("%s_%s", userID, characterID)
	
	if history, exists := em.emotionHistory[key]; exists {
		snapshot := EmotionSnapshot{
			Timestamp: time.Now(),
			Emotion:   newEmotion,
			Trigger:   trigger,
			Change:    newEmotion.Affection - oldEmotion.Affection,
			Context:   context,
		}
		
		history.EmotionTimeline = append(history.EmotionTimeline, snapshot)
		history.CurrentEmotion = newEmotion
		history.TotalInteractions++
		history.LastInteraction = time.Now()
		
		// 限制歷史記錄長度
		if len(history.EmotionTimeline) > 100 {
			history.EmotionTimeline = history.EmotionTimeline[len(history.EmotionTimeline)-100:]
		}
	}
}