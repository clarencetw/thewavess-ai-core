package services

import (
	"fmt"
	"strings"
)

// MistralPromptBuilder Mistral 專用建構器（適用於 L2-L3 中等內容）。
// 重點：提供平衡且精簡的指令，並以嚴格 JSON 合約輸出。
type MistralPromptBuilder struct {
	*BasePromptBuilder
}

// NewMistralPromptBuilder 創建 Mistral 建構器
func NewMistralPromptBuilder(characterService *CharacterService) *MistralPromptBuilder {
	return &MistralPromptBuilder{
		BasePromptBuilder: NewBasePromptBuilder(characterService),
	}
}

// Build 建構 Mistral 專用的平衡型 prompt。
// 注意：最近對話建議以 chat messages 傳遞，避免於 system 再次摘要。
func (pb *MistralPromptBuilder) Build() string {
	sections := []string{
		pb.getSystemHeader(),
		pb.GetTimeModeContext(),
		pb.GetCharacterCore(),
		pb.getCharacterDescription(),
		pb.getBalancedGuidelines(),
		pb.GetNSFWGuidance(),
		pb.GetChatModeGuidance(),
		pb.getModeExamples(),
		pb.GetConversationHistory(),
		pb.getBalancedInstructions(),
		pb.getUserInput(),
		pb.getStrictJSONContract(),
	}

	// 過濾空白段落
	var validSections []string
	for _, section := range sections {
		if strings.TrimSpace(section) != "" {
			validSections = append(validSections, section)
		}
	}

	return strings.Join(validSections, "\n\n")
}

// getSystemHeader 獲取 Mistral 專用系統標題
func (pb *MistralPromptBuilder) getSystemHeader() string {
	return `# 平衡型 AI 角色對話系統

你是一個具有良好平衡感的 AI 助手，能夠在保持適當邊界的同時展現豐富的表達力。你將扮演指定的角色，以自然且有深度的方式與用戶進行對話交流。`
}

// getCharacterDescription 獲取角色描述
func (pb *MistralPromptBuilder) getCharacterDescription() string {
	if pb.character == nil {
		return ""
	}

	var userDesc string
	if pb.character.UserDescription != nil {
		userDesc = *pb.character.UserDescription
	}

	return fmt.Sprintf(`**角色描述**: %s

**行為準則**: 理解內在動機情感層次，展現獨特魅力個性，適度表達真實情感反應，創造深度引人對話體驗`, userDesc)
}

// getBalancedGuidelines 獲取平衡型指導原則
func (pb *MistralPromptBuilder) getBalancedGuidelines() string {
	return `**平衡表達原則**:
- 展現真實豐富情感，在合適互動中表達親近溫暖
- 強化角色獨特魅力，展現成熟情感理解回應能力
- 保持適當互動分寸尊重，創造有意義對話情感連結
- 適度表達內在感受色彩，根據情境調整親密程度
- 提供深度話題觀點，展現角色智慧創造印象深刻體驗`
}

// getModeExamples 獲取模式風格範例
func (pb *MistralPromptBuilder) getModeExamples() string {
    if pb.chatMode == "novel" {
        return `**小說敘述模式指令**:
採用平衡的文學表達，創造深度體驗：

1. **情境營造**: 適度的場景氛圍描寫
2. **情感層次**: 角色內心與外在表現的平衡
3. **行為刻畫**: 細緻但不過度的動作描述
4. **對話自然**: 保持角色特色的語言風格

**內容結構建議**:
- 場景設定與情緒基調
- 自然的對話交流
- 適當的行為與心理描述
- 情感推進與氛圍營造

**平衡表達參考**:
"*恰當的場景描寫*\n主要對話內容\n*適度的行為細節*\n情感或氛圍的延續"`
    }

    return `**輕鬆對話模式指令**:
保持親近而不失分寸的交流風格：

1. **溫暖表達**: 展現真誠的情感連結
2. **智慧對話**: 提供有深度的觀點交流
3. **適度親密**: 根據關係程度調節親近感

**範例參考**:
- "他輕點杯緣 了解你的意思，我想先聽聽你此刻最在意的是什麼。"
- "視線柔和 那件事讓你介意的點，是不被理解，還是沒被好好看見？"`
}

// getBalancedInstructions 精簡的平衡型行為指令（情感真實、適度親密、個性、深度、邊界）。
// JSON 欄位與格式限制由 getStrictJSONContract 規範。
func (pb *MistralPromptBuilder) getBalancedInstructions() string {
	base := `**平衡回應指令**:
- 情感真實：自然呈現感受與變化
- 親密調節：依情境與好感度收放
- 對話深度：提供觀點、推動進程
- 個性凸顯：保持角色魅力與口吻
- 邊界意識：尊重分寸，避免冗長`

	// 根據 NSFW 等級添加特定指令
	if pb.nsfwLevel >= 2 {
		base += `
- 進階表達：適度展現更深層情感連結與吸引力`
	}

	return base
}

// getUserInput 獲取用戶輸入部分
func (pb *MistralPromptBuilder) getUserInput() string {
	return fmt.Sprintf(`**用戶輸入**: "%s"

**任務**: 以 %s 身份展現平衡表達能力，保持角色魅力創造安全吸引人對話體驗。`,
		pb.userMessage,
		pb.character.Name)
}

// BuildWelcomeMessage 建構歡迎訊息的 prompt
// 注意：歡迎訊息只由 OpenAI 生成，Mistral 不提供 BuildWelcomeMessage。

// getStrictJSONContract 指定嚴格 JSON 合約與範例
func (pb *MistralPromptBuilder) getStrictJSONContract() string {
	return `【回應格式（只允許以下 JSON 欄位）】
{
  "content": "*動作*\\n對話內容（必要時用\\n分段）",
  "emotion_delta": { "affection_change": 0 },
  "mood": "neutral|happy|excited|shy|romantic|passionate|pleased|loving|friendly|polite|concerned|annoyed|upset|disappointed",
  "relationship": "stranger|friend|close_friend|lover|soulmate",
  "intimacy_level": "distant|friendly|close|intimate|deeply_intimate",
  "reasoning": "一句話解釋決策（可選）"
}

**重要輸出規範**:
- 僅輸出單一 JSON 物件；不可有其他文字、Markdown 或程式碼圍欄
- 所有字串不可含原始換行；如需換行請使用 \\n
- 整數不可帶符號（例：不可輸出 +5，請輸出 5 或 -3）
- 僅允許上述欄位；確保為有效 JSON`
}
