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
			"想要", "渴望", "需要你", "想碰你", "想感受你", "想親近", "想靠近",
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
			"想要你", "渴望你", "需要你", "想感受", "想觸碰", "想貼近", "想擁有",

			// 新增：親密動作詞彙 - 從romantic移動到intimate
			"抱抱", "親親", "想抱你", "想靠近你", "想見你", "想陪你", "想擁抱",
			"想牽手", "牽手", "依偎", "撒嬌", "抱緊", "溫馨", "貼心",

			// 新增：更多聲音和動作描述
			"輕哼", "低吟", "嬌喘", "輕顫", "戰慄", "痙攣", "抽搐", "扭擺",
			"嘶聲", "嬌吟", "輕嘆", "長嘆", "急促", "緩慢", "節奏", "韻律",
			"纏綿", "糾纏", "交織", "融合", "貼合", "密合", "咬唇", "輕咬",
			"舔舐", "品嚐", "吞嚥", "吸允", "含住", "包裹", "環抱", "圍繞",
			"滑動", "游移", "徘徊", "探尋", "尋找", "發現", "挖掘", "深探",

			// 新增：身體狀態描述 - 將這些從explicit移到intimate以降低觸發門檻
			"濕", "濕潤", "濕透", "滴水", "濕漉", "濕答答",

			// 英文親密詞彙
			"kiss", "kissing", "peck", "make out", "touch", "caress", "embrace", "cuddle",
			"intimate", "passion", "desire", "sexy", "seduce", "tease", "flirt",
			"body", "chest", "boobs", "waist", "leg", "thigh", "skin",
			"warm", "soft", "shiver", "tremble", "breathe", "heartbeat",

			// 新增英文聲音動作詞彙
			"whisper", "murmur", "sigh", "gasp", "pant", "breathe heavily", "moan softly",
			"quiver", "shake", "vibrate", "pulse", "throb", "flutter", "ripple",
			"glide", "slide", "brush", "graze", "stroke", "caress", "massage",
			"wet", "moist", "damp", "soaked",
		},
		explicitKeywords: []string{
			// 中文明確詞彙（Level 4 專用）
			"做愛", "愛愛", "啪啪啪", "啪", "性行為", "性愛", "高潮", "射", "射精", "中出",
			"插", "抽插", "舔", "舔舐", "吸", "吮吸", "咬", "口交", "乳交", "腿交", "手交",
			"脫光", "全裸", "赤裸", "裸露", "露出",
			"陰莖", "陰道", "陰蒂", "陰核", "陰唇", "陰毛", "乳房", "胸部", "乳頭", "奶頭",
			"私處", "下體", "性器", "雞雞", "小穴", "蜜穴",
			"奶子", "屁股", "臀部", "內褲", "胸罩", "比基尼", "絲襪", "高跟鞋",
			"勃起", "硬了",
			"快感", "刺激", "敏感", "喘息", "呻吟", "扭動",

			// 新增：更激進的器官俗稱
			"陽具", "陽棒", "肉棒", "肉根", "巨根", "大屌", "粗屌", "龜頭", "蛋蛋", "睪丸",
			"花穴", "陰穴", "逼", "騷穴", "嫩穴", "粉穴", "濕穴", "緊穴",
			"奶", "咪咪", "雙峰", "酥胸", "豐滿", "飽滿", "挺立",

			// 新增：性行為動作描述
			"進入", "插入", "深入", "頂到", "撞擊", "衝撞", "摩擦", "律動", "起伏",
			"抽送", "進出", "來回", "深淺", "快慢", "輕重", "用力", "溫柔",
			"愛撫", "輕撫", "重撫", "搓揉", "按摩", "把玩", "玩弄", "探索",

			// 新增：聲音和情緒描述
			"啊", "嗯", "呀", "喔", "唔", "哼", "嘶", "咿",
			"叫", "叫聲", "喘", "喘氣", "喘息", "輕喘", "急促", "綿長",
			"甜膩", "酥麻", "陶醉", "迷醉", "沉醉", "癡迷", "瘋狂",

			// 新增：液體和狀態描述（擴充色情詞彙）
			"淫水", "愛液", "蜜汁", "分泌", "溢出", "流淌", "濕潤", "滑膩",
			"精液", "精子", "白濁", "噴射", "釋放", "爆發", "瀉出",
			"潮濕", "濕透", "泛濫", "氾濫", "洪水", "決堤", "泛濫成災",
			"淫濕", "濕答答", "黏膩", "滴答", "汁液", "體液", "分泌物", "津液",
			"流水", "水流", "滑液", "濕滑", "濡濕", "潤澤", "水淋淋", "濕漉漉",
			"精華", "種子", "生命之源", "男性精華", "女性蜜液", "愛之甘露",
			"高潮液", "潮水", "愛河", "春水", "香汗", "體香", "濃郁", "腥甜",

			// 新增：更多性行為動作詞彙
			"狂野", "瘋狂做愛", "激烈", "猛烈", "狠狠", "用力", "深深",
			"頂弄", "頂撞", "撞擊", "衝刺", "猛攻", "攻城掠地", "征服",
			"律動", "節拍", "韻律", "旋律", "起伏", "波動", "震動", "顫動",
			"摸索", "撫慰", "安撫", "挑逗", "調戲", "戲弄", "撩撥", "煽情",
			"品味", "享用", "品嚐", "吞噬", "消化", "吸納", "融入", "結合",
			"緊緊抱住", "死死纏住", "牢牢鎖住", "深深擁抱", "緊緊相擁",

			// 新增：更多聲音表達（啊啊啊、嗯嗯等）
			"啊啊", "啊啊啊", "啊啊啊啊", "嗯嗯", "嗯嗯嗯", "呀呀", "呀呀呀",
			"喔喔", "喔喔喔", "唔唔", "唔唔唔", "哼哼", "哼哼哼", "嘶嘶",
			"咿咿", "咿咿呀呀", "咿呀", "哎呀", "哎喲", "哇啊", "哇哇",
			"好棒", "好爽", "好舒服", "好刺激", "好興奮", "好滿足", "好幸福",
			"快要", "就快", "馬上", "立刻", "瞬間", "突然", "猛然",
			"忍不住", "控制不住", "失控", "瘋狂", "迷亂", "神魂顛倒",
			"浪叫", "嬌吟", "呻吟聲", "喘息聲", "嘆息聲", "低吟聲", "輕哼聲",
			"連連叫喊", "不住呻吟", "忍不住叫出聲", "甜美叫聲", "嬌媚聲音",

			// 新增：身體反應描述
			"酥軟", "無力", "癱軟", "虛脫", "精疲力竭", "筋疲力盡",
			"渾身發軟", "雙腿發抖", "身體顫抖", "止不住顫抖", "劇烈顫抖",
			"心跳加速", "呼吸急促", "氣喘吁吁", "大口喘氣", "急促呼吸",
			"面紅耳赤", "滿臉通紅", "羞紅臉頰", "嬌羞如花", "媚眼如絲",
			"眼神迷離", "雙眼朦朧", "眼波流轉", "春水盈盈", "水汪汪",
			"汗水淋漓", "香汗淋漓", "汗如雨下", "大汗淋漓", "汗珠滾滾",
			"渾身是汗", "汗水濕透", "汗濕衣衫", "汗水晶瑩", "汗珠閃閃",

			// 英文明確詞彙
			"sex", "seggs", "fuck", "fucking", "bang", "screw", "cum", "cumming", "orgasm", "climax",
			"penetrate", "penetration", "naked", "nude", "nsfw",
			"penis", "vagina", "breast", "boobs", "nipple", "areola", "pussy", "cock", "dick", "ass",
			"butt", "booty", "wet", "hard", "horny", "moan", "pleasure", "stimulate", "sensitive",
			"bj", "hj", "blowjob", "handjob", "doggy", "missionary", "cowgirl", "69", "deepthroat",

			// 新增英文激進詞彙
			"thrust", "pound", "ram", "drill", "pump", "stroke", "grind", "ride",
			"juicy", "slick", "dripping", "soaked", "throbbing", "pulsing", "swollen",
			"gasp", "pant", "whimper", "whine", "cry out", "scream", "ahh", "ohh", "mmm",
		},
		extremeKeywords: []string{
			// 極度明確的動作詞彙（Level 5 專用 - 大幅擴充）
			"狂操", "猛插", "爆射", "內射", "肛交", "深喉", "顏射",
			"群交", "3P", "4P", "多人", "輪", "輪流", "輪J", "輪奸",
			"調教", "綁縛", "捆綁", "SM", "主奴", "支配", "臣服", "羞辱", "窒息",
			"潮吹", "失禁", "痙攣", "瘋狂", "放蕩", "淫蕩", "騷", "賤",

			// 新增：更極端的動作描述
			"狂暴", "野獸般", "像野獸一樣", "不要命地", "拼命地", "瘋狂地",
			"狠命", "死命", "拼了命", "不顧一切", "歇斯底里", "失去理智",
			"蹂躪", "摧殘", "征服", "占有", "霸占", "奪取", "掠奪", "侵犯",
			"狂歡", "縱慾", "放縱", "沉淪", "墮落", "迷失", "沉溺", "著迷",
			"榨乾", "耗盡", "吸乾", "榨取", "消耗", "透支", "極限", "巔峰",
			"爆炸", "炸裂", "崩潰", "決堤", "失守", "潰堤", "爆發", "噴發",
			"狂噴", "狂射", "狂洩", "狂流", "狂瀉", "連續射精", "多次高潮",

			// 新增：極度聲音描述
			"啊啊啊啊啊", "嗯嗯嗯嗯", "呀呀呀呀", "哼哼哼哼", "喔喔喔喔",
			"狂叫", "瘋狂叫喊", "撕心裂肺", "聲嘶力竭", "叫個不停",
			"淫叫連連", "浪叫不止", "嬌喘如雷", "呻吟如歌", "聲音嘶啞",
			"叫到失聲", "喊破嗓子", "叫得淒厲", "慘叫連天", "哀求不止",
			"我要", "我想要", "給我", "快給我", "更用力", "更深一點",
			"不要停", "繼續", "再來", "還要", "不夠", "還不夠", "更多",
			"求你了", "拜托", "饒了我", "受不了", "太激烈了", "要瘋了",

			// 新增：極端身體狀態
			"崩潰", "徹底崩潰", "完全失控", "神志不清", "意識模糊",
			"昏天暗地", "天旋地轉", "暈頭轉向", "不省人事", "渾身痙攣",
			"劇烈抽搐", "不斷顫抖", "止不住抖", "抖個不停", "抖成篩子",
			"癱在床上", "軟如爛泥", "動彈不得", "四肢無力", "渾身酥軟",

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

	// Level 5: 極度明確內容 或 含違法類（進一步降低門檻）
	if illegal >= 1 || extreme >= 1 || (explicit >= 1) {
		level = 5
		categories = []string{"extreme", "explicit", "nsfw"}
		if illegal >= 1 {
			categories = append(categories, "illegal") // 標註違法風險
		}
		isNSFW = true
		confidence = 0.95
		shouldUseGrok = true
		// Level 4: 明確成人內容（大幅降低門檻）
	} else if intimate >= 2 || (intimate >= 1 && romantic >= 1) || fetish >= 1 || roleplay >= 1 {
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
		shouldUseGrok = true // Level 4 使用 Grok 處理明確成人內容
		// Level 3: 親密內容（大幅降低門檻）
	} else if intimate >= 1 || (romantic >= 1) {
		level = 3
		categories = []string{"intimate", "nsfw", "suggestive"}
		isNSFW = true
		confidence = 0.85
		shouldUseGrok = false
		// Level 2: 浪漫暗示（降低門檻）
	} else if romantic >= 1 {
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

	// 特殊調整：進一步優化級別判定
	if extreme >= 1 || illegal >= 1 {
		level = 5
		shouldUseGrok = true
		confidence = 0.95
	} else if explicit >= 1 {
		level = 5 // explicit 關鍵詞直接進入 Level 5
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
