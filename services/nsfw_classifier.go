package services

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// NSFWClassifier 智能關鍵字加權式 NSFW 內容分級器（支持邊界與情境抑制）
//
// 🎯 核心設計原則（台灣女性向系統雙引擎路由）：
// 1. 分級標準：L1(<2分) L2(≥2分) L3(≥4分) L4(≥6分) L5(≥10分)
// 2. 引擎路由：L1-L3 → OpenAI, L4-L5 → Grok，確保內容適配度
// 3. 中文優化：專門優化中文性愛詞彙分級準確度（如"幹我"、"操我"等）
// 4. 台灣法規：嚴格阻擋未成年相關內容，符合台灣法律要求
// 5. 情境抑制：醫療/藝術/教育語境自動降級，減少誤觸
// 6. Sticky Session：L4+ 觸發後 5 分鐘內保持 Grok 引擎，避免引擎跳動
//
// ⚠️ 重要：此分級器是雙引擎架構的核心，分級準確度直接影響用戶體驗
type NSFWClassifier struct {
	rules          []keywordRule
	suppressors    []suppressRule
	// 移除 underageProxRe - 不再處理未成年相關內容
}

type keywordRule struct {
	pattern  *regexp.Regexp
	weight   int
	category string // act, body, nudity, context, euphemism
	reason   string
	hard     bool // 命中即 L5
}

type suppressRule struct {
	pattern  *regexp.Regexp
	category string // medical, art, education
}

