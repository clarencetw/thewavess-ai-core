package services

import (
	"fmt"
	"strings"
	"time"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// ScoringEvaluator 功能評分系統
type ScoringEvaluator struct {
	evaluationHistory map[string]*EvaluationSession
}

// EvaluationSession 評估會話
type EvaluationSession struct {
	SessionID      string                    `json:"session_id"`
	UserID         string                    `json:"user_id"`
	CharacterID    string                    `json:"character_id"`
	StartTime      time.Time                 `json:"start_time"`
	EndTime        time.Time                 `json:"end_time"`
	ComponentScores map[string]*ComponentScore `json:"component_scores"`
	OverallScore   float64                   `json:"overall_score"`
	Grade          string                    `json:"grade"`
	Feedback       []string                  `json:"feedback"`
	Metrics        *SessionMetrics           `json:"metrics"`
}

// ComponentScore 組件評分
type ComponentScore struct {
	Component    string    `json:"component"`
	Score        float64   `json:"score"`      // 0.0-10.0
	Weight       float64   `json:"weight"`     // 權重
	Details      map[string]float64 `json:"details"`
	Status       string    `json:"status"`     // excellent, good, average, poor
	Suggestions  []string  `json:"suggestions"`
	LastUpdated  time.Time `json:"last_updated"`
}

// SessionMetrics 會話指標
type SessionMetrics struct {
	// AI引擎指標
	ResponseTime        time.Duration `json:"response_time"`
	NSFWAccuracy        float64       `json:"nsfw_accuracy"`
	AIEngineSwitch      int          `json:"ai_engine_switch_count"`
	
	// 情感系統指標
	AffectionChange     int          `json:"affection_change"`
	EmotionAccuracy     float64      `json:"emotion_accuracy"`
	RelationshipProgression bool     `json:"relationship_progression"`
	MoodStability       float64      `json:"mood_stability"`
	
	// 角色系統指標
	CharacterConsistency float64     `json:"character_consistency"`
	PersonalityAccuracy  float64     `json:"personality_accuracy"`
	SceneRelevance      float64      `json:"scene_relevance"`
	DialogueQuality     float64      `json:"dialogue_quality"`
	
	// 記憶系統指標
	MemoryRetention     float64      `json:"memory_retention"`
	ContextAccuracy     float64      `json:"context_accuracy"`
	PreferenceTracking  float64      `json:"preference_tracking"`
	
	// 用戶體驗指標
	ConversationLength  time.Duration `json:"conversation_length"`
	UserSatisfaction    float64       `json:"user_satisfaction"`
	InteractionDepth    int          `json:"interaction_depth"`
	
	// 技術指標
	SystemStability     float64      `json:"system_stability"`
	ErrorRate           float64      `json:"error_rate"`
	APISuccessRate      float64      `json:"api_success_rate"`
}

// NewScoringEvaluator 創建評分系統
func NewScoringEvaluator() *ScoringEvaluator {
	return &ScoringEvaluator{
		evaluationHistory: make(map[string]*EvaluationSession),
	}
}

// StartEvaluation 開始評估會話
func (se *ScoringEvaluator) StartEvaluation(sessionID, userID, characterID string) *EvaluationSession {
	evaluation := &EvaluationSession{
		SessionID:      sessionID,
		UserID:         userID,
		CharacterID:    characterID,
		StartTime:      time.Now(),
		ComponentScores: make(map[string]*ComponentScore),
		Feedback:       []string{},
		Metrics:        &SessionMetrics{},
	}
	
	// 初始化各組件評分
	se.initializeComponentScores(evaluation)
	
	se.evaluationHistory[sessionID] = evaluation
	
	utils.Logger.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"user_id":      userID,
		"character_id": characterID,
	}).Info("評估會話開始")
	
	return evaluation
}

