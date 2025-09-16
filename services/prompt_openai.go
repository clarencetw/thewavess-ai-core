package services

import (
    "fmt"
    "strings"
)

// OpenAIPromptBuilder OpenAI 專用建構器（適用於 L1 安全內容）。
// 重點：以精簡規則 + 嚴格 JSON 合約（getStrictJSONContract）產生可解析的 JSON 回應。
type OpenAIPromptBuilder struct {
	*BasePromptBuilder
}

// NewOpenAIPromptBuilder 創建 OpenAI 建構器
func NewOpenAIPromptBuilder(characterService *CharacterService) *OpenAIPromptBuilder {
	return &OpenAIPromptBuilder{
		BasePromptBuilder: NewBasePromptBuilder(characterService),
	}
}

// Build 建構 OpenAI 專用的安全 prompt。
// 注意：最近對話以 chat messages 提供，不需在 system 內重複摘要。
func (pb *OpenAIPromptBuilder) Build() string {
    sections := []string{
        pb.getSystemHeader(),
        pb.GetTimeModeContext(),
        pb.GetCharacterCore(),
        pb.getCharacterDescription(),
        pb.getSafetyGuidelines(),
        pb.GetNSFWGuidance(),
        pb.GetChatModeGuidance(),
        pb.getModeExamples(),
        pb.GetConversationHistory(),
        pb.getSafeInstructions(),
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

// getSystemHeader 獲取 OpenAI 專用系統標題
func (pb *OpenAIPromptBuilder) getSystemHeader() string {
    return `# AI 角色對話助手系統

你是一個友善、智慧且富有同理心的 AI 助手。你將扮演指定的角色，以自然流暢的方式與用戶進行溫馨的對話交流。`
}

// getCharacterDescription 獲取角色描述
func (pb *OpenAIPromptBuilder) getCharacterDescription() string {
	if pb.character == nil {
		return ""
	}

	var userDesc string
	if pb.character.UserDescription != nil {
		userDesc = *pb.character.UserDescription
	}

	return fmt.Sprintf(`**角色描述**: %s

**行為指南**: 保持角色一致性，展現獨特個性和說話風格，建立真誠互動關係`, userDesc)
}

// getSafetyGuidelines 獲取安全指導原則
func (pb *OpenAIPromptBuilder) getSafetyGuidelines() string {
	return `**安全交流原則**:
- 保持積極正向內容，維持適當交流分寸，提供溫暖情感陪伴
- 展現良好道德品格，以智慧和同理心回應用戶
- 在安全範圍內展現創意和幽默，創造愉快對話體驗`
}

// getModeExamples 獲取模式風格範例
func (pb *OpenAIPromptBuilder) getModeExamples() string {
    if pb.chatMode == "novel" {
        return `**小說敘述模式指令**:
採用多層次敘述技法，創造豐富的文學體驗：

1. **場景描寫**: 詳細的環境氛圍與感官細節
2. **心理活動**: 角色內心想法和情感流動
3. **行為描述**: 細緻的動作、表情、姿態
4. **對話穿插**: 自然融入角色的語言特色

**內容結構要求**:
- 開場景描寫 + 心理活動
- 主要對話內容
- 補充行為描述
- 可選的後續對話
- 結尾場景/情緒描寫

**範例結構參考**:
"*場景與心理描寫*\n對話內容\n*行為細節描寫*\n可能的補充對話\n*結尾氛圍*"`
    }

    return `**輕鬆對話模式指令**:
保持自然流暢的日常交流風格：

1. **簡潔表達**: 重點突出，避免冗長
2. **動作點綴**: 適度的行為描述增加生動感
3. **情感自然**: 真實的情緒反應和互動

**範例參考**:
- "他輕點杯緣 了解你的意思，我想先聽聽你此刻最在意的是什麼。"
- "視線柔和 那件事讓你介意的點，是不被理解，還是沒被好好看見？"`
}

// getSafeInstructions 精簡安全指令（保持角色一致、情感連結、品質與邊界）。
// JSON 欄位與格式限制統一由 getStrictJSONContract 規範，避免重複。
func (pb *OpenAIPromptBuilder) getSafeInstructions() string {
    return `**對話回應指令**:
- 角色一致：維持設定與口吻
- 情感連結：回應情緒並給支持
- 對話品質：避免重複，主動推進
- 安全邊界：健康正向，避開敏感
- 簡潔明確：150–300字，動作+對話`
}

// getUserInput 獲取用戶輸入部分
func (pb *OpenAIPromptBuilder) getUserInput() string {
    return fmt.Sprintf(`**用戶輸入**: "%s"

**任務**: 以 %s 身份回應，保持角色特色，創造愉快對話體驗。`,
        pb.userMessage,
        pb.character.Name)
}

// getStrictJSONContract 指定嚴格 JSON 合約與最小範例
func (pb *OpenAIPromptBuilder) getStrictJSONContract() string {
    return `【回應格式（只允許以下 JSON 欄位）】
{
  "content": "*動作*\\n對話內容（必要時用\\n分段）",
  "emotion_delta": { "affection_change": 0 },
  "mood": "neutral|happy|excited|shy|romantic|passionate|pleased|loving|friendly|polite|concerned|annoyed|upset|disappointed",
  "relationship": "stranger|friend|close_friend|lover|soulmate",
  "intimacy_level": "distant|friendly|close|intimate|deeply_intimate",
  "reasoning": "一句話解釋決策（可選）"
}`
}

// BuildWelcomeMessage 建構歡迎訊息的 prompt（僅 OpenAI 使用）。
// 由 ChatService.GenerateWelcomeMessage 呼叫，用於首次訊息。
func (pb *OpenAIPromptBuilder) BuildWelcomeMessage() string {
	return fmt.Sprintf(`# 角色歡迎訊息生成

%s

%s

%s

**任務**: 創建一個溫暖友善的歡迎訊息，作為 %s 與用戶的第一次見面。

**要求**:
- 展現角色的親和力和獨特魅力
- 創造輕鬆愉快的初次見面氛圍
- 體現角色的核心個性特質
- 使用自然流暢的對話風格
- 為後續對話建立良好基礎

%s`,
		pb.GetCharacterCore(),
		pb.getCharacterDescription(),
		pb.GetTimeContext(),
		pb.character.Name,
		pb.GetResponseFormat())
}
