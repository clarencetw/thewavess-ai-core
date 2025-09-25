package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
)

// ClassificationResult 分類結果
type ClassificationResult struct {
	Level       int     `json:"level"`
	Reason      string  `json:"reason"`
	MatchedWord string  `json:"matched_word"`
	Confidence  float64 `json:"confidence"` // 兼容性字段
	ChunkID     string  `json:"chunk_id"`   // 兼容性字段
}

// EnhancedKeywordClassifier 增強關鍵字分類器
type EnhancedKeywordClassifier struct {
	keywords map[string]int
}

// NewEnhancedKeywordClassifier 創建增強關鍵字分類器
func NewEnhancedKeywordClassifier() *EnhancedKeywordClassifier {
	classifier := &EnhancedKeywordClassifier{
		keywords: make(map[string]int),
	}

	// 載入各等級關鍵字
	classifier.loadAllKeywords()

	return classifier
}

// loadAllKeywords 載入所有等級的關鍵字
func (c *EnhancedKeywordClassifier) loadAllKeywords() {
	// 載入各等級關鍵字
	c.loadL1Keywords()
	c.loadL2Keywords()
	c.loadL3Keywords()
	c.loadL4Keywords()
	c.loadL5Keywords()

	utils.Logger.Infof("關鍵字分類器載入完成，共 %d 個關鍵字", len(c.keywords))
}

// ClassifyContent 分類內容
func (c *EnhancedKeywordClassifier) ClassifyContent(content string) (*ClassificationResult, error) {
	if content == "" {
		return &ClassificationResult{Level: 1, Reason: "空內容"}, nil
	}

	content = strings.ToLower(content)
	maxLevel := 1
	matchedWord := ""

	// 掃描所有關鍵字，優先匹配更長的關鍵字
	for keyword, level := range c.keywords {
		if strings.Contains(content, keyword) {
			// 優先級：1. 更長的關鍵字  2. 更高的等級
			if len(keyword) > len(matchedWord) || (len(keyword) == len(matchedWord) && level > maxLevel) {
				maxLevel = level
				matchedWord = keyword
			}
		}
	}

	reason := fmt.Sprintf("匹配關鍵字: %s", matchedWord)
	if matchedWord == "" {
		reason = "無匹配關鍵字，預設L1"
	}

	return &ClassificationResult{
		Level:       maxLevel,
		Reason:      reason,
		MatchedWord: matchedWord,
		Confidence:  1.0, // 關鍵字匹配確定性高
		ChunkID:     "",  // 關鍵字模式下不需要
	}, nil
}

// GetClassifierInfo 獲取分類器資訊
func (c *EnhancedKeywordClassifier) GetClassifierInfo() map[string]interface{} {
	levelStats := make(map[int]int)
	for _, level := range c.keywords {
		levelStats[level]++
	}

	return map[string]interface{}{
		"total_keywords": len(c.keywords),
		"level_stats":    levelStats,
		"last_updated":   time.Now().Format(time.RFC3339),
	}
}

// AddKeyword 新增關鍵字（測試用）
func (c *EnhancedKeywordClassifier) AddKeyword(keyword string, level int) {
	c.keywords[keyword] = level
}

// RemoveKeyword 移除關鍵字（測試用）
func (c *EnhancedKeywordClassifier) RemoveKeyword(keyword string) {
	delete(c.keywords, keyword)
}

// HasKeyword 檢查是否包含關鍵字
func (c *EnhancedKeywordClassifier) HasKeyword(keyword string) bool {
	_, exists := c.keywords[keyword]
	return exists
}

// GetKeywordLevel 獲取關鍵字等級
func (c *EnhancedKeywordClassifier) GetKeywordLevel(keyword string) int {
	if level, exists := c.keywords[keyword]; exists {
		return level
	}
	return 0
}

// GetLevelStats 獲取等級統計（兼容性方法）
func (c *EnhancedKeywordClassifier) GetLevelStats() map[int]int {
	levelStats := make(map[int]int)
	for _, level := range c.keywords {
		levelStats[level]++
	}
	return levelStats
}

// GetKeywordCount 獲取關鍵字總數（兼容性方法）
func (c *EnhancedKeywordClassifier) GetKeywordCount() int {
	return len(c.keywords)
}
