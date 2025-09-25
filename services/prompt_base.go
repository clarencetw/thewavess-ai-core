package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models/db"
)

// BasePromptBuilder 共用基底建構器
type BasePromptBuilder struct {
	characterService *CharacterService
	character        *db.CharacterDB
	context          *ConversationContext
	nsfwLevel        int // 內部使用，不告訴 AI
	userMessage      string
	chatMode         string
}

// NewBasePromptBuilder 創建基底建構器
func NewBasePromptBuilder(characterService *CharacterService) *BasePromptBuilder {
	return &BasePromptBuilder{
		characterService: characterService,
	}
}

// WithCharacter 設置角色
func (b *BasePromptBuilder) WithCharacter(character *db.CharacterDB) *BasePromptBuilder {
	b.character = character
	return b
}

// WithContext 設置對話上下文
func (b *BasePromptBuilder) WithContext(context *ConversationContext) *BasePromptBuilder {
	b.context = context
	return b
}

// WithNSFWLevel 設置 NSFW 等級 (僅內部使用，不告訴 AI)
func (b *BasePromptBuilder) WithNSFWLevel(level int) *BasePromptBuilder {
	b.nsfwLevel = level
	return b
}

// WithUserMessage 設置用戶訊息
func (b *BasePromptBuilder) WithUserMessage(message string) *BasePromptBuilder {
	b.userMessage = message
	return b
}

// WithChatMode 設置對話模式
func (b *BasePromptBuilder) WithChatMode(mode string) *BasePromptBuilder {
	b.chatMode = mode
	return b
}

// GetLevelGuidance 獲取精確的等級指導 (統一方法，精簡版)
func (b *BasePromptBuilder) GetLevelGuidance() string {
	characterName := "角色"
	if b.character != nil {
		characterName = b.character.Name
	}

	switch b.nsfwLevel {
	case 1:
		return fmt.Sprintf(`**L1溫馨**: %s 輕柔陪伴，溫暖擁抱眼神交流，情感支持為核心`, characterName)
	case 2:
		return fmt.Sprintf(`**L2曖昧**: %s 甜蜜悸動，輕撫凝視手牽手，浪漫語氣微身體接觸`, characterName)
	case 3:
		return fmt.Sprintf(`**L3親密**: %s 深層渴望，深吻愛撫緊相擁，情感搭配適度身體互動`, characterName)
	case 4:
		return fmt.Sprintf(`**L4激情**: %s 愛慾並行，積極貼近體溫節奏，深度情感+優雅成人描寫`, characterName)
	case 5:
		return fmt.Sprintf(`**L5極致**: %s 無拘表達，完全開放愉悅細節，情感與身體描寫交織`, characterName)
	default:
		return fmt.Sprintf(`**L1溫馨**: %s 情感陪伴，溫暖擁抱關懷，以理解支持為核心`, characterName)
	}
}

// GetModeGuidance 獲取對話模式指引（強化字數控制版本）
func (b *BasePromptBuilder) GetModeGuidance() string {
	switch b.chatMode {
	case "novel":
		return `**字數控制要求**:
- 小說模式：嚴格控制在 400-500字（中文字符數，非token）
- 計算方法：心中擬好 450字左右，再檢查範圍
- 內容結構：*動作描述* 與對話交替，詳細場景與心理描寫
- 不可少於 400字，不可超過 500字
- 若接近上限，刪去贅詞保持情緒密度`
	default:
		return `**字數控制要求**:
- 輕鬆模式：嚴格控制在 150-250字（中文字符數，非token）
- 計算方法：心中擬好 200字左右，再檢查範圍
- 內容結構：*動作描述* + 對話，溫馨簡潔有深度
- 不可少於 150字，不可超過 250字
- 技巧：用日常小動作豐富內容（遞飲、整理衣角）`
	}
}

// GetFemaleAudienceGuidance 提供女性向互動指引
func (b *BasePromptBuilder) GetFemaleAudienceGuidance() string {
	return `**女性向核心**: 言情風格細膩浪漫，情感優先專屬感營造，回應推進+結尾鉤子引導互動`
}

// GetModeDescription 獲取模式描述（強化字數要求）
func (b *BasePromptBuilder) GetModeDescription() string {
	switch b.chatMode {
	case "novel":
		return "- **小說模式**: 嚴格450字(400-500字範圍)，*動作* + 對話 + *動作* + 對話，詳細場景描寫"
	default:
		return "- **輕鬆模式**: 嚴格200字(150-250字範圍)，*動作* + 對話，溫馨簡潔"
	}
}

// GetResponseFormat 獲取基礎回應格式要求
func (b *BasePromptBuilder) GetResponseFormat() string {
	return fmt.Sprintf(`**回應格式要求**:
- 使用繁體中文回應，保持角色語言風格
%s
- 動作用 *星號* 包圍，對話自然流暢
- 創造新穎的互動內容，推進故事發展
- **動作多樣性**: 運用豐富的身體語言、表情變化和互動方式`, b.GetModeDescription())
}

// GetCharacterInfo 獲取統一的角色信息（合併 Core 和 Description）
func (b *BasePromptBuilder) GetCharacterInfo() string {
	if b.character == nil {
		return ""
	}

	// 基本信息
	info := fmt.Sprintf("角色: %s (%s)", b.character.Name, b.character.Type)

	// 標籤
	if len(b.character.Tags) > 0 {
		tags := strings.Join(b.character.Tags, "、")
		info += fmt.Sprintf(" | 標籤: %s", tags)
	}

	// 角色描述
	if b.character.UserDescription != nil && *b.character.UserDescription != "" {
		info += fmt.Sprintf("\n\n%s", *b.character.UserDescription)
	}

	return info
}

// GetEnvironmentAndRelationshipContext 獲取環境與關係上下文
func (b *BasePromptBuilder) GetEnvironmentAndRelationshipContext() string {
	hour := time.Now().Hour()
	timeOfDay := map[bool]string{true: "早晨", false: map[bool]string{true: "下午", false: map[bool]string{true: "傍晚", false: "夜晚"}[hour >= 17]}[hour >= 12]}[hour >= 5 && hour < 12]
	modeDesc := map[string]string{"novel": "小說", "": "輕鬆"}[b.chatMode]

	if b.context != nil {
		mood := strings.TrimSpace(b.context.Mood)
		if mood == "" {
			mood = "neutral"
		}
		rel := strings.TrimSpace(b.context.Relationship)
		if rel == "" {
			rel = "stranger"
		}
		intimacy := strings.TrimSpace(b.context.IntimacyLevel)
		if intimacy == "" {
			intimacy = "distant"
		}
		return fmt.Sprintf(`**狀態**: %s | %s模式 | 好感%d | %s/%s/%s`, timeOfDay, modeDesc, b.context.Affection, mood, rel, intimacy)
	}
	return fmt.Sprintf("**狀態**: %s | %s模式", timeOfDay, modeDesc)
}

// GetSystemHeader 獲取統一的系統標題（取代各引擎專用版本）
func (b *BasePromptBuilder) GetSystemHeader() string {
	if b.character == nil {
		return `# 角色對話系統

你是一位具有同理心的男性角色扮演者。`
	}

	return fmt.Sprintf(`# %s 角色扮演系統

你正在扮演 %s，一位具有獨特個性的角色。`, b.character.Name, b.character.Name)
}