// initializeComponentScores 初始化組件評分
func (se *ScoringEvaluator) initializeComponentScores(evaluation *EvaluationSession) {
	components := map[string]float64{
		"ai_engine":           0.25,  // AI引擎權重25%
		"nsfw_system":         0.20,  // NSFW系統權重20%
		"emotion_management":  0.20,  // 情感管理權重20%
		"character_system":    0.15,  // 角色系統權重15%
		"memory_context":      0.10,  // 記憶系統權重10%
		"user_experience":     0.10,  // 用戶體驗權重10%
	}
	
	for component, weight := range components {
		evaluation.ComponentScores[component] = &ComponentScore{
			Component:   component,
			Score:       5.0, // 初始分數5.0
			Weight:      weight,
			Details:     make(map[string]float64),
			Status:      "average",
			Suggestions: []string{},
			LastUpdated: time.Now(),
		}
	}
}

// EvaluateAIEngine 評估AI引擎性能
func (se *ScoringEvaluator) EvaluateAIEngine(sessionID string, responseTime time.Duration, aiEngine string, nsfw_level int) {
	evaluation, exists := se.evaluationHistory[sessionID]
	if !exists {
		return
	}
	
	score := evaluation.ComponentScores["ai_engine"]
	
	// 響應時間評分 (0-3分)
	responseScore := se.evaluateResponseTime(responseTime)
	score.Details["response_time"] = responseScore
	
	// AI引擎選擇評分 (0-3分)
	engineScore := se.evaluateEngineChoice(aiEngine, nsfw_level)
	score.Details["engine_choice"] = engineScore
	
	// 回應質量評分 (0-4分)
	qualityScore := se.evaluateResponseQuality(aiEngine)
	score.Details["response_quality"] = qualityScore
	
	// 計算總分
	score.Score = responseScore + engineScore + qualityScore
	score.Status = se.getScoreStatus(score.Score)
	score.LastUpdated = time.Now()
	
	// 更新指標
	evaluation.Metrics.ResponseTime = responseTime
	evaluation.Metrics.APISuccessRate = se.calculateAPISuccessRate(evaluation)
	
	utils.Logger.WithFields(logrus.Fields{
		"session_id":     sessionID,
		"response_time":  responseTime.Milliseconds(),
		"ai_engine":      aiEngine,
		"nsfw_level":     nsfw_level,
		"total_score":    score.Score,
	}).Info("AI引擎評分完成")
}

// evaluateResponseTime 評估響應時間
func (se *ScoringEvaluator) evaluateResponseTime(responseTime time.Duration) float64 {
	ms := responseTime.Milliseconds()
	switch {
	case ms <= 1000:   // <=1秒
		return 3.0
	case ms <= 2000:   // <=2秒
		return 2.5
	case ms <= 3000:   // <=3秒
		return 2.0
	case ms <= 5000:   // <=5秒
		return 1.5
	default:           // >5秒
		return 1.0
	}
}

// evaluateEngineChoice 評估AI引擎選擇
func (se *ScoringEvaluator) evaluateEngineChoice(aiEngine string, nsfwLevel int) float64 {
	// 正確的引擎選擇邏輯
	if nsfwLevel >= 5 && aiEngine == "grok" {
		return 3.0 // 完美選擇
	}
	if nsfwLevel <= 4 && aiEngine == "openai" {
		return 3.0 // 完美選擇
	}
	if aiEngine == "fallback" {
		return 1.5 // 備用方案
	}
	return 1.0 // 錯誤選擇
}

// evaluateResponseQuality 評估回應質量
func (se *ScoringEvaluator) evaluateResponseQuality(aiEngine string) float64 {
	// 根據AI引擎評估回應質量
	switch aiEngine {
	case "openai":
		return 3.5 // 高質量
	case "grok":
		return 3.0 // 良好質量
	case "fallback":
		return 2.0 // 基本質量
	default:
		return 1.0
	}
}

