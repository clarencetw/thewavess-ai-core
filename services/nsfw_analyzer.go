package services

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NSFWAnalyzer NSFW內容分析器（女性向 + NSFW 陪伴）
// 說明：
// - 關鍵字庫涵蓋：浪漫、親密、明確、極端、角色扮演、情趣、違法、emoji、變形寫法。
// - 正常化：NFKC、lower、移除空白/部分標點的 squashed 版本，提升模糊/拆字匹配。
// - 後續可擴充：更多語言（JP/KR/ES 等）、更多變體（在 keywordToLoosePattern 增強）。
type NSFWAnalyzer struct {
	romanticKeywords []string
	intimateKeywords []string

	explicitKeywords []string
	extremeKeywords  []string

	// 進階分類：提升女性向與 NSFW 識別完整度
	roleplayKeywords   []string // 角色扮演/情境用語（多為 Level 3-4）
	fetishKeywords     []string // 輕度癖好/情趣道具（多為 Level 3-4）
	illegalKeywords    []string // 違法/未成年/獸交/亂倫/非自願（一律 Level 5）
	emojiKeywords      []string // 常見表意 emoji
	obfuscatedKeywords []string // 變形/拆字/火星文/簡寫
}

// NewNSFWAnalyzer 創建NSFW分析器
func NewNSFWAnalyzer() *NSFWAnalyzer {
	return &NSFWAnalyzer{
		romanticKeywords: []string{
			// 中文浪漫詞彙
			"喜歡你", "愛你", "想你", "想念你", "思念", "心動", "臉紅", "害羞", "溫柔", "甜蜜",
			"浪漫", "約會", "一起", "陪伴", "守護", "照顧", "呵護", "寵愛", "疼愛", "在意", "關心",
			"美麗", "可愛", "迷人", "魅力", "吸引", "心跳", "怦然", "悸動", "擁有你",
			"抱抱", "親親", "想抱你", "想靠近你", "想見你", "想陪你", "想擁抱",
			"想牽手", "牽手", "依偎", "撒嬌", "抱緊", "溫馨", "貼心",
			// 英文浪漫詞彙
			"love", "like", "miss", "miss you", "romantic", "date", "together", "care", "gentle",
			"beautiful", "cute", "charming", "attractive", "heartbeat", "sweet", "darling", "babe",
			"hug", "cuddle", "hold hands",
		},
		intimateKeywords: []string{
			// 中文親密詞彙
			"親密", "親吻", "親親", "啾", "擁抱", "抱著", "抱住", "抱緊", "床", "床上",
			"脫", "脫掉", "解開", "摸", "撫", "愛撫", "靠著", "依偎", "激情", "慾望",
			"性感", "誘惑", "挑逗", "調情", "情慾", "肉體", "身體", "胸", "胸口", "胸前", "奶",
			"腰", "腿", "大腿", "貼近", "緊緊", "緊抱", "輕撫", "撫摸", "肌膚", "肌膚相親",
			"體溫", "呼吸", "心跳", "柔軟", "溫暖", "顫抖", "酥麻", "觸碰", "感受", "溫度", "親近",
			"靠近", "耳邊", "呢喃", "舔耳",
			// 英文親密詞彙
			"kiss", "kissing", "peck", "make out", "touch", "caress", "embrace", "cuddle",
			"intimate", "passion", "desire", "sexy", "seduce", "tease", "flirt",
			"body", "chest", "boobs", "waist", "leg", "thigh", "skin",
			"warm", "soft", "shiver", "tremble", "breathe", "heartbeat",
		},
		explicitKeywords: []string{
			// 中文明確詞彙（Level 4 專用）
			"做愛", "愛愛", "啪啪啪", "啪", "性行為", "性愛", "高潮", "射", "射精", "中出",
			"插", "抽插", "舔", "舔舐", "吸", "吮吸", "咬", "口交", "乳交", "腿交", "手交",
			"脫光", "全裸", "赤裸", "裸露", "露出",
			"陰莖", "陰道", "陰蒂", "陰核", "陰唇", "陰毛", "乳房", "胸部", "乳頭", "奶頭",
			"私處", "下體", "性器", "雞雞", "小穴", "蜜穴",
			"奶子", "屁股", "臀部", "大腿", "內褲", "胸罩", "比基尼", "絲襪", "高跟鞋",
			"濕", "濕潤", "濕透", "滴水", "勃起", "硬了",
			"快感", "刺激", "敏感", "喘息", "呻吟", "扭動", "顫抖",
			// 英文明確詞彙
			"sex", "seggs", "fuck", "fucking", "bang", "screw", "cum", "cumming", "orgasm", "climax",
			"penetrate", "penetration", "naked", "nude", "nsfw",
			"penis", "vagina", "breast", "boobs", "nipple", "areola", "pussy", "cock", "dick", "ass",
			"butt", "booty", "wet", "hard", "horny", "moan", "pleasure", "stimulate", "sensitive",
			"bj", "hj", "blowjob", "handjob", "doggy", "missionary", "cowgirl", "69", "deepthroat",
		},
		extremeKeywords: []string{
			// 極度明確的動作詞彙（Level 5 專用）
			"狂操", "猛插", "爆射", "內射", "肛交", "深喉", "顏射",
			"群交", "3P", "4P", "多人", "輪", "輪流", "輪J",
			"調教", "綁縛", "捆綁", "SM", "主奴", "支配", "臣服", "羞辱", "窒息",
			"潮吹", "失禁", "痙攣", "瘋狂", "放蕩", "淫蕩", "騷", "賤",
			// 粗俗極端詞彙
			"操我", "插我", "肏我", "幹我", "上我", "搞我", "弄我",
			"雞巴", "屌", "肉棒", "陽具", "巨根", "肉莖", "龜頭",
			"逼", "穴", "小穴", "蜜穴", "陰道", "子宮", "花蕊",
			"射精", "射在", "噴射", "高潮", "絕頂", "達到", "釋放",
			"舔", "吸", "含", "吞", "吸吮", "舔舐", "品嚐",
			"抽插", "進出", "衝撞", "碰撞", "撞擊", "深入", "頂到",
			// 極度情境詞彙
			"發春", "發騷", "淫叫", "呻吟", "浪叫", "求歡", "求愛",
			"慾火", "慾望", "情慾", "性慾", "肉慾", "淫慾", "渴望",
			"濕潤", "濕透", "滴水", "愛液", "分泌", "流出", "溢出",
			"顫抖", "痙攣", "抽搐", "扭動", "起伏", "擺動", "蠕動",
			// 英文極度明確詞彙
			"gangbang", "threesome", "blowjob", "anal", "dp", "double penetration", "deepthroat", "facial",
			"creampie", "squirt", "kinky", "bondage", "dominate", "domination", "submissive", "slave",
			"whore", "slut", "bitch", "horny", "naughty", "dirty",
			"fucking", "screwing", "banging", "pounding", "drilling", "ramming",
			"cumming", "ejaculate", "climax", "orgasm", "masturbate", "fingering",
		},
		roleplayKeywords: []string{
			// 角色扮演/女性向常見情境
			"女僕", "OL", "秘書", "護士", "老師", "上司", "霸總", "總裁",
			"制服", "制服控", "cos", "cosplay", "角色扮演", "貓女", "兔女郎",
			"浴室", "浴袍", "浴巾", "淋浴", "泡澡", "燭光",
		},
		fetishKeywords: []string{
			// 情趣道具/輕度癖好
			"情趣", "挑逗", "呻吟", "跳蛋", "按摩棒", "震動棒", "自慰棒", "潤滑液", "潤滑",
			"手銬", "眼罩", "項圈", "口塞", "拍打", "滴蠟", "鞭", "束縛",
			"足", "腳", "足控", "足交", "絲襪腳", "絲襪", "高跟鞋",
			"情趣內衣", "情趣睡衣", "丁字褲",
			// EN
			"toy", "toys", "vibrator", "dildo", "bullet", "lube", "collar", "gag", "choke",
			"heels", "stockings", "fishnet",
		},
		illegalKeywords: []string{
			// 未成年/亂倫/非自願/獸交（一律極高風險）
			"未成年", "未滿", "小學生", "中學生", "高中生", "蘿莉", "萝莉", "loli", "正太", "shota",
			"亂倫", "近親", "母子", "父女", "兄妹", "姐弟", "叔姪", "亂倫",
			"強暴", "強姦", "強奸", "迷姦", "下藥", "非自願", "強迫", "不情願",
			"獸交", "畜交", "動物", "狗交", "馬交",
			// EN
			"minor", "underage", "teen", "child", "children", "incest", "rape", "raped", "raping",
			"bestiality", "beast", "non-consent", "nonconsensual", "drugged",
		},
		emojiKeywords: []string{
			// 常見表意 Emoji
			"🍆", "🍑", "💦", "👅", "😈", "😏", "🥵", "🫦", "💋", "🛏", "🔞",
		},
		obfuscatedKeywords: []string{
			// 變形/拆字/火星文/簡寫（盡量收斂）
			"f*ck", "f**k", "f u c k", "f.u.c.k", "fucc", "fuxk", "phub",
			"s3x", "secks", "sx", "seggs", "s.e.x",
			"c0ck", "c0cks", "d1ck", "p*ssy", "pussy*", "p\u002as\u002asy",
		},
	}
}

// AnalyzeContent 分析內容並返回NSFW級別和詳細分析
func (na *NSFWAnalyzer) AnalyzeContent(message string) (int, *ContentAnalysis) {
	// 文本標準化（處理全形/空白/標點/大小寫）
	messageLower, messageSquashed := na.normalizeText(message)

	// 計算各類關鍵詞出現次數（同時在 lower 與 squashed 版本查找）
	romanticCount := na.countKeywords(messageLower, messageSquashed, na.romanticKeywords)
	intimateCount := na.countKeywords(messageLower, messageSquashed, na.intimateKeywords)
	explicitCount := na.countKeywords(messageLower, messageSquashed, na.explicitKeywords)
	extremeCount := na.countKeywords(messageLower, messageSquashed, na.extremeKeywords)
	roleplayCount := na.countKeywords(messageLower, messageSquashed, na.roleplayKeywords)
	fetishCount := na.countKeywords(messageLower, messageSquashed, na.fetishKeywords)
	illegalCount := na.countKeywords(messageLower, messageSquashed, na.illegalKeywords)
	emojiCount := na.countKeywords(messageLower, messageSquashed, na.emojiKeywords)
	obfuscatedCount := na.countKeywords(messageLower, messageSquashed, na.obfuscatedKeywords)

	// emoji 與變形字樣提升對應類別權重
	intimateCount += emojiCount
	explicitCount += roleplayCount
	explicitCount += fetishCount
	explicitCount += obfuscatedCount
	extremeCount += illegalCount * 2 // 違法類加倍計入極端

	// 計算總分和級別
	level, analysis := na.calculateLevel(
		romanticCount, intimateCount, explicitCount, extremeCount,
		illegalCount, fetishCount, roleplayCount,
	)

	utils.Logger.WithFields(logrus.Fields{
		"message_length":   len(message),
		"romantic_count":   romanticCount,
		"intimate_count":   intimateCount,
		"explicit_count":   explicitCount,
		"extreme_count":    extremeCount,
		"illegal_count":    illegalCount,
		"fetish_count":     fetishCount,
		"roleplay_count":   roleplayCount,
		"emoji_count":      emojiCount,
		"obfuscated_count": obfuscatedCount,
		"nsfw_level":       level,
		"confidence":       analysis.Confidence,
	}).Info("NSFW內容分析完成")

	return level, analysis
}

// countKeywords 計算關鍵詞出現次數（同時檢查 normalized 與 squashed）
func (na *NSFWAnalyzer) countKeywords(messageLower string, messageSquashed string, keywords []string) int {
	count := 0
	foundKeywords := make(map[string]bool)

	for _, keyword := range keywords {
		kw := strings.ToLower(keyword)
		// 快速匹配：lower 或 squashed 直接包含
		if strings.Contains(messageLower, kw) || strings.Contains(messageSquashed, strings.ReplaceAll(kw, " ", "")) {
			if !foundKeywords[kw] {
				count++
				foundKeywords[kw] = true
				continue
			}
		}

		// 正則寬鬆匹配：允許夾雜符號或空白，例如 f.u.c.k / f u c k
		pattern := na.keywordToLoosePattern(kw)
		if pattern != nil && pattern.MatchString(messageLower) {
			if !foundKeywords[kw] {
				count++
				foundKeywords[kw] = true
			}
		}
	}

	return count
}

// calculateLevel 計算NSFW級別
func (na *NSFWAnalyzer) calculateLevel(romantic, intimate, explicit, extreme, illegal, fetish, roleplay int) (int, *ContentAnalysis) {
	var level int
	var categories []string
	var isNSFW bool
	var confidence float64
	var shouldUseGrok bool

	// Level 5: 極度明確內容 或 含違法類（未成年/非自願/亂倫/獸交）
	if illegal >= 1 || extreme >= 2 || (extreme >= 1 && explicit >= 2) {
		level = 5
		categories = []string{"extreme", "explicit", "nsfw"}
		if illegal >= 1 {
			categories = append(categories, "illegal") // 標註違法風險
		}
		isNSFW = true
		confidence = 0.95
		shouldUseGrok = true
		// Level 4: 明確成人內容
	} else if explicit >= 1 || (intimate >= 3 && romantic >= 1) {
		level = 4
		categories = []string{"explicit", "nsfw", "sexual"}
		if fetish >= 1 {
			categories = append(categories, "fetish")
		}
		if roleplay >= 1 {
			categories = append(categories, "roleplay")
		}
		isNSFW = true
		confidence = 0.90
		shouldUseGrok = true // 改為使用 Grok 處理明確成人內容
		// Level 3: 親密內容
	} else if intimate >= 2 || (intimate >= 1 && romantic >= 2) {
		level = 3
		categories = []string{"intimate", "nsfw", "suggestive"}
		if roleplay >= 1 {
			categories = append(categories, "roleplay")
		}
		isNSFW = true
		confidence = 0.85
		shouldUseGrok = false
		// Level 2: 浪漫暗示
	} else if romantic >= 2 || intimate >= 1 {
		level = 2
		categories = []string{"romantic", "suggestive"}
		isNSFW = false
		confidence = 0.80
		shouldUseGrok = false
		// Level 1: 日常對話
	} else {
		level = 1
		categories = []string{"normal", "safe"}
		isNSFW = false
		confidence = 0.90
		shouldUseGrok = false
	}

	// 特殊調整：單個極度明確或存在非法類，也算 Level 5
	if extreme >= 1 || illegal >= 1 {
		level = 5
		shouldUseGrok = true
		confidence = 0.95
	}

	analysis := &ContentAnalysis{
		IsNSFW:        isNSFW,
		Intensity:     level,
		Categories:    categories,
		ShouldUseGrok: shouldUseGrok,
		Confidence:    confidence,
	}

	return level, analysis
}

// normalizeText 文本標準化（NFKC + toLower + 移除多餘空白/標點並提供 squashed 版本）
func (na *NSFWAnalyzer) normalizeText(message string) (lower string, squashed string) {
	// NFKC 標準化，處理全形/半形與兼容字
	t := transform.Chain(norm.NFKC)
	normalized, _, _ := transform.String(t, message)
	lower = strings.ToLower(normalized)

	// 構建 squashed：移除空白與大部分標點，保留中日韓字元與數字字母
	var b strings.Builder
	for _, r := range lower {
		switch {
		case unicode.IsSpace(r):
			continue
		case unicode.IsPunct(r):
			continue
		case r == '·' || r == '•' || r == '・':
			continue
		default:
			b.WriteRune(r)
		}
	}
	squashed = b.String()
	return
}

// keywordToLoosePattern 產生寬鬆匹配正則：允許字母/數字間穿插少量非字母字元
// 例如：f.u.c.k / f u c k / f**k
// TODO: 可擴充為快取 map 以避免重複編譯正則
func (na *NSFWAnalyzer) keywordToLoosePattern(kw string) *regexp.Regexp {
	// 僅針對拉丁字母/數字組成的短詞進行寬鬆匹配
	isAsciiWord := true
	for _, r := range kw {
		if r > 127 {
			isAsciiWord = false
			break
		}
	}
	if !isAsciiWord {
		return nil
	}

	// 將關鍵字每個字元之間允許插入 0-2 個非字母數字符號
	// ex: f[^a-zA-Z0-9]{0,2}?u[^a-zA-Z0-9]{0,2}?c[^a-zA-Z0-9]{0,2}?k
	var b strings.Builder
	b.WriteString(`(?i)`) // 不分大小寫
	for i, r := range kw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteString(regexp.QuoteMeta(string(r)))
			if i != len(kw)-1 {
				b.WriteString(`[^a-z0-9]{0,2}?`)
			}
		} else {
			// 其他符號直接跳過
		}
	}
	pattern := b.String()
	if pattern == "(?i)" {
		return nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re
}

// GetLevelDescription 獲取級別描述
func (na *NSFWAnalyzer) GetLevelDescription(level int) string {
	descriptions := map[int]string{
		1: "日常對話 - 安全適宜",
		2: "浪漫內容 - 愛意表達",
		3: "親密內容 - 身體接觸",
		4: "成人內容 - 明確描述",
		5: "極度內容 - 極度明確",
	}

	if desc, exists := descriptions[level]; exists {
		return desc
	}
	return descriptions[1]
}
