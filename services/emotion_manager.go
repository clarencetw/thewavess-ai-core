package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// 全局情感管理器實例
var (
	globalEmotionManager *EmotionManager
	emotionManagerOnce   sync.Once
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

// GetEmotionState 獲取當前情感狀態（支援資料庫持久化）
func (em *EmotionManager) GetEmotionState(userID, characterID string) *EmotionState {
	key := fmt.Sprintf("%s_%s", userID, characterID)
	
	// 先從內存檢查
	if history, exists := em.emotionHistory[key]; exists {
		return history.CurrentEmotion
	}
	
	// 從資料庫載入情感狀態
	ctx := context.Background()
	var dbEmotion models.EmotionState
	err := database.DB.NewSelect().
		Model(&dbEmotion).
		Where("user_id = ? AND character_id = ?", userID, characterID).
		Scan(ctx)
	
	var currentEmotion *EmotionState
	
	if err == nil {
		// 從資料庫載入成功，轉換為內存結構
		currentEmotion = &EmotionState{
			Affection:     dbEmotion.Affection,
			Mood:          dbEmotion.Mood,
			Relationship:  dbEmotion.Relationship,
			IntimacyLevel: dbEmotion.IntimacyLevel,
		}
		
		utils.Logger.WithFields(logrus.Fields{
			"user_id":     userID,
			"character_id": characterID,
			"affection":   currentEmotion.Affection,
		}).Info("從資料庫載入情感狀態")
	} else {
		// 資料庫中沒有記錄，創建新的情感狀態
		currentEmotion = &EmotionState{
			Affection:     em.getInitialAffection(userID, characterID),
			Mood:          "neutral",
			Relationship:  "stranger",
			IntimacyLevel: "distant",
		}
		
		// 保存到資料庫
		newDbEmotion := &models.EmotionState{
			ID:                utils.GenerateID(16),
			UserID:           userID,
			CharacterID:      characterID,
			Affection:        currentEmotion.Affection,
			Mood:             currentEmotion.Mood,
			Relationship:     currentEmotion.Relationship,
			IntimacyLevel:    currentEmotion.IntimacyLevel,
			TotalInteractions: 0,
			LastInteraction:  time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		
		_, err = database.DB.NewInsert().Model(newDbEmotion).Exec(ctx)
		if err != nil {
			utils.Logger.WithError(err).Error("保存新情感狀態到資料庫失敗")
		} else {
			utils.Logger.WithFields(logrus.Fields{
				"user_id":     userID,
				"character_id": characterID,
				"affection":   currentEmotion.Affection,
			}).Info("創建新情感狀態並保存到資料庫")
		}
	}
	
	// 更新內存快取
	em.emotionHistory[key] = &EmotionHistory{
		UserID:            userID,
		CharacterID:       characterID,
		CurrentEmotion:    currentEmotion,
		EmotionTimeline:   []EmotionSnapshot{},
		Milestones:        []RelationshipMilestone{},
		LastInteraction:   time.Now(),
		TotalInteractions: 0,
	}
	
	return currentEmotion
}

// UpdateEmotion 更新情感狀態
func (em *EmotionManager) UpdateEmotion(currentEmotion *EmotionState, userMessage string, contentAnalysis *ContentAnalysis) *EmotionState {
    // NOTE(擴充說明): 這裡是所有「情感狀態更新規則」的匯集點。
    // 如需新增/調整功能，建議從以下方向擴充：
    // 1) 權重/詞庫配置化：將關鍵字、加減分權重抽到可配置來源（DB/檔案/環境變數）。
    // 2) 動量/冷卻時間：連續正面/負面對話加乘，或在一次大波動後進入冷卻期，避免分數快速抖動。
    // 3) 時間衰減：長時間未互動時，對好感度做緩慢衰減（如每日-1，設下限）。
    // 4) 人設加權：依角色(character)或使用者偏好給不同詞類不同加權（例如醫生角色對「健康關懷」詞更敏感）。
    // 5) 事件觸發：將某些行為封裝為事件（如完成任務、達成節日），在此加分。
    // 6) 模型情緒輸出融合：可加入情緒分析模型(或LLM判斷)的情感分數，再與規則權重融合。
    // 7) 上限/下限規範：依當前關係階段限制單回合最大變動，防止跳級過快。
    // 8) 反饋解釋：記錄本次命中的規則與加減分明細，用於API回傳debug。

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
    // TODO(可擴充): 將 calculateAffectionChange 改為回傳 (delta, reasons []string)，
    // 以便保存到情感歷史中，對前端/數據分析展示「為什麼會加/減分」。
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
    change := 2 // 基礎增長（每次互動）

    // 正面詞彙影響
    // TODO(關鍵字擴充):
    // - 若要加入更多同義詞或詞根，可改為 map[string]int 設定權重，或採用詞典+模糊匹配。
    // - 建議改為從可配置來源載入（DB/JSON/ENV），允許運營無需上線即可調整。
    // - 可依角色/使用者偏好拆分不同詞庫（ex: per-character/per-user overrides）。
    positiveWords := []string{
        "喜歡", "愛", "謝謝", "感謝", "開心", "高興", "快樂", "想念", "關心", "在意",
        "美好", "溫暖", "舒服", "安心", "放心", "信任", "依賴", "需要", "重要",
        "love", "like", "thank", "happy", "miss", "care", "warm", "trust", "need",
    }
	
	for _, word := range positiveWords {
		if strings.Contains(strings.ToLower(userMessage), word) {
			change += 3
			break
		}
	}
	
    // NSFW內容的信任加成
    // TODO(規則調整): 可依關係階段或使用者設定調整權重；
    // 例如在低關係階段降低 NSFW 加成，在高關係階段提高加成；
    // 或針對 intensity 不同區間使用非線性曲線（如Sigmoid）計算加分。
    if analysis.IsNSFW {
        switch analysis.Intensity {
        case 2, 3: // 浪漫和親密內容
            change += 1
        case 4, 5: // 明確成人內容
            change += 2 // 表示高度信任
        }
    }

    // 負面詞彙影響
    // TODO(關鍵字擴充): 同正向詞彙，建議可配置；另可加入「強負面詞」給更大權重（如-5）。
    negativeWords := []string{
        "討厭", "煩", "不喜歡", "離開", "再見", "結束", "分手", "不要", "停止",
        "無聊", "失望", "生氣", "憤怒", "傷心", "難過", "痛苦", "後悔",
        "hate", "annoying", "dislike", "leave", "bye", "stop", "boring", "angry",
    }
	
	for _, word := range negativeWords {
		if strings.Contains(strings.ToLower(userMessage), word) {
			change -= 2
			break
		}
	}
	
    // 長度獎勵（表示投入度）
    // TODO(行為獎懲): 可加上「連續對話回合數」或「回覆速度」等行為指標加權，
    // 例如連續N回合正面互動 +1，若中斷超過M小時則減少動量或觸發微弱衰減。
    if len(userMessage) > 50 {
        change += 1
    }

    return change
}

// determineRelationship 確定關係狀態
func (em *EmotionManager) determineRelationship(affection int) string {
    // TODO(擴充):
    // - 可將各階段的閾值改為可配置表；
    // - 可加入「降級保護/緩衝區」避免臨界值來回抖動（hysteresis）。
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
    // TODO(擴充): 親密度可依場景或角色特性另行偏移，例如醫生角色在專業場景下上限為 close。
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
    // TODO(關鍵字擴充): 建議整理成 map["mood"][]string 並可配置；
    // 也可接入情緒分析模型，將模型輸出映射到內建 mood 類別。
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
    // TODO(規則調整): 可依當前關係階段為 romantic/passionate 設定不同門檻；
    // 亦可在 intensity 與 affection 之間使用對映表或曲線，而非if/else。
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
    // TODO(擴充): 可將里程碑點位、名稱與可解鎖內容(對話/場景/稱呼)配置化；
    // 亦可針對特定角色定義專屬里程碑。
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


// TODO: 實現簡單的情感統計功能
// 需要實現以下統計數據：
// 1. 基本統計：當前好感度、關係狀態、總互動次數、最後互動時間
// 2. 好感度趨勢：最高/最低好感度、好感度變化總數
// 3. 里程碑統計：已達成里程碑數量、最近里程碑
// 4. 互動分析：平均每日互動次數、連續互動天數
// 5. 心情分析：最常見心情、心情變化次數

// GetSimpleEmotionStats 獲取簡單情感統計（暫時實現）
func (em *EmotionManager) GetSimpleEmotionStats(userID, characterID string) map[string]interface{} {
	key := fmt.Sprintf("%s_%s", userID, characterID)
	
	history, exists := em.emotionHistory[key]
	if !exists {
		return map[string]interface{}{
			"error": "no_emotion_data",
		}
	}
	
	// 基本統計
	stats := map[string]interface{}{
		"current_affection":    history.CurrentEmotion.Affection,
		"current_relationship": history.CurrentEmotion.Relationship,
		"current_mood":         history.CurrentEmotion.Mood,
		"total_interactions":   history.TotalInteractions,
		"total_milestones":     len(history.Milestones),
		"timeline_length":      len(history.EmotionTimeline),
		"last_interaction":     history.LastInteraction,
	}
	
	// 好感度趨勢分析
	if len(history.EmotionTimeline) > 0 {
		maxAffection := history.CurrentEmotion.Affection
		minAffection := history.CurrentEmotion.Affection
		positiveChanges := 0
		negativeChanges := 0
		
		for _, snapshot := range history.EmotionTimeline {
			affection := snapshot.Emotion.Affection
			if affection > maxAffection {
				maxAffection = affection
			}
			if affection < minAffection {
				minAffection = affection
			}
			
			if snapshot.Change > 0 {
				positiveChanges++
			} else if snapshot.Change < 0 {
				negativeChanges++
			}
		}
		
		stats["max_affection"] = maxAffection
		stats["min_affection"] = minAffection
		stats["positive_changes"] = positiveChanges
		stats["negative_changes"] = negativeChanges
		stats["first_interaction"] = history.EmotionTimeline[0].Timestamp
		
		// 活躍天數計算
		daysSinceFirst := time.Since(history.EmotionTimeline[0].Timestamp).Hours() / 24
		stats["days_since_first"] = int(daysSinceFirst)
	}
	
	// 最近里程碑
	if len(history.Milestones) > 0 {
		recentMilestones := history.Milestones
		if len(recentMilestones) > 3 {
			recentMilestones = recentMilestones[len(recentMilestones)-3:]
		}
		stats["recent_milestones"] = recentMilestones
	}
	
	return stats
}

// SaveEmotionSnapshot 保存情感快照（支援資料庫持久化）
func (em *EmotionManager) SaveEmotionSnapshot(userID, characterID, trigger, contextContent string, oldEmotion, newEmotion *EmotionState) {
    key := fmt.Sprintf("%s_%s", userID, characterID)
    ctx := context.Background()

    // 1. 更新情感狀態到資料庫
    // TODO(數據擴充): 可另建一張 explanation 表或在 emotion_history.context 中寫入「規則命中明細」，
    // 以便於前端或BI系統完整還原每次變動的原因（關鍵字、NSFW 等級、長度獎勵、動量/冷卻等）。
    _, err := database.DB.NewUpdate().
        Model((*models.EmotionState)(nil)).
        Set("affection = ?", newEmotion.Affection).
		Set("mood = ?", newEmotion.Mood).
		Set("relationship = ?", newEmotion.Relationship).
		Set("intimacy_level = ?", newEmotion.IntimacyLevel).
		Set("total_interactions = total_interactions + 1").
		Set("last_interaction = ?", time.Now()).
		Set("updated_at = ?", time.Now()).
		Where("user_id = ? AND character_id = ?", userID, characterID).
		Exec(ctx)
	
	if err != nil {
		utils.Logger.WithError(err).Error("更新情感狀態到資料庫失敗")
	}
	
	// 2. 保存情感歷史記錄到資料庫
	emotionHistory := &models.EmotionHistory{
		ID:              utils.GenerateID(16),
		UserID:          userID,
		CharacterID:     characterID,
		OldAffection:    oldEmotion.Affection,
		NewAffection:    newEmotion.Affection,
		AffectionChange: newEmotion.Affection - oldEmotion.Affection,
		OldMood:         oldEmotion.Mood,
		NewMood:         newEmotion.Mood,
		TriggerType:     trigger,
		TriggerContent:  contextContent,
		Context:         map[string]interface{}{
			"old_relationship":  oldEmotion.Relationship,
			"new_relationship":  newEmotion.Relationship,
			"old_intimacy":      oldEmotion.IntimacyLevel,
			"new_intimacy":      newEmotion.IntimacyLevel,
		},
		CreatedAt:       time.Now(),
	}
	
	_, err = database.DB.NewInsert().Model(emotionHistory).Exec(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("保存情感歷史記錄到資料庫失敗")
	}
	
	// 3. 檢查並保存新的里程碑到資料庫
	newMilestones := em.CheckMilestones(userID, characterID, oldEmotion, newEmotion)
	for _, milestone := range newMilestones {
		dbMilestone := &models.EmotionMilestone{
			ID:             utils.GenerateID(16),
			UserID:         userID,
			CharacterID:    characterID,
			MilestoneType:  milestone.Type,
			Description:    milestone.Description,
			AffectionLevel: milestone.Affection,
			AchievedAt:     milestone.AchievedAt,
		}
		
		_, err = database.DB.NewInsert().Model(dbMilestone).Exec(ctx)
		if err != nil {
			utils.Logger.WithError(err).Error("保存情感里程碑到資料庫失敗")
		} else {
			utils.Logger.WithFields(logrus.Fields{
				"user_id":     userID,
				"character_id": characterID,
				"milestone":   milestone.Type,
				"affection":   milestone.Affection,
			}).Info("情感里程碑已保存到資料庫")
		}
	}
	
	// 4. 更新內存快取
	if history, exists := em.emotionHistory[key]; exists {
		snapshot := EmotionSnapshot{
			Timestamp: time.Now(),
			Emotion:   newEmotion,
			Trigger:   trigger,
			Change:    newEmotion.Affection - oldEmotion.Affection,
			Context:   contextContent,
		}
		
		history.EmotionTimeline = append(history.EmotionTimeline, snapshot)
		history.CurrentEmotion = newEmotion
		history.TotalInteractions++
		history.LastInteraction = time.Now()
		history.Milestones = append(history.Milestones, newMilestones...)
		
		// 限制內存中的歷史記錄長度
		if len(history.EmotionTimeline) > 100 {
			history.EmotionTimeline = history.EmotionTimeline[len(history.EmotionTimeline)-100:]
		}
		
		// 限制內存中的里程碑記錄長度
		if len(history.Milestones) > 50 {
			history.Milestones = history.Milestones[len(history.Milestones)-50:]
		}
	}
	
	utils.Logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"character_id":    characterID,
		"affection_change": newEmotion.Affection - oldEmotion.Affection,
		"new_affection":   newEmotion.Affection,
		"trigger":         trigger,
	}).Info("情感快照已保存完成")
}

// TODO: 實現以下簡單的情感趨勢和報告功能
// GetEmotionTrend - 獲取最近N天的情感變化趨勢
// GetEmotionReport - 生成簡單的情感報告
// GetMoodAnalysis - 分析心情變化模式
// GetInteractionPattern - 分析互動模式

// GetEmotionTrend 獲取情感變化趨勢（最近N天）
func (em *EmotionManager) GetEmotionTrend(userID, characterID string, days int) []EmotionSnapshot {
	key := fmt.Sprintf("%s_%s", userID, characterID)
	
	history, exists := em.emotionHistory[key]
	if !exists {
		return []EmotionSnapshot{}
	}
	
	cutoff := time.Now().AddDate(0, 0, -days)
	var trend []EmotionSnapshot
	
	for _, snapshot := range history.EmotionTimeline {
		if snapshot.Timestamp.After(cutoff) {
			trend = append(trend, snapshot)
		}
	}
	
	return trend
}

// ClearEmotionHistory 清空情感歷史（用於測試或重置）
func (em *EmotionManager) ClearEmotionHistory(userID, characterID string) {
	key := fmt.Sprintf("%s_%s", userID, characterID)
	delete(em.emotionHistory, key)
}

// GetAllEmotionKeys 獲取所有情感記錄的鍵值（用於統計和管理）
func (em *EmotionManager) GetAllEmotionKeys() []string {
	var keys []string
	for key := range em.emotionHistory {
		keys = append(keys, key)
	}
	return keys
}

// GetEmotionHistoryCount 獲取情感歷史記錄總數
func (em *EmotionManager) GetEmotionHistoryCount() int {
	return len(em.emotionHistory)
}

// GetEmotionManager 獲取全局情感管理器實例（單例模式）
func GetEmotionManager() *EmotionManager {
	emotionManagerOnce.Do(func() {
		globalEmotionManager = NewEmotionManager()
		utils.Logger.Info("全局情感管理器初始化完成")
	})
	return globalEmotionManager
}