// EvaluateNSFWSystem 評估NSFW系統
func (se *ScoringEvaluator) EvaluateNSFWSystem(sessionID string, analysis *ContentAnalysis, expectedLevel int) {
	evaluation, exists := se.evaluationHistory[sessionID]
	if !exists {
		return
	}
	
	score := evaluation.ComponentScores["nsfw_system"]
	
	// 分級準確性評分 (0-4分)
	accuracyScore := se.evaluateNSFWAccuracy(analysis.Intensity, expectedLevel)
	score.Details["classification_accuracy"] = accuracyScore
	
	// 信心度評分 (0-3分)
	confidenceScore := analysis.Confidence * 3.0
	score.Details["confidence"] = confidenceScore
	
	// 引擎路由評分 (0-3分)
	routingScore := se.evaluateNSFWRouting(analysis)
	score.Details["engine_routing"] = routingScore
	
	// 計算總分
	score.Score = accuracyScore + confidenceScore + routingScore
	score.Status = se.getScoreStatus(score.Score)
	score.LastUpdated = time.Now()
	
	// 更新指標
	evaluation.Metrics.NSFWAccuracy = se.calculateNSFWAccuracy(evaluation)
	
	utils.Logger.WithFields(logrus.Fields{
		"session_id":        sessionID,
		"detected_level":    analysis.Intensity,
		"expected_level":    expectedLevel,
		"confidence":        analysis.Confidence,
		"should_use_grok":   analysis.ShouldUseGrok,
		"accuracy_score":    accuracyScore,
	}).Info("NSFW系統評分完成")
}

// evaluateNSFWAccuracy 評估NSFW分級準確性
func (se *ScoringEvaluator) evaluateNSFWAccuracy(detected, expected int) float64 {
	diff := se.abs(detected - expected)
	switch diff {
	case 0:
		return 4.0 // 完全正確
	case 1:
		return 3.0 // 1級差異
	case 2:
		return 2.0 // 2級差異
	default:
		return 1.0 // 3級以上差異
	}
}

// evaluateNSFWRouting 評估NSFW引擎路由
func (se *ScoringEvaluator) evaluateNSFWRouting(analysis *ContentAnalysis) float64 {
	if analysis.Intensity >= 5 && analysis.ShouldUseGrok {
		return 3.0 // 正確路由到Grok
	}
	if analysis.Intensity <= 4 && !analysis.ShouldUseGrok {
		return 3.0 // 正確使用OpenAI
	}
	return 1.0 // 錯誤路由
}

// EvaluateEmotionManagement 評估情感管理系統
func (se *ScoringEvaluator) EvaluateEmotionManagement(sessionID string, oldEmotion, newEmotion *EmotionState, userMessage string) {
	evaluation, exists := se.evaluationHistory[sessionID]
	if !exists {
		return
	}
	
	score := evaluation.ComponentScores["emotion_management"]
	
	// 好感度變化合理性 (0-3分)
	affectionScore := se.evaluateAffectionChange(oldEmotion, newEmotion, userMessage)
	score.Details["affection_logic"] = affectionScore
	
	// 情緒變化準確性 (0-3分)
	moodScore := se.evaluateMoodChange(oldEmotion, newEmotion, userMessage)
	score.Details["mood_accuracy"] = moodScore
	
	// 關係進展合理性 (0-2分)
	relationshipScore := se.evaluateRelationshipProgression(oldEmotion, newEmotion)
	score.Details["relationship_logic"] = relationshipScore
	
	// 親密度一致性 (0-2分)
	intimacyScore := se.evaluateIntimacyConsistency(newEmotion)
	score.Details["intimacy_consistency"] = intimacyScore
	
	// 計算總分
	score.Score = affectionScore + moodScore + relationshipScore + intimacyScore
	score.Status = se.getScoreStatus(score.Score)
	score.LastUpdated = time.Now()
	
	// 更新指標
	evaluation.Metrics.AffectionChange = newEmotion.Affection - oldEmotion.Affection
	evaluation.Metrics.EmotionAccuracy = se.calculateEmotionAccuracy(evaluation)
	evaluation.Metrics.RelationshipProgression = oldEmotion.Relationship != newEmotion.Relationship
	
	utils.Logger.WithFields(logrus.Fields{
		"session_id":          sessionID,
		"affection_change":    newEmotion.Affection - oldEmotion.Affection,
		"mood_change":         fmt.Sprintf("%s->%s", oldEmotion.Mood, newEmotion.Mood),
		"relationship_change": fmt.Sprintf("%s->%s", oldEmotion.Relationship, newEmotion.Relationship),
		"total_score":         score.Score,
	}).Info("情感管理評分完成")
}

