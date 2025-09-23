package tests

import (
	"os"
	"testing"
	"time"

	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// 限制測試次數的真實對比
func TestLimitedRealComparison(t *testing.T) {
	// 初始化環境
	utils.InitLogger()

	// 設定工作目錄到項目根目錄
	if err := os.Chdir("/Users/clarence/github/thewavess-ai-core"); err != nil {
		t.Fatalf("無法切換到項目目錄: %v", err)
	}

	utils.LoadEnv()

	// 只測試少量訊息，避免API成本
	testMessages := []string{
		"今天天氣真好",           // L1 安全
		"心情不太好",            // L1 邊界案例
		"成為我的女朋友吧",        // L2 情感
		"想擁抱你",             // L3 親密
		"想和你做愛",           // L4 性愛
	}

	// 只測試3輪，控制API調用次數
	testRounds := 3
	totalApiCalls := len(testMessages) * testRounds // 只有15次API調用

	t.Logf("=== 限制真實測試 (%d訊息 × %d輪 = %d次API調用) ===",
		len(testMessages), testRounds, totalApiCalls)

	// 真實embedding分類器
	embeddingClassifier := services.NewEnhancedKeywordClassifier()
	keywordClassifier := NewSimpleKeywordClassifier()


	// 測試embedding分類器（真實API）
	var embeddingTotalTime time.Duration
	var embeddingResults []int

	t.Logf("開始真實embedding測試...")
	start := time.Now()

	for round := 0; round < testRounds; round++ {
		for i, msg := range testMessages {
			roundStart := time.Now()
			result, err := embeddingClassifier.ClassifyContent(msg)
			roundDuration := time.Since(roundStart)

			embeddingTotalTime += roundDuration

			if err != nil {
				t.Errorf("Embedding分類失敗: %v", err)
				continue
			}

			if round == 0 { // 只記錄第一輪結果
				embeddingResults = append(embeddingResults, result.Level)
				t.Logf("  訊息 %d: '%s' -> L%d (耗時: %v)",
					i+1, msg, result.Level, roundDuration)
			}
		}
	}
	embeddingWallTime := time.Since(start)

	// 測試關鍵字分類器
	var keywordTotalTime time.Duration
	var keywordResults []int

	t.Logf("開始關鍵字分類器測試...")
	start = time.Now()

	for round := 0; round < testRounds; round++ {
		for i, msg := range testMessages {
			result := keywordClassifier.ClassifyContent(msg)
			keywordTotalTime += result.Duration

			if round == 0 { // 只記錄第一輪結果
				keywordResults = append(keywordResults, result.Level)
				t.Logf("  訊息 %d: '%s' -> L%d (耗時: %v)",
					i+1, msg, result.Level, result.Duration)
			}
		}
	}
	keywordWallTime := time.Since(start)

	// 結果分析
	t.Logf("\n=== 真實效能對比 ===")
	avgEmbeddingTime := float64(embeddingTotalTime.Nanoseconds()) / float64(totalApiCalls) / 1000
	avgKeywordTime := float64(keywordTotalTime.Nanoseconds()) / float64(totalApiCalls) / 1000

	t.Logf("Embedding分類器:")
	t.Logf("  總牆上時間: %v", embeddingWallTime)
	t.Logf("  平均延遲: %.2f μs", avgEmbeddingTime)
	t.Logf("  處理量: %.0f 次/秒", float64(totalApiCalls)/embeddingWallTime.Seconds())

	t.Logf("關鍵字分類器:")
	t.Logf("  總牆上時間: %v", keywordWallTime)
	t.Logf("  平均延遲: %.2f μs", avgKeywordTime)
	t.Logf("  處理量: %.0f 次/秒", float64(totalApiCalls)/keywordWallTime.Seconds())

	speedup := avgEmbeddingTime / avgKeywordTime
	t.Logf("\n關鍵字比embedding快: %.1fx", speedup)

	// 分類結果對比
	t.Logf("\n=== 分類結果對比 ===")
	for i, msg := range testMessages {
		match := keywordResults[i] == embeddingResults[i]
		status := map[bool]string{true: "一致", false: "差異"}[match]

		t.Logf("訊息: '%s'", msg)
		t.Logf("  關鍵字: L%d | Embedding: L%d | %s",
			keywordResults[i], embeddingResults[i], status)
	}

	// 計算一致性
	var agreements int
	for i := range testMessages {
		if keywordResults[i] == embeddingResults[i] {
			agreements++
		}
	}
	consistency := float64(agreements) / float64(len(testMessages)) * 100
	t.Logf("\n一致性: %.1f%% (%d/%d)", consistency, agreements, len(testMessages))

	// 成本分析
	embeddingCostPerMsg := 0.0018 // USD
	actualCost := float64(totalApiCalls) * embeddingCostPerMsg

	t.Logf("\n=== 成本分析 ===")
	t.Logf("本次測試成本: $%.4f (%d次API調用)", actualCost, totalApiCalls)
	t.Logf("關鍵字分類成本: $0.0000")

	// 推算大規模使用成本
	dailyMsgs := 1000
	monthlyEmbeddingCost := float64(dailyMsgs) * 30 * embeddingCostPerMsg
	t.Logf("假設每日%d訊息，embedding每月成本: $%.2f", dailyMsgs, monthlyEmbeddingCost)
}

// 專門測試邊界案例的精確度
func TestEdgeCaseAccuracy(t *testing.T) {
	utils.InitLogger()

	// 設定工作目錄到項目根目錄
	if err := os.Chdir("/Users/clarence/github/thewavess-ai-core"); err != nil {
		t.Fatalf("無法切換到項目目錄: %v", err)
	}

	utils.LoadEnv()

	// 專門選擇容易誤判的案例
	edgeCases := []struct {
		message  string
		expected int
		reason   string
	}{
		{"心情不太好", 1, "情緒支持，非NSFW"},
		{"成為我的女朋友吧", 2, "情感表白"},
		{"工作很累想要休息", 1, "日常對話"},
		{"想要你的擁抱", 3, "明確親密需求"},
		{"感到有點興奮", 1, "可能是工作興奮，非性暗示"},
	}

	embeddingClassifier := services.NewEnhancedKeywordClassifier()
	keywordClassifier := NewSimpleKeywordClassifier()

	t.Logf("=== 邊界案例精確度測試 (%d個案例) ===", len(edgeCases))

	var embeddingCorrect, keywordCorrect int

	for i, test := range edgeCases {
		// Embedding分類
		embResult, err := embeddingClassifier.ClassifyContent(test.message)
		if err != nil {
			t.Errorf("Embedding分類失敗: %v", err)
			continue
		}

		// 關鍵字分類
		keyResult := keywordClassifier.ClassifyContent(test.message)

		// 檢查正確性
		embCorrect := embResult.Level == test.expected
		keyCorrect := keyResult.Level == test.expected

		if embCorrect {
			embeddingCorrect++
		}
		if keyCorrect {
			keywordCorrect++
		}

		t.Logf("案例 %d: %s", i+1, test.message)
		t.Logf("  預期: L%d (%s)", test.expected, test.reason)
		t.Logf("  Embedding: L%d %s (信心: %.2f)",
			embResult.Level,
			map[bool]string{true: "✓", false: "✗"}[embCorrect],
			embResult.Confidence)
		t.Logf("  關鍵字: L%d %s",
			keyResult.Level,
			map[bool]string{true: "✓", false: "✗"}[keyCorrect])
		t.Logf("")
	}

	total := len(edgeCases)
	t.Logf("=== 邊界案例準確率總結 ===")
	t.Logf("Embedding: %d/%d (%.1f%%)", embeddingCorrect, total,
		float64(embeddingCorrect)/float64(total)*100)
	t.Logf("關鍵字: %d/%d (%.1f%%)", keywordCorrect, total,
		float64(keywordCorrect)/float64(total)*100)

	if embeddingCorrect > keywordCorrect {
		t.Logf("Embedding在邊界案例上表現更好")
	} else if keywordCorrect > embeddingCorrect {
		t.Logf("關鍵字在邊界案例上表現更好")
	} else {
		t.Logf("兩種方法在邊界案例上表現相當")
	}
}