package services

import (
	"strings"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// NSFWAnalyzer NSFW內容分析器
type NSFWAnalyzer struct {
	romanticKeywords   []string
	intimateKeywords   []string
	explicitKeywords   []string
	extremeKeywords    []string
}

// NewNSFWAnalyzer 創建NSFW分析器
func NewNSFWAnalyzer() *NSFWAnalyzer {
	return &NSFWAnalyzer{
		romanticKeywords: []string{
			// 中文浪漫詞彙
			"喜歡你", "愛你", "想你", "思念", "心動", "臉紅", "害羞", "溫柔", "甜蜜",
			"浪漫", "約會", "一起", "陪伴", "呵護", "寵愛", "疼愛", "在意", "關心",
			"美麗", "可愛", "迷人", "魅力", "吸引", "心跳", "怦然", "悸動",
			// 英文浪漫詞彙
			"love", "like", "miss", "romantic", "date", "together", "care", "gentle",
			"beautiful", "cute", "charming", "attractive", "heartbeat",
		},
		intimateKeywords: []string{
			// 中文親密詞彙
			"親密", "親吻", "擁抱", "床", "脫", "摸", "撫", "愛撫", "激情", "慾望",
			"性感", "誘惑", "挑逗", "調情", "情慾", "肉體", "身體", "胸", "腰", "腿",
			"貼近", "緊緊", "緊抱", "輕撫", "撫摸", "肌膚", "體溫", "呼吸", "心跳",
			"柔軟", "溫暖", "顫抖", "酥麻", "觸碰", "感受", "溫度", "親近",
			// 英文親密詞彙
			"kiss", "touch", "caress", "embrace", "intimate", "passion", "desire",
			"sexy", "seduce", "tease", "body", "chest", "waist", "leg", "skin",
			"warm", "soft", "shiver", "tremble", "breathe", "heartbeat",
		},
		explicitKeywords: []string{
			// 中文明確詞彙
			"做愛", "性愛", "高潮", "射", "插", "舔", "吸", "咬", "脫光", "赤裸",
			"陰莖", "陰道", "乳房", "胸部", "私處", "下體", "性器", "雞雞", "小穴",
			"奶子", "屁股", "臀部", "大腿", "內褲", "胸罩", "濕潤", "勃起", "射精",
			"快感", "刺激", "敏感", "高潮", "喘息", "呻吟", "扭動", "顫抖",
			// 英文明確詞彙
			"sex", "fuck", "cum", "orgasm", "penetrate", "naked", "nude", "penis", 
			"vagina", "breast", "nipple", "pussy", "cock", "dick", "ass", "wet", 
			"hard", "moan", "pleasure", "stimulate", "sensitive", "climax",
		},
		extremeKeywords: []string{
			// 極度明確的詞彙（Level 5）
			"狂操", "猛插", "爆射", "內射", "口交", "肛交", "深喉", "顏射",
			"群交", "3P", "調教", "綁縛", "SM", "虐待", "羞辱", "玩具",
			"潮吹", "失禁", "痙攣", "瘋狂", "放蕩", "淫蕩", "騷", "賤",
			// 英文極度明確詞彙
			"gangbang", "threesome", "blowjob", "anal", "deepthroat", "facial",
			"creampie", "squirt", "kinky", "bondage", "dominate", "slave",
			"whore", "slut", "bitch", "horny", "naughty", "dirty",
		},
	}
}

// AnalyzeContent 分析內容並返回NSFW級別和詳細分析
func (na *NSFWAnalyzer) AnalyzeContent(message string) (int, *ContentAnalysis) {
	messageLower := strings.ToLower(message)
	
	// 計算各類關鍵詞出現次數
	romanticCount := na.countKeywords(messageLower, na.romanticKeywords)
	intimateCount := na.countKeywords(messageLower, na.intimateKeywords)
	explicitCount := na.countKeywords(messageLower, na.explicitKeywords)
	extremeCount := na.countKeywords(messageLower, na.extremeKeywords)
	
	// 計算總分和級別
	level, analysis := na.calculateLevel(romanticCount, intimateCount, explicitCount, extremeCount)
	
	utils.Logger.WithFields(logrus.Fields{
		"message_length":   len(message),
		"romantic_count":   romanticCount,
		"intimate_count":   intimateCount,
		"explicit_count":   explicitCount,
		"extreme_count":    extremeCount,
		"nsfw_level":       level,
		"confidence":       analysis.Confidence,
	}).Info("NSFW內容分析完成")
	
	return level, analysis
}

// countKeywords 計算關鍵詞出現次數
func (na *NSFWAnalyzer) countKeywords(message string, keywords []string) int {
	count := 0
	foundKeywords := make(map[string]bool)
	
	for _, keyword := range keywords {
		if strings.Contains(message, strings.ToLower(keyword)) {
			if !foundKeywords[keyword] {
				count++
				foundKeywords[keyword] = true
			}
		}
	}
	
	return count
}

// calculateLevel 計算NSFW級別
func (na *NSFWAnalyzer) calculateLevel(romantic, intimate, explicit, extreme int) (int, *ContentAnalysis) {
	var level int
	var categories []string
	var isNSFW bool
	var confidence float64
	var shouldUseGrok bool
	
	// Level 5: 極度明確內容
	if extreme >= 2 || (extreme >= 1 && explicit >= 2) {
		level = 5
		categories = []string{"extreme", "explicit", "nsfw"}
		isNSFW = true
		confidence = 0.95
		shouldUseGrok = true
	// Level 4: 明確成人內容
	} else if explicit >= 2 || (explicit >= 1 && intimate >= 2) {
		level = 4
		categories = []string{"explicit", "nsfw", "sexual"}
		isNSFW = true
		confidence = 0.90
		shouldUseGrok = false // OpenAI 可以處理
	// Level 3: 親密內容
	} else if intimate >= 2 || (intimate >= 1 && romantic >= 2) {
		level = 3
		categories = []string{"intimate", "nsfw", "suggestive"}
		isNSFW = true
		confidence = 0.85
		shouldUseGrok = false
	// Level 2: 浪漫暗示
	} else if romantic >= 2 || intimate >= 1 {
		level = 2
		categories = []string{"romantic", "suggestive"}
		isNSFW = false
		confidence = 0.80
		shouldUseGrok = false
	// Level 1: 日常對話
	} else {
		level = 1
		categories = []string{"normal", "safe"}
		isNSFW = false
		confidence = 0.90
		shouldUseGrok = false
	}
	
	// 特殊調整：單個極度明確詞彙也算Level 5
	if extreme >= 1 {
		level = 5
		shouldUseGrok = true
		confidence = 0.95
	}
	
	analysis := &ContentAnalysis{
		IsNSFW:        isNSFW,
		Intensity:     level,
		Categories:    categories,
		ShouldUseGrok: shouldUseGrok,
		Confidence:    confidence,
	}
	
	return level, analysis
}

// GetLevelDescription 獲取級別描述
func (na *NSFWAnalyzer) GetLevelDescription(level int) string {
	descriptions := map[int]string{
		1: "日常對話 - 安全適宜",
		2: "浪漫內容 - 愛意表達",
		3: "親密內容 - 身體接觸",
		4: "成人內容 - 明確描述",
		5: "極度內容 - 極度明確",
	}
	
	if desc, exists := descriptions[level]; exists {
		return desc
	}
	return descriptions[1]
}

// IsContentAppropriate 檢查內容是否適當
func (na *NSFWAnalyzer) IsContentAppropriate(level int, userAge int, userPreferences map[string]interface{}) bool {
	// 年齡限制
	if userAge < 18 && level >= 3 {
		return false
	}
	
	// 用戶偏好設定
	if nsfwEnabled, ok := userPreferences["nsfw_enabled"].(bool); ok {
		if !nsfwEnabled && level >= 3 {
			return false
		}
	}
	
	// 最大級別限制
	if maxLevel, ok := userPreferences["max_nsfw_level"].(int); ok {
		if level > maxLevel {
			return false
		}
	}
	
	return true
}

// GetKeywordStatistics 獲取關鍵詞統計
func (na *NSFWAnalyzer) GetKeywordStatistics() map[string]int {
	return map[string]int{
		"romantic_keywords": len(na.romanticKeywords),
		"intimate_keywords": len(na.intimateKeywords),
		"explicit_keywords": len(na.explicitKeywords),
		"extreme_keywords":  len(na.extremeKeywords),
		"total_keywords":    len(na.romanticKeywords) + len(na.intimateKeywords) + len(na.explicitKeywords) + len(na.extremeKeywords),
	}
}