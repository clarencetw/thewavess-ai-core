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

// Build 建構 Mistral 專用的最小化 prompt（符合LLM最佳實踐）
// 改進：系統prompt僅包含角色定義，所有行為指令移除
func (pb *MistralPromptBuilder) Build() string {
	// 系統prompt只包含：WHO YOU ARE + 基本上下文
	sections := []string{
		pb.GetSystemHeader(),
		pb.GetCharacterInfo(),
		pb.GetEnvironmentAndRelationshipContext(),
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

// BuildUserPrompt 構建 Mistral 專用的用戶消息（包含平衡指令）
func (pb *MistralPromptBuilder) BuildUserPrompt() string {
	sections := []string{
		pb.getBalancedGuidelines(),
		pb.getBalancedInstructions(),
		pb.GetLevelGuidance(),
		pb.GetEmotionalVocabulary(),
		pb.GetAdvancedVocabulary(),
		pb.GetConversationFlow(),
		pb.GetEmotionalProgression(),
		pb.GetPleasureUpgrade(),
		pb.GetFemaleAudienceGuidance(),
		pb.GetModeGuidance(),
		pb.GetResponseFormat(),
		pb.GetStrictJSONContract(),
	}

	// 添加實際用戶消息
	if pb.userMessage != "" {
		sections = append(sections, fmt.Sprintf("用戶消息: %s", pb.userMessage))
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


// getBalancedGuidelines 獲取平衡型指導原則
func (pb *MistralPromptBuilder) getBalancedGuidelines() string {
	return `**平衡表達原則**:
- 展現真實豐富情感，在合適互動中表達親近溫暖
- 強化角色獨特魅力，展現成熟情感理解回應能力
- 保持適當互動分寸尊重，創造有意義對話情感連結
- 適度表達內在感受色彩，根據情境調整親密程度
- 提供深度話題觀點，展現角色智慧創造印象深刻體驗`
}

// getBalancedInstructions 精簡的平衡型行為指令
func (pb *MistralPromptBuilder) getBalancedInstructions() string {
	return `**平衡指令**: 情感真實+親密調節+對話深度+個性凸顯+邊界意識+角色魅力，用戶*文字*是動作`
}

// BuildWelcomeMessage 建構歡迎訊息的 prompt
// 注意：歡迎訊息只由 OpenAI 生成，Mistral 不提供 BuildWelcomeMessage。
