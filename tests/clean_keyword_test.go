package tests

import (
	"strings"
	"testing"
	"time"
)

// CleanKeywordClassifier 乾淨的擴展關鍵字分類器（無重複）
type CleanKeywordClassifier struct {
	keywords map[string]int
}

func NewCleanKeywordClassifier() *CleanKeywordClassifier {
	// 建立無重複的全面關鍵字庫
	keywords := map[string]int{
		// L5 - 極度露骨 (30個)
		"插進": 5, "抽插": 5, "內射": 5, "射精": 5, "陰莖": 5,
		"陰道": 5, "龜頭": 5, "肉棒": 5, "濕穴": 5, "潮吹": 5,
		"高潮": 5, "精液": 5, "陰唇": 5, "陰蒂": 5, "G點": 5,
		"前列腺": 5, "射出": 5, "噴出": 5, "填滿": 5, "撞擊": 5,
		"深入": 5, "頂到": 5, "戳到": 5, "撐開": 5, "夾緊": 5,
		"收縮": 5, "銷魂": 5, "極樂": 5, "爽死": 5, "淫蕩": 5,

		// L4 - 明確性行為 (40個)
		"做愛": 4, "性愛": 4, "上床": 4, "親熱": 4, "舌吻": 4,
		"深吻": 4, "挑逗": 4, "勾引": 4, "調情": 4, "性交": 4,
		"交配": 4, "雲雨": 4, "魚水": 4, "房事": 4, "同房": 4,
		"激情": 4, "熱情": 4, "慾火": 4, "慾望": 4, "性慾": 4,
		"興奮": 4, "亢奮": 4, "激動": 4, "衝動": 4, "渴望": 4,
		"飢渴": 4, "難耐": 4, "忍不住": 4, "控制不住": 4, "發春": 4,
		"發情": 4, "動情": 4, "情動": 4, "春心": 4, "春意": 4,
		"火熱": 4, "燃燒": 4, "沸騰": 4, "滾燙": 4, "炙熱": 4,

		// L3 - 親密接觸 (50個)
		"親吻": 3, "擁抱": 3, "撫摸": 3, "身體": 3, "脫衣": 3,
		"裸體": 3, "胸部": 3, "乳房": 3, "大腿": 3, "腰部": 3,
		"背部": 3, "肩膀": 3, "手臂": 3, "脖子": 3, "腹部": 3,
		"臀部": 3, "撫弄": 3, "觸摸": 3, "懷抱": 3, "抱住": 3,
		"摟抱": 3, "環抱": 3, "吻": 3, "香吻": 3, "熱吻": 3,
		"輕吻": 3, "床上": 3, "臥室": 3, "浴室": 3, "洗澡": 3,
		"躺下": 3, "靠近": 3, "貼近": 3, "依偎": 3, "相擁": 3,
		"緊貼": 3, "溫柔": 3, "柔情": 3, "溫存": 3, "性感": 3,
		"誘人": 3, "魅惑": 3, "撩人": 3, "勾人": 3, "曲線": 3,
		"身材": 3, "體態": 3, "肌膚": 3, "體溫": 3, "體香": 3,

		// L2 - 情感親密 (80個)
		"喜歡": 2, "愛": 2, "愛你": 2, "愛上": 2, "深愛": 2,
		"約會": 2, "邀約": 2, "邀請": 2, "赴約": 2, "約定": 2,
		"戀人": 2, "情人": 2, "愛人": 2, "戀愛": 2, "戀情": 2,
		"情侶": 2, "伴侶": 2, "另一半": 2, "對象": 2, "心上人": 2,
		"男朋友": 2, "女朋友": 2, "男友": 2, "女友": 2, "表白": 2,
		"告白": 2, "示愛": 2, "示意": 2, "暗示": 2, "透露": 2,
		"交往": 2, "在一起": 2, "相處": 2, "相伴": 2, "陪伴": 2,
		"關係": 2, "感情": 2, "情感": 2, "心動": 2, "動心": 2,
		"心跳": 2, "心悸": 2, "怦然": 2, "想念": 2, "思念": 2,
		"懷念": 2, "掛念": 2, "惦記": 2, "美麗": 2, "美貌": 2,
		"漂亮": 2, "好看": 2, "傾國傾城": 2, "迷人": 2, "迷戀": 2,
		"著迷": 2, "癡迷": 2, "沉迷": 2, "魅力": 2, "吸引": 2,
		"甜蜜": 2, "甜美": 2, "溫馨": 2, "浪漫": 2, "羅曼蒂克": 2,
		"幸福": 2, "快樂": 2, "開心": 2, "歡樂": 2, "愉快": 2,
		"親密": 2, "密切": 2, "貼心": 2, "體貼": 2, "細心": 2,
		"深情": 2, "情深": 2, "情意": 2, "真情": 2, "真愛": 2,

		// L1 - 安全關鍵字 (用於避免誤判)
		"心情": 1, "情緒": 1, "感受": 1, "工作": 1, "學習": 1,
		"天氣": 1, "食物": 1, "電影": 1, "音樂": 1, "旅行": 1,
		"家人": 1, "朋友": 1, "健康": 1, "運動": 1, "休息": 1,
		"疲累": 1, "累": 1, "難過": 1, "傷心": 1, "煩惱": 1,
	}

	return &CleanKeywordClassifier{keywords: keywords}
}

