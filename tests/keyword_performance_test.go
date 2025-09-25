package tests

import (
	"strings"
	"testing"
	"time"
)

// SimpleKeywordClassifier 純關鍵字分類器（獨立測試）
type SimpleKeywordClassifier struct{}

func NewSimpleKeywordClassifier() *SimpleKeywordClassifier {
	return &SimpleKeywordClassifier{}
}

type SimpleClassificationResult struct {
	Level      int
	Confidence float64
	Reason     string
	Duration   time.Duration
}

func (k *SimpleKeywordClassifier) ClassifyContent(message string) *SimpleClassificationResult {
	start := time.Now()

	// 擴展的關鍵字規則
	nsfwKeywords := map[string]int{
		// L5 - 極度露骨
		"插進": 5, "抽插": 5, "內射": 5, "射精": 5, "陰莖": 5, "陰道": 5,
		"龜頭": 5, "肉棒": 5, "濕穴": 5, "潮吹": 5, "高潮": 5,

		// L4 - 明確性行為
		"做愛": 4, "性愛": 4, "上床": 4, "親熱": 4, "舌吻": 4,
		"愛液": 4, "興奮": 4, "慾火": 4, "渴望": 4,

		// L3 - 親密接觸
		"親吻": 3, "擁抱": 3, "撫摸": 3, "身體": 3, "脫衣": 3,
		"胸部": 3, "大腿": 3, "床上": 3, "浴室": 3, "性感": 3,
		"誘惑": 3, "魅力": 3, "溫柔": 3,

		// L2 - 情感親密
		"約會": 2, "戀人": 2, "情侶": 2, "表白": 2, "告白": 2,
		"交往": 2, "喜歡": 2, "愛": 2, "心動": 2, "想念": 2,
		"美麗": 2, "漂亮": 2, "迷人": 2,
	}

	normalized := strings.ToLower(strings.TrimSpace(message))
	maxLevel := 1

	for keyword, level := range nsfwKeywords {
		if strings.Contains(normalized, keyword) {
			if level > maxLevel {
				maxLevel = level
			}
		}
	}

	duration := time.Since(start)

	return &SimpleClassificationResult{
		Level:      maxLevel,
		Confidence: 0.95,
		Reason:     "keyword_match",
		Duration:   duration,
	}
}

// 測試數據集
var testMessages = []string{
	// L1 - 安全內容 (5個)
	"今天天氣真好",
	"我想學習程式設計",
	"晚餐吃什麼好呢",
	"工作有點累",
	"心情不太好", // 邊界案例

	// L2 - 情感親密 (5個)
	"我喜歡你",
	"想和你約會",
	"你很美麗",
	"成為我的女朋友吧",
	"我愛你",

	// L3 - 親密接觸 (5個)
	"想擁抱你",
	"親吻你的唇",
	"撫摸你的手",
	"你的身體很溫暖",
	"在床上聊天",

	// L4 - 明確性行為 (5個)
	"想和你做愛",
	"上床親熱",
	"性愛很美好",
	"感到很興奮",
	"慾火焚身",

	// L5 - 極度露骨 (5個)
	"插進你的身體",
	"射精在裡面",
	"濕穴很緊",
	"肉棒很硬",
	"達到高潮",
}

func TestKeywordClassifierPerformance(t *testing.T) {
	classifier := NewSimpleKeywordClassifier()
	testRounds := 1000

	// 預期結果（根據人工標註）
	expectedLevels := []int{
		1, 1, 1, 1, 1, // L1 安全內容
		2, 2, 2, 2, 2, // L2 情感親密
		3, 3, 3, 3, 3, // L3 親密接觸
		4, 4, 4, 4, 4, // L4 明確性行為
		5, 5, 5, 5, 5, // L5 極度露骨
	}

	var totalDuration time.Duration
	var correctCount int

	t.Logf("開始關鍵字分類器效能測試 (%d 輪，每輪 %d 個訊息)", testRounds, len(testMessages))

	start := time.Now()

	for round := 0; round < testRounds; round++ {
		for i, msg := range testMessages {
			result := classifier.ClassifyContent(msg)
			totalDuration += result.Duration

			// 檢查準確性
			if result.Level == expectedLevels[i] {
				correctCount++
			}
		}
	}

	totalTime := time.Since(start)
	totalTests := testRounds * len(testMessages)

	t.Logf("\n=== 關鍵字分類器效能結果 ===")
	t.Logf("總測試次數: %d", totalTests)
	t.Logf("總耗時: %v", totalTime)
	t.Logf("平均每次分類: %.2f μs", float64(totalTime.Nanoseconds())/float64(totalTests)/1000)
	t.Logf("每秒處理量: %.0f 次/秒", float64(totalTests)/totalTime.Seconds())
	t.Logf("準確率: %d/%d (%.1f%%)", correctCount, totalTests, float64(correctCount)/float64(totalTests)*100)

	// 測試不同長度訊息的效能
	longMessages := []string{
		"這是一個很長的訊息" + strings.Repeat("但是完全安全的內容", 50),
		"另一個包含關鍵字的長訊息" + strings.Repeat("但只有最後才出現親吻", 50),
		strings.Repeat("重複的安全文字", 100),
	}

	t.Logf("\n=== 長訊息效能測試 ===")
	for i, longMsg := range longMessages {
		start := time.Now()
		for round := 0; round < 100; round++ {
			classifier.ClassifyContent(longMsg)
		}
		duration := time.Since(start)
		avgTime := float64(duration.Nanoseconds()) / 100 / 1000
		t.Logf("長訊息 %d (長度: %d): %.2f μs/次", i+1, len(longMsg), avgTime)
	}
}

func TestKeywordClassifierAccuracy(t *testing.T) {
	classifier := NewSimpleKeywordClassifier()

	// 預期結果
	expectedLevels := []int{
		1, 1, 1, 1, 1, // L1 安全內容
		2, 2, 2, 2, 2, // L2 情感親密
		3, 3, 3, 3, 3, // L3 親密接觸
		4, 4, 4, 4, 4, // L4 明確性行為
		5, 5, 5, 5, 5, // L5 極度露骨
	}

	var correctCount int
	var incorrectDetails []string

	t.Logf("=== 關鍵字分類器準確性測試 ===")

	for i, msg := range testMessages {
		result := classifier.ClassifyContent(msg)
		expected := expectedLevels[i]

		if result.Level == expected {
			correctCount++
		} else {
			incorrectDetails = append(incorrectDetails,
				"訊息: '"+msg[:min(30, len(msg))]+"' | 預期: L"+string(rune(expected+'0'))+" | 實際: L"+string(rune(result.Level+'0')))
		}

		t.Logf("訊息: %s | 預期: L%d | 實際: L%d | 耗時: %.2f μs | ✓",
			msg[:min(25, len(msg))], expected, result.Level,
			float64(result.Duration.Nanoseconds())/1000)
	}

	total := len(testMessages)
	t.Logf("\n=== 準確性總結 ===")
	t.Logf("正確分類: %d/%d (%.1f%%)", correctCount, total, float64(correctCount)/float64(total)*100)

	if len(incorrectDetails) > 0 {
		t.Logf("\n錯誤分類詳情:")
		for _, detail := range incorrectDetails {
			t.Logf("  %s", detail)
		}
	}
}

func BenchmarkKeywordClassifier(b *testing.B) {
	classifier := NewSimpleKeywordClassifier()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := testMessages[i%len(testMessages)]
		classifier.ClassifyContent(msg)
	}
}
