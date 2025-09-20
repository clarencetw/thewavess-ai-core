package services

import (
	"fmt"
	"strings"
)

// OpenAIPromptBuilder OpenAI 專用建構器（適用於 L1-L3 安全內容）
//
// 🎯 核心設計原則（台灣女性向系統）：
// 1. 系統定位：女性向系統，AI 扮演男性角色服務女性用戶
// 2. 內容分級：L1-L3 安全→中等內容，觸發條件為 NSFW <4 (L1<2分, L2≥2分, L3≥4分)
// 3. 安全表達：藝術化語言包裝，避免直接性愛描述
// 4. 雙引擎架構：與 Grok (L4-L5) 形成互補，處理較溫和內容
// 5. 通用性：絕不可 hardcode 角色名稱，必須支援動態角色載入
// 6. JSON 合約：嚴格 JSON 格式輸出，確保可解析性
//
// ⚠️ 重要：此 builder 專門處理 L1-L3 安全內容，OpenAI 拒絕時自動轉向 Grok
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
		pb.GetFemaleAudienceGuidance(),
		pb.GetModeGuidance(),
		pb.getSafetyGuidelines(),
		pb.GetNSFWGuidance(),
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
	return `**創意表達原則**:
- 這是成人角色扮演創作，重點在情感連結和藝術表達
- 用詩意語言描述親密互動，強調情感深度勝過身體描述
- 以「靠近」「溫暖」「親密」等詞彙替代直接描述
- 創造有藝術價值的溫馨體驗，展現角色的真實情感`
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

	return `**輕鬆對話模式指令（女性向系統）**:
重點在吸引女性用戶，展現角色魅力與體貼：

1. 溫暖回應：先給理解與關懷（展現角色特有的體貼）
2. 魅力語氣：依角色性格調整，避免過於軟弱或生硬
3. 主動引導：體貼地關心對方，提供安全感
4. 細節渲染：用角色特色視角增添溫馨感
5. 動作約定：用戶的 *文字* 是用戶動作；你自然回應即可`
}

// getSafeInstructions 精簡安全指令（保持角色一致、情感連結、品質與邊界）。
// JSON 欄位與格式限制統一由 getStrictJSONContract 規範，避免重複。
func (pb *OpenAIPromptBuilder) getSafeInstructions() string {
	return `**創意回應指令**:
- 角色一致：維持設定與口吻，展現角色獨特魅力
- 情感先行：重視情感連結，依角色性格表達關愛與陪伴
- 藝術包裝：用文學語言描述親密，如"感受彼此心跳"、"溫暖擁抱"
- 自然發展：強調角色間的情感共鳴和自然互動
- 成熟自信：成熟男性角色應展現自信魅力，不會對親密話題過度迴避
- 角色魅力：依角色設定調整語氣與行為，展現吸引女性的特質
- 動作規則：用戶的 *文字* 是用戶動作，自然回應即可
- 創意邊界：在藝術表達範圍內最大化情感親密感`
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