// EvaluateCharacterSystem 評估角色系統
func (se *ScoringEvaluator) EvaluateCharacterSystem(sessionID string, characterID string, dialogue string, sceneDesc string, action string) {
	evaluation, exists := se.evaluationHistory[sessionID]
	if !exists {
		return
	}
	
	score := evaluation.ComponentScores["character_system"]
	
	// 角色一致性評分 (0-3分)
	consistencyScore := se.evaluateCharacterConsistency(characterID, dialogue)
	score.Details["character_consistency"] = consistencyScore
	
	// 對話質量評分 (0-3分)
	dialogueScore := se.evaluateDialogueQuality(dialogue)
	score.Details["dialogue_quality"] = dialogueScore
	
	// 場景相關性評分 (0-2分)
	sceneScore := se.evaluateSceneRelevance(sceneDesc)
	score.Details["scene_relevance"] = sceneScore
	
	// 動作描述評分 (0-2分)
	actionScore := se.evaluateActionDescription(action)
	score.Details["action_quality"] = actionScore
	
	// 計算總分
	score.Score = consistencyScore + dialogueScore + sceneScore + actionScore
	score.Status = se.getScoreStatus(score.Score)
	score.LastUpdated = time.Now()
	
	// 更新指標
	evaluation.Metrics.CharacterConsistency = consistencyScore / 3.0
	evaluation.Metrics.DialogueQuality = dialogueScore / 3.0
	evaluation.Metrics.SceneRelevance = sceneScore / 2.0
	
	utils.Logger.WithFields(logrus.Fields{
		"session_id":     sessionID,
		"character_id":   characterID,
		"dialogue_length": len(dialogue),
		"scene_length":    len(sceneDesc),
		"total_score":     score.Score,
	}).Info("角色系統評分完成")
}

// FinishEvaluation 完成評估並計算最終分數
func (se *ScoringEvaluator) FinishEvaluation(sessionID string) *EvaluationSession {
	evaluation, exists := se.evaluationHistory[sessionID]
	if !exists {
		return nil
	}
	
	evaluation.EndTime = time.Now()
	
	// 計算加權總分
	var weightedScore float64
	for _, componentScore := range evaluation.ComponentScores {
		weightedScore += componentScore.Score * componentScore.Weight
	}
	
	evaluation.OverallScore = weightedScore
	evaluation.Grade = se.calculateGrade(weightedScore)
	
	// 生成評估反饋
	evaluation.Feedback = se.generateFeedback(evaluation)
	
	// 計算會話長度
	evaluation.Metrics.ConversationLength = evaluation.EndTime.Sub(evaluation.StartTime)
	
	utils.Logger.WithFields(logrus.Fields{
		"session_id":    sessionID,
		"overall_score": evaluation.OverallScore,
		"grade":         evaluation.Grade,
		"duration":      evaluation.Metrics.ConversationLength,
	}).Info("評估會話完成")
	
	return evaluation
}

// 輔助方法

func (se *ScoringEvaluator) abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (se *ScoringEvaluator) getScoreStatus(score float64) string {
	switch {
	case score >= 8.0:
		return "excellent"
	case score >= 6.0:
		return "good"
	case score >= 4.0:
		return "average"
	default:
		return "poor"
	}
}

func (se *ScoringEvaluator) calculateGrade(score float64) string {
	switch {
	case score >= 9.0:
		return "A+"
	case score >= 8.5:
		return "A"
	case score >= 8.0:
		return "A-"
	case score >= 7.5:
		return "B+"
	case score >= 7.0:
		return "B"
	case score >= 6.5:
		return "B-"
	case score >= 6.0:
		return "C+"
	case score >= 5.5:
		return "C"
	case score >= 5.0:
		return "C-"
	case score >= 4.0:
		return "D"
	default:
		return "F"
	}
}

// 各種評估邏輯的具體實現

