package services

import (
	"fmt"
	"strings"
	"time"
	"sync"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// MemoryManager 記憶管理器 - 管理短期和長期記憶
type MemoryManager struct {
	shortTermMemory map[string]*ShortTermMemory  // 短期記憶（會話級）
	longTermMemory  map[string]*LongTermMemory   // 長期記憶（跨會話）
	mu              sync.RWMutex
}

// ShortTermMemory 短期記憶 - 當次對話上下文
type ShortTermMemory struct {
	SessionID      string                 `json:"session_id"`
	UserID         string                 `json:"user_id"`
	CharacterID    string                 `json:"character_id"`
	RecentMessages []MessageSummary       `json:"recent_messages"`
	CurrentTopic   string                 `json:"current_topic"`
	LastEmotion    string                 `json:"last_emotion"`
	UnfinishedTask string                 `json:"unfinished_task"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// MessageSummary 消息摘要
type MessageSummary struct {
	Role      string    `json:"role"`      // user/assistant
	Summary   string    `json:"summary"`    // 摘要內容（限100字）
	Emotion   string    `json:"emotion"`    // 情緒標記
	Keywords  []string  `json:"keywords"`   // 關鍵詞
	Timestamp time.Time `json:"timestamp"`
}

// LongTermMemory 長期記憶 - 跨會話持久化
type LongTermMemory struct {
	UserID        string              `json:"user_id"`
	CharacterID   string              `json:"character_id"`
	Preferences   []Preference        `json:"preferences"`      // 偏好
	Nicknames     []Nickname          `json:"nicknames"`        // 稱呼
	Milestones    []Milestone         `json:"milestones"`       // 里程碑
	Dislikes      []Dislike           `json:"dislikes"`         // 禁忌
	PersonalInfo  map[string]string   `json:"personal_info"`    // 個人信息
	LastUpdated   time.Time           `json:"last_updated"`
}

// Preference 用戶偏好
type Preference struct {
	ID          string    `json:"id"`
	Category    string    `json:"category"`    // 類別：稱呼/活動/話題等
	Content     string    `json:"content"`     // 偏好內容
	Importance  int       `json:"importance"`  // 重要度 1-5
	Evidence    string    `json:"evidence"`    // 證據來源
	CreatedAt   time.Time `json:"created_at"`
}

// Nickname 稱呼記錄
type Nickname struct {
	Name       string    `json:"name"`        // 稱呼
	Context    string    `json:"context"`     // 使用場景
	Frequency  int       `json:"frequency"`   // 使用頻率
	LastUsed   time.Time `json:"last_used"`
}

// Milestone 關係里程碑
type Milestone struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // 類型：首次牽手/告白/約會等
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Affection   int       `json:"affection"`   // 當時的好感度
}

// Dislike 禁忌話題
type Dislike struct {
	Topic      string    `json:"topic"`       // 話題/行為
	Severity   int       `json:"severity"`    // 嚴重程度 1-5
	Evidence   string    `json:"evidence"`    // 證據
	RecordedAt time.Time `json:"recorded_at"`
}

// NewMemoryManager 創建記憶管理器
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		shortTermMemory: make(map[string]*ShortTermMemory),
		longTermMemory:  make(map[string]*LongTermMemory),
	}
}

// GetShortTermMemory 獲取短期記憶
func (mm *MemoryManager) GetShortTermMemory(sessionID string) *ShortTermMemory {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	if memory, exists := mm.shortTermMemory[sessionID]; exists {
		return memory
	}
	return nil
}

// UpdateShortTermMemory 更新短期記憶
func (mm *MemoryManager) UpdateShortTermMemory(sessionID, userID, characterID string, messages []ChatMessage) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	// 獲取或創建短期記憶
	memory := mm.shortTermMemory[sessionID]
	if memory == nil {
		memory = &ShortTermMemory{
			SessionID:      sessionID,
			UserID:         userID,
			CharacterID:    characterID,
			RecentMessages: []MessageSummary{},
		}
		mm.shortTermMemory[sessionID] = memory
	}
	
	// 處理最近的消息（保留最近5條）
	for _, msg := range messages {
		summary := mm.summarizeMessage(msg)
		memory.RecentMessages = append(memory.RecentMessages, summary)
	}
	
	// 限制記憶長度
	if len(memory.RecentMessages) > 5 {
		memory.RecentMessages = memory.RecentMessages[len(memory.RecentMessages)-5:]
	}
	
	// 更新當前話題和情緒
	if len(memory.RecentMessages) > 0 {
		lastMsg := memory.RecentMessages[len(memory.RecentMessages)-1]
		memory.LastEmotion = lastMsg.Emotion
		memory.CurrentTopic = mm.extractTopic(lastMsg.Keywords)
	}
	
	memory.UpdatedAt = time.Now()
	
	utils.Logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"messages":   len(memory.RecentMessages),
		"topic":      memory.CurrentTopic,
	}).Info("短期記憶更新完成")
}

// summarizeMessage 摘要消息
func (mm *MemoryManager) summarizeMessage(msg ChatMessage) MessageSummary {
	content := msg.Content
	
	// 限制長度為100字
	if len(content) > 100 {
		content = content[:97] + "..."
	}
	
	return MessageSummary{
		Role:      msg.Role,
		Summary:   content,
		Emotion:   mm.detectEmotion(msg.Content),
		Keywords:  mm.extractKeywords(msg.Content),
		Timestamp: msg.CreatedAt,
	}
}

// detectEmotion 檢測情緒
func (mm *MemoryManager) detectEmotion(content string) string {
	content = strings.ToLower(content)
	
	if strings.Contains(content, "開心") || strings.Contains(content, "高興") || strings.Contains(content, "快樂") {
		return "happy"
	}
	if strings.Contains(content, "難過") || strings.Contains(content, "傷心") || strings.Contains(content, "痛苦") {
		return "sad"
	}
	if strings.Contains(content, "生氣") || strings.Contains(content, "憤怒") || strings.Contains(content, "煩") {
		return "angry"
	}
	if strings.Contains(content, "害羞") || strings.Contains(content, "臉紅") || strings.Contains(content, "不好意思") {
		return "shy"
	}
	if strings.Contains(content, "興奮") || strings.Contains(content, "激動") || strings.Contains(content, "期待") {
		return "excited"
	}
	if strings.Contains(content, "擔心") || strings.Contains(content, "緊張") || strings.Contains(content, "焦慮") {
		return "worried"
	}
	
	return "neutral"
}

// extractKeywords 提取關鍵詞
func (mm *MemoryManager) extractKeywords(content string) []string {
	keywords := []string{}
	
	// 提取重要詞彙
	importantWords := []string{
		"喜歡", "愛", "想念", "關心", "擔心", "害怕",
		"生日", "紀念日", "約會", "工作", "家人", "朋友",
		"累", "開心", "難過", "生氣", "興奮", "緊張",
	}
	
	for _, word := range importantWords {
		if strings.Contains(content, word) {
			keywords = append(keywords, word)
		}
	}
	
	// 限制關鍵詞數量
	if len(keywords) > 3 {
		keywords = keywords[:3]
	}
	
	return keywords
}

// extractTopic 從關鍵詞提取話題
func (mm *MemoryManager) extractTopic(keywords []string) string {
	if len(keywords) == 0 {
		return "閒聊"
	}
	
	// 話題分類
	topicMap := map[string][]string{
		"情感表達": {"喜歡", "愛", "想念", "關心"},
		"情緒狀態": {"開心", "難過", "生氣", "興奮", "緊張", "累"},
		"重要日子": {"生日", "紀念日", "約會"},
		"生活工作": {"工作", "家人", "朋友"},
	}
	
	for topic, words := range topicMap {
		for _, keyword := range keywords {
			for _, word := range words {
				if keyword == word {
					return topic
				}
			}
		}
	}
	
	return "日常對話"
}

// GetLongTermMemory 獲取長期記憶
func (mm *MemoryManager) GetLongTermMemory(userID, characterID string) *LongTermMemory {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	key := fmt.Sprintf("%s_%s", userID, characterID)
	if memory, exists := mm.longTermMemory[key]; exists {
		return memory
	}
	
	// 創建新的長期記憶
	return &LongTermMemory{
		UserID:       userID,
		CharacterID:  characterID,
		Preferences:  []Preference{},
		Nicknames:    []Nickname{},
		Milestones:   []Milestone{},
		Dislikes:     []Dislike{},
		PersonalInfo: make(map[string]string),
		LastUpdated:  time.Now(),
	}
}

// ExtractAndUpdateLongTermMemory 從對話中提取並更新長期記憶
func (mm *MemoryManager) ExtractAndUpdateLongTermMemory(userID, characterID, userMessage, aiResponse string, emotion *EmotionState) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	key := fmt.Sprintf("%s_%s", userID, characterID)
	memory := mm.longTermMemory[key]
	if memory == nil {
		memory = &LongTermMemory{
			UserID:       userID,
			CharacterID:  characterID,
			Preferences:  []Preference{},
			Nicknames:    []Nickname{},
			Milestones:   []Milestone{},
			Dislikes:     []Dislike{},
			PersonalInfo: make(map[string]string),
		}
		mm.longTermMemory[key] = memory
	}
	
	// 提取偏好
	mm.extractPreferences(memory, userMessage)
	
	// 提取稱呼
	mm.extractNicknames(memory, userMessage, aiResponse)
	
	// 檢測里程碑
	mm.detectMilestones(memory, userMessage, aiResponse, emotion)
	
	// 提取禁忌
	mm.extractDislikes(memory, userMessage)
	
	// 提取個人信息
	mm.extractPersonalInfo(memory, userMessage)
	
	memory.LastUpdated = time.Now()
	
	utils.Logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"character_id": characterID,
		"preferences":  len(memory.Preferences),
		"milestones":   len(memory.Milestones),
	}).Info("長期記憶更新完成")
}

// extractPreferences 提取偏好
func (mm *MemoryManager) extractPreferences(memory *LongTermMemory, message string) {
	// 偏好模板
	templates := []struct {
		pattern  string
		category string
	}{
		{"我喜歡", "preference"},
		{"我愛", "preference"},
		{"我最喜歡", "strong_preference"},
		{"我希望", "wish"},
		{"我想要", "desire"},
	}
	
	for _, template := range templates {
		if idx := strings.Index(message, template.pattern); idx != -1 {
			// 提取偏好內容
			content := message[idx:]
			if len(content) > 50 {
				content = content[:50]
			}
			
			// 計算重要度
			importance := mm.calculateImportance(content)
			
			// 檢查是否已存在
			exists := false
			for _, pref := range memory.Preferences {
				if strings.Contains(pref.Content, content) {
					exists = true
					break
				}
			}
			
			if !exists {
				preference := Preference{
					ID:         utils.GenerateID(8),
					Category:   template.category,
					Content:    content,
					Importance: importance,
					Evidence:   message,
					CreatedAt:  time.Now(),
				}
				memory.Preferences = append(memory.Preferences, preference)
			}
		}
	}
}

// extractNicknames 提取稱呼
func (mm *MemoryManager) extractNicknames(memory *LongTermMemory, userMessage, aiResponse string) {
	// 從AI回應中提取稱呼
	nicknames := []string{"寶貝", "乖", "小傻瓜", "親愛的", "小可愛"}
	
	for _, nickname := range nicknames {
		if strings.Contains(aiResponse, nickname) {
			// 更新或添加稱呼記錄
			found := false
			for i, n := range memory.Nicknames {
				if n.Name == nickname {
					memory.Nicknames[i].Frequency++
					memory.Nicknames[i].LastUsed = time.Now()
					found = true
					break
				}
			}
			
			if !found {
				memory.Nicknames = append(memory.Nicknames, Nickname{
					Name:      nickname,
					Context:   "親密對話",
					Frequency: 1,
					LastUsed:  time.Now(),
				})
			}
		}
	}
}

// detectMilestones 檢測里程碑
func (mm *MemoryManager) detectMilestones(memory *LongTermMemory, userMessage, aiResponse string, emotion *EmotionState) {
	// 里程碑關鍵詞
	milestoneKeywords := map[string]string{
		"第一次":  "first_time",
		"告白":   "confession",
		"在一起":  "together",
		"我愛你":  "love_declaration",
		"想見你":  "miss_you",
		"約會":   "date",
	}
	
	for keyword, mType := range milestoneKeywords {
		if strings.Contains(userMessage, keyword) || strings.Contains(aiResponse, keyword) {
			// 檢查是否已存在同類型里程碑
			exists := false
			for _, m := range memory.Milestones {
				if m.Type == mType {
					exists = true
					break
				}
			}
			
			if !exists {
				milestone := Milestone{
					ID:          utils.GenerateID(8),
					Type:        mType,
					Description: fmt.Sprintf("達成里程碑：%s", keyword),
					Date:        time.Now(),
					Affection:   emotion.Affection,
				}
				memory.Milestones = append(memory.Milestones, milestone)
			}
		}
	}
}

// extractDislikes 提取禁忌
func (mm *MemoryManager) extractDislikes(memory *LongTermMemory, message string) {
	// 禁忌模板
	dislikePatterns := []string{
		"我不喜歡", "我討厭", "不要", "別", "我害怕", "我不想",
	}
	
	for _, pattern := range dislikePatterns {
		if idx := strings.Index(message, pattern); idx != -1 {
			// 提取禁忌內容
			content := message[idx:]
			if len(content) > 50 {
				content = content[:50]
			}
			
			// 檢查是否已存在
			exists := false
			for _, dislike := range memory.Dislikes {
				if strings.Contains(dislike.Topic, content) {
					exists = true
					break
				}
			}
			
			if !exists {
				dislike := Dislike{
					Topic:      content,
					Severity:   3, // 默認中等嚴重
					Evidence:   message,
					RecordedAt: time.Now(),
				}
				memory.Dislikes = append(memory.Dislikes, dislike)
			}
		}
	}
}

// extractPersonalInfo 提取個人信息
func (mm *MemoryManager) extractPersonalInfo(memory *LongTermMemory, message string) {
	// 個人信息模板
	infoPatterns := map[string][]string{
		"birthday":   {"我的生日", "生日是"},
		"age":        {"我今年", "歲"},
		"occupation": {"我是", "我的工作", "職業"},
		"hobby":      {"我的愛好", "我喜歡"},
	}
	
	for infoType, patterns := range infoPatterns {
		for _, pattern := range patterns {
			if strings.Contains(message, pattern) {
				// 簡單提取信息（實際應該更智能）
				if _, exists := memory.PersonalInfo[infoType]; !exists {
					memory.PersonalInfo[infoType] = message
				}
			}
		}
	}
}

// calculateImportance 計算重要度
func (mm *MemoryManager) calculateImportance(content string) int {
	importance := 2 // 基礎重要度
	
	// 包含強調詞增加重要度
	emphasisWords := []string{"非常", "特別", "超級", "最", "一直", "永遠"}
	for _, word := range emphasisWords {
		if strings.Contains(content, word) {
			importance++
			break
		}
	}
	
	// 包含情感詞增加重要度
	emotionWords := []string{"愛", "喜歡", "討厭", "害怕"}
	for _, word := range emotionWords {
		if strings.Contains(content, word) {
			importance++
			break
		}
	}
	
	// 限制最大重要度
	if importance > 5 {
		importance = 5
	}
	
	return importance
}

// GetMemoryPrompt 生成記憶提示詞
func (mm *MemoryManager) GetMemoryPrompt(sessionID, userID, characterID string) string {
	shortTerm := mm.GetShortTermMemory(sessionID)
	longTerm := mm.GetLongTermMemory(userID, characterID)
	
	var prompt strings.Builder
	
	// 長期記憶部分
	prompt.WriteString("# Long-Term Memory (summary)\n")
	
	// 添加偏好
	if len(longTerm.Preferences) > 0 {
		prompt.WriteString("- 偏好：")
		for i, pref := range longTerm.Preferences {
			if i < 3 { // 限制數量
				prompt.WriteString(pref.Content)
				if i < len(longTerm.Preferences)-1 && i < 2 {
					prompt.WriteString("、")
				}
			}
		}
		prompt.WriteString("\n")
	}
	
	// 添加里程碑
	if len(longTerm.Milestones) > 0 {
		prompt.WriteString("- 里程碑：")
		for i, milestone := range longTerm.Milestones {
			if i < 2 { // 限制數量
				prompt.WriteString(milestone.Description)
				if i == 0 && len(longTerm.Milestones) > 1 {
					prompt.WriteString("；")
				}
			}
		}
		prompt.WriteString("\n")
	}
	
	// 添加禁忌
	if len(longTerm.Dislikes) > 0 {
		prompt.WriteString("- 禁忌：")
		for i, dislike := range longTerm.Dislikes {
			if i < 2 { // 限制數量
				prompt.WriteString(dislike.Topic)
				if i == 0 && len(longTerm.Dislikes) > 1 {
					prompt.WriteString("、")
				}
			}
		}
		prompt.WriteString("\n")
	}
	
	// 短期記憶部分
	if shortTerm != nil && len(shortTerm.RecentMessages) > 0 {
		prompt.WriteString("\n# Recent Context (last 3-5 turns)\n")
		for i, msg := range shortTerm.RecentMessages {
			if i < 5 { // 限制數量
				prompt.WriteString(fmt.Sprintf("- %s\n", msg.Summary))
			}
		}
		
		if shortTerm.CurrentTopic != "" {
			prompt.WriteString(fmt.Sprintf("- 當前話題：%s\n", shortTerm.CurrentTopic))
		}
	}
	
	return prompt.String()
}

// CleanupOldMemory 清理舊記憶
func (mm *MemoryManager) CleanupOldMemory() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	now := time.Now()
	
	// 清理超過24小時的短期記憶
	for sessionID, memory := range mm.shortTermMemory {
		if now.Sub(memory.UpdatedAt) > 24*time.Hour {
			delete(mm.shortTermMemory, sessionID)
			utils.Logger.WithField("session_id", sessionID).Info("清理過期短期記憶")
		}
	}
	
	// 清理長期記憶中的過期數據
	for key, memory := range mm.longTermMemory {
		// 限制偏好數量
		if len(memory.Preferences) > 20 {
			memory.Preferences = memory.Preferences[len(memory.Preferences)-20:]
		}
		
		// 限制里程碑數量
		if len(memory.Milestones) > 10 {
			memory.Milestones = memory.Milestones[len(memory.Milestones)-10:]
		}
		
		// 更新記憶
		mm.longTermMemory[key] = memory
	}
}

// GetMemoryStatistics 獲取記憶統計
func (mm *MemoryManager) GetMemoryStatistics() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	stats := map[string]interface{}{
		"short_term_sessions": len(mm.shortTermMemory),
		"long_term_users":     len(mm.longTermMemory),
		"total_preferences":   0,
		"total_milestones":    0,
		"total_dislikes":      0,
	}
	
	// 統計長期記憶數據
	for _, memory := range mm.longTermMemory {
		stats["total_preferences"] = stats["total_preferences"].(int) + len(memory.Preferences)
		stats["total_milestones"] = stats["total_milestones"].(int) + len(memory.Milestones)
		stats["total_dislikes"] = stats["total_dislikes"].(int) + len(memory.Dislikes)
	}
	
	return stats
}