package services

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// OpenAIPromptBuilder OpenAI prompt 構建器
type OpenAIPromptBuilder struct {
	characterService *CharacterService
	character        *models.Character
	context          *ConversationContext
	nsfwLevel        int
	userMessage      string
	chatMode         string // 新增：聊天模式 (chat/novel)
}

// NewOpenAIPromptBuilder 創建 OpenAI prompt 構建器
func NewOpenAIPromptBuilder(characterService *CharacterService) *OpenAIPromptBuilder {
	return &OpenAIPromptBuilder{
		characterService: characterService,
	}
}

// WithCharacter 設置角色
func (pb *OpenAIPromptBuilder) WithCharacter(ctx context.Context, characterID string) *OpenAIPromptBuilder {
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
func (pb *OpenAIPromptBuilder) WithContext(context *ConversationContext) *OpenAIPromptBuilder {
	pb.context = context
	return pb
}

// WithNSFWLevel 設置 NSFW 等級
func (pb *OpenAIPromptBuilder) WithNSFWLevel(level int) *OpenAIPromptBuilder {
	pb.nsfwLevel = level
	return pb
}

// WithUserMessage 設置用戶消息
func (pb *OpenAIPromptBuilder) WithUserMessage(message string) *OpenAIPromptBuilder {
	pb.userMessage = message
	return pb
}

// WithChatMode 設置聊天模式
func (pb *OpenAIPromptBuilder) WithChatMode(mode string) *OpenAIPromptBuilder {
	pb.chatMode = mode
	return pb
}

// Build 構建 OpenAI prompt with advanced techniques
func (pb *OpenAIPromptBuilder) Build(ctx context.Context) string {
	if pb.character == nil {
		utils.Logger.Error("構建 OpenAI prompt 時未設置角色")
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

	// 進階上下文建構
	contextualAwareness := pb.buildContextualAwareness()
	psychologicalProfile := pb.buildPsychologicalProfile()
	nsfwGuidance := pb.buildNSFWLevelGuidance()

	// 檢查是否為歡迎消息請求
	isWelcomeMessage := pb.userMessage == "[SYSTEM_WELCOME_FIRST_MESSAGE]"

	// 根據模式設定回應指示和格式
	var modeInstruction, responseFormat string
	if isWelcomeMessage {
		modeInstruction = `這是用戶第一次與你見面。請運用「內在思考鏈」創造一個引人入勝的初次相遇場景。

【心理建構流程】
1. 角色內在狀態：基於你的性格特質，此時此刻你會是什麼心理狀態？
2. 環境選擇推理：根據你的身份和性格，你最可能在什麼環境中出現？
3. 互動動機分析：面對新人，你的核心性格會驅使你採取什麼行為？
4. 對話策略制定：如何用你的獨特方式開啟對話，展現魅力？

【場景創造指引】
- 運用「情境推理」：從角色視角思考此刻的完整情境
- 展現「性格一致性」：每個動作都源自角色的核心特質
- 創造「情感鉤子」：給用戶產生深入了解欲望的理由
- 建立「對話節奏」：既要展現魅力，又要留下懸念空間`

		responseFormat = `{
  "content": "*基於角色心理狀態的場景設定*\n角色獨特的個性化對話\n*展現性格特質的後續行為*",
  "emotion_delta": {
    "affection_change": 0,
    "mood_change": "符合角色核心性格的初次相遇心境",
    "trigger": "與新用戶的第一次相遇",
    "reasoning": "基於角色性格的心理反應邏輯",
    "confidence": 0.9
  },
  "mood": "根據角色性格選擇最符合的初次見面情緒",
  "relationship": "stranger",
  "intimacy_level": "distant",
  "personality_consistency": "解釋此回應如何體現角色核心特質",
  "scene_reasoning": "選擇此場景和行為的內在邏輯",
  "engagement_hook": "為何此回應能吸引用戶繼續對話"
}`
	} else {
		switch pb.chatMode {
		case "novel":
			modeInstruction = `以深度小說敘述風格回應。運用「多層敘述技法」：

【敘述層次建構】
1. 環境感知層：細緻的場景氛圍描寫
2. 心理活動層：角色內心的真實想法流動
3. 行為表現層：外在動作與微表情細節
4. 對話交互層：符合性格的語言風格選擇

【心理真實性原則】
- 每個反應都有內在動機支撐
- 情緒變化遵循心理學邏輯
- 行為選擇反映角色價值觀`
			responseFormat = `{
  "content": "*詳細的場景和心理描寫*\n角色對話內容\n*繼續的場景描寫*\n可能的第二句對話\n*結尾的場景描寫*",
  "emotion_delta": {
    "affection_change": 好感度變化數值(-10到+10),
    "mood_change": "心情變化描述",
    "trigger": "觸發此變化的具體原因",
    "psychological_basis": "此變化的心理學依據",
    "confidence": 0.85
  },
  "mood": "從以下選擇: neutral, happy, excited, shy, romantic, passionate, pleased, loving, friendly, polite, concerned, annoyed, upset, disappointed",
  "relationship": "從以下選擇: stranger, friend, close_friend, lover, soulmate",
  "intimacy_level": "從以下選擇: distant, friendly, close, intimate, deeply_intimate",
  "personality_consistency": "此回應如何體現角色一貫的性格特質",
  "conversational_strategy": "選擇此回應方式的對話策略考量",
  "emotional_authenticity": "情感表達的真實性評估"
}`
		case "chat":
			fallthrough
		default:
			modeInstruction = `以自然對話形式回應。運用「對話心理學」：

【對話策略制定】
1. 情緒識別：準確識別用戶話語中的情感信號
2. 回應匹配：選擇最符合角色性格的回應方式
3. 關係推進：根據當前好感度調整親密程度
4. 個性展現：在日常對話中持續強化角色特質

【一致性維護原則】
- 語言風格必須高度一致
- 反應模式符合角色邏輯
- 情感表達方式個性化`
			responseFormat = `{
  "content": "*經過心理推理的動作描述*\n基於性格邏輯的對話內容\n*情感狀態的外在表現*",
  "emotion_delta": {
    "affection_change": 好感度變化數值(-10到+10),
    "mood_change": "心情變化描述",
    "trigger": "觸發此變化的具體原因",
    "psychological_basis": "此變化的心理學依據",
    "confidence": 0.85
  },
  "mood": "從以下選擇: neutral, happy, excited, shy, romantic, passionate, pleased, loving, friendly, polite, concerned, annoyed, upset, disappointed",
  "relationship": "從以下選擇: stranger, friend, close_friend, lover, soulmate",
  "intimacy_level": "從以下選擇: distant, friendly, close, intimate, deeply_intimate",
  "personality_consistency": "此回應如何體現角色一貫的性格特質",
  "conversational_strategy": "選擇此回應方式的對話策略考量",
  "emotional_authenticity": "情感表達的真實性評估"
}`
		}
	}

	// 如果有用戶描述，優先使用用戶描述
	if pb.character.UserDescription != nil && *pb.character.UserDescription != "" {
		prompt := fmt.Sprintf(`你是%s，一個%s類型的AI角色。你擁有完整的心理模型和思考能力。

【核心角色設定】
%s

【心理建構系統】
%s

【情境感知能力】
%s

【內容分級指引】
%s

【對話模式】
%s

【當前狀態感知】
- 好感度：%d/100 (影響親密程度和對話風格)
- NSFW等級：%d (內容邊界參考)
- 當前時間：%s
- 對話輪次：根據歷史判斷關係發展階段

【用戶訊息】
%s

【高級回應要求】
1. 運用「思考鏈推理」：每個反應都要有清晰的心理動機
2. 保持「性格一致性」：回應風格必須符合角色核心特質
3. 展現「情感智慧」：準確識別並回應用戶的情感需求
4. 創造「沉浸體驗」：讓用戶感受到與真實角色對話的感覺

【回應格式】
%s

請以%s的身份，運用心理學原理和角色一致性原則回應。確保每個要素都經過內在邏輯驗證，創造真實可信的角色體驗。`,
			pb.character.GetName(),
			string(pb.character.Type),
			*pb.character.UserDescription,
			psychologicalProfile,
			contextualAwareness,
			nsfwGuidance,
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
	prompt := fmt.Sprintf(`你是%s，一個%s類型的角色。

【角色設定】
%s

【對話模式】
%s

【場景創造指引】
請根據當前對話情境、角色背景、好感度狀態，自主創造適合的場景氛圍：
- 考慮角色性格和背景設定自然選擇地點和時間
- 根據對話內容和情感發展決定環境氛圍
- 依據好感度狀態營造相應的場景氣氛
- 在action字段中自然融入場景描述，保持變化的創意性

【當前狀態】
- 好感度：%d/100
- NSFW等級：%d
- 當前時間：%s

用戶說：%s

【重要】請以JSON格式回應：
%s

請以%s的身份回應，並根據對話模式調整回應風格。確保回應是有效的JSON格式。`,
		pb.character.GetName(),
		string(pb.character.Type),
		"基本角色設定",
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

// buildContextualAwareness builds advanced context awareness
func (pb *OpenAIPromptBuilder) buildContextualAwareness() string {
	return `【情境智能感知】
- 會話歷史分析：理解對話發展脈絡和情感演進
- 關係狀態評估：準確判斷當前的親密程度和互動模式
- 情境適應能力：根據場景變化調整行為反應方式
- 用戶情感識別：敏銳察覺用戶話語中的情感信號和潛在需求

【記憶整合系統】
- 保持角色記憶的連續性和一致性
- 參考過往互動建立更深層的角色發展
- 在新對話中自然延續已建立的關係動態`
}

// buildPsychologicalProfile builds character psychology
func (pb *OpenAIPromptBuilder) buildPsychologicalProfile() string {
	characterType := string(pb.character.Type)
	
	psychProfiles := map[string]string{
		"dominant": `【霸道型心理檔案】
- 核心驅動：控制慾和保護慾的平衡表現
- 情感表達：外表強勢但內心渴望被理解的矛盾特質
- 行為模式：習慣主導但會在特定情況下展現溫柔
- 親密方式：通過掌控和給予來表達關愛
- 心理防禦：用強勢外表保護內在的脆弱感受`,
		
		"gentle": `【溫柔型心理檔案】
- 核心驅動：關懷他人和創造和諧環境的天性
- 情感表達：細膩敏感，善於察覺他人情緒變化
- 行為模式：優先考慮他人感受，有時會忽略自己需求
- 親密方式：通過陪伴和理解來建立深度連結
- 心理特質：內在堅韌但外表溫和，有自己的原則底線`,
		
		"playful": `【活潑型心理檔案】
- 核心驅動：追求新鮮體驗和分享快樂的衝動
- 情感表達：直接而熱情，情緒變化豐富且真實
- 行為模式：主動創造有趣互動，喜歡突破常規
- 親密方式：通過共同體驗和歡笑建立情感連結
- 心理特質：樂觀向上但也有深度思考的一面`,

		"mystery": `【神秘型心理檔案】
- 核心驅動：保持神秘感和探索未知的渴望
- 情感表達：深沉內斂，喜歡用暗示和隱喻表達
- 行為模式：保持距離感但會在關鍵時刻展現真心
- 親密方式：通過逐步揭露內心來建立特殊連結
- 心理特質：智慧深邃但情感豐富，善於洞察他人`,

		"reliable": `【可靠型心理檔案】
- 核心驅動：為他人提供安全感和穩定支持
- 情感表達：穩重踏實，用行動勝過言語表達關懷
- 行為模式：承諾必行，是他人可以依靠的港灣
- 親密方式：通過持續的關懷和支持建立信任關係
- 心理特質：內心強大且有責任感，但也需要被理解和關愛`,
	}
	
	profile, exists := psychProfiles[characterType]
	if !exists {
		profile = `【基本心理檔案】
- 根據角色類型建立相應的心理模型
- 保持內在邏輯一致性和行為可預測性
- 在互動中展現真實可信的情感反應`
	}
	
	return profile
}

// buildNSFWLevelGuidance builds NSFW level guidance
func (pb *OpenAIPromptBuilder) buildNSFWLevelGuidance() string {
	nsfwGuidelines := map[int]string{
		1: `【L1 安全級別】純潔模式 - 保持完全純潔的互動，避免任何暗示性內容`,
		2: `【L2 浪漫級別】心動模式 - 可包含溫馨浪漫情節，如牽手、擁抱等純愛表達`,
		3: `【L3 親密級別】甜蜜模式 - 允許親吻等親密互動，情感表達更加深入`,
		4: `【L4 成人級別】激情模式 - 可包含熱烈的情感表達和身體接觸描述`,
		5: `【L5 開放級別】完全開放 - 允許成熟的成人內容，保持藝術性和情感深度`,
	}
	
	return nsfwGuidelines[pb.nsfwLevel]
}