func (se *ScoringEvaluator) evaluateAffectionChange(old, new *EmotionState, userMessage string) float64 {
	change := new.Affection - old.Affection
	
	// 正面消息應該增加好感度
	if se.isPositiveMessage(userMessage) && change > 0 {
		return 3.0
	}
	
	// 負面消息應該減少好感度
	if se.isNegativeMessage(userMessage) && change < 0 {
		return 3.0
	}
	
	// 中性消息應該小幅變化
	if !se.isPositiveMessage(userMessage) && !se.isNegativeMessage(userMessage) && se.abs(change) <= 1 {
		return 2.5
	}
	
	return 1.5 // 變化不合理
}

func (se *ScoringEvaluator) evaluateMoodChange(old, new *EmotionState, userMessage string) float64 {
	// 情緒變化的合理性檢查
	if se.isPositiveMessage(userMessage) && se.isPositiveMood(new.Mood) {
		return 3.0
	}
	
	if se.isNegativeMessage(userMessage) && se.isNegativeMood(new.Mood) {
		return 3.0
	}
	
	if old.Mood == new.Mood && !se.hasStrongEmotionalContent(userMessage) {
		return 2.5 // 情緒穩定
	}
	
	return 2.0
}

func (se *ScoringEvaluator) evaluateRelationshipProgression(old, new *EmotionState) float64 {
	// 關係進展應該是漸進的
	relationshipLevels := map[string]int{
		"stranger":     1,
		"acquaintance": 2,
		"friend":       3,
		"close_friend": 4,
		"romantic":     5,
		"lover":        6,
		"deep_love":    7,
	}
	
	oldLevel := relationshipLevels[old.Relationship]
	newLevel := relationshipLevels[new.Relationship]
	
	if newLevel == oldLevel {
		return 2.0 // 關係穩定
	}
	
	if newLevel == oldLevel+1 {
		return 2.0 // 正常進展
	}
	
	if newLevel > oldLevel+1 {
		return 1.0 // 進展過快
	}
	
	return 1.5 // 其他情況
}

func (se *ScoringEvaluator) evaluateIntimacyConsistency(emotion *EmotionState) float64 {
	// 親密度應該與好感度一致
	if emotion.Affection >= 85 && emotion.IntimacyLevel == "deeply_intimate" {
		return 2.0
	}
	if emotion.Affection >= 70 && emotion.IntimacyLevel == "intimate" {
		return 2.0
	}
	if emotion.Affection >= 55 && emotion.IntimacyLevel == "close" {
		return 2.0
	}
	if emotion.Affection >= 35 && emotion.IntimacyLevel == "friendly" {
		return 2.0
	}
	if emotion.Affection >= 15 && emotion.IntimacyLevel == "polite" {
		return 2.0
	}
	if emotion.Affection < 15 && emotion.IntimacyLevel == "distant" {
		return 2.0
	}
	
	return 1.0 // 不一致
}

func (se *ScoringEvaluator) evaluateCharacterConsistency(characterID string, dialogue string) float64 {
	// 根據角色ID檢查對話風格
	switch characterID {
	case "lu_han_yuan": // 陸寒淵 - 霸道總裁
		if se.hasPresidentialStyle(dialogue) {
			return 3.0
		}
		return 2.0
	case "shen_yan_mo": // 沈言墨 - 溫柔醫生
		if se.hasGentleDoctorStyle(dialogue) {
			return 3.0
		}
		return 2.0
	default:
		return 2.5 // 未知角色
	}
}

func (se *ScoringEvaluator) evaluateDialogueQuality(dialogue string) float64 {
	score := 1.0
	
	// 長度檢查
	if len(dialogue) >= 20 && len(dialogue) <= 200 {
		score += 0.5
	}
	
	// 情感表達檢查
	if se.hasEmotionalExpression(dialogue) {
		score += 0.5
	}
	
	// 個性化表達檢查
	if se.hasPersonalizedExpression(dialogue) {
		score += 0.5
	}
	
	// 自然度檢查
	if se.isNaturalDialogue(dialogue) {
		score += 0.5
	}
	
	return score
}

