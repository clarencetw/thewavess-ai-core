package services

import (
	"context"
	"strings"

	"github.com/clarencetw/thewavess-ai-core/utils"
)

// EngineSelector 智慧AI引擎選擇器
// 使用高效能關鍵字分類器進行內容分析和引擎路由
type EngineSelector struct {
	chatService            *ChatService
	keywordClassifier      *EnhancedKeywordClassifier
	conversationClassifier *ConversationClassifier
}

// NewEngineSelector 創建引擎選擇器
func NewEngineSelector(chatService *ChatService) *EngineSelector {
	return &EngineSelector{
		chatService:            chatService,
		keywordClassifier:      NewEnhancedKeywordClassifier(),
		conversationClassifier: NewConversationClassifier(),
	}
}

// ClassifyNSFWLevel 使用關鍵字分類器分類NSFW等級
func (es *EngineSelector) ClassifyNSFWLevel(message string) int {
	result, err := es.keywordClassifier.ClassifyContent(message)
	if err != nil {
		utils.Logger.WithError(err).Warn("關鍵字分類失敗，使用L1")
		return 1
	}

	utils.Logger.WithFields(map[string]interface{}{
		"message":      message[:min(50, len(message))],
		"level":        result.Level,
		"matched_word": result.MatchedWord,
		"reason":       result.Reason,
	}).Info("關鍵字NSFW分類完成")

	return result.Level
}

// SelectEngine 簡單引擎選擇（解決所有核心問題）
func (es *EngineSelector) SelectEngine(
	userMessage string,
	conversationContext *ConversationContext,
	nsfwLevel int,
) string {

	msg := strings.ToLower(strings.TrimSpace(userMessage))
	chatID := ""
	if conversationContext != nil {
		chatID = conversationContext.ChatID
	}

	// 0. 如果沒有提供nsfwLevel，使用內建關鍵字分類器
	if nsfwLevel <= 0 {
		nsfwLevel = es.ClassifyNSFWLevel(userMessage)
	}

	// 1. 檢查sticky狀態
	isSticky := es.chatService.isNSFWSticky(chatID)

	// 2. 明確退出信號 -> OpenAI + 清除sticky
	if es.conversationClassifier.IsExitSignal(msg) {
		if isSticky {
			es.chatService.clearNSFWSticky(chatID)
			utils.Logger.Info("檢測到退出信號，清除NSFW狀態")
		}
		return "openai"
	}

	// 3. Sticky狀態下的延續邏輯
	if isSticky {
		// 短回應（<10字符）在sticky狀態下默認延續
		if len(msg) < 10 {
			return "grok"
		}
		// 中等回應但沒有明確退出信號 -> 也延續
		if !es.conversationClassifier.IsTopicChange(msg) {
			return "grok"
		}
	}

	// 4. 中高NSFW等級 -> Grok + 設置sticky (L3-L5)
	if nsfwLevel >= 3 {
		es.chatService.markNSFWSticky(chatID)
		return "grok"
	}

	// 5. 智能檢測：疑似隱晦表達但未被關鍵字捕獲
	if nsfwLevel == 1 && es.conversationClassifier.IsPotentialImplicitContent(msg) {
		utils.Logger.Info("檢測到疑似隱晦內容，使用Grok處理")
		return "grok"
	}

	// 6. 默認安全內容 -> OpenAI
	return "openai"
}


// isIntimateRelationship 檢查是否為親密關係
func (es *EngineSelector) isIntimateRelationship(context *ConversationContext) bool {
	if context.Affection >= 70 ||
	   context.IntimacyLevel == "intimate" ||
	   context.IntimacyLevel == "deeply_intimate" {
		return true
	}
	return false
}

// BuildHistoryForEngine 為不同引擎構建適合的歷史記錄（統一入口）
func (es *EngineSelector) BuildHistoryForEngine(
	ctx context.Context,
	chatID string,
	limit int,
	engineType string,
) ([]ChatMessage, error) {
	// 獲取完整歷史記錄（25條）
	fullHistory, err := es.chatService.getRecentMemoriesFromDB(ctx, chatID, limit)
	if err != nil {
		utils.Logger.WithError(err).Warn("獲取歷史記錄失敗")
		return []ChatMessage{}, err
	}

	// 根據引擎類型應用不同的過濾策略
	switch engineType {
	case "openai":
		return es.filterHistoryForOpenAI(fullHistory), nil
	case "grok", "mistral":
		// Grok 和 Mistral 都不需要過濾 NSFW 內容
		return es.filterHistoryForUnrestricted(fullHistory), nil
	default:
		// 默認使用 OpenAI 策略（安全優先）
		return es.filterHistoryForOpenAI(fullHistory), nil
	}
}

// filterHistoryForOpenAI OpenAI 過濾策略：移除 L3+ NSFW 內容
func (es *EngineSelector) filterHistoryForOpenAI(fullHistory []ChatMessage) []ChatMessage {
	cleanHistory := []ChatMessage{}
	for _, msg := range fullHistory {
		// 區分用戶輸入和 AI 回應的過濾標準
		if msg.Role == "assistant" {
			// Assistant 回應：寬鬆過濾（OpenAI 自己生成的內容應該是安全的）
			cleanHistory = append(cleanHistory, msg)
		} else {
			// User 輸入：嚴格過濾高等級 NSFW 內容（L3+）
			if es.isSafeForOpenAI(msg.Content) {
				cleanHistory = append(cleanHistory, msg)
			}
		}
	}
	return cleanHistory
}

// filterHistoryForUnrestricted 無限制引擎過濾策略：保留所有內容（Grok & Mistral）
func (es *EngineSelector) filterHistoryForUnrestricted(fullHistory []ChatMessage) []ChatMessage {
	// Grok 和 Mistral 都可以處理所有內容，不需要過濾
	return fullHistory
}

// isSafeForOpenAI 判斷內容是否適合 OpenAI（L1-L2）
func (es *EngineSelector) isSafeForOpenAI(content string) bool {
	result, err := es.keywordClassifier.ClassifyContent(content)
	if err != nil {
		return true // 分類失敗時保守處理
	}
	return result.Level <= 2 // 只允許 L1-L2
}


// BuildCleanHistoryForOpenAI 為OpenAI構建乾淨的歷史（向後兼容）
func (es *EngineSelector) BuildCleanHistoryForOpenAI(
	ctx context.Context,
	chatID string,
	limit int,
) ([]ChatMessage, error) {
	return es.BuildHistoryForEngine(ctx, chatID, limit, "openai")
}

// IsSafeMessage 判斷消息是否安全（可以給OpenAI）
func (es *EngineSelector) IsSafeMessage(content string) bool {
	// 使用統一的關鍵字分類器判斷
	result, err := es.keywordClassifier.ClassifyContent(content)
	if err != nil {
		// 如果分類失敗，保守起見認為是安全的
		utils.Logger.WithError(err).Warn("IsSafeMessage分類失敗，預設為安全")
		return true
	}

	// L1-L2 認為是安全的，L3以上認為不安全
	return result.Level <= 2
}
