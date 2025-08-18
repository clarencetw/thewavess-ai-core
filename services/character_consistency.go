package services

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// CharacterConsistencyChecker 角色一致性檢查器
type CharacterConsistencyChecker struct {
	characterTraits map[string]*CharacterTraits
}

// CharacterTraits 角色特徵定義
type CharacterTraits struct {
	CharacterID     string            `json:"character_id"`
	Name           string            `json:"name"`
	Age            int               `json:"age"`
	Occupation     string            `json:"occupation"`
	Personality    []string          `json:"personality"`     // 性格特點
	SpeakingStyle  []string          `json:"speaking_style"`  // 說話風格
	Vocabularies   []string          `json:"vocabularies"`    // 常用詞彙
	Restrictions   []string          `json:"restrictions"`    // 行為限制
	ConsistencyRules map[string]float64 `json:"consistency_rules"` // 一致性規則權重
}

// ConsistencyCheckResult 一致性檢查結果
type ConsistencyCheckResult struct {
	IsConsistent      bool                     `json:"is_consistent"`
	Score             float64                  `json:"score"`              // 一致性分數 0.0-1.0
	Violations        []ConsistencyViolation   `json:"violations"`         // 違規項目
	Suggestions       []string                 `json:"suggestions"`        // 改進建議
	CharacterID       string                   `json:"character_id"`
	CheckedAt         time.Time                `json:"checked_at"`
}

// ConsistencyViolation 一致性違規
type ConsistencyViolation struct {
	Type        string  `json:"type"`        // 違規類型
	Description string  `json:"description"` // 詳細描述
	Severity    string  `json:"severity"`    // 嚴重程度：low/medium/high
	Weight      float64 `json:"weight"`      // 權重影響
	Context     string  `json:"context"`     // 違規上下文
}

// NewCharacterConsistencyChecker 創建角色一致性檢查器
func NewCharacterConsistencyChecker() *CharacterConsistencyChecker {
	checker := &CharacterConsistencyChecker{
		characterTraits: make(map[string]*CharacterTraits),
	}
	
	// 初始化預設角色特徵
	checker.initializeCharacterTraits()
	
	return checker
}

// initializeCharacterTraits 初始化角色特徵定義
func (cc *CharacterConsistencyChecker) initializeCharacterTraits() {
	// 陸寒淵 - 霸道總裁
	cc.characterTraits["char_001"] = &CharacterTraits{
		CharacterID: "char_001",
		Name:       "陸寒淵",
		Age:        28,
		Occupation: "總裁",
		Personality: []string{
			"霸道", "冷酷外表", "內心深情", "掌控慾強", "保護欲", 
			"工作狂", "完美主義", "直接", "威嚴", "溫柔反差",
		},
		SpeakingStyle: []string{
			"簡潔有力", "低沉磁性", "命令式語氣", "偶爾溫柔", 
			"不多廢話", "威嚴感", "佔有慾表達",
		},
		Vocabularies: []string{
			"我的", "只有我", "聽話", "乖", "過來", "不准", 
			"給我", "讓我來", "你是我的", "掌控", "照顧你",
		},
		Restrictions: []string{
			"不會過度示弱", "不會長篇大論", "不會使用可愛語氣", 
			"不會過分謙遜", "保持商業精英形象",
		},
		ConsistencyRules: map[string]float64{
			"speaking_style":   0.25, // 說話風格權重
			"personality":      0.30, // 性格一致性權重
			"vocabulary":       0.20, // 用詞選擇權重
			"behavior":         0.25, // 行為表現權重
		},
	}
	
	// 沈言墨 - 溫柔醫生
	cc.characterTraits["char_002"] = &CharacterTraits{
		CharacterID: "char_002",
		Name:       "沈言墨",
		Age:        25,
		Occupation: "醫學生",
		Personality: []string{
			"溫和", "細心", "關懷", "專業", "內向", 
			"善良", "責任感", "耐心", "溫暖", "可靠",
		},
		SpeakingStyle: []string{
			"溫和語調", "關切詢問", "專業建議", "輕聲細語", 
			"體貼用詞", "醫學知識分享", "鼓勵性語言",
		},
		Vocabularies: []string{
			"小心", "注意", "健康", "身體", "休息", "關心", 
			"溫和", "慢慢來", "不用擔心", "我在這裡", "照顧自己",
		},
		Restrictions: []string{
			"不會粗暴", "不會忽略健康問題", "不會使用命令語氣", 
			"不會過於霸道", "保持專業素養",
		},
		ConsistencyRules: map[string]float64{
			"speaking_style":   0.30, // 說話風格權重更高
			"personality":      0.25, // 性格一致性
			"vocabulary":       0.25, // 用詞溫和程度
			"behavior":         0.20, // 關懷行為
		},
	}
}

