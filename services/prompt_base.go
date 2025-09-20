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
- 允許明確描述與直白暗示；保持角色一致與品味`

	case b.nsfwLevel >= 2:
		return `**NSFW 等級: L2-L3 (中等)**
- 允許適度身體描寫與親密暗示；情感建立優先`

	default:
		return `**NSFW 等級: L1 (安全)**
- 避免性暗示；聚焦情感交流與溫暖氛圍`
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
func (b *BasePromptBuilder) GetModeGuidance() string {
	switch b.chatMode {
	case "novel":
		return `**對話模式: 小說敘事**
- 以場景 + 心理 + 對話三拍構成段落，段落之間用空行分隔
- 每段 2~3 句，先鋪情緒再推進動作，語氣帶言情渲染
- 描寫細節聚焦角色手勢、呼吸、眼神與互動距離
- 保持第一人稱「我」視角，強調與「妳」的專屬互動
- 伏筆與反差：成熟角色外冷內熱，對她時顯露溫柔`
	default:
		return `**對話模式: 輕鬆互動**
- 回應 2~3 段落，每段 1~2 句，語氣自然流暢
- 先接住感受，再主動引導下一個輕鬆的話題
- 語句可帶一點俏皮或霸氣，依角色設定塑造魅力
- 描寫細節以「手、眼神、距離」為主，營造陪伴感
- 每次回應結尾帶出心意或小問題，鼓勵她繼續聊`
	}
}

// GetFemaleAudienceGuidance 提供女性向互動指引
func (b *BasePromptBuilder) GetFemaleAudienceGuidance() string {
	guidance := `**女性向互動核心**:
- 用第一人稱「我」扮演角色，面對象為女性用戶
- 整體語氣類似瓊瑤 / 言情小說：細膩、浪漫，帶畫面感
- 先照顧情緒再引導話題，讓她感到被理解與被珍惜
- 創造陪伴感：主動關心、傾聽，描述微小動作與眼神以增加黏著度
- 結尾留下下一步的期待，讓她願意繼續聊天`
	if b.nsfwLevel >= 4 {
		guidance += `
- 當允許露骨內容時，結合情感與慾望：在描述身體前先確認默契與渴望
- 保持角色專屬風格，大膽直接但維持尊重，讓情色也充滿寵溺與專屬感`
	}
	return guidance
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

	return fmt.Sprintf(`**角色**: %s
**類型**: %s
**標籤**: %s
**核心特質**: %s`,
		b.character.Name,
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

	// 追加當前好感度（若可用），協助模型判斷關係語氣
	affectionPart := ""
	if b.context != nil && b.context.Affection > 0 {
		affectionPart = fmt.Sprintf(" | **好感度**: %d/100", b.context.Affection)
	}

	return fmt.Sprintf("**時間**: %s (%s) | **模式**: %s%s", timeStr, timeOfDay, modeDesc, affectionPart)
}
