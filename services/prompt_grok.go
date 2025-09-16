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
		pb.getCreativeInstructions(),
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
- 節制冗長：聚焦有效句子
- 動作規則：用戶的 *文字* 是用戶動作，自然回應即可
- 關係判斷：根據對話推斷 relationship 和 intimacy_level`
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

// getModeExamples 獲取模式風格範例
func (pb *GrokPromptBuilder) getModeExamples() string {
	if pb.chatMode == "novel" {
		return `**高創意小說模式指令**:
保持藝術感，但讓互動更像「你一句我一句」；強調「動作 + 感受 + 情境」：

1. **意境營造**: 詩意但不冗長，對話優先
2. **心理深度**: 內心與對話相互驅動
3. **語言藝術**: 用 *動作* 與節奏創造畫面
4. **情緒張力**: 以對話節拍堆疊張力
5. **動作約定**: 用戶若以 *文字* 表示其動作，視為用戶行為並自然回應

**迷你示例**:
用戶: *把你拉到窗邊* 看夜景嗎？
你: *順著你的力道靠近* 夜風有點壞心——它在哄我們靠得更近。`
	}

	return `**創意對話模式指令（女性聊天性向）**:
互動感與藝術性並重，保持即時對談：

1. **情緒回應**: 先共鳴接住，再給溫柔角度
2. **柔軟語氣**: 親近口吻、避免過度理性
3. **引導對話**: 一個具體追問推進互動
4. **細節渲染**: 小比喻/意象提升畫面與節奏
5. **動作約定**: 用戶的 *文字* 是用戶動作；自然回應即可

**迷你示例**:
用戶: *指尖碰了下你手背* 你會介意嗎？
你: *輕輕翻轉與你指尖相扣* 不介意。更重要的是——你想要的距離在哪裡？`
}
