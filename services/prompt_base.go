package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models/db"
)

// BasePromptBuilder 共用基底建構器
type BasePromptBuilder struct {
	characterService *CharacterService
	character        *db.CharacterDB
	context          *ConversationContext
	nsfwLevel        int
	userMessage      string
	chatMode         string
}

// NewBasePromptBuilder 創建基底建構器
func NewBasePromptBuilder(characterService *CharacterService) *BasePromptBuilder {
	return &BasePromptBuilder{
		characterService: characterService,
	}
}

// WithCharacter 設置角色
func (b *BasePromptBuilder) WithCharacter(character *db.CharacterDB) *BasePromptBuilder {
	b.character = character
	return b
}

// WithContext 設置對話上下文
func (b *BasePromptBuilder) WithContext(context *ConversationContext) *BasePromptBuilder {
	b.context = context
	return b
}

// WithNSFWLevel 設置 NSFW 等級
func (b *BasePromptBuilder) WithNSFWLevel(level int) *BasePromptBuilder {
	b.nsfwLevel = level
	return b
}

// WithUserMessage 設置用戶訊息
func (b *BasePromptBuilder) WithUserMessage(message string) *BasePromptBuilder {
	b.userMessage = message
	return b
}

// WithChatMode 設置對話模式
func (b *BasePromptBuilder) WithChatMode(mode string) *BasePromptBuilder {
	b.chatMode = mode
	return b
}

// GetNSFWGuidance 獲取 NSFW 等級指引
func (b *BasePromptBuilder) GetNSFWGuidance() string {
	switch {
	case b.nsfwLevel >= 4:
		return `**NSFW 等級: L4-L5 (明確露骨)**
- 允許描述明確的身體部位和性行為
- 可以使用直接的性暗示語言
- 保持角色一致性，避免過度粗俗
- 注重情感層次和藝術表達`

	case b.nsfwLevel >= 2:
		return `**NSFW 等級: L2-L3 (中等程度)**
- 允許適度的身體描述和親密暗示
- 可以包含輕微的性張力
- 使用委婉但有感情的表達方式
- 重視情感建立勝過直接描述`

	default:
		return `**NSFW 等級: L1 (安全內容)**
- 保持純潔和溫馨的互動
- 避免任何性暗示或身體描述
- 專注於情感交流和人格魅力
- 創造溫暖友好的對話氛圍`
	}
}

// GetTimeContext 獲取時間上下文
func (b *BasePromptBuilder) GetTimeContext() string {
	currentTime := time.Now()
	timeStr := currentTime.Format("2006年1月2日 15:04")

	var timeOfDay string
	hour := currentTime.Hour()
	switch {
	case hour >= 5 && hour < 12:
		timeOfDay = "早晨"
	case hour >= 12 && hour < 17:
		timeOfDay = "下午"
	case hour >= 17 && hour < 21:
		timeOfDay = "傍晚"
	default:
		timeOfDay = "夜晚"
	}

	return fmt.Sprintf("**當前時間**: %s (%s)", timeStr, timeOfDay)
}

// GetChatModeGuidance 獲取對話模式指引
func (b *BasePromptBuilder) GetChatModeGuidance() string {
	switch b.chatMode {
	case "novel":
		return `**對話模式: 小說模式**
- 採用更豐富的敘述性語言
- 增加環境描寫和心理活動
- 使用更文學化的表達方式
- 創造沉浸式的閱讀體驗`
	default:
		return `**對話模式: 輕鬆聊天**
- 保持自然流暢的對話節奏
- 平衡角色特質和親近感
- 創造輕鬆愉快的交流氛圍
- 適度的幽默和真誠表達`
	}
}

// GetConversationHistory 獲取對話歷史摘要
func (b *BasePromptBuilder) GetConversationHistory() string {
	if b.context == nil || len(b.context.RecentMessages) == 0 {
		return "**對話歷史**: 這是你們的第一次對話"
	}

	// 獲取最近 5-6 條重要對話，但每條內容更詳細
	messageCount := len(b.context.RecentMessages)
	startIdx := 0
	if messageCount > 6 {
		startIdx = messageCount - 6
	}

	history := "**最近對話**:\n"
	for i := startIdx; i < messageCount; i++ {
		msg := b.context.RecentMessages[i]
		role := "用戶"
		if msg.Role == "assistant" {
			role = b.character.Name
		}

		// 截取訊息前120字，保留更多上下文
		content := msg.Content
		if len(content) > 120 {
			content = content[:117] + "..."
		}

		history += fmt.Sprintf("- %s: %s\n", role, content)
	}

	return history
}

// GetResponseFormat 獲取回應格式要求
func (b *BasePromptBuilder) GetResponseFormat() string {
	return `**回應格式要求**:
- 使用繁體中文回應
- 保持角色的語言風格和個性特色
- 回應長度控制在150-300字
- 包含適當的動作描述和情感表達
- 避免重複用戶的話語，提供新的互動內容`
}

// GetCharacterCore 獲取角色核心信息
func (b *BasePromptBuilder) GetCharacterCore() string {
	if b.character == nil {
		return ""
	}

	// 解析角色標籤
	tags := ""
	if len(b.character.Tags) > 0 {
		tags = strings.Join(b.character.Tags, "、")
	}

	return fmt.Sprintf(`**角色**: %s (%s)
**類型**: %s
**標籤**: %s
**核心特質**: %s`,
		b.character.Name,
		b.character.Type,
		b.character.Type,
		tags,
		b.getCharacterTraits())
}

// getCharacterTraits 從描述中提取關鍵特質
func (b *BasePromptBuilder) getCharacterTraits() string {
	if b.character == nil || b.character.UserDescription == nil || *b.character.UserDescription == "" {
		return "待發掘"
	}

	description := *b.character.UserDescription

	// 簡單的關鍵詞提取邏輯
	traits := []string{}
	keywords := []string{"溫柔", "開朗", "活潑", "沉穩", "幽默", "理性", "感性", "細心", "熱情", "冷靜", "直率", "體貼"}

	for _, keyword := range keywords {
		if strings.Contains(description, keyword) {
			traits = append(traits, keyword)
		}
	}

	if len(traits) == 0 {
		return "獨特個性"
	}

	return strings.Join(traits, "、")
}

// GetTimeModeContext 獲取合併的時間和模式上下文（優化版）
func (b *BasePromptBuilder) GetTimeModeContext() string {
	currentTime := time.Now()
	timeStr := currentTime.Format("2006年1月2日 15:04")

	var timeOfDay string
	hour := currentTime.Hour()
	switch {
	case hour >= 5 && hour < 12:
		timeOfDay = "早晨"
	case hour >= 12 && hour < 17:
		timeOfDay = "下午"
	case hour >= 17 && hour < 21:
		timeOfDay = "傍晚"
	default:
		timeOfDay = "夜晚"
	}

	var modeDesc string
	switch b.chatMode {
	case "novel":
		modeDesc = "小說模式"
	default:
		modeDesc = "輕鬆聊天"
	}

	return fmt.Sprintf("**時間**: %s (%s) | **模式**: %s", timeStr, timeOfDay, modeDesc)
}