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

	// 4. 高NSFW等級 -> Grok + 設置sticky
	if nsfwLevel >= 3 {
		es.chatService.markNSFWSticky(chatID)
		return "grok"
	}

	// 5. 關係因素：親密關係 + 中等NSFW -> Grok
	if conversationContext != nil && nsfwLevel >= 2 {
		if es.isIntimateRelationship(conversationContext) {
			return "grok"
		}
	}

	// 6. 智能檢測：疑似隱晦表達但未被關鍵字捕獲
	if nsfwLevel == 1 && es.conversationClassifier.IsPotentialImplicitContent(msg) {
		utils.Logger.Info("檢測到疑似隱晦內容，使用Grok處理")
		return "grok"
	}

	// 7. 默認安全內容 -> OpenAI
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

// BuildCleanHistoryForOpenAI 為OpenAI構建乾淨的歷史（解決歷史污染）
func (es *EngineSelector) BuildCleanHistoryForOpenAI(
	ctx context.Context,
	chatID string,
	limit int,
) ([]ChatMessage, error) {

	// 獲取完整歷史
	fullHistory, err := es.chatService.getRecentMemoriesFromDB(ctx, chatID, limit)
	if err != nil {
		return []ChatMessage{}, nil
	}

	// 過濾NSFW內容，保留安全對話
	cleanHistory := []ChatMessage{}
	for _, msg := range fullHistory {
		// 簡單過濾：不包含NSFW關鍵字的消息
		if es.IsSafeMessage(msg.Content) {
			cleanHistory = append(cleanHistory, msg)
		}
	}

	return cleanHistory, nil
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
