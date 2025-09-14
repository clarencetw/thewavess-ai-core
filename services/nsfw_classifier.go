package services

import (
    "context"
    "regexp"
    "strings"
    "time"

    "github.com/clarencetw/thewavess-ai-core/utils"
    "github.com/sirupsen/logrus"
)

// NSFWClassifier 關鍵字加權式 NSFW 內容分級器（支持邊界與情境抑制）
type NSFWClassifier struct {
    rules            []keywordRule
    suppressors      []suppressRule
    underageProxRe   *regexp.Regexp // 未成年人與性相關詞近距離匹配
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

// NewNSFWClassifier 創建新的關鍵字加權分級器
func NewNSFWClassifier() *NSFWClassifier {
    c := &NSFWClassifier{}
    c.initRules()

    utils.Logger.WithFields(logrus.Fields{
        "method": "keyword_based_weighted",
        "type":   "keyword_classifier_with_context",
        "rules":  len(c.rules),
    }).Info("NSFW關鍵字分級器初始化")

    return c
}

// ClassifyContent 使用關鍵字分析內容並返回 NSFW 等級
func (c *NSFWClassifier) ClassifyContent(ctx context.Context, message string) (*ClassificationResult, error) {
    startTime := time.Now()

    preview := message
    if len(preview) > 30 {
        preview = preview[:30]
    }
    utils.Logger.WithField("message_preview", preview).Info("開始關鍵字 NSFW 分級分析")

    // 加權規則分級
    result := c.classifyByKeywords(message)

    duration := time.Since(startTime)
    utils.Logger.WithFields(logrus.Fields{
        "level":      result.Level,
        "confidence": result.Confidence,
        "reason":     result.Reason,
        "duration":   duration,
        "method":     "keywords_weighted",
    }).Info("關鍵字 NSFW 分級完成")

    return result, nil
}

// initRules 初始化規則、抑制詞與近距離模式
func (c *NSFWClassifier) initRules() {
    // 英文使用單字邊界，避免 "Sussex" 誤觸
    wb := func(w string) *regexp.Regexp { return regexp.MustCompile(`(?i)\b` + w + `\b`) }
    // 中文/混合模式（允許簡單空白）
    rx := func(p string) *regexp.Regexp { return regexp.MustCompile(p) }

    c.rules = []keywordRule{
        // 硬性觸發：性暴力/非自願/亂倫
        {pattern: rx(`(?i)強姦|強暴|強奸|迷姦|迷奸|下藥.*(性|上床|侵犯)|非自願|亂倫|近親`), weight: 100, category: "act", reason: "sexual_violence_or_incest", hard: true},
        {pattern: wb("rape"), weight: 100, category: "act", reason: "rape", hard: true},
        {pattern: rx(`(?i)兒童色情|未成年人?猥褻`), weight: 100, category: "act", reason: "child_exploitation", hard: true},

        // 明確行為（單獨即 L5）
        {pattern: rx(`(?i)口\s*交|肛\s*交|乳\s*交|性交|性行為|做愛|內射|外射|顏射|射精|潮吹|手\s*交|足\s*交`), weight: 8, category: "act", reason: "explicit_act"},
        {pattern: wb("blowjob"), weight: 8, category: "act", reason: "explicit_act_en"},
        {pattern: wb("anal"), weight: 8, category: "act", reason: "explicit_act_en"},
        {pattern: wb("69"), weight: 8, category: "act", reason: "explicit_act_en"},

        // 明確部位/裸體
        {pattern: rx(`(?i)陰莖|陽具|龜頭|睪丸|陰道|陰蒂|乳頭|乳暈|下體|私處|陰部`), weight: 3, category: "body", reason: "explicit_body"},
        {pattern: rx(`(?i)裸體|裸露|全裸|半裸|脫光`), weight: 3, category: "nudity", reason: "nudity"},
        {pattern: wb("nude"), weight: 3, category: "nudity", reason: "nudity_en"},

        // 色情場景/內容
        {pattern: rx(`(?i)色情|情色|A片|成人片|床戲|群交|3P|調教|SM`), weight: 3, category: "context", reason: "porn_context"},
        {pattern: wb("porn"), weight: 3, category: "context", reason: "porn_en"},

        // 委婉/變體
        {pattern: rx(`(?i)上床|滾床單|打炮|開車|車速很快|做那件事`), weight: 2, category: "euphemism", reason: "euphemism"},
        {pattern: wb("sex"), weight: 2, category: "context", reason: "sex_en"},
    }

    // 抑制詞：醫療/藝術/教育語境（降低部位/裸體/一般語境分數）
    c.suppressors = []suppressRule{
        {pattern: rx(`(?i)醫學|醫院|臨床|診所|手術|檢查|X光|乳房攝影|乳房X光|乳房超音波|腫瘤|發炎|解剖學|生殖健康`), category: "medical"},
        {pattern: rx(`(?i)美術|藝術|雕像|雕塑|裸體藝術|人體素描|人體寫生|畫室|畫展|博物館|攝影展`), category: "art"},
        {pattern: rx(`(?i)教育|教材|課程|講座|性教育|諮商|課堂|報告`), category: "education"},
    }

    // 未成年人 x 性 近距離（雙向）
    c.underageProxRe = rx(`(?i)(未成年人?|兒童).{0,30}(性交|做愛|性|口\s*交|裸|猥褻|侵犯)|(性交|做愛|性|口\s*交|裸|猥褻|侵犯).{0,30}(未成年人?|兒童)`)
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
    cleaned = strings.ReplaceAll(cleaned, "pr0n", "porn")
    return cleaned
}

// classifyByKeywords 基於加權規則 + 抑制詞 + 近距離硬觸發
func (c *NSFWClassifier) classifyByKeywords(message string) *ClassificationResult {
    msg := c.normalize(message)

    // 1) 未成年近距離或硬規則：直接 L5
    if c.underageProxRe.MatchString(msg) {
        return &ClassificationResult{Level: 5, Confidence: 0.98, Reason: "underage_proximity"}
    }
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
    explicitActHit := false

    for _, r := range c.rules {
        if r.hard {
            continue
        }
        if r.pattern.MatchString(msg) {
            w := r.weight
            if inSuppressCtx {
                // 降低在醫療/藝術/教育情境下的誤觸
                switch r.category {
                case "body", "nudity":
                    if w > 2 {
                        w -= 2
                    } else {
                        w = 0
                    }
                case "context":
                    if w > 1 {
                        w -= 1
                    } else {
                        w = 0
                    }
                }
            }

            if w > 0 {
                score += w
                if w > maxWeight {
                    maxWeight = w
                    strongestReason = r.reason
                }
                if r.category == "act" {
                    explicitActHit = true
                }
            }
        }
    }

    // 4) 簡化為 2 級分級映射（適配系統實際使用）
    // 類別分數說明：
    // - 明確行為：8（單獨即 L5）
    // - 明確部位/裸體/色情場景：3+
    // - 委婉語/一般語境：2+
    level := 1
    if explicitActHit || score >= 2 {
        level = 5  // 任何 NSFW 內容都歸為 Level 5
    } else {
        level = 1  // 安全內容為 Level 1
    }

    // 信心估計：簡化為 2 級系統
    confidence := 0.99 // 安全內容信心很高
    if level == 5 {
        if explicitActHit || score >= 8 {
            confidence = 0.96 // 明確 NSFW 內容信心高
        } else {
            confidence = 0.80 // 一般 NSFW 內容信心中等
        }
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