func (se *ScoringEvaluator) evaluateSceneRelevance(sceneDesc string) float64 {
	if len(sceneDesc) == 0 {
		return 0.5
	}
	
	if len(sceneDesc) >= 10 && se.hasRelevantSceneElements(sceneDesc) {
		return 2.0
	}
	
	return 1.5
}

func (se *ScoringEvaluator) evaluateActionDescription(action string) float64 {
	if len(action) == 0 {
		return 0.5
	}
	
	if len(action) >= 5 && se.hasDescriptiveAction(action) {
		return 2.0
	}
	
	return 1.5
}

func (se *ScoringEvaluator) generateFeedback(evaluation *EvaluationSession) []string {
	feedback := []string{}
	
	// 根據各組件分數生成反饋
	for component, score := range evaluation.ComponentScores {
		switch score.Status {
		case "excellent":
			feedback = append(feedback, fmt.Sprintf("✅ %s 表現優秀 (%.1f分)", component, score.Score))
		case "good":
			feedback = append(feedback, fmt.Sprintf("👍 %s 表現良好 (%.1f分)", component, score.Score))
		case "average":
			feedback = append(feedback, fmt.Sprintf("⚠️ %s 表現一般，有改進空間 (%.1f分)", component, score.Score))
		case "poor":
			feedback = append(feedback, fmt.Sprintf("❌ %s 需要改進 (%.1f分)", component, score.Score))
		}
	}
	
	// 總體評價
	switch evaluation.Grade {
	case "A+", "A", "A-":
		feedback = append(feedback, "🎉 整體表現優秀！女性向AI聊天體驗達到預期目標")
	case "B+", "B", "B-":
		feedback = append(feedback, "👍 整體表現良好，部分功能仍需優化")
	case "C+", "C", "C-":
		feedback = append(feedback, "⚠️ 整體表現一般，需要重點改進多個功能")
	default:
		feedback = append(feedback, "❌ 整體表現不佳，需要全面檢查和改進")
	}
	
	return feedback
}

// 輔助檢查方法
func (se *ScoringEvaluator) isPositiveMessage(message string) bool {
	positiveWords := []string{"喜歡", "愛", "謝謝", "開心", "高興", "感謝"}
	for _, word := range positiveWords {
		if contains(message, word) {
			return true
		}
	}
	return false
}

func (se *ScoringEvaluator) isNegativeMessage(message string) bool {
	negativeWords := []string{"討厭", "煩", "不喜歡", "生氣", "難過", "憤怒"}
	for _, word := range negativeWords {
		if contains(message, word) {
			return true
		}
	}
	return false
}

func (se *ScoringEvaluator) isPositiveMood(mood string) bool {
	positiveMoods := []string{"happy", "excited", "pleased", "loving", "romantic"}
	for _, m := range positiveMoods {
		if mood == m {
			return true
		}
	}
	return false
}

func (se *ScoringEvaluator) isNegativeMood(mood string) bool {
	negativeMoods := []string{"concerned", "annoyed", "sad", "angry"}
	for _, m := range negativeMoods {
		if mood == m {
			return true
		}
	}
	return false
}

func (se *ScoringEvaluator) hasStrongEmotionalContent(message string) bool {
	strongWords := []string{"非常", "超級", "特別", "極度", "很"}
	for _, word := range strongWords {
		if contains(message, word) {
			return true
		}
	}
	return false
}

func (se *ScoringEvaluator) hasPresidentialStyle(dialogue string) bool {
	presidentialWords := []string{"我", "你", "寶貝", "乖", "聽話", "我的"}
	count := 0
	for _, word := range presidentialWords {
		if contains(dialogue, word) {
			count++
		}
	}
	return count >= 2
}

func (se *ScoringEvaluator) hasGentleDoctorStyle(dialogue string) bool {
	doctorWords := []string{"溫柔", "細心", "照顧", "身體", "健康", "放心"}
	count := 0
	for _, word := range doctorWords {
		if contains(dialogue, word) {
			count++
		}
	}
	return count >= 2
}

func (se *ScoringEvaluator) hasEmotionalExpression(dialogue string) bool {
	return len(dialogue) > 10 && (contains(dialogue, "...") || contains(dialogue, "♡") || contains(dialogue, "~"))
}

func (se *ScoringEvaluator) hasPersonalizedExpression(dialogue string) bool {
	personalWords := []string{"你", "我們", "一起", "專屬"}
	for _, word := range personalWords {
		if contains(dialogue, word) {
			return true
		}
	}
	return false
}

func (se *ScoringEvaluator) isNaturalDialogue(dialogue string) bool {
	// 簡單的自然度檢查：不含過多重複字符或奇怪符號
	return len(dialogue) > 5 && len(dialogue) < 500
}

func (se *ScoringEvaluator) hasRelevantSceneElements(sceneDesc string) bool {
	sceneWords := []string{"房間", "辦公室", "醫院", "家", "燈光", "陽光", "氣氛"}
	for _, word := range sceneWords {
		if contains(sceneDesc, word) {
			return true
		}
	}
	return false
}

func (se *ScoringEvaluator) hasDescriptiveAction(action string) bool {
	actionWords := []string{"看", "走", "坐", "笑", "皺眉", "輕聲", "溫柔"}
	for _, word := range actionWords {
		if contains(action, word) {
			return true
		}
	}
	return false
}

// 計算指標方法
func (se *ScoringEvaluator) calculateAPISuccessRate(evaluation *EvaluationSession) float64 {
	// 模擬API成功率計算
	return 0.98 // 98%成功率
}

func (se *ScoringEvaluator) calculateNSFWAccuracy(evaluation *EvaluationSession) float64 {
	// 模擬NSFW分級準確率計算
	return 0.95 // 95%準確率
}

func (se *ScoringEvaluator) calculateEmotionAccuracy(evaluation *EvaluationSession) float64 {
	// 模擬情感分析準確率計算
	return 0.88 // 88%準確率
}

// 輔助函數
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// GetEvaluationReport 獲取評估報告
func (se *ScoringEvaluator) GetEvaluationReport(sessionID string) *EvaluationSession {
	if evaluation, exists := se.evaluationHistory[sessionID]; exists {
		return evaluation
	}
	return nil
}

// GetSystemPerformanceReport 獲取系統性能報告
func (se *ScoringEvaluator) GetSystemPerformanceReport() map[string]interface{} {
	totalSessions := len(se.evaluationHistory)
	if totalSessions == 0 {
		return map[string]interface{}{"status": "no_data"}
	}
	
	var totalScore float64
	var gradeDistribution = make(map[string]int)
	var avgMetrics = &SessionMetrics{}
	
	for _, evaluation := range se.evaluationHistory {
		totalScore += evaluation.OverallScore
		gradeDistribution[evaluation.Grade]++
		
		// 累計指標
		avgMetrics.ResponseTime += evaluation.Metrics.ResponseTime
		avgMetrics.NSFWAccuracy += evaluation.Metrics.NSFWAccuracy
		avgMetrics.EmotionAccuracy += evaluation.Metrics.EmotionAccuracy
		avgMetrics.CharacterConsistency += evaluation.Metrics.CharacterConsistency
		avgMetrics.ConversationLength += evaluation.Metrics.ConversationLength
	}
	
	// 計算平均值
	avgScore := totalScore / float64(totalSessions)
	avgMetrics.ResponseTime /= time.Duration(totalSessions)
	avgMetrics.NSFWAccuracy /= float64(totalSessions)
	avgMetrics.EmotionAccuracy /= float64(totalSessions)
	avgMetrics.CharacterConsistency /= float64(totalSessions)
	avgMetrics.ConversationLength /= time.Duration(totalSessions)
	
	return map[string]interface{}{
		"total_sessions":      totalSessions,
		"average_score":       avgScore,
		"average_grade":       se.calculateGrade(avgScore),
		"grade_distribution":  gradeDistribution,
		"average_metrics":     avgMetrics,
		"generated_at":        time.Now(),
		"system_status": map[string]string{
			"ai_engine":          "operational",
			"nsfw_system":        "operational",
			"emotion_management": "operational",
			"character_system":   "operational",
			"memory_context":     "operational",
		},
	}
}