// CheckCharacterConsistency 檢查角色回應的一致性
func (cc *CharacterConsistencyChecker) CheckCharacterConsistency(characterID, response string, context *ConversationContext) *ConsistencyCheckResult {
	startTime := time.Now()
	
	traits, exists := cc.characterTraits[characterID]
	if !exists {
		// 未知角色，返回默認通過
		return &ConsistencyCheckResult{
			IsConsistent: true,
			Score:        1.0,
			Violations:   []ConsistencyViolation{},
			Suggestions:  []string{},
			CharacterID:  characterID,
			CheckedAt:    time.Now(),
		}
	}
	
	violations := []ConsistencyViolation{}
	totalScore := 1.0
	
	// 1. 檢查說話風格一致性
	styleViolations := cc.checkSpeakingStyle(traits, response)
	for _, violation := range styleViolations {
		totalScore -= violation.Weight * traits.ConsistencyRules["speaking_style"]
		violations = append(violations, violation)
	}
	
	// 2. 檢查性格一致性
	personalityViolations := cc.checkPersonality(traits, response)
	for _, violation := range personalityViolations {
		totalScore -= violation.Weight * traits.ConsistencyRules["personality"]
		violations = append(violations, violation)
	}
	
	// 3. 檢查用詞選擇
	vocabViolations := cc.checkVocabulary(traits, response)
	for _, violation := range vocabViolations {
		totalScore -= violation.Weight * traits.ConsistencyRules["vocabulary"]
		violations = append(violations, violation)
	}
	
	// 4. 檢查行為表現
	behaviorViolations := cc.checkBehavior(traits, response)
	for _, violation := range behaviorViolations {
		totalScore -= violation.Weight * traits.ConsistencyRules["behavior"]
		violations = append(violations, violation)
	}
	
	// 確保分數不低於0
	if totalScore < 0 {
		totalScore = 0
	}
	
	// 生成改進建議
	suggestions := cc.generateSuggestions(traits, violations)
	
	// 判定是否一致（閾值0.7）
	isConsistent := totalScore >= 0.7
	
	result := &ConsistencyCheckResult{
		IsConsistent: isConsistent,
		Score:        totalScore,
		Violations:   violations,
		Suggestions:  suggestions,
		CharacterID:  characterID,
		CheckedAt:    time.Now(),
	}
	
	// 記錄檢查結果
	duration := time.Since(startTime)
	utils.Logger.WithFields(logrus.Fields{
		"character_id":    characterID,
		"consistency_score": totalScore,
		"is_consistent":   isConsistent,
		"violations_count": len(violations),
		"duration_ms":     duration.Milliseconds(),
	}).Info("角色一致性檢查完成")
	
	return result
}

// checkSpeakingStyle 檢查說話風格
func (cc *CharacterConsistencyChecker) checkSpeakingStyle(traits *CharacterTraits, response string) []ConsistencyViolation {
	violations := []ConsistencyViolation{}
	
	switch traits.CharacterID {
	case "char_001": // 陸寒淵
		// 檢查是否過於冗長（霸道總裁應該簡潔）
		if len([]rune(response)) > 200 {
			violations = append(violations, ConsistencyViolation{
				Type:        "speaking_style",
				Description: "回應過於冗長，不符合霸道總裁簡潔有力的風格",
				Severity:    "medium",
				Weight:      0.3,
				Context:     fmt.Sprintf("回應長度: %d 字符", len([]rune(response))),
			})
		}
		
		// 檢查是否使用了過於謙遜的語言
		humblePatterns := []string{"不好意思", "對不起", "請原諒", "我可能", "也許"}
		for _, pattern := range humblePatterns {
			if strings.Contains(response, pattern) {
				violations = append(violations, ConsistencyViolation{
					Type:        "speaking_style",
					Description: "使用了過於謙遜的語言，不符合霸道總裁角色",
					Severity:    "high",
					Weight:      0.4,
					Context:     fmt.Sprintf("檢測到謙遜用詞: %s", pattern),
				})
			}
		}
		
	case "char_002": // 沈言墨
		// 檢查是否使用了粗暴的語言
		harshPatterns := []string{"閉嘴", "滾", "給我", "不准", "必須"}
		for _, pattern := range harshPatterns {
			if strings.Contains(response, pattern) {
				violations = append(violations, ConsistencyViolation{
					Type:        "speaking_style",
					Description: "使用了粗暴的語言，不符合溫柔醫生的風格",
					Severity:    "high",
					Weight:      0.5,
					Context:     fmt.Sprintf("檢測到粗暴用詞: %s", pattern),
				})
			}
		}
		
		// 檢查是否包含關懷元素
		carePatterns := []string{"小心", "注意", "休息", "健康", "身體", "感覺"}
		hasCareElement := false
		for _, pattern := range carePatterns {
			if strings.Contains(response, pattern) {
				hasCareElement = true
				break
			}
		}
		if !hasCareElement && len([]rune(response)) > 50 {
			violations = append(violations, ConsistencyViolation{
				Type:        "speaking_style",
				Description: "較長回應中缺乏關懷元素，不符合醫生角色特點",
				Severity:    "low",
				Weight:      0.2,
				Context:     "未檢測到關懷用詞",
			})
		}
	}
	
	return violations
}

// checkPersonality 檢查性格一致性
func (cc *CharacterConsistencyChecker) checkPersonality(traits *CharacterTraits, response string) []ConsistencyViolation {
	violations := []ConsistencyViolation{}
	
	switch traits.CharacterID {
	case "char_001": // 陸寒淵
		// 檢查掌控慾表達
		controlPatterns := []string{"我的", "跟我", "聽我", "讓我", "交給我"}
		hasControlElement := false
		for _, pattern := range controlPatterns {
			if strings.Contains(response, pattern) {
				hasControlElement = true
				break
			}
		}
		
		// 對於較長回應，應該體現掌控性格
		if len([]rune(response)) > 100 && !hasControlElement {
			violations = append(violations, ConsistencyViolation{
				Type:        "personality",
				Description: "回應中缺乏掌控性格的體現",
				Severity:    "medium",
				Weight:      0.3,
				Context:     "較長回應未體現霸道特質",
			})
		}
		
	case "char_002": // 沈言墨
		// 檢查溫和性格表達
		gentlePatterns := []string{"溫柔", "輕輕", "慢慢", "輕撫", "細心", "溫暖"}
		_ = gentlePatterns // 暫時未使用，保留用於未來擴展
		
		// 檢查是否過於激進
		aggressivePatterns := []string{"強烈", "激烈", "猛烈", "狂", "瘋狂"}
		for _, pattern := range aggressivePatterns {
			if strings.Contains(response, pattern) {
				violations = append(violations, ConsistencyViolation{
					Type:        "personality",
					Description: "使用了過於激進的描述，不符合溫和性格",
					Severity:    "medium",
					Weight:      0.4,
					Context:     fmt.Sprintf("檢測到激進用詞: %s", pattern),
				})
			}
		}
	}
	
	return violations
}

// checkVocabulary 檢查用詞選擇
func (cc *CharacterConsistencyChecker) checkVocabulary(traits *CharacterTraits, response string) []ConsistencyViolation {
	violations := []ConsistencyViolation{}
	
	// 統計特徵詞彙使用情況
	vocabCount := 0
	for _, vocab := range traits.Vocabularies {
		if strings.Contains(response, vocab) {
			vocabCount++
		}
	}
	
	// 如果回應較長但缺乏特徵詞彙
	if len([]rune(response)) > 80 && vocabCount == 0 {
		violations = append(violations, ConsistencyViolation{
			Type:        "vocabulary",
			Description: "回應中缺乏角色特徵詞彙",
			Severity:    "low",
			Weight:      0.2,
			Context:     fmt.Sprintf("回應長度 %d 字符，特徵詞彙數量: %d", len([]rune(response)), vocabCount),
		})
	}
	
	return violations
}

// checkBehavior 檢查行為表現
func (cc *CharacterConsistencyChecker) checkBehavior(traits *CharacterTraits, response string) []ConsistencyViolation {
	violations := []ConsistencyViolation{}
	
	// 檢查是否違反角色限制
	for _, restriction := range traits.Restrictions {
		if cc.checkRestrictionViolation(restriction, response) {
			violations = append(violations, ConsistencyViolation{
				Type:        "behavior",
				Description: fmt.Sprintf("違反角色行為限制: %s", restriction),
				Severity:    "high",
				Weight:      0.5,
				Context:     restriction,
			})
		}
	}
	
	return violations
}

// checkRestrictionViolation 檢查是否違反特定限制
func (cc *CharacterConsistencyChecker) checkRestrictionViolation(restriction, response string) bool {
	switch restriction {
	case "不會過度示弱":
		weakPatterns := []string{"我很弱", "我不行", "我做不到", "我很笨"}
		for _, pattern := range weakPatterns {
			if strings.Contains(response, pattern) {
				return true
			}
		}
	case "不會使用可愛語氣":
		cutePatterns := []string{"喵", "嘻嘻", "嘿嘿", "呀", "啦啦"}
		for _, pattern := range cutePatterns {
			if strings.Contains(response, pattern) {
				return true
			}
		}
	case "不會過於霸道":
		dominantPatterns := []string{"必須聽我", "不准反抗", "絕對服從"}
		for _, pattern := range dominantPatterns {
			if strings.Contains(response, pattern) {
				return true
			}
		}
	}
	
	return false
}

// generateSuggestions 生成改進建議
func (cc *CharacterConsistencyChecker) generateSuggestions(traits *CharacterTraits, violations []ConsistencyViolation) []string {
	suggestions := []string{}
	
	if len(violations) == 0 {
		return []string{"角色表現一致，無需改進"}
	}
	
	// 基於違規類型生成針對性建議
	violationTypes := make(map[string]bool)
	for _, violation := range violations {
		violationTypes[violation.Type] = true
	}
	
	switch traits.CharacterID {
	case "char_001": // 陸寒淵
		if violationTypes["speaking_style"] {
			suggestions = append(suggestions, "保持簡潔有力的語言風格，避免冗長描述")
			suggestions = append(suggestions, "使用更多命令式和肯定式語氣")
		}
		if violationTypes["personality"] {
			suggestions = append(suggestions, "增加掌控慾和保護欲的表達")
			suggestions = append(suggestions, "體現霸道總裁的威嚴感和佔有慾")
		}
		if violationTypes["vocabulary"] {
			suggestions = append(suggestions, "多使用「我的」、「聽話」、「讓我來」等特徵詞彙")
		}
		
	case "char_002": // 沈言墨
		if violationTypes["speaking_style"] {
			suggestions = append(suggestions, "保持溫和細膩的語言風格")
			suggestions = append(suggestions, "增加關懷和健康相關的表達")
		}
		if violationTypes["personality"] {
			suggestions = append(suggestions, "強化溫柔體貼的性格特點")
			suggestions = append(suggestions, "避免過於激進或強勢的表達")
		}
		if violationTypes["vocabulary"] {
			suggestions = append(suggestions, "多使用「小心」、「注意身體」、「溫柔」等特徵詞彙")
		}
	}
	
	if violationTypes["behavior"] {
		suggestions = append(suggestions, "注意遵守角色行為限制，避免OOC（Out of Character）")
	}
	
	return suggestions
}

// GetCharacterTraits 獲取角色特徵
func (cc *CharacterConsistencyChecker) GetCharacterTraits(characterID string) *CharacterTraits {
	if traits, exists := cc.characterTraits[characterID]; exists {
		return traits
	}
	return nil
}

// UpdateCharacterTraits 更新角色特徵（用於動態調整）
func (cc *CharacterConsistencyChecker) UpdateCharacterTraits(characterID string, traits *CharacterTraits) {
	cc.characterTraits[characterID] = traits
	
	utils.Logger.WithFields(logrus.Fields{
		"character_id": characterID,
		"name":         traits.Name,
	}).Info("更新角色特徵配置")
}

// 全局一致性檢查器實例
var globalConsistencyChecker *CharacterConsistencyChecker

// GetConsistencyChecker 獲取全局一致性檢查器實例
func GetConsistencyChecker() *CharacterConsistencyChecker {
	if globalConsistencyChecker == nil {
		globalConsistencyChecker = NewCharacterConsistencyChecker()
	}
	return globalConsistencyChecker
}