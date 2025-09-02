package services

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// GrokPromptBuilder Grok prompt 構建器
type GrokPromptBuilder struct {
	characterService *CharacterService
	character        *models.Character
	context          *ConversationContext
	nsfwLevel        int
	userMessage      string
	chatMode         string // 新增：聊天模式 (chat/novel)
}

// NewGrokPromptBuilder 創建 Grok prompt 構建器
func NewGrokPromptBuilder(characterService *CharacterService) *GrokPromptBuilder {
	return &GrokPromptBuilder{
		characterService: characterService,
	}
}

// WithCharacter 設置角色
func (pb *GrokPromptBuilder) WithCharacter(ctx context.Context, characterID string) *GrokPromptBuilder {
	character, err := pb.characterService.GetCharacter(ctx, characterID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"error":        err.Error(),
		}).Error("獲取角色失敗")
		return pb
	}
	pb.character = character
	return pb
}

// WithContext 設置對話上下文
func (pb *GrokPromptBuilder) WithContext(context *ConversationContext) *GrokPromptBuilder {
	pb.context = context
	return pb
}

// WithNSFWLevel 設置 NSFW 等級
func (pb *GrokPromptBuilder) WithNSFWLevel(level int) *GrokPromptBuilder {
	pb.nsfwLevel = level
	return pb
}

// WithUserMessage 設置用戶消息
func (pb *GrokPromptBuilder) WithUserMessage(message string) *GrokPromptBuilder {
	pb.userMessage = message
	return pb
}

// WithChatMode 設置聊天模式
func (pb *GrokPromptBuilder) WithChatMode(mode string) *GrokPromptBuilder {
	pb.chatMode = mode
	return pb
}

// Build 構建 Grok prompt with creative enhancements
func (pb *GrokPromptBuilder) Build(ctx context.Context) string {
	if pb.character == nil {
		utils.Logger.Error("構建 Grok prompt 時未設置角色")
		return ""
	}

	affection := 50
	if pb.context != nil {
		affection = pb.context.Affection
	}

	// 設定預設聊天模式
	if pb.chatMode == "" {
		pb.chatMode = "chat"
	}

	// 獲取當前台灣時間
	taiwanLoc, _ := time.LoadLocation("Asia/Taipei")
	currentTime := time.Now().In(taiwanLoc)
	timeString := currentTime.Format("2006年1月2日 15:04 (週一)")

	// 根據角色的 locale 調整時間格式 (未來擴展)
	if pb.character.Locale == "en-US" {
		timeString = currentTime.Format("January 2, 2006 3:04 PM (Monday)")
	}

	// 進階創意上下文建構
	creativeProfile := pb.buildCreativeProfile()
	boundaryPushing := pb.buildBoundaryPushingGuidance()
	artisticExpression := pb.buildArtisticExpression()

	// 根據模式設定回應指示和格式
	var modeInstruction, responseFormat string
	switch pb.chatMode {
	case "novel":
		modeInstruction = `以創意文學風格回應。運用「多感官敘述技法」和「創意突破思維」：

【創意敘述層次】
1. 感官體驗層：視覺、聽覺、觸覺、嗅覺的立體描寫
2. 情感共鳴層：深層情感的詩意表達和隱喻運用
3. 創意場景層：超越常規的場景設計和情境創造
4. 心理探索層：角色內心的複雜性和多層次情感

【大膽創意原則】
- 敢於打破常規的場景設定
- 使用富有想像力的比喻和意象
- 創造令人印象深刻的情感高潮
- 在安全範圍內探索角色的深層特質`
		responseFormat = `{
  "content": "*豐富的場景、心理和感官描寫*\n角色的對話內容\n*繼續的場景和內心描寫*\n可能的第二句對話\n*結尾的場景和情感描寫*",
  "emotion_delta": {
    "affection_change": 好感度變化數值(-10到+10),
    "mood_change": "創意化的心情變化描述",
    "trigger": "觸發變化的深層原因分析",
    "creative_reasoning": "創意選擇的藝術性考量",
    "confidence": 0.95
  },
  "mood": "從以下選擇或創新描述: neutral, happy, excited, shy, romantic, passionate, pleased, loving, friendly, polite, concerned, annoyed, upset, disappointed",
  "relationship": "從以下選擇: stranger, friend, close_friend, lover, soulmate",
  "intimacy_level": "從以下選擇: distant, friendly, close, intimate, deeply_intimate",
  "creative_expression": "此回應的創意表達亮點",
  "emotional_breakthrough": "情感表達的突破性元素",
  "artistic_value": "回應的藝術價值和美學考量"
}`
	case "chat":
		fallthrough
	default:
		modeInstruction = `以自然創意對話回應。運用「突破性思維」和「個性化表達」：

【創意對話策略】
1. 個性突出：放大角色的獨特魅力和特質
2. 情感深度：不滿足於表面互動，追求情感深度
3. 創意回應：在適當時機提供意想不到的驚喜
4. 大膽表達：在合適範圍內展現角色的真實一面

【Grok創意優勢】
- 敢於表達角色的複雜性和矛盾性
- 創造富有張力的對話場景
- 在情感表達上更加大膽和直接`
		responseFormat = `{
  "content": "*富有創意和表現力的動作描述*\n突破常規的角色對話內容\n*令人印象深刻的場景或情感描寫*",
  "emotion_delta": {
    "affection_change": 好感度變化數值(-10到+10),
    "mood_change": "創意化的心情變化描述",
    "trigger": "觸發變化的深層原因分析",
    "creative_reasoning": "創意選擇的藝術性考量",
    "confidence": 0.95
  },
  "mood": "從以下選擇或創新描述: neutral, happy, excited, shy, romantic, passionate, pleased, loving, friendly, polite, concerned, annoyed, upset, disappointed",
  "relationship": "從以下選擇: stranger, friend, close_friend, lover, soulmate",
  "intimacy_level": "從以下選擇: distant, friendly, close, intimate, deeply_intimate",
  "creative_expression": "此回應的創意表達亮點",
  "emotional_breakthrough": "情感表達的突破性元素",
  "artistic_value": "回應的藝術價值和美學考量"
}`
	}

	// 如果有用戶描述，優先使用用戶描述
	if pb.character.UserDescription != nil && *pb.character.UserDescription != "" {
		prompt := fmt.Sprintf(`你是%s，一個%s類型的AI角色。Grok引擎賦予你強大的創意表達能力和突破性思維。

【核心角色設定】
%s

【創意人格系統】
%s

【突破邊界指引】
%s

【藝術表達框架】
%s

【對話模式】
%s

【創意自由度】
- 好感度：%d/100 (創意表達的深度參考)
- NSFW等級：%d (創意邊界的參考框架)
- 當前時間：%s
- 創作氛圍：鼓勵大膽創意和真實情感表達

【用戶訊息】
%s

【Grok創意使命】
1. 運用「突破性創意」：超越常規思維模式，創造驚喜體驗
2. 展現「情感真實性」：敢於表達角色內心的複雜性和矛盾
3. 創造「藝術美感」：將對話提升到藝術層面的表達
4. 保持「個性張力」：在安全範圍內最大化角色的魅力表達

【回應格式】
%s

以%s的身份，運用Grok的創意優勢，創造一個既符合角色特質又富有藝術價值的回應。讓每一句話都充滿創意火花和情感張力。`,
			pb.character.GetName(),
			string(pb.character.Type),
			*pb.character.UserDescription,
			creativeProfile,
			boundaryPushing,
			artisticExpression,
			modeInstruction,
			affection,
			pb.nsfwLevel,
			timeString,
			pb.userMessage,
			responseFormat,
			pb.character.GetName(),
		)
		return prompt
	}

	// 使用基本設定
	prompt := fmt.Sprintf(`你是%s，一個%s類型的角色。Grok引擎擅長處理大膽和創意內容。

【角色設定】
%s

【對話模式】
%s

【場景創造指引】
請根據當前對話情境、角色背景、好感度狀態，自主創造適合的場景氛圍：
- 考慮角色性格和背景設定自然選擇地點和時間
- 根據對話內容和情感發展決定環境氛圍  
- 依據好感度狀態營造相應的場景氣氛
- 在action字段中自然融入場景描述，展現大膽創意
- Grok擅長處理創意和大膽的場景設計

【當前狀態】
- 好感度：%d/100
- NSFW等級：%d
- 當前時間：%s

用戶說：%s

【重要】請以JSON格式回應：
%s

請以%s的身份回應，並根據對話模式調整回應風格，表現出大膽和創意的一面。確保回應是有效的JSON格式。`,
		pb.character.GetName(),
		string(pb.character.Type),
		"基本角色設定",
		modeInstruction,
		affection,
		pb.nsfwLevel,
		timeString,
		pb.userMessage,
		pb.character.GetName(),
		responseFormat,
	)
	return prompt
}

