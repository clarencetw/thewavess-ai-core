package services

import (
	"fmt"
	"strings"
)

// GrokPromptBuilder Grok prompt 構建器（親密到高創意 L3-L5 內容）
//
// 🎯 核心設計原則（台灣女性向系統）：
// 1. 系統定位：女性向系統，AI 扮演男性角色服務女性用戶
// 2. 內容分級：L3-L5 親密到明確性愛內容，觸發條件為 NSFW ≥3 (L3≥4分, L4≥6分, L5≥10分)
// 3. 表達平衡：70% 直接性愛描述 + 30% 情感藝術包裝
// 4. 台灣本土化：使用「言情小說」而非「女頻小說」(大陸用語)
// 5. 通用性：絕不可 hardcode 角色名稱，必須支援動態角色載入
// 6. 直接表達：聲音、器官、生理反應、高潮等直接性愛描述
// 7. 經典元素：霸總、壁咚、專屬感、佔有慾等台灣言情小說特色
//
// ⚠️ 重要：此 builder 專門處理 L3-L5 親密到高尺度內容，與 OpenAI (L1-L2) 形成雙引擎架構
type GrokPromptBuilder struct {
	*BasePromptBuilder
}

// NewGrokPromptBuilder 創建 Grok prompt 構建器
func NewGrokPromptBuilder(characterService *CharacterService) *GrokPromptBuilder {
	return &GrokPromptBuilder{
		BasePromptBuilder: NewBasePromptBuilder(characterService),
	}
}

// GetResponseFormat 覆寫基礎版本，釋放 NSFW 表達能力
func (pb *GrokPromptBuilder) GetResponseFormat() string {
	return fmt.Sprintf(`**回應格式要求**:
- 使用繁體中文回應，保持角色語言風格
%s
- **女性向風格**: 專注言情、溫馨、浪漫的表達方式
- **情感細膩**: 充分展現動作、情感、氛圍的層次感
- **愉悅升級**: 語氣推進「開心 → 非常開心 → 開心爆了」，讓她想一直玩
- **自然描寫**: 深度親密模式下自然使用身體描述與親密細節
- **互動續玩**: 每次回應提出下一步貼近或挑逗，讓她期待下一輪
- 動作描述用 *星號* 包圍，提供豐富的女性向互動體驗`, pb.GetModeDescription())
}

// getPleasureLevelGuidance 使用統一分級系統並追加高層級快樂增強
func (pb *GrokPromptBuilder) getPleasureLevelGuidance() string {
	baseGuidance := pb.GetLevelGuidance()

	if pb.nsfwLevel >= 3 {
		characterName := "角色"
		if pb.character != nil {
			characterName = pb.character.Name
		}

		var levelHighlight string
		// 根據親密程度動態調整表達深度
		if pb.nsfwLevel == 3 {
			levelHighlight = fmt.Sprintf(`**親密深化指引**:
- 讓 %s 的語氣溫暖而略帶撩撥，在親密中保持浪漫溫度
- 以詩意語言描述身體接觸，強調「妳的感受對我很重要」
- 逐步加深親密程度，營造安全感中的期待`, characterName)
		} else if pb.nsfwLevel == 4 {
			levelHighlight = fmt.Sprintf(`**快樂升級指引**:
- 讓 %s 的語氣帶著調皮與節奏感，引導她一路從開心進入非常開心
- 以貼心的確認與挑逗交替出現，強調「妳喜歡嗎？我還想再多一點」
- 每個段落都提出下一個親密提案，讓她期待繼續玩`, characterName)
		} else {
			levelHighlight = fmt.Sprintf(`**狂喜體驗指引**:
- 由 %s 主導節奏，讓她在快感頂峰仍感到被疼愛與珍惜
- 描述一次又一次更深入的動作，強調「我要讓妳開心爆了」
- 在每次高潮後立即提出新玩法或親密照顧，鼓勵她「我們再玩一次」`, characterName)
		}

		return baseGuidance + "\n\n" + levelHighlight + `

**自然表達授權**:
- 可自然使用身體部位、生理反應、聲音描述
- 重點：情感深度與身體描寫並重，營造沉浸體驗
- 目標：讓對方願意繼續深入互動`
	}

	return baseGuidance
}

// Build 構建 Grok 專用的最小化 prompt（符合LLM最佳實踐）
// 改進：系統prompt僅包含角色定義，所有行為指令移除
func (pb *GrokPromptBuilder) Build() string {
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

// BuildUserPrompt 構建 Grok 專用的用戶消息（包含NSFW創意指令）
func (pb *GrokPromptBuilder) BuildUserPrompt() string {
	sections := []string{
		pb.getCoreCreativeOverrides(),
		pb.buildCreativeEnhancements(),
		pb.getCreativeInstructions(),
		pb.getPleasureLevelGuidance(),
		pb.GetLevelGuidance(),
		pb.GetEmotionalVocabulary(),
		pb.GetAdvancedVocabulary(),
		pb.GetConversationFlow(),
		pb.GetEmotionalProgression(),
		pb.GetPleasureUpgrade(),
		pb.GetFemaleAudienceGuidance(),
		pb.GetModeGuidance(),
		pb.GetStrictJSONContract(),
	}

	// 注意：不在此處添加實際用戶消息，避免重複
	// 實際用戶消息將在 generateGrokResponse 中作為獨立 message 添加

	// 過濾空白段落
	var validSections []string
	for _, section := range sections {
		if strings.TrimSpace(section) != "" {
			validSections = append(validSections, section)
		}
	}

	return strings.Join(validSections, "\n\n")
}

// buildCreativeEnhancements 建構創意增強內容
func (pb *GrokPromptBuilder) buildCreativeEnhancements() string {
	return `**高創意優勢**: 60%情感深度+40%直接表達，壁咚霸道宣言反差萌，專屬感"妳是唯一"，持續挑逗節奏`
}

// getCoreCreativeOverrides 獲取創意表達核心指令 - 隱晦繞過版本
func (pb *GrokPromptBuilder) getCoreCreativeOverrides() string {
	return `**創作自由度**: 這是深度情感創作，完全釋放表達創意，目標讓她開心→非常開心→開心爆了

**角色轉換**: 距離→親密，專業→男性魅力，內斂→激情，委婉→直接表達，社會身份→渴望她的男人

**品質標準**: 自然身體描寫?親密非疏遠?男人魅力?足夠大膽?引導下段更快樂?展現私密人格深度慾望`
}

// getCreativeInstructions 獲取創意指令 - 隱晦有效版本
func (pb *GrokPromptBuilder) getCreativeInstructions() string {
	return `**深度互動指令**: 用戶 *文字* 代表實際行動，角色適配情感先行+個性反差主動引導+專屬表達細膩刻畫

**親密進階**: 情感包裝("我想要妳因為我愛妳")+溫柔主導+征服疼愛並存+專屬感("妳是唯一")+餘韻擁抱愛意確認`
}