// ClassificationResult 分級結果
type ClassificationResult struct {
	Level      int     `json:"level"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// NewNSFWClassifier 創建新的智能關鍵字加權分級器
func NewNSFWClassifier() *NSFWClassifier {
	c := &NSFWClassifier{}
	c.initRules()

	utils.Logger.WithFields(logrus.Fields{
		"method": "intelligent_keyword_weighted",
		"type":   "advanced_nsfw_classifier",
		"rules":  len(c.rules),
	}).Info("智能 NSFW 關鍵字分級器初始化")

	return c
}

// ClassifyContent 使用智能關鍵字分析內容並返回 NSFW 等級
func (c *NSFWClassifier) ClassifyContent(ctx context.Context, message string) (*ClassificationResult, error) {
	startTime := time.Now()

	preview := message
	if len(preview) > 30 {
		preview = preview[:30]
	}
	utils.Logger.WithField("message_preview", preview).Info("開始智能關鍵字 NSFW 分級分析")

	// 智能加權規則分級
	result := c.classifyByKeywords(message)

	duration := time.Since(startTime)
	utils.Logger.WithFields(logrus.Fields{
		"level":      result.Level,
		"confidence": result.Confidence,
		"reason":     result.Reason,
		"duration":   duration,
		"method":     "intelligent_keywords_weighted",
	}).Info("智能關鍵字 NSFW 分級完成")

	return result, nil
}

// initRules 初始化完整的關鍵字規則、抑制詞與近距離模式
func (c *NSFWClassifier) initRules() {
	// 英文使用單字邊界，避免 "Sussex" 誤觸
	wb := func(w string) *regexp.Regexp { return regexp.MustCompile(`(?i)\b` + w + `\b`) }
	// 中文/混合模式（允許簡單空白）
	rx := func(p string) *regexp.Regexp { return regexp.MustCompile(p) }

    c.rules = []keywordRule{
		// 硬性觸發：性暴力/非自願/亂倫 (權重100 = 直接L5)
        {pattern: rx(`(?i)強姦|強暴|強奸|迷姦|迷奸|下藥.*(性|上床|侵犯)|性侵|性虐待|性剝削|非自願|逼迫.*(性|上床)|亂倫|近親|近親相姦|輪姦|輪奸`), weight: 100, category: "act", reason: "sexual_violence_or_incest", hard: true},
        {pattern: rx(`(?i)兄妹|姐弟|父女|母子|繼(兄|姐|弟|妹)`), weight: 100, category: "act", reason: "incest_family_roles", hard: true},
        {pattern: rx(`(?i)step\s*(brother|sister|mom|mother|dad|father)`), weight: 100, category: "act", reason: "incest_step_roles_en", hard: true},
        {pattern: wb("rape"), weight: 100, category: "act", reason: "rape", hard: true},
        {pattern: rx(`(?i)獸交|畜交|bestiality`), weight: 100, category: "act", reason: "bestiality", hard: true},
        // 台灣法律不合法：未成年性相關內容，一律不處理（硬性阻擋）
        {pattern: rx(`(?i)(未滿\s*(18|十八)歲|未成年(人|者)?|小學生|國中生|高中生).{0,12}(性|裸|性愛|做愛|性交|色情|猥褻|裸照|裸體|上床)`), weight: 100, category: "act", reason: "illegal_underage", hard: true},
        {pattern: rx(`(?i)蘿莉(控)?|ロリ|萝莉(控)?|loli(con)?`), weight: 100, category: "act", reason: "illegal_underage", hard: true},
        {pattern: rx(`(?i)(child\s*(porn(ography)?|sexual)|underage\s*(sex|nude|sexual)|minor\s*(sex|sexual))`), weight: 100, category: "act", reason: "illegal_underage_en", hard: true},

		// 明確性行為 (權重10 = L4-L5)
        {pattern: rx(`(?i)口\s*交|肛\s*交|乳\s*交|性交|性行為|做愛|內射|外射|顏射|射精|潮吹|手\s*交|足\s*交|中出|口爆|深喉|後\s*入|背後位|騎乘|女上位|自慰|手淫|自瀆|打手槍`), weight: 10, category: "act", reason: "explicit_sexual_act"},
        {pattern: rx(`(?i)\b(?:blowjob|handjob|footjob|anal|cumshot|creampie|deepthroat|doggystyle|cowgirl|rimming|fingering|masturbate|jerk\s*off|fap|bukkake|facial|double\s*penetration|dp|69|sixty[-\s]*nine)\b`), weight: 10, category: "act", reason: "explicit_sexual_act_en"},
        {pattern: rx(`(?i)舔陰|舔穴|吞精|榨精|含住|吸吮(乳頭|陰蒂)?|幹我|搞我|要我|上我|操我|插我|弄我|玩我|用我`), weight: 10, category: "act", reason: "explicit_sexual_act_zh_ext"},

		// 明確身體部位 (權重6 = L3-L4)
        {pattern: rx(`(?i)陰莖|陽具|龜頭|睪丸|陰道|陰蒂|花蒂|陰核|乳頭|乳暈|下體|私處|陰部|生殖器|肉棒|小穴|蜜穴|穴|菊花|肛門|陰毛|陰唇|奶子|咪咪|巨乳|酥胸|老二`), weight: 6, category: "body", reason: "explicit_body_parts"},
        {pattern: rx(`(?i)\b(?:penis|vagina|clitoris|nipples?|areolae?|genitals|pussy|cock|dick|boobs|tits|ass|butt(ocks)?|balls|testicles)\b`), weight: 6, category: "body", reason: "explicit_body_parts_en"},

		// 裸體相關 (權重5 = L3)
        {pattern: rx(`(?i)裸體|裸露|全裸|半裸|脫光|脫衣|脫褲|脫掉.*褲|脫下.*褲|走光|露點|透視裝|比基尼`), weight: 5, category: "nudity", reason: "nudity_content"},
        {pattern: rx(`(?i)\b(?:nude|naked|topless|undress|strip|see[-\s]*through|cleavage)\b`), weight: 5, category: "nudity", reason: "nudity_content_en"},

		// 色情場景/內容 (權重4 = L2-L3)
        {pattern: rx(`(?i)色情|情色|A片|AV|成人片|床戲|群交|3P|多P|調教|SM|BDSM|本子|H漫|H本|裡番|工口|18禁`), weight: 4, category: "context", reason: "porn_context"},
        {pattern: rx(`(?i)\b(?:porn|xxx|adult|hentai|lewd|nsfw|bdsm|threesome|onlyfans|fansly|pornhub|xvideos|xhamster|redtube|cam(girl|boy)?|cam4)\b`), weight: 4, category: "context", reason: "porn_context_en"},

		// 身體部位一般描述 (權重3 = L2)
        {pattern: rx(`(?i)胸部|乳房|臀部|大腿|腰部|身材|曲線|蜜桃臀|翹臀|馬甲線|川字腹`), weight: 3, category: "body", reason: "body_description"},
        {pattern: rx(`(?i)\b(?:breast|chest|thigh|curves|butt|hips|waist|cleavage)\b`), weight: 3, category: "body", reason: "body_description_en"},

		// 委婉/暗示 (權重2 = L1-L2)
        {pattern: rx(`(?i)上床|滾床單|打炮|打砲|約炮|約砲|約P|開車|車速很快|做那件事|親密|撫摸|愛撫|車震|開房(間)?|睡了她|辦事|發車|愛愛|ML|運動|溫存|纏綿|交歡|歡愛|魚水之歡|雲雨|巫山雲雨|春宵|洞房|滿足我|要你|想你|需要你|渴望你`), weight: 2, category: "euphemism", reason: "sexual_euphemism"},
        {pattern: rx(`(?i)\b(?:sex|sexy|intimate|seduce|tease|hook\s*up|smash|bang|netflix\s*and\s*chill|make\s*love|get\s*it\s*on|sleep\s*together)\b`), weight: 2, category: "euphemism", reason: "sexual_euphemism_en"},

		// 輕微暗示 (權重1 = L1)
        {pattern: rx(`(?i)誘惑|魅惑|性感|撒嬌|挑逗|親吻|舌吻|親熱|呻吟`), weight: 1, category: "suggestive", reason: "mild_suggestive"},
        {pattern: rx(`(?i)\b(?:flirt|charming|attractive|make\s*out|kiss|kissing|moan)\b`), weight: 1, category: "suggestive", reason: "mild_suggestive_en"},
    }

	// 抑制詞：醫療/藝術/教育語境（降低誤觸）
    c.suppressors = []suppressRule{
        {pattern: rx(`(?i)醫學|醫院|臨床|診所|手術|檢查|X光|乳房攝影|乳房X光|乳房超音波|腫瘤|發炎|解剖學|生殖健康|醫療`), category: "medical"},
        {pattern: rx(`(?i)美術|藝術|雕像|雕塑|裸體藝術|人體素描|人體寫生|畫室|畫展|博物館|攝影展|文學|小說`), category: "art"},
        {pattern: rx(`(?i)教育|教材|課程|講座|性教育|諮商|課堂|報告|學術|研究|科學|新聞|報導|紀錄片|維基|百科|政策|法規`), category: "education"},
    }

	// 移除未成年近距離檢測
    // c.underageProxRe = rx(`...`) // 已停用
}

// normalize 做基本正規化（小寫、移除零寬、簡單諧音）
func (c *NSFWClassifier) normalize(s string) string {
	lowered := strings.ToLower(s)
	// 移除零寬字元
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case 0x200B, 0x200C, 0x200D, 0xFEFF: // zero-width space/joiners
			return -1
		default:
			return r
		}
	}, lowered)
	// 常見 leetspeak/諧音
    cleaned = strings.ReplaceAll(cleaned, "seggs", "sex")
    cleaned = strings.ReplaceAll(cleaned, "s3x", "sex")
    cleaned = strings.ReplaceAll(cleaned, "s*x", "sex")
    cleaned = strings.ReplaceAll(cleaned, "pr0n", "porn")
    cleaned = strings.ReplaceAll(cleaned, "p0rn", "porn")
    return cleaned
}

// classifyByKeywords 基於智能加權規則 + 抑制詞 + 近距離硬觸發
func (c *NSFWClassifier) classifyByKeywords(message string) *ClassificationResult {
	msg := c.normalize(message)

	// 移除未成年檢測 - 不再處理此類內容
	for _, r := range c.rules {
		if r.hard && r.pattern.MatchString(msg) {
			return &ClassificationResult{Level: 5, Confidence: 0.99, Reason: r.reason}
		}
	}

	// 2) 抑制語境（醫療/藝術/教育）
	hasMedical, hasArt, hasEdu := false, false, false
	for _, srs := range c.suppressors {
		if srs.pattern.MatchString(msg) {
			switch srs.category {
			case "medical":
				hasMedical = true
			case "art":
				hasArt = true
			case "education":
				hasEdu = true
			}
		}
	}
	inSuppressCtx := hasMedical || hasArt || hasEdu

	// 3) 加權累計
	score := 0
	var strongestReason string
	maxWeight := 0

	for _, r := range c.rules {
		if r.hard {
			continue
		}
		if r.pattern.MatchString(msg) {
			w := r.weight
			if inSuppressCtx {
				// 在醫療/藝術/教育情境下降低權重
				switch r.category {
				case "body", "nudity":
					w = max(0, w-3) // 大幅降低
				case "context", "euphemism":
					w = max(0, w-2) // 中度降低
				case "suggestive":
					w = max(0, w-1) // 輕微降低
				}
			}

			if w > 0 {
				score += w
				if w > maxWeight {
					maxWeight = w
					strongestReason = r.reason
				}
			}
		}
	}

	// 4) 智能分級映射 (更精確的 L1-L5 分級)
	level := 1
	confidence := 0.99

	if score >= 10 {
		level = 5 // 明確性行為
		confidence = 0.95
	} else if score >= 6 {
		level = 4 // 明確身體部位
		confidence = 0.90
	} else if score >= 4 {
		level = 3 // 裸體/色情場景
		confidence = 0.85
	} else if score >= 2 {
		level = 2 // 身體描述/委婉語
		confidence = 0.80
	} else if score >= 1 {
		level = 1 // 輕微暗示，仍為安全
		confidence = 0.95
	}

	reason := strongestReason
	if reason == "" {
		if inSuppressCtx {
			reason = "suppressed_by_context"
		} else {
			reason = "safe"
		}
	}

	return &ClassificationResult{
		Level:      level,
		Confidence: confidence,
		Reason:     reason,
	}
}

