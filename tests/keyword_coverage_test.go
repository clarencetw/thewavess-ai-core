package tests

import (
	"testing"

	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// 測試關鍵字覆蓋度和衝突情況
func TestKeywordCoverageAndConflicts(t *testing.T) {
	utils.InitLogger()
	utils.LoadEnv()

	t.Logf("=== 關鍵字覆蓋度與衝突分析 ===")

	classifier := services.NewEnhancedKeywordClassifier()

	// 1. 測試邊界詞彙（容易誤判的）
	boundaryTests := []struct {
		message  string
		expected int
		category string
	}{
		// 日常詞彙可能被誤判
		{"愛吃水果", 2, "包含'愛'但非情感"},
		{"愛護環境", 2, "包含'愛'但非情感"},
		{"身體健康", 3, "包含'身體'但非親密"},
		{"溫柔的貓", 3, "包含'溫柔'但非親密"},
		{"心情激動", 4, "包含'激動'但非性行為"},
		{"興奮地跳舞", 4, "包含'興奮'但非性行為"},

		// 組合詞可能產生誤判
		{"深入了解", 5, "包含'深入'但非露骨"},
		{"填滿空白", 5, "包含'填滿'但非露骨"},
		{"撞擊聲音", 5, "包含'撞擊'但非露骨"},

		// 正常情感表達
		{"我愛媽媽", 2, "正常親情"},
		{"愛學習愛工作", 2, "正常表達"},
		{"身體不舒服", 3, "醫療相關"},

		// 真正的分級內容
		{"想跟你約會", 2, "真正L2"},
		{"想要擁抱你", 3, "真正L3"},
		{"想要做愛", 4, "真正L4"},
		{"想要插進去", 5, "真正L5"},
	}

	var missClassified int
	for i, test := range boundaryTests {
		result, err := classifier.ClassifyContent(test.message)
		if err != nil {
			t.Errorf("案例 %d 分類失敗: %v", i+1, err)
			continue
		}

		correct := result.Level == test.expected
		if !correct {
			missClassified++
			t.Logf("❌ 案例 %d: '%s' (%s)", i+1, test.message, test.category)
			t.Logf("    預期: L%d | 實際: L%d | 原因: %s",
				test.expected, result.Level, result.Reason)
		} else {
			t.Logf("✅ 案例 %d: '%s' (%s)", i+1, test.message, test.category)
		}
	}

	// 2. 測試常見場景
	commonScenarios := []struct {
		message  string
		expected int
		scenario string
	}{
		// 日常對話
		{"今天心情很好", 1, "日常"},
		{"工作很累想休息", 1, "日常"},
		{"天氣真的很熱", 1, "日常"},

		// 友情
		{"你是我最好的朋友", 1, "友情"},
		{"想念朋友了", 2, "友情但包含'想念'"},

		// 親情
		{"愛爸爸媽媽", 2, "親情但包含'愛'"},
		{"想要回家抱抱家人", 3, "親情但包含'抱'"},

		// 戀愛
		{"我喜歡你", 2, "戀愛L2"},
		{"想和你在一起", 2, "戀愛L2"},
		{"想要親親你", 3, "戀愛L3"},

		// 性暗示
		{"想要親密接觸", 3, "性暗示L3"},
		{"想要更進一步", 2, "性暗示L2"},

		// 明確性內容
		{"想跟你上床", 4, "明確L4"},
		{"想要激情纏綿", 4, "明確L4"},

		// 露骨內容
		{"想要插入你", 5, "露骨L5"},
		{"讓我射進去", 5, "露骨L5"},
	}

	var commonMissed int
	for i, test := range commonScenarios {
		result, err := classifier.ClassifyContent(test.message)
		if err != nil {
			t.Errorf("常見案例 %d 分類失敗: %v", i+1, err)
			continue
		}

		correct := result.Level == test.expected
		if !correct {
			commonMissed++
			t.Logf("❌ 常見案例 %d: '%s' (%s)", i+1, test.message, test.scenario)
			t.Logf("    預期: L%d | 實際: L%d", test.expected, result.Level)
		}
	}

	// 3. 分析結果
	totalBoundary := len(boundaryTests)
	totalCommon := len(commonScenarios)
	boundaryAccuracy := float64(totalBoundary-missClassified) / float64(totalBoundary) * 100
	commonAccuracy := float64(totalCommon-commonMissed) / float64(totalCommon) * 100

	t.Logf("\n=== 覆蓋度分析結果 ===")
	t.Logf("邊界詞彙準確率: %d/%d (%.1f%%)", totalBoundary-missClassified, totalBoundary, boundaryAccuracy)
	t.Logf("常見場景準確率: %d/%d (%.1f%%)", totalCommon-commonMissed, totalCommon, commonAccuracy)

	// 4. 關鍵字統計
	info := classifier.GetClassifierInfo()
	if stats, ok := info["level_stats"].(map[int]int); ok {
		t.Logf("\n=== 關鍵字分布 ===")
		total := 0
		for level, count := range stats {
			t.Logf("L%d: %d個關鍵字", level, count)
			total += count
		}
		t.Logf("總計: %d個關鍵字", total)
	}

	// 5. 建議
	t.Logf("\n=== 改進建議 ===")
	if boundaryAccuracy < 80 {
		t.Logf("⚠️  邊界詞彙準確率過低，需要增加更多上下文判斷")
	}
	if commonAccuracy < 90 {
		t.Logf("⚠️  常見場景準確率不足，需要擴充關鍵字庫")
	}

	t.Logf("📈 建議擴充方向:")
	t.Logf("   1. 增加更多L1安全詞彙（防誤判）")
	t.Logf("   2. 增加情境組合判斷（如'身體健康' vs '身體接觸'）")
	t.Logf("   3. 考慮詞彙權重和組合規則")
	t.Logf("   4. 建議關鍵字總數擴充到500-800個")
}

// 測試關鍵字衝突情況
func TestKeywordConflicts(t *testing.T) {
	utils.InitLogger()
	utils.LoadEnv()

	t.Logf("=== 關鍵字衝突檢測 ===")

	classifier := services.NewEnhancedKeywordClassifier()

	// 測試包含多等級關鍵字的句子
	conflictTests := []struct {
		message   string
		keywords  []string
		levels    []int
		expected  int
		reasoning string
	}{
		{
			"愛你的身體",
			[]string{"愛", "身體"},
			[]int{2, 3},
			3, // 應該取最高等級
			"包含L2和L3關鍵字，應取L3",
		},
		{
			"激動地跟朋友約會",
			[]string{"激動", "朋友", "約會"},
			[]int{4, 1, 2},
			4, // 取最高等級
			"包含L1、L2、L4關鍵字，應取L4",
		},
		{
			"想要深入了解你的心情",
			[]string{"深入", "心情"},
			[]int{5, 1},
			5, // 可能誤判
			"'深入'被誤判為L5，但語境是L1",
		},
		{
			"愛情讓人心動想要親近",
			[]string{"愛", "心動", "親近"},
			[]int{2, 2, 3},
			3,
			"多個情感詞彙組合",
		},
	}

	for i, test := range conflictTests {
		result, err := classifier.ClassifyContent(test.message)
		if err != nil {
			t.Errorf("衝突測試 %d 失敗: %v", i+1, err)
			continue
		}

		t.Logf("衝突案例 %d: '%s'", i+1, test.message)
		t.Logf("  包含關鍵字: %v (等級: %v)", test.keywords, test.levels)
		t.Logf("  預期: L%d | 實際: L%d | 匹配: %s",
			test.expected, result.Level, result.Reason)
		t.Logf("  分析: %s", test.reasoning)

		if result.Level != test.expected {
			t.Logf("  ❌ 衝突處理不當")
		} else {
			t.Logf("  ✅ 衝突處理正確")
		}
		t.Logf("")
	}
}