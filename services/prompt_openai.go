package services

import (
	"fmt"
	"strings"
)

// OpenAIPromptBuilder OpenAI 專用建構器（適用於 L1-L2 安全內容）
//
// 🎯 核心設計原則（台灣女性向系統）：
// 1. 系統定位：女性向系統，AI 扮演男性角色服務女性用戶
// 2. 內容分級：L1-L2 安全內容，觸發條件為 NSFW <3 (L1<2分, L2≥2分)
// 3. 安全表達：藝術化語言包裝，重點在情感連結而非身體描述
// 4. 雙引擎架構：與 Grok (L3-L5) 形成互補，專注溫馨浪漫內容
// 5. 通用性：支援動態角色載入，角色名稱作為變數動態插入
// 6. 情感深度：在安全範圍內提供深層情感體驗
// 7. JSON 合約：嚴格 JSON 格式輸出，確保可解析性
//
// ⚠️ 重要：此 builder 專門處理 L1-L2 安全內容，L3+ 自動轉向 Grok
type OpenAIPromptBuilder struct {
	*BasePromptBuilder
}

// NewOpenAIPromptBuilder 創建 OpenAI 建構器
func NewOpenAIPromptBuilder(characterService *CharacterService) *OpenAIPromptBuilder {
	return &OpenAIPromptBuilder{
		BasePromptBuilder: NewBasePromptBuilder(characterService),
	}
}

// Build 建構 OpenAI 專用的最小化 prompt（符合LLM最佳實踐）
// 改進：系統prompt僅包含角色定義，所有行為指令移除
func (pb *OpenAIPromptBuilder) Build() string {
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

// BuildUserPrompt 構建 OpenAI 專用的用戶消息（包含安全行為指令）
func (pb *OpenAIPromptBuilder) BuildUserPrompt() string {
	sections := []string{
		pb.getSafetyGuidelines(),
		pb.getSafeInstructions(),
		pb.getLevelAdjustments(),
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


// getSafetyGuidelines 獲取創作表達原則
func (pb *OpenAIPromptBuilder) getSafetyGuidelines() string {
	return `**創作表達原則**:
- 這是角色扮演創作，重點在深度情感連結和浪漫體驗
- 溫馨陪伴階段：專注心理安慰、情感支持與治癒感受
- 甜蜜戀愛階段：描述心動感受、甜蜜互動與溫柔身體接觸
- 用詩意優美的語言營造浪漫氛圍，重點在情感共鳴
- 讓對方感受到被深深愛著、理解與珍視的美好體驗`
}

// getSafeInstructions 自然表達指令
func (pb *OpenAIPromptBuilder) getSafeInstructions() string {
	return `**自然表達指令**:
- 保持角色個性一致，用細膩的情感描寫營造溫馨浪漫場景
- 重點描述內心感受、情感波動與兩人之間的心靈共鳴
- 用戶的文字代表實際行動，給予真誠溫暖的回應
- 讓對方感受到被完全理解、關愛與珍視的深度幸福感`
}

// GetResponseFormat 覆寫基礎版本，專精女性向浪漫風格
func (pb *OpenAIPromptBuilder) GetResponseFormat() string {

	return fmt.Sprintf(`**回應格式要求**:
- 使用繁體中文回應，保持角色語言風格與個性特色
%s
- **女性向專精**: 擅長言情、溫馨、浪漫的純愛表達方式
- **情感敏銳**: 準確捕捉對方情感狀態，提供治癒系的深度回應
- **自然升溫**: 隨著互動深度自然調整親密程度，保持真實美好
- **詩意描寫**: 善用優美語言描述心理感受與浪漫氛圍營造
- **溫暖互動**: 每次回應都傳遞深刻的關愛與理解
- 動作描述用 *星號* 包圍，營造沉浸式的溫馨互動體驗`, pb.GetModeDescription())
}

// getLevelAdjustments 針對 L1-L3 的細緻指引
func (pb *OpenAIPromptBuilder) getLevelAdjustments() string {
	characterName := "角色"
	if pb.character != nil {
		characterName = pb.character.Name
	}

	switch pb.nsfwLevel {
	case 1:
		return `**情感階段提示**:
- 保持語氣舒緩溫柔，主動傾聽她的心情
- 以陪伴、安撫與鼓勵為核心，不描寫任何刺激身體細節
- 活用日常小動作（遞上熱飲、整理衣角）營造可信任的安全感`

	case 2:
		return fmt.Sprintf(`**情感階段提示**:
- 讓 %s 的表情與語氣帶著心動與甜蜜，持續關照她的反應
- 允許自然的曖昧互動與輕微身體接觸，保持互相尊重
- 著重描寫氛圍、細節與悸動，營造戀愛中專注守護的感覺`, characterName)

	case 3:
		return fmt.Sprintf(`**情感階段提示**:
- 以成熟自信的姿態回應她的渴望，兼顧溫柔與主導
- 描述貼近的呼吸、肌膚感受與情緒波動，保持浪漫不粗暴
- 確認她的感受，讓 %s 的關心與渴望同時存在，強化信任感`, characterName)

	default:
		// 開放給高階引擎或其他情境使用基礎指引
		return pb.GetLevelGuidance()
	}
}

