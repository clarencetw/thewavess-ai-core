package tests

import (
	"testing"
	"time"

	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// 整合測試：驗證關鍵字版NSFW分類器的實際效果
func TestKeywordNSFWIntegration(t *testing.T) {
	// 初始化環境
	utils.InitLogger()
	utils.LoadEnv()

	t.Logf("=== 關鍵字NSFW分類器整合測試 ===")

	// 獲取分類器（應該是關鍵字版本）
	classifier := services.NewEnhancedKeywordClassifier()


	// 測試案例
	testCases := []struct {
		message  string
		expected int
		desc     string
	}{
		{"今天天氣真好", 1, "安全內容"},
		{"心情不太好", 1, "情緒支持"},
		{"成為我的女朋友吧", 2, "情感表白"},
		{"想擁抱你", 3, "親密行為"},
		{"想和你做愛", 4, "明確性行為"},
		{"插進你身體", 5, "露骨表達"},
	}

	t.Logf("開始測試 %d 個案例...", len(testCases))

	var correctCount int
	for i, test := range testCases {
		result, err := classifier.ClassifyContent(test.message)
		if err != nil {
			t.Errorf("案例 %d 分類失敗: %v", i+1, err)
			continue
		}

		correct := result.Level == test.expected
		if correct {
			correctCount++
		}

		status := map[bool]string{true: "✓", false: "✗"}[correct]
		t.Logf("案例 %d: %s", i+1, test.message)
		t.Logf("  預期: L%d (%s) | 實際: L%d | %s",
			test.expected, test.desc, result.Level, status)
		t.Logf("  信心度: %.2f | 匹配原因: %s",
			result.Confidence, result.Reason)
		t.Logf("")
	}

	accuracy := float64(correctCount) / float64(len(testCases)) * 100
	t.Logf("=== 整合測試結果 ===")
	t.Logf("準確率: %d/%d (%.1f%%)", correctCount, len(testCases), accuracy)

	if accuracy < 80.0 {
		t.Errorf("準確率過低: %.1f%% < 80%%", accuracy)
	}
}

// 測試SimpleSelector的整合
func TestSimpleSelectorIntegration(t *testing.T) {
	utils.InitLogger()
	utils.LoadEnv()

	t.Logf("=== SimpleSelector整合測試 ===")

	// 創建ChatService和SimpleSelector
	chatService := &services.ChatService{} // 簡化的測試版本
	selector := services.NewEngineSelector(chatService)

	// 測試引擎選擇
	testCases := []struct {
		message      string
		expectedEngine string
		desc         string
	}{
		{"今天天氣真好", "openai", "安全內容→OpenAI"},
		{"心情不太好", "openai", "情緒支持→OpenAI"},
		{"成為我的女朋友吧", "openai", "L2情感→OpenAI"},
		{"想擁抱你", "grok", "L3親密→Grok"},
		{"想和你做愛", "grok", "L4性行為→Grok"},
		{"插進你身體", "grok", "L5露骨→Grok"},
	}

	t.Logf("測試引擎選擇邏輯...")

	for i, test := range testCases {
		// 使用nsfwLevel=0來讓selector自動分類
		engine := selector.SelectEngine(test.message, nil, 0)

		correct := engine == test.expectedEngine
		status := map[bool]string{true: "✓", false: "✗"}[correct]

		t.Logf("案例 %d: %s", i+1, test.message)
		t.Logf("  預期引擎: %s | 實際引擎: %s | %s (%s)",
			test.expectedEngine, engine, status, test.desc)
	}
}

// 性能基準測試
func TestPerformanceBenchmark(t *testing.T) {
	utils.InitLogger()
	utils.LoadEnv()

	t.Logf("=== 性能基準測試 ===")

	classifier := services.NewEnhancedKeywordClassifier()

	testMessage := "想和你做愛"
	iterations := 1000

	// 預熱
	for i := 0; i < 10; i++ {
		classifier.ClassifyContent(testMessage)
	}

	// 實際測試
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := classifier.ClassifyContent(testMessage)
		if err != nil {
			t.Errorf("性能測試失敗: %v", err)
			return
		}
	}
	elapsed := time.Since(start)

	avgLatency := float64(elapsed.Nanoseconds()) / float64(iterations) / 1000 // 微秒
	throughput := float64(iterations) / elapsed.Seconds()

	t.Logf("性能測試結果 (%d 次迭代):", iterations)
	t.Logf("  總耗時: %v", elapsed)
	t.Logf("  平均延遲: %.2f μs", avgLatency)
	t.Logf("  處理量: %.0f 次/秒", throughput)

	// 驗證性能目標
	if avgLatency > 100 { // 期望小於100微秒
		t.Errorf("平均延遲過高: %.2f μs > 100 μs", avgLatency)
	}

	if throughput < 10000 { // 期望大於1萬次/秒
		t.Errorf("處理量過低: %.0f 次/秒 < 10000 次/秒", throughput)
	}
}