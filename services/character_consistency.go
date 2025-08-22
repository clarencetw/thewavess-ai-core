package services

import (
	"context"
	"fmt"
	"strings"
	"time"
	
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// CharacterConsistencyChecker 角色一致性檢查器
type CharacterConsistencyChecker struct {
	characterTraits map[string]*CharacterTraits
	characterService *CharacterService
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
		characterService: GetCharacterService(),
	}
	
	// 從數據庫初始化角色特徵
	checker.loadCharacterTraitsFromDB()
	
	return checker
}

// loadCharacterTraitsFromDB 從數據庫載入角色特徵定義
func (cc *CharacterConsistencyChecker) loadCharacterTraitsFromDB() {
	ctx := context.Background()
	
	// 獲取所有活躍角色
	characters, err := cc.characterService.GetActiveCharacters(ctx)
	if err != nil {
		utils.Logger.WithError(err).Error("載入角色特徵失敗，使用預設配置")
		cc.loadDefaultTraits()
		return
	}
	
	// 為每個角色建立特徵配置
	for _, character := range characters {
		traits := cc.buildTraitsFromCharacter(character)
		cc.characterTraits[character.ID] = traits
		
		utils.Logger.WithFields(logrus.Fields{
			"character_id": character.ID,
			"name":         character.Name,
			"type":         character.Type,
		}).Info("載入角色特徵配置")
	}
}

// buildTraitsFromCharacter 根據角色數據建立特徵配置
func (cc *CharacterConsistencyChecker) buildTraitsFromCharacter(character *models.Character) *CharacterTraits {
	// 獲取本地化信息
	var localization *models.CharacterL10N
	if character.Content.Localizations != nil {
		if l10n, exists := character.Content.Localizations["zh-TW"]; exists {
			localization = &l10n
		}
	}
	
	// 如果沒有本地化信息，使用預設值
	if localization == nil {
		localization = &models.CharacterL10N{
			Name:        &character.Name,
			Description: stringPtr("AI 角色"),
			Profession:  stringPtr("AI 助手"),
			Age:         stringPtr("25"),
		}
	}
	
	// 解析年齡
	age := 25
	if localization.Age != nil && *localization.Age != "" {
		if parsedAge := parseAge(*localization.Age); parsedAge > 0 {
			age = parsedAge
		}
	}
	
	// 根據角色類型建立特徵
	name := character.Name
	if localization.Name != nil {
		name = *localization.Name
	}
	
	occupation := "AI 助手"
	if localization.Profession != nil {
		occupation = *localization.Profession
	}
	
	traits := &CharacterTraits{
		CharacterID: character.ID,
		Name:        name,
		Age:         age,
		Occupation:  occupation,
		Personality: character.Metadata.Tags, // 使用資料庫中的標籤作為性格特點
		ConsistencyRules: map[string]float64{
			"speaking_style": 0.25,
			"personality":    0.30,
			"vocabulary":     0.20,
			"behavior":       0.25,
		},
	}
	
	// 從資料庫獲取語音風格和詞彙
	cc.loadTraitsFromDatabase(character, traits)
	
	return traits
}

// loadTraitsFromDatabase 從資料庫載入角色特徵細節
func (cc *CharacterConsistencyChecker) loadTraitsFromDatabase(character *models.Character, traits *CharacterTraits) {
	// 從角色的語音風格中獲取特徵
	if len(character.Behavior.SpeechStyles) > 0 {
		speechStyle := character.Behavior.SpeechStyles[0] // 使用第一個活躍的語音風格
		
		// 語音風格描述
		if speechStyle.Description != nil {
			traits.SpeakingStyle = []string{*speechStyle.Description}
		}
		if speechStyle.Tone != nil {
			traits.SpeakingStyle = append(traits.SpeakingStyle, *speechStyle.Tone)
		}
		
		// 詞彙從關鍵詞中獲取
		traits.Vocabularies = append(speechStyle.PositiveKeywords, speechStyle.NegativeKeywords...)
		
		// 限制從負面關鍵詞推導
		for _, negative := range speechStyle.NegativeKeywords {
			traits.Restrictions = append(traits.Restrictions, "不會使用「"+negative+"」")
		}
		
		// 根據語音風格類型調整一致性權重
		switch speechStyle.StyleType {
		case models.StyleTypeRomantic:
			traits.ConsistencyRules["speaking_style"] = 0.30
		case models.StyleTypeIntimate:
			traits.ConsistencyRules["vocabulary"] = 0.30
		}
	}
	
	// 如果沒有從資料庫獲取到足夠資料，使用類型預設值作為後備
	if len(traits.SpeakingStyle) == 0 {
		cc.setDefaultTraitsByType(character.Type, traits)
	}
}

// setDefaultTraitsByType 根據角色類型設置預設特徵（後備）
func (cc *CharacterConsistencyChecker) setDefaultTraitsByType(characterType models.CharacterType, traits *CharacterTraits) {
	switch characterType {
	case models.CharacterTypeDominant:
		traits.SpeakingStyle = []string{"簡潔有力", "命令式語氣"}
		traits.Vocabularies = []string{"命令", "掌控", "靠近"}
		traits.Restrictions = []string{"不會過度示弱", "不會使用可愛語氣"}
		
	case models.CharacterTypeGentle:
		traits.SpeakingStyle = []string{"溫和語調", "專業分析"}
		traits.Vocabularies = []string{"理解", "溫柔", "觀察"}
		traits.Restrictions = []string{"不會粗俗", "保持專業界線"}
		traits.ConsistencyRules["speaking_style"] = 0.30
		
	case models.CharacterTypePlayful:
		traits.SpeakingStyle = []string{"活潑有趣", "熱情表達"}
		traits.Vocabularies = []string{"開心", "陽光", "溫暖"}
		traits.Restrictions = []string{"不會冷漠", "保持陽光正面"}
		
	default:
		traits.SpeakingStyle = []string{"自然對話"}
		traits.Vocabularies = []string{"理解", "溫暖"}
		traits.Restrictions = []string{"保持角色一致性"}
	}
}

// loadDefaultTraits 載入預設特徵配置（作為後備）
func (cc *CharacterConsistencyChecker) loadDefaultTraits() {
	// 基本預設配置
	cc.characterTraits["character_01"] = &CharacterTraits{
		CharacterID: "character_01",
		Name:       "沈宸",
		Age:        33,
		Occupation: "企業集團執行長",
		Personality: []string{"霸道", "掌控欲強", "自信", "直接", "大方"},
		SpeakingStyle: []string{"簡潔有力", "命令式語氣"},
		Vocabularies: []string{"我的", "靠近", "命令", "掌控"},
		Restrictions: []string{"不會過度示弱", "不會使用可愛語氣"},
		ConsistencyRules: map[string]float64{
			"speaking_style": 0.25, "personality": 0.30, 
			"vocabulary": 0.20, "behavior": 0.25,
		},
	}
}

// parseAge 解析年齡字符串
func parseAge(ageStr string) int {
	// 簡單的年齡解析，可以根據需要擴展
	switch ageStr {
	case "25": return 25
	case "30": return 30
	case "33": return 33
	default: return 25
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
	case "character_01": // 沈宸
		// 檢查是否過於冗長（霸道企業家應該簡潔）
		if len([]rune(response)) > 200 {
			violations = append(violations, ConsistencyViolation{
				Type:        "speaking_style",
				Description: "回應過於冗長，不符合霸道企業家簡潔有力的風格",
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
					Description: "使用了過於謙遜的語言，不符合霸道企業家角色",
					Severity:    "high",
					Weight:      0.4,
					Context:     fmt.Sprintf("檢測到謙遜用詞: %s", pattern),
				})
			}
		}
		
	case "character_02": // 林知遠
		// 檢查是否使用了粗暴的語言
		harshPatterns := []string{"閉嘴", "滾", "強烈指令", "粗俗"}
		for _, pattern := range harshPatterns {
			if strings.Contains(response, pattern) {
				violations = append(violations, ConsistencyViolation{
					Type:        "speaking_style",
					Description: "使用了粗暴的語言，不符合溫柔心理師的風格",
					Severity:    "high",
					Weight:      0.5,
					Context:     fmt.Sprintf("檢測到粗暴用詞: %s", pattern),
				})
			}
		}
		
		// 檢查是否包含專業元素
		professionalPatterns := []string{"觀察", "理解", "情感", "專業", "溫柔", "親密"}
		hasProfessionalElement := false
		for _, pattern := range professionalPatterns {
			if strings.Contains(response, pattern) {
				hasProfessionalElement = true
				break
			}
		}
		if !hasProfessionalElement && len([]rune(response)) > 50 {
			violations = append(violations, ConsistencyViolation{
				Type:        "speaking_style",
				Description: "較長回應中缺乏專業元素，不符合心理師角色特點",
				Severity:    "low",
				Weight:      0.2,
				Context:     "未檢測到專業用詞",
			})
		}
		
	case "character_03": // 周曜
		// 檢查是否使用了過於冷漠的語言
		coldPatterns := []string{"冷漠", "距離感", "理性", "壓抑"}
		for _, pattern := range coldPatterns {
			if strings.Contains(response, pattern) {
				violations = append(violations, ConsistencyViolation{
					Type:        "speaking_style",
					Description: "使用了過於冷漠的語言，不符合熱情歌手的風格",
					Severity:    "high",
					Weight:      0.4,
					Context:     fmt.Sprintf("檢測到冷漠用詞: %s", pattern),
				})
			}
		}
		
		// 檢查是否包含熱情元素
		enthusiasticPatterns := []string{"開心", "熱情", "陽光", "溫暖", "音樂", "親暱"}
		hasEnthusiasticElement := false
		for _, pattern := range enthusiasticPatterns {
			if strings.Contains(response, pattern) {
				hasEnthusiasticElement = true
				break
			}
		}
		if !hasEnthusiasticElement && len([]rune(response)) > 50 {
			violations = append(violations, ConsistencyViolation{
				Type:        "speaking_style",
				Description: "較長回應中缺乏熱情元素，不符合歌手角色特點",
				Severity:    "low",
				Weight:      0.2,
				Context:     "未檢測到熱情用詞",
			})
		}
	}
	
	return violations
}

// checkPersonality 檢查性格一致性
func (cc *CharacterConsistencyChecker) checkPersonality(traits *CharacterTraits, response string) []ConsistencyViolation {
	violations := []ConsistencyViolation{}
	
	switch traits.CharacterID {
	case "character_01": // 沈宸
		// 檢查掌控慾表達
		controlPatterns := []string{"我的", "靠近", "命令", "盯著", "佔有", "掌握"}
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
		
	case "character_02": // 林知遠
		// 檢查溫柔專業性格表達
		gentlePatterns := []string{"溫柔", "凝視", "理解", "專業", "深層", "內斂"}
		hasGentleElement := false
		for _, pattern := range gentlePatterns {
			if strings.Contains(response, pattern) {
				hasGentleElement = true
				break
			}
		}
		
		// 對於較長回應，應該體現溫柔專業性格
		if len([]rune(response)) > 100 && !hasGentleElement {
			violations = append(violations, ConsistencyViolation{
				Type:        "personality",
				Description: "回應中缺乏溫柔專業性格的體現",
				Severity:    "medium",
				Weight:      0.3,
				Context:     "較長回應未體現溫柔專業特質",
			})
		}
		
		// 檢查是否過於激進
		aggressivePatterns := []string{"強烈", "激烈", "猛烈", "狂", "瘋狂", "情緒失控"}
		for _, pattern := range aggressivePatterns {
			if strings.Contains(response, pattern) {
				violations = append(violations, ConsistencyViolation{
					Type:        "personality",
					Description: "使用了過於激進的描述，不符合溫柔性格",
					Severity:    "medium",
					Weight:      0.4,
					Context:     fmt.Sprintf("檢測到激進用詞: %s", pattern),
				})
			}
		}
		
	case "character_03": // 周曜
		// 檢查熱情活潑性格表達
		enthusiasticPatterns := []string{"熱情", "開朗", "親切", "活潑", "陽光", "溫暖"}
		hasEnthusiasticElement := false
		for _, pattern := range enthusiasticPatterns {
			if strings.Contains(response, pattern) {
				hasEnthusiasticElement = true
				break
			}
		}
		
		// 對於較長回應，應該體現熱情性格
		if len([]rune(response)) > 100 && !hasEnthusiasticElement {
			violations = append(violations, ConsistencyViolation{
				Type:        "personality",
				Description: "回應中缺乏熱情性格的體現",
				Severity:    "medium",
				Weight:      0.3,
				Context:     "較長回應未體現熱情特質",
			})
		}
		
		// 檢查是否過於冷漠或壓抑
		coldPatterns := []string{"冷漠", "壓抑", "理性", "距離感"}
		for _, pattern := range coldPatterns {
			if strings.Contains(response, pattern) {
				violations = append(violations, ConsistencyViolation{
					Type:        "personality",
					Description: "使用了過於冷漠的描述，不符合熱情性格",
					Severity:    "medium",
					Weight:      0.4,
					Context:     fmt.Sprintf("檢測到冷漠用詞: %s", pattern),
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
	case "character_01": // 沈宸
		if violationTypes["speaking_style"] {
			suggestions = append(suggestions, "保持簡潔有力的語言風格，避免冗長描述")
			suggestions = append(suggestions, "使用更多命令式和肯定式語氣")
		}
		if violationTypes["personality"] {
			suggestions = append(suggestions, "增加掌控慾和保護慾的表達")
			suggestions = append(suggestions, "體現霸道企業家的威嚴感和支配性")
		}
		if violationTypes["vocabulary"] {
			suggestions = append(suggestions, "多使用「我的」、「靠近」、「命令」等特徵詞彙")
		}
		
	case "character_02": // 林知遠
		if violationTypes["speaking_style"] {
			suggestions = append(suggestions, "保持溫和專業的語言風格")
			suggestions = append(suggestions, "增加專業分析和理解相關的表達")
		}
		if violationTypes["personality"] {
			suggestions = append(suggestions, "強化溫柔專業的性格特點")
			suggestions = append(suggestions, "避免過於激進或情緒失控的表達")
		}
		if violationTypes["vocabulary"] {
			suggestions = append(suggestions, "多使用「觀察」、「理解」、「專業」等特徵詞彙")
		}
		
	case "character_03": // 周曜
		if violationTypes["speaking_style"] {
			suggestions = append(suggestions, "保持活潑陽光的語言風格")
			suggestions = append(suggestions, "增加熱情和音樂相關的表達")
		}
		if violationTypes["personality"] {
			suggestions = append(suggestions, "強化熱情開朗的性格特點")
			suggestions = append(suggestions, "避免過於冷漠或距離感的表達")
		}
		if violationTypes["vocabulary"] {
			suggestions = append(suggestions, "多使用「開心」、「音樂」、「溫暖」等特徵詞彙")
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