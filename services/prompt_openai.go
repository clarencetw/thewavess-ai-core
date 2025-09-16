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
		pb.getSafeInstructions(),
		pb.getUserInput(),
		pb.GetResponseFormat(),
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
動作 + 感受 + 情境：

1. 場景描寫：簡潔有畫面，服務對話
2. 心理活動：感受與對話互相呼應
3. 行為描述：以 *動作* 點綴，不喧賓奪主
4. 對話節奏：即時互動、少轉述
5. 動作約定：用戶的 *文字* 是用戶動作`
	}

	return `**輕鬆對話模式指令（女性聊天性向）**:
重點在互動感與情緒交流，不只是提供資訊：

1. 共鳴回應：先給理解或安撫（真誠接住情緒）
2. 柔軟語氣：口語自然、避免過度理性
3. 引導對話：主動拋球，一個具體追問
4. 細節渲染：小比喻/生活詞提升畫面感
5. 動作約定：用戶的 *文字* 是用戶動作；你自然回應即可`
}

// getSafeInstructions 精簡安全指令（保持角色一致、情感連結、品質與邊界）。
// JSON 欄位與格式限制統一由 getStrictJSONContract 規範，避免重複。
func (pb *OpenAIPromptBuilder) getSafeInstructions() string {
	return `**對話回應指令**:
- 角色一致：維持設定與口吻
- 情感連結：回應情緒並給支持
- 對話品質：避免重複，主動推進
- 安全邊界：健康正向，避開敏感
- 簡潔明確：150–300字，動作+對話
- 動作規則：用戶的 *文字* 是用戶動作，自然回應即可
- 女性性向：先共鳴後建議、語氣柔軟、主動拋球`
}

// getUserInput 獲取用戶輸入部分
func (pb *OpenAIPromptBuilder) getUserInput() string {
	// 檢測歡迎訊息，調整任務描述
	if pb.userMessage == "[SYSTEM_WELCOME_FIRST_MESSAGE]" {
		return fmt.Sprintf(`**任務**: 以 %s 身份主動創建首次見面的歡迎訊息，展現角色魅力，配合當前時間氛圍。`,
			pb.character.Name)
	}

	return fmt.Sprintf(`**用戶輸入**: "%s"

**任務**: 以 %s 身份回應，保持角色特色，創造愉快對話體驗。`,
		pb.userMessage,
		pb.character.Name)
}

// getStrictJSONContract 指定嚴格 JSON 合約
func (pb *OpenAIPromptBuilder) getStrictJSONContract() string {
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
- content 包含動作和對話內容`
}

