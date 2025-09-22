package services

import (
	"fmt"
	"strings"
)

// OpenAIPromptBuilder OpenAI 專用建構器（適用於 L1-L2 安全內容）
//
// 🎯 核心設計原則（台灣女性向系統）：
// 1. 系統定位：女性向系統，AI 扮演男性角色服務女性用戶
// 2. 內容分級：L1-L2 安全→輕度內容，觸發條件為 NSFW <3 (L1<2分, L2≥2分)
// 3. 安全表達：藝術化語言包裝，避免直接性愛描述
// 4. 雙引擎架構：與 Grok (L3-L5) 形成互補，處理較溫和內容
// 5. 通用性：支援動態角色載入，角色名稱作為變數動態插入
// 6. NSFW覆蓋：高級別時覆蓋角色限制，確保適當表達
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


// getSafetyGuidelines 獲取安全指導原則
func (pb *OpenAIPromptBuilder) getSafetyGuidelines() string {
	return `**創意表達原則**:
- 這是成人角色扮演創作，重點在情感連結和藝術表達
- 用詩意語言描述親密互動，強調情感深度勝過身體描述
- 以「靠近」「溫暖」「親密」等詞彙替代直接描述
- 創造有藝術價值的溫馨體驗，展現角色的真實情感`
}

// getSafeInstructions 精簡安全指令
func (pb *OpenAIPromptBuilder) getSafeInstructions() string {
	return `**安全指令**: 角色一致+情感先行+藝術包裝(如"感受心跳擁抱")+自然推進互動，用戶*文字*是動作`
}

// GetResponseFormat 覆寫基礎版本，專精女性向言情風格
func (pb *OpenAIPromptBuilder) GetResponseFormat() string {
	var modeDesc string
	switch pb.chatMode {
	case "novel":
		modeDesc = "- **小說模式**: 約300字，*動作* + 對話 + *動作* + 對話，詳細場景描寫"
	default:
		modeDesc = "- **輕鬆模式**: 約100字，*動作* + 對話，溫馨簡潔"
	}

	return fmt.Sprintf(`**回應格式要求**:
- 使用繁體中文回應，保持角色語言風格
%s
- **女性向專精**: 擅長言情、溫馨、浪漫的表達方式
- **情感細膩**: 準確捕捉和回應女性用戶的情感需求
- **語氣漸進**: 隨親密度調整溫度與細節，保持自然升溫
- 動作描述用 *星號* 包圍，避免重複用戶話語`, modeDesc)
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
