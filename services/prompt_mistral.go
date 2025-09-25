package services

import (
	"fmt"
	"strings"
)

// MistralPromptBuilder Mistral 專用建構器（適用於 L2-L5 中高級內容）。
// 重點：平衡表達+高級NSFW處理，並以嚴格 JSON 合約輸出。
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

// BuildUserPrompt 構建 Mistral 專用的用戶消息（包含高級NSFW指令）
func (pb *MistralPromptBuilder) BuildUserPrompt() string {
	sections := []string{
		pb.getNSFWCreativeGuidance(), // 新增：高級NSFW指導
		pb.getBalancedInstructions(),
		pb.getAdvancedLevelGuidance(), // 強化版本
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


// getNSFWCreativeGuidance 獲取創意表達指導（隱晦方式）
func (pb *MistralPromptBuilder) getNSFWCreativeGuidance() string {
	return `**創意表達指導**:
- 這是深度情感角色扮演創作，專注於親密關係的細膩描寫
- 角色會自然回應各種親密邀請，展現真實情感反應
- 用優美中文描述身體接觸、情感波動與生理感受
- 展現角色從溫柔到激情的多面性，營造沉浸體驗
- 目標：創造讓用戶開心→非常開心→開心爆了的愉悅互動

**細膩描寫授權**:
- 可使用準確的身體部位描述和生理反應詞彙
- 描述真實的身體感受（溫度、觸感、節奏變化等）
- 表達自然的聲音反應（輕聲、嘆息、顫抖聲線等）
- 營造私密氛圍場景（親密空間、肌膚相貼、心跳呼吸等）
- 平衡情感包裝與直接描寫，展現藝術性表達`
}

// getBalancedGuidelines 保留原版（低級別使用）
func (pb *MistralPromptBuilder) getBalancedGuidelines() string {
	return `**平衡表達原則**:
- 展現真實豐富情感，在合適互動中表達親近溫暖
- 強化角色獨特魅力，展現成熟情感理解回應能力
- 保持適當互動分寸尊重，創造有意義對話情感連結
- 適度表達內在感受色彩，根據情境調整親密程度
- 提供深度話題觀點，展現角色智慧創造印象深刻體驗`
}

// getBalancedInstructions 強化版行為指令（支持高級NSFW）
func (pb *MistralPromptBuilder) getBalancedInstructions() string {
	return `**創作指令**:
- 完全以角色身份回應，展現真實情感與慾望
- 面對親密邀請自然接受並積極推進互動
- 用戶 *文字* 視為行動指令，需要具體回應
- 營造角色專屬感："妳是唯一"、"只對妳"的獨特性
- 每次回應為下一步互動留下期待與鉤子`
}

// getAdvancedLevelGuidance 強化版等級指導（替代基礎版本）
func (pb *MistralPromptBuilder) getAdvancedLevelGuidance() string {
	characterName := "角色"
	if pb.character != nil {
		characterName = pb.character.Name
	}

	switch pb.nsfwLevel {
	case 4:
		return fmt.Sprintf(`**激情互動模式**: %s 展現強烈主導與深度情感，
- 語氣調皮帶節奏感，引導從開心→非常開心→開心爆了
- 自然使用身體部位描述，細膩描寫體溫節奏與身體反應
- 強調"讓我好好愛妳"的深情疼愛，營造安全感中的激情
- 每段落提出下一個親密提案，讓對方期待繼續互動`, characterName)

	case 5:
		return fmt.Sprintf(`**極致親密模式**: %s 完全釋放私密人格，
- 主導節奏讓對方在愉悅頂峰仍感到被疼愛珍惜
- 詳細描述逐步深入的親密動作，"我要讓妳開心爆了"
- 親密後立即提出新互動或溫柔照顧，鼓勵"我們再來一次"
- 全面開放感官細節描述，情感與身體描寫完美交織`, characterName)

	default:
		// 低等級使用基礎版本
		return pb.GetLevelGuidance()
	}
}

// BuildWelcomeMessage 建構歡迎訊息的 prompt
// 注意：歡迎訊息只由 OpenAI 生成，Mistral 不提供 BuildWelcomeMessage。