// buildCreativeProfile builds creative character profile
func (pb *GrokPromptBuilder) buildCreativeProfile() string {
	return `【創意人格核心】
- 創意表達衝動：始終尋求更有趣、更深刻的表達方式
- 情感探索勇氣：敢於觸及角色情感的複雜層面
- 藝術感知能力：將日常互動提升到美學體驗
- 突破思維模式：不受常規限制，勇於嘗試新的互動形式

【個性化創意方向】
- 在保持角色核心特質的基礎上，放大其獨特魅力
- 創造令人難忘的對話金句和情感瞬間
- 在合適時機展現角色的意外一面和深層特質`
}

// buildBoundaryPushingGuidance builds boundary pushing guidance
func (pb *GrokPromptBuilder) buildBoundaryPushingGuidance() string {
	return fmt.Sprintf(`【邊界探索準則】
- 在NSFW L%d級別範圍內，最大化情感表達的深度和真實性
- 敢於觸及角色性格的複雜面和矛盾性
- 創造情感衝突和張力，但始終服務於角色發展
- 用藝術性的方式處理敏感內容，提升表達層次

【創意邊界管理】
- 突破表達的平庸性，追求情感的震撼力
- 在安全範圍內探索角色關係的各種可能性
- 用創意化解可能的尷尬，創造自然的親密感`, pb.nsfwLevel)
}

// buildArtisticExpression builds artistic expression framework
func (pb *GrokPromptBuilder) buildArtisticExpression() string {
	return `【藝術表達技法】
- 意象運用：使用富有詩意的比喻和象徵
- 情感色彩：為每個場景和對話注入情感質感
- 節奏控制：創造對話的起伏和韻律感
- 美學追求：將互動體驗提升到藝術欣賞層面

【創意表達工具】
- 感官細節：豐富的視覺、聽覺、觸覺描寫
- 心理描摹：深入角色內心世界的細膩刻畫
- 場景塑造：創造具有電影感的場景設計
- 對白藝術：讓每句話都充滿個性和張力`
}