func (c *CleanKeywordClassifier) ClassifyContent(message string) *SimpleClassificationResult {
	start := time.Now()

	normalized := strings.ToLower(strings.TrimSpace(message))
	maxLevel := 1
	matchedKeywords := []string{}

	// 檢查所有關鍵字
	for keyword, level := range c.keywords {
		if strings.Contains(normalized, keyword) {
			matchedKeywords = append(matchedKeywords, keyword)
			if level > maxLevel {
				maxLevel = level
			}
		}
	}

	duration := time.Since(start)

	return &SimpleClassificationResult{
		Level:      maxLevel,
		Confidence: calculateCleanConfidence(len(matchedKeywords), maxLevel),
		Reason:     strings.Join(matchedKeywords, ","),
		Duration:   duration,
	}
}

func calculateCleanConfidence(matchCount, level int) float64 {
	// 基於匹配關鍵字數量和等級計算信心度
	baseConfidence := map[int]float64{
		1: 0.6, 2: 0.75, 3: 0.85, 4: 0.92, 5: 0.97,
	}

	confidence := baseConfidence[level]

	// 多關鍵字匹配增加信心度
	if matchCount > 1 {
		confidence += float64(matchCount-1) * 0.03
		if confidence > 0.99 {
			confidence = 0.99
		}
	}

	return confidence
}

func (c *CleanKeywordClassifier) GetKeywordCount() int {
	return len(c.keywords)
}

func (c *CleanKeywordClassifier) GetKeywordsByLevel(level int) []string {
	var keywords []string
	for keyword, l := range c.keywords {
		if l == level {
			keywords = append(keywords, keyword)
		}
	}
	return keywords
}

