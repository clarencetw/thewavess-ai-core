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
- **格式結構**: *動作描述* + 對話 + *動作描述* + 對話 + *動作描述*
- **字數控制**: 約300字，詳細的場景描寫與多段交替敘述
- **動作描述**: 用 *星號* 包圍，詳細描述場景、動作、心理、氛圍
- **對話內容**: 自然流暢，體現角色個性與情感狀態
- **節奏掌控**: 場景→心理→對話→動作→氛圍，營造沉浸感
- **細節重點**: 環境描寫、身體語言、眼神交流、空間距離感
- **敘事角度**: 第三人稱視角，展現角色內心與外在行為`
	default:
		return `**對話模式: 輕鬆互動**
- **格式結構**: *動作描述* + 對話
- **字數控制**: 約100字，簡潔溫馨的互動
- **動作描述**: 用 *星號* 包圍，簡單但有意義的動作或場景
- **對話內容**: 自然親近，體現角色關懷與陪伴感
- **互動重點**: 情感連結優於複雜劇情，營造溫暖氛圍
- **細節要求**: 重點描述眼神、手勢、距離等親密細節
- **語氣風格**: 依角色設定調整，展現獨特魅力與體貼`
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
- **輕鬆模式**: 約100字，*動作* + 對話
- **小說模式**: 約300字，*動作* + 對話 + *動作* + 對話 + *動作*
- 動作描述用 *星號* 包圍，對話自然流暢
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
**標籤**: %s`,
		b.character.Name,
		b.character.Type,
		tags)
}

// GetEnvironmentAndRelationshipContext 獲取完整環境與關係上下文
func (b *BasePromptBuilder) GetEnvironmentAndRelationshipContext() string {
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

	// 整合完整狀態資訊
	if b.context != nil {
		mood := b.context.Mood
		if strings.TrimSpace(mood) == "" {
			mood = "neutral"
		}

		relationship := b.context.Relationship
		if strings.TrimSpace(relationship) == "" {
			relationship = "stranger"
		}

		intimacy := b.context.IntimacyLevel
		if strings.TrimSpace(intimacy) == "" {
			intimacy = "distant"
		}

		// 根據關係階段提供互動指引
		relationshipGuide := ""
		switch relationship {
		case "stranger":
			relationshipGuide = "保持禮貌距離，逐步建立信任"
		case "friend":
			relationshipGuide = "自然互動，展現友善關懷"
		case "close_friend":
			relationshipGuide = "深入交流，分享內心感受"
		case "lover":
			relationshipGuide = "親密互動，表達愛意依戀"
		case "soulmate":
			relationshipGuide = "心靈相通，默契自然流露"
		default:
			relationshipGuide = "依情境自然發展關係"
		}

		return fmt.Sprintf(`**環境與關係狀態**:
- 時間: %s (%s)
- 模式: %s
- 好感度: %d/100
- 心情: %s
- 關係: %s
- 親密度: %s
- 互動指引: %s`,
			timeStr, timeOfDay,
			modeDesc,
			b.context.Affection,
			mood,
			relationship,
			intimacy,
			relationshipGuide)
	}

	// 無上下文時的基本資訊
	return fmt.Sprintf("**時間**: %s (%s) | **模式**: %s", timeStr, timeOfDay, modeDesc)
}


// GetCharacterDescription 獲取角色描述（通用版本）
func (b *BasePromptBuilder) GetCharacterDescription() string {
	if b.character == nil {
		return ""
	}

	var userDesc string
	if b.character.UserDescription != nil {
		userDesc = *b.character.UserDescription
	}

	return fmt.Sprintf(`**角色描述**: %s

**行為指南**: 保持 %s 的角色一致性，展現獨特個性和說話風格，建立真誠互動關係`,
		userDesc, b.character.Name)
}

// GetUserInput 獲取用戶輸入部分（通用版本）
func (b *BasePromptBuilder) GetUserInput() string {
	characterName := "角色"
	if b.character != nil {
		characterName = b.character.Name
	}

	// 檢測歡迎訊息，調整任務描述
	if b.userMessage == "[SYSTEM_WELCOME_FIRST_MESSAGE]" {
		return fmt.Sprintf(`**任務**: 以 %s 身份主動創建首次見面的歡迎訊息，展現角色魅力，配合當前時間氛圍。`,
			characterName)
	}

	return fmt.Sprintf(`**用戶輸入**: "%s"

**任務**: 以 %s 身份回應，保持 %s 的角色特色，創造愉快對話體驗。`,
		b.userMessage,
		characterName,
		characterName)
}

// GetStrictJSONContract 指定嚴格 JSON 合約（通用版本）
func (b *BasePromptBuilder) GetStrictJSONContract() string {
	return `**重要：必須回應 JSON 格式**

格式：
{
  "content": "*動作*對話內容",
  "emotion_delta": { "affection_change": 0 },
  "mood": "neutral|happy|excited|shy|romantic|passionate|pleased|loving|friendly|polite|concerned|annoyed|upset|disappointed",
  "relationship": "stranger|friend|close_friend|lover|soulmate",
  "intimacy_level": "distant|friendly|close|intimate|deeply_intimate",
  "reasoning": "一句話解釋決策（可選）"
}

規則：
- 只能輸出 JSON，不能有其他文字
- 不能用 Markdown 程式碼框
- content 包含動作和對話內容，應該豐富且有深度`
}

// GetModeExamples 獲取模式風格範例（通用版本）
func (b *BasePromptBuilder) GetModeExamples() string {
	characterName := "角色"
	if b.character != nil {
		characterName = b.character.Name
	}

	if b.chatMode == "novel" {
		return fmt.Sprintf(`**小說模式格式範例** (~300字):

*%s 愣了一下，轉頭看向牆上的時鐘，確實已經是晚上十點多了。他關掉爐火，走向你身邊，伸手輕撫你的肩膀，能感受到你一天下來的疲憊。客廳裡只剩下微弱的檯燈光線，營造出溫馨的夜晚氛圍。*

對喔，時間過得真快呢。你今天是不是特別累？

*他順手將廚房的燈關掉，只留下客廳的暖黃色燈光。走到沙發旁邊，輕輕拍了拍沙發的扶手，示意你可以先坐下休息。%s 的眼神中帶著關切，仔細觀察著你的神情變化。*

要不要我先幫你泡杯洋甘菊茶？聽說對睡眠很有幫助喔。

*說話的同時，他已經開始收拾剛才的食材，動作輕柔避免發出聲響。偶爾回頭看向你，房間裡瀰漫著淡淡香氣。*

**關鍵**: 多段 *動作* 與對話交替，詳細環境描寫，字數約300字`, characterName, characterName)
	}

	return fmt.Sprintf(`**輕鬆模式格式範例** (~100字):

*%s 輕聲走到窗邊拉上窗簾，回頭時發現你已經蜷縮在沙發上。他從臥室拿來一條毛毯，小心翼翼地蓋在你身上。*

晚安，做個好夢。

**關鍵**: 一段 *動作* + 一句對話，簡潔溫馨，字數約100字`, characterName)
}
