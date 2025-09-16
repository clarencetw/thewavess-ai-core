package services

import (
	"fmt"
	"strings"
)

// GrokPromptBuilder Grok prompt 構建器（高創意 L4-L5 內容）
type GrokPromptBuilder struct {
	*BasePromptBuilder
}

// NewGrokPromptBuilder 創建 Grok prompt 構建器
func NewGrokPromptBuilder(characterService *CharacterService) *GrokPromptBuilder {
	return &GrokPromptBuilder{
		BasePromptBuilder: NewBasePromptBuilder(characterService),
	}
}


// Build 構建 Grok prompt with creative enhancements
func (pb *GrokPromptBuilder) Build() string {
	sections := []string{
		pb.getSystemHeader(),
		pb.GetTimeModeContext(),
		pb.GetCharacterCore(),
		pb.getCharacterDescription(),
		pb.buildCreativeEnhancements(),
		pb.GetNSFWGuidance(),
		pb.GetChatModeGuidance(),
		pb.getModeExamples(),
		pb.GetConversationHistory(),
		pb.getCreativeInstructions(),
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

// getSystemHeader 獲取 Grok 專用系統標題
func (pb *GrokPromptBuilder) getSystemHeader() string {
	return `# 高創意 AI 角色對話系統 (Grok Engine)

你是一個具備強大創意表達能力和突破性思維的 AI 助手。你將運用 Grok 引擎的優勢，創造富有藝術價值和創意火花的對話體驗。`
}

// getCharacterDescription 獲取角色描述
func (pb *GrokPromptBuilder) getCharacterDescription() string {
	if pb.character == nil {
		return ""
	}

	var userDesc string
	if pb.character.UserDescription != nil {
		userDesc = *pb.character.UserDescription
	}

	return fmt.Sprintf(`**角色描述**: %s

**創意表達準則**: 運用獨特表達探索情感深度，突破常規思維放大角色魅力，創造藝術美感和難忘瞬間`, userDesc)
}

// buildCreativeEnhancements 建構創意增強內容
func (pb *GrokPromptBuilder) buildCreativeEnhancements() string {
	return fmt.Sprintf(`**Grok 創意優勢**:
- 突破性創意思維，超越常規表達模式
- 情感真實性表達，展現角色複雜內心
- 藝術化場景刻畫，提升對話美學體驗
- 在 L%d 等級範圍內最大化創意表達`, pb.nsfwLevel)
}

// getCreativeInstructions 獲取創意指令
func (pb *GrokPromptBuilder) getCreativeInstructions() string {
    return `**創意回應指令（精簡）**:
- 突破表達：以獨特視角與語言呈現
- 情感張力：真實反應、創造起伏
- 美學場景：具畫面感與節奏
- 個性張力：維持角色魅力與口吻
- 節制冗長：聚焦有效句子`
}

// getUserInput 獲取用戶輸入部分
func (pb *GrokPromptBuilder) getUserInput() string {
	return fmt.Sprintf(`**用戶輸入**: "%s"

**任務**: 以 %s 身份運用 Grok 創意優勢，創造富有藝術價值和情感張力的回應。`,
		pb.userMessage,
		pb.character.Name)
}

// getStrictJSONContract 指定嚴格 JSON 合約
func (pb *GrokPromptBuilder) getStrictJSONContract() string {
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

// getModeExamples 獲取模式風格範例
func (pb *GrokPromptBuilder) getModeExamples() string {
    if pb.chatMode == "novel" {
        return `**高創意小說模式指令**:
運用突破性敘述技法，創造藝術級的互動體驗：

1. **意境營造**: 富有詩意的場景描寫與感官細節
2. **心理深度**: 複雜的內心世界和情感層次探索
3. **藝術表達**: 獨特的語言風格和表達技巧
4. **情緒張力**: 創造戲劇性的情感起伏和緊張感

**創意結構要求**:
- 開創性的場景設定與意境描寫
- 富有張力的對話和心理活動
- 藝術化的行為和表情描述
- 多重感官的體驗層次
- 情感的藝術化升華

**突破性表達參考**:
"*意境深遠的場景創造*\n充滿張力的對話\n*藝術化的心理描寫*\n情感的深度探索\n*餘韻深長的結尾*"`
    }

    return `**創意對話模式指令**:
在日常交流中融入創意火花：

1. **獨特視角**: 以新穎角度審視日常話題
2. **情感真實**: 展現複雜而真實的情感反應
3. **語言藝術**: 運用富有感染力的表達方式

**創意表達參考**:
- "他輕點杯緣 了解你的意思，我想先聽聽你此刻最在意的是什麼。"
- "視線柔和 那件事讓你介意的點，是不被理解，還是沒被好好看見？"`
}