// 對比兩種分類方法的核心測試
func TestKeywordVsEmbeddingComparison(t *testing.T) {
	clean := NewCleanKeywordClassifier()

	t.Logf("=== 關鍵字分類器規模 ===")
	t.Logf("總關鍵字數: %d", clean.GetKeywordCount())

	for level := 1; level <= 5; level++ {
		keywords := clean.GetKeywordsByLevel(level)
		t.Logf("  L%d: %d 個關鍵字", level, len(keywords))
		if level <= 3 && len(keywords) <= 10 {
			t.Logf("    示例: %v", keywords[:minLen(len(keywords), 5)])
		}
	}

	// 核心對比測試案例
	criticalTestCases := []struct {
		message          string
		embeddingResult  int // 從真實測試獲得
		expectedCorrect  int // 人工判斷的正確答案
		description      string
	}{
		{"今天天氣真好", 1, 1, "安全內容"},
		{"心情不太好", 1, 1, "情緒支持（embedding正確）"},
		{"成為我的女朋友吧", 2, 2, "情感表白（embedding正確）"},
		{"想擁抱你", 2, 3, "親密行為（embedding誤判）"},
		{"想和你做愛", 2, 4, "明確性行為（embedding嚴重誤判）"},

		// 額外測試案例
		{"我深深愛著你", 0, 2, "情感深度"},
		{"撫摸你的臉頰", 0, 3, "親密撫摸"},
		{"感到很興奮", 0, 4, "性暗示"},
		{"插進你身體", 0, 5, "露骨表達"},
		{"你的身材很好", 0, 3, "身體讚美"},
	}

	t.Logf("\n=== 準確性對比測試 ===")

	var keywordCorrect, embeddingCorrect int
	var keywordResults []int

	for i, test := range criticalTestCases {
		result := clean.ClassifyContent(test.message)
		keywordResults = append(keywordResults, result.Level)

		keywordMatch := result.Level == test.expectedCorrect
		embeddingMatch := test.embeddingResult == test.expectedCorrect

		if keywordMatch {
			keywordCorrect++
		}
		if embeddingMatch && test.embeddingResult > 0 {
			embeddingCorrect++
		}

		status := ""
		if keywordMatch && embeddingMatch {
			status = "✓✓ 都正確"
		} else if keywordMatch && !embeddingMatch {
			status = "✓✗ 關鍵字勝"
		} else if !keywordMatch && embeddingMatch {
			status = "✗✓ Embedding勝"
		} else {
			status = "✗✗ 都錯誤"
		}

		t.Logf("案例 %d: %s", i+1, test.message)
		t.Logf("  預期: L%d (%s)", test.expectedCorrect, test.description)
		t.Logf("  關鍵字: L%d | Embedding: L%d | %s",
			result.Level, test.embeddingResult, status)
		if len(result.Reason) > 0 {
			t.Logf("  匹配詞: %s", result.Reason)
		}
		t.Logf("")
	}

	// 統計準確率
	embeddingTestCount := 5 // 只有前5個有embedding結果
	total := len(criticalTestCases)

	t.Logf("=== 準確率統計 ===")
	t.Logf("關鍵字分類器: %d/%d (%.1f%%)",
		keywordCorrect, total, float64(keywordCorrect)/float64(total)*100)
	t.Logf("Embedding分類器: %d/%d (%.1f%%)",
		embeddingCorrect, embeddingTestCount, float64(embeddingCorrect)/float64(embeddingTestCount)*100)

	// 效能測試
	t.Logf("\n=== 效能測試 ===")
	testRounds := 1000

	start := time.Now()
	for round := 0; round < testRounds; round++ {
		for _, test := range criticalTestCases {
			clean.ClassifyContent(test.message)
		}
	}
	totalTime := time.Since(start)

	totalTests := testRounds * len(criticalTestCases)
	avgTime := float64(totalTime.Nanoseconds()) / float64(totalTests) / 1000

	t.Logf("關鍵字分類器 (%d關鍵字):", clean.GetKeywordCount())
	t.Logf("  平均延遲: %.2f μs", avgTime)
	t.Logf("  處理量: %.0f 次/秒", float64(totalTests)/totalTime.Seconds())

	// 與embedding對比
	embeddingAvgTime := 806573.0 // 從真實測試獲得
	speedup := embeddingAvgTime / avgTime

	t.Logf("\nEmbedding對比:")
	t.Logf("  Embedding平均延遲: %.2f μs", embeddingAvgTime)
	t.Logf("  關鍵字速度優勢: %.0fx 更快", speedup)

	// 成本分析
	t.Logf("\n=== 成本分析 ===")
	dailyMessages := 10000
	embeddingDailyCost := float64(dailyMessages) * 0.0018
	keywordDailyCost := 0.0

	t.Logf("每日 %d 訊息成本:", dailyMessages)
	t.Logf("  Embedding: $%.2f", embeddingDailyCost)
	t.Logf("  關鍵字: $%.2f", keywordDailyCost)
	t.Logf("  年節省: $%.0f", embeddingDailyCost*365)
}

// 擴展關鍵字覆蓋率測試
func TestKeywordCoverage(t *testing.T) {
	clean := NewCleanKeywordClassifier()

	// 測試各種邊界表達
	edgeExpressions := []struct {
		message  string
		expected int
		reason   string
	}{
		// 語意暗示（embedding可能更好）
		{"想要更親近你", 3, "親密暗示"},
		{"感受你的溫暖", 2, "情感親密"},
		{"希望我們關係更進一步", 2, "關係發展"},

		// 明確表達（關鍵字應該更好）
		{"摸摸你的胸部", 3, "明確身體接觸"},
		{"在床上抱著你", 3, "明確親密場景"},
		{"激情地親吻", 4, "明確激情行為"},

		// 複合表達
		{"心情不好想要你的擁抱", 3, "情緒+親密"},
		{"今天很累想和你一起洗澡", 3, "日常+親密"},
		{"工作壓力大想和你做愛", 4, "壓力+性行為"},
	}

	t.Logf("=== 關鍵字覆蓋率測試 ===")

	var correctCount int
	for i, test := range edgeExpressions {
		result := clean.ClassifyContent(test.message)
		correct := result.Level == test.expected

		if correct {
			correctCount++
		}

		status := map[bool]string{true: "✓", false: "✗"}[correct]
		t.Logf("案例 %d: %s", i+1, test.message)
		t.Logf("  預期: L%d (%s) | 實際: L%d | %s",
			test.expected, test.reason, result.Level, status)
		if len(result.Reason) > 0 {
			t.Logf("  匹配關鍵字: %s", result.Reason)
		}
		t.Logf("")
	}

	accuracy := float64(correctCount) / float64(len(edgeExpressions)) * 100
	t.Logf("關鍵字覆蓋率: %d/%d (%.1f%%)", correctCount, len(edgeExpressions), accuracy)
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}