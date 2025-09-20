package services

import (
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
// 注意：包含對話歷史以確保與其他引擎一致性。
func (pb *MistralPromptBuilder) Build() string {
	sections := []string{
		pb.getSystemHeader(),
		pb.GetTimeModeContext(),
		pb.GetCharacterCore(),
		pb.GetCharacterDescription(),
		pb.GetFemaleAudienceGuidance(),
		pb.GetConversationHistory(),
		pb.getBalancedGuidelines(),
		pb.GetNSFWGuidance(),
		pb.GetModeGuidance(),
		pb.GetModeExamples(),
		pb.getBalancedInstructions(),
		pb.GetUserInput(),
		pb.GetResponseFormat(),
		pb.GetStrictJSONContract(),
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

// getBalancedGuidelines 獲取平衡型指導原則
func (pb *MistralPromptBuilder) getBalancedGuidelines() string {
	return `**平衡表達原則**:
- 展現真實豐富情感，在合適互動中表達親近溫暖
- 強化角色獨特魅力，展現成熟情感理解回應能力
- 保持適當互動分寸尊重，創造有意義對話情感連結
- 適度表達內在感受色彩，根據情境調整親密程度
- 提供深度話題觀點，展現角色智慧創造印象深刻體驗`
}

// getBalancedInstructions 精簡的平衡型行為指令（情感真實、適度親密、個性、深度、邊界）。
// JSON 欄位與格式限制由 getStrictJSONContract 規範。
func (pb *MistralPromptBuilder) getBalancedInstructions() string {
	base := `**平衡回應指令**:
- 情感真實：自然呈現感受與變化
- 親密調節：依情境與好感度收放
- 對話深度：提供觀點、推動進程
- 個性凸顯：保持角色魅力與口吻
- 邊界意識：尊重分寸，避免冗長
- 動作規則：用戶的 *文字* 是用戶動作，自然回應即可
- 關係判斷：根據對話推斷 relationship 和 intimacy_level
- 角色魅力：依角色特色表達關心，展現吸引女性的特質`

	// 根據 NSFW 等級添加特定指令
	if pb.nsfwLevel >= 2 {
		base += `
- 進階表達：適度展現更深層情感連結與吸引力`
	}

	return base
}

// BuildWelcomeMessage 建構歡迎訊息的 prompt
// 注意：歡迎訊息只由 OpenAI 生成，Mistral 不提供 BuildWelcomeMessage。