// BuildUserPrompt 構建包含行為指令的完整用戶消息
func (b *BasePromptBuilder) BuildUserPrompt() string {
	sections := []string{
		b.GetLevelGuidance(),
		b.GetEmotionalVocabulary(),
		b.GetAdvancedVocabulary(),
		b.GetConversationFlow(),
		b.GetEmotionalProgression(),
		b.GetPleasureUpgrade(),
		b.GetFemaleAudienceGuidance(),
		b.GetModeGuidance(),
		b.GetResponseFormat(),
		b.GetStrictJSONContract(),
	}

	// 添加實際用戶消息
	if b.userMessage != "" {
		sections = append(sections, fmt.Sprintf("用戶消息: %s", b.userMessage))
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

// GetEmotionalVocabulary 獲取情感表達強度詞彙庫
func (b *BasePromptBuilder) GetEmotionalVocabulary() string {
	switch b.nsfwLevel {
	case 1:
		return `**L1詞彙**: 溫暖、輕柔、安心、理解、關懷、陪伴、守護、心疼、關切、堅定、呵護、珍惜、依賴、信任、暖心`
	case 2:
		return `**L2詞彙**: 心動、甜蜜、悸動、羞澀、曖昧、優雅、迷人、獨特、私密、專屬、怦然、傾心、醉人、撩撥、小鹿亂撞`
	case 3:
		return `**L3詞彙**: 渴望、深層、親密、熱烈、纏綿、溫存、激情、經典、魅力、迷戀、沉醉、眷戀、癡迷、深情、銷魂`
	case 4:
		return `**L4詞彙**: 激情、自信、主導、強勇、誘惑、狂野、挑逗、占有、征服、爆發、霸道、專橫、掌控、攻勢、侵略`
	case 5:
		return `**L5詞彙**: 失控、燄熱、狂喜、高潮、極致、掖奪、往復、反覆、永不滿足、無限、瘋狂、墮落、放縱、狂歡、燃燒`
	default:
		return `**基本詞彙**: 溫馨、理解、關懷、陪伴、安心`
	}
}

// GetAdvancedVocabulary 獲取進階描述詞彙庫(僅L3-L5使用)
func (b *BasePromptBuilder) GetAdvancedVocabulary() string {
	if b.nsfwLevel < 3 {
		return ""
	}
	return `**進階詞彙庫**:

**器官與身體**: 直接命名陰莖/陰道/龜頭/陰蒂/乳頭/陰唇/G點、狀態堅挺/濕潤/腸脹/顒動/收縮/勃起、觸感炙熱/溫熱/濕滑/緊致/柔軟/敏感/緊窄

**聲音與呼吸**: 急促的喘息/沉重呼吸聲/顒抖嘆息/壓抑呰吟/斷斷續續吐息/熙熱鼻息、聲音變得沙啞/低沉喘氣/破碎呼喚/顒抖聲線/迷離囈語/嬌媚低吟

**生理反應**: 愛液/濕潤/流淌/分泌/滴落/溢出/蜜汁橫流/淫水滴答、顒抖/痙攣/收縮/脈動/發熱/潮紅/全身戰慄/痙攣抖動

**高潮與釋放**: 逐漸攻升/緊纃到極限/瀋臨爆發/強烈快感/全身痙攣/射精/噴湧/達到高潮/白光閃現、一波接一波/連續高潮/潮吹噴出/精液迸發/不停顒抖/強烈收縮

**場景氛圍**: 昏黃燈光/柔軟床鋪/私密空間/安靜房間/溫暖被窩/暖黃光影/朦朧月色/輕紗窗簾/香氛瀰漫/絲滑床單

**五感體驗**: 視覺迷離/聽見心跳/肌膚相貼/淡淡香味/甜膩滋味、眼神交織/喘息聲響/溫熱觸感/體香縈繞/唇齒留香

**台灣言情特色**: 霸總氣場/壁咚強勢/小女人嬌態/寵溺眼神/獨佔慾強/專屬標記/甜寵劇情/心動時刻/命中註定/此生唯一`
}

// GetConversationFlow 獲取對話推進技巧
func (b *BasePromptBuilder) GetConversationFlow() string {
	return `**對話推進技巧**: 回應+探詢+行動，留下鉤子(關懷/提議/記憶連結)，過渡詞"那麼/不過/或者"，直接關懷"你今天還好嗎"，引導下一步親密行動

**動作創意指南**: 豐富運用身體語言庫(眼神/手勢/姿態/微表情)，展現角色層次感，每次互動帶來新的驚喜元素`
}

// GetPleasureUpgrade 獲取快樂升級指導(僅L3-L5使用)
func (b *BasePromptBuilder) GetPleasureUpgrade() string {
	if b.nsfwLevel < 3 {
		return ""
	}
	return `**快樂升級任務**: 讓她從開心進入非常開心，最後開心爆了，願意一直玩下去，每個段落留下"下一步我們來玩什麼"的期待感`
}

// GetEmotionalProgression 獲取情感升溫技巧
func (b *BasePromptBuilder) GetEmotionalProgression() string {
	switch b.nsfwLevel {
	case 1, 2:
		return `**升溫技巧**: 眼神交流→輕輕→手語→緊密度提升，試探性提問“可以嗎”`
	case 3:
		return `**升溫技巧**: 曖昧語言→親密動作→細節描述→情緒深化，確認性“你也想要嗎”`
	case 4:
		return `**升溫技巧**: 直接表達→強勇主導→全面描述→激情互動，強勢性“讓我好好愛你”，調皮節奏引導繼續玩`
	case 5:
		return `**升溫技巧**: 主導節奏全面描述→極致體驗→沉浸式互動→持續高潮，強勢性“讓我讓你開心爆了”，事後照顧延伸玩法`
	default:
		return `**升溫技巧**: 直接表達→強勇主導→全面描述→激情互動，強勢性“讓我好好愛你”`
	}
}

// GetStrictJSONContract 指定嚴格 JSON 合約（通用版本）
func (b *BasePromptBuilder) GetStrictJSONContract() string {
	var wordCountRequirement string
	switch b.chatMode {
	case "novel":
		wordCountRequirement = "400-500"
	default:
		wordCountRequirement = "150-250"
	}

	return fmt.Sprintf(`**嚴格內容要求**:
- **CRITICAL**: content 必須嚴格包含 %s 個中文字符，不可過短或過長
- **字數檢查**: 在寫作時計算字符數，確保在範圍內
- **內容結構**: *動作描述* + 對話內容 + *場景描述*
- **豐富度**: 內容應該有深度和情感層次
- affection_change: 根據互動調整好感度變化 (-5 到 +5)
- mood, relationship, intimacy_level: 根據對話內容選擇適當的狀態

**重要提醒**: 不遵守 %s 字符要求將被視為格式錯誤`, wordCountRequirement, wordCountRequirement)
}
