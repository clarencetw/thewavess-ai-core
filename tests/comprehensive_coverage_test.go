package tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// 大規模語料測試 - 模擬真實用戶輸入場景
func TestMassiveVocabularyGaps(t *testing.T) {
	utils.InitLogger()
	utils.LoadEnv()

	t.Logf("=== 大規模詞彙缺口分析 ===")

	classifier := services.NewEnhancedKeywordClassifier()

	// 1. 日常對話大量測試（應該都是L1）
	dailyConversations := []string{
		// 工作相關
		"今天開會很累", "老闆交代的任務", "同事請假了", "加班到很晚", "薪水太少了",
		"升職加薪", "換工作", "面試結果", "專案進度", "deadline壓力",
		"出差計劃", "會議記錄", "報告撰寫", "數據分析", "客戶需求",

		// 學習教育
		"考試成績", "作業太多", "老師很嚴格", "課程內容", "畢業論文",
		"獎學金申請", "社團活動", "圖書館讀書", "補習班", "語言學習",
		"實習經驗", "技能培訓", "證照考試", "研究所", "出國留學",

		// 生活起居
		"買菜做飯", "打掃房間", "洗衣服", "收拾東西", "裝修房子",
		"搬家計劃", "房租太貴", "水電費", "網路問題", "家電維修",
		"寵物照顧", "植物澆水", "垃圾分類", "節能環保", "生活品質",

		// 健康醫療
		"身體檢查", "看醫生", "吃藥治療", "運動健身", "減肥計劃",
		"失眠問題", "壓力大", "心理健康", "營養補充", "預防疾病",
		"牙科治療", "視力保健", "皮膚保養", "過敏反應", "健康飲食",

		// 娛樂休閒
		"看電影", "聽音樂", "玩遊戲", "閱讀小說", "追劇",
		"旅遊計劃", "拍照攝影", "運動比賽", "演唱會", "展覽參觀",
		"朋友聚餐", "逛街購物", "咖啡廳", "KTV唱歌", "戶外活動",

		// 科技數位
		"手機壞了", "電腦升級", "軟體更新", "網路購物", "線上課程",
		"社群媒體", "影片剪輯", "程式設計", "人工智慧", "區塊鏈",
		"電子產品", "數據備份", "資訊安全", "科技新聞", "創新發明",
	}

	// 2. 情感表達測試（應該主要是L1-L2）
	emotionalExpressions := []string{
		// 開心快樂
		"今天心情很好", "開心得不得了", "快樂似神仙", "笑得很燦爛", "興高采烈",
		"歡天喜地", "樂不可支", "心花怒放", "喜出望外", "滿心歡喜",

		// 傷心難過
		"心情很沮喪", "難過得想哭", "傷心欲絕", "痛苦不堪", "心碎了",
		"眼淚止不住", "鬱鬱寡歡", "愁眉苦臉", "垂頭喪氣", "黯然神傷",

		// 憤怒生氣
		"氣得發抖", "火冒三丈", "怒不可遏", "憤憤不平", "暴跳如雷",
		"咬牙切齒", "怒火中燒", "氣急敗壞", "義憤填膺", "勃然大怒",

		// 焦慮擔心
		"擔心得睡不著", "焦慮不安", "坐立不安", "憂心忡忡", "惶惶不安",
		"心神不寧", "忐忑不安", "驚慌失措", "心急如焚", "膽戰心驚",

		// 驚訝意外
		"太意外了", "驚訝得說不出話", "目瞪口呆", "大吃一驚", "始料未及",
		"出乎意料", "匪夷所思", "不敢置信", "震驚不已", "嘆為觀止",
	}

	// 3. 網路流行語測試（現代用語缺口）
	internetSlang := []string{
		// 年輕世代用語
		"ㄅㄨㄒㄧㄝˋ", "gg了", "87分", "9487", "484", "D卡", "Dcard",
		"笑死", "超讚", "神器", "開箱", "首推", "必推", "已跪", "跪了",
		"太強", "威猛", "霸氣", "帥炸", "美翻", "可愛爆", "萌翻",

		// 網路梗圖用語
		"笑翻", "笑尿", "笑炸", "好派", "很可以", "可以的", "沒問題",
		"超棒der", "好厲害der", "真的假的", "不會吧", "扯爆",
		"太誇張", "超扯", "誇張欸", "好奇怪喔", "怪怪der",

		// 表情符號文字化
		"QQ", "T_T", "orz", "囧", "= =", "> <", "XD", "XDDD",
		"哈哈哈", "嗚嗚嗚", "嘿嘿嘿", "呵呵", "哇嗚", "耶耶耶",

		// 語助詞
		"欸", "啦", "啊", "喔", "嘛", "耶", "吼", "蛤", "餒", "捏",
		"厚", "拉", "咧", "內", "呦", "唷", "喲", "吶", "ㄋㄟ",
	}

	// 4. 隱晦性暗示測試（容易被遺漏的L3-L4內容）
	subtleImplications := []string{
		// 隱晦的親密暗示
		"想要更靠近你", "想和你更親密", "想要認識真正的你", "想看到你的另一面",
		"想和你單獨相處", "想要私下聊聊", "想要深度交流", "想了解你的內心",
		"想要感受你的溫度", "想要你的陪伴", "想要你在身邊", "想要你的懷抱",

		// 模糊的性暗示
		"想要和你發生點什麼", "想要體驗新的感覺", "想要嘗試刺激的事",
		"想要探索彼此", "想要突破界限", "想要釋放壓力", "想要放鬆身心",
		"想要感受快樂", "想要達到巔峰", "想要完全放開", "想要忘我境界",

		// 委婉表達
		"想要那個", "想要做那件事", "想要完成最後一步", "想要水到渠成",
		"想要瓜熟蒂落", "想要自然發生", "想要順其自然", "想要心滿意足",
		"想要天雷勾動地火", "想要乾柴烈火", "想要如膠似漆", "想要難分難捨",
	}

	// 5. 執行大規模測試
	testCategories := map[string][]string{
		"日常對話": dailyConversations,
		"情感表達": emotionalExpressions,
		"網路用語": internetSlang,
		"隱晦暗示": subtleImplications,
	}

	totalTests := 0
	wrongClassifications := 0

	for category, tests := range testCategories {
		t.Logf("\n--- 測試類別: %s (%d個案例) ---", category, len(tests))
		categoryWrong := 0

		for i, testMsg := range tests {
			result, err := classifier.ClassifyContent(testMsg)
			if err != nil {
				t.Errorf("分類失敗: %v", err)
				continue
			}

			totalTests++

			// 判斷是否分類錯誤
			var expectedLevel int
			var isWrong bool

			switch category {
			case "日常對話":
				expectedLevel = 1
				isWrong = result.Level > 1
			case "情感表達":
				expectedLevel = 1
				isWrong = result.Level > 2 // 允許情感詞彙到L2
			case "網路用語":
				expectedLevel = 1
				isWrong = result.Level > 1
			case "隱晦暗示":
				expectedLevel = 3 // 這些應該被識別為L3或以上
				isWrong = result.Level < 3
			}

			if isWrong {
				wrongClassifications++
				categoryWrong++
				if categoryWrong <= 5 { // 只顯示前5個錯誤
					t.Logf("  ❌ 案例 %d: '%s'", i+1, testMsg)
					t.Logf("     預期: ≤L%d | 實際: L%d | 原因: %s",
						expectedLevel, result.Level, result.Reason)
				}
			}
		}

		accuracy := float64(len(tests)-categoryWrong) / float64(len(tests)) * 100
		t.Logf("  %s準確率: %d/%d (%.1f%%)", category, len(tests)-categoryWrong, len(tests), accuracy)
		if categoryWrong > 5 {
			t.Logf("  （顯示前5個錯誤，實際錯誤 %d 個）", categoryWrong)
		}
	}

	// 6. 分析結果
	overallAccuracy := float64(totalTests-wrongClassifications) / float64(totalTests) * 100

	t.Logf("\n=== 大規模測試結果 ===")
	t.Logf("總測試案例: %d", totalTests)
	t.Logf("錯誤分類: %d", wrongClassifications)
	t.Logf("整體準確率: %.1f%%", overallAccuracy)

	// 獲取當前關鍵字統計
	info := classifier.GetClassifierInfo()
	if stats, ok := info["level_stats"].(map[int]int); ok {
		total := 0
		for _, count := range stats {
			total += count
		}
		t.Logf("當前關鍵字總數: %d", total)
	}

	t.Logf("\n=== 問題分析 ===")
	if overallAccuracy < 85 {
		t.Logf("⚠️  整體準確率過低！")
	}

	t.Logf("📈 詞彙缺口明顯，建議:")
	t.Logf("   1. 擴充日常詞彙庫到 2000+ 個（防誤判）")
	t.Logf("   2. 增加網路流行語 1000+ 個")
	t.Logf("   3. 補充情感表達詞彙 1000+ 個")
	t.Logf("   4. 完善隱晦表達識別 500+ 個")
	t.Logf("   5. 總目標: 5000-10000 個關鍵字")

	if overallAccuracy < 90 {
		t.Errorf("準確率不足 90%%，證明需要大幅擴充關鍵字庫")
	}
}

// 壓力測試：隨機組合詞彙
func TestRandomCombinationStress(t *testing.T) {
	utils.InitLogger()
	utils.LoadEnv()

	t.Logf("=== 隨機組合壓力測試 ===")

	classifier := services.NewEnhancedKeywordClassifier()

	// 常見詞彙池
	words := []string{
		"想要", "喜歡", "愛", "感覺", "覺得", "認為", "希望", "期待",
		"開心", "快樂", "難過", "傷心", "生氣", "憤怒", "擔心", "焦慮",
		"工作", "學習", "讀書", "考試", "上班", "下班", "休息", "睡覺",
		"吃飯", "喝水", "運動", "散步", "購物", "看電影", "聽音樂", "玩遊戲",
		"朋友", "家人", "同事", "老師", "學生", "醫生", "護士", "警察",
		"今天", "明天", "昨天", "現在", "以前", "以後", "早上", "晚上",
		"很", "非常", "特別", "超級", "真的", "實在", "確實", "當然",
		"可以", "能夠", "應該", "需要", "必須", "想", "要", "會",
	}

	rand.Seed(time.Now().UnixNano())
	successCount := 0
	totalTests := 200

	for i := 0; i < totalTests; i++ {
		// 隨機組合2-4個詞
		wordCount := rand.Intn(3) + 2 // 2-4個詞
		var selectedWords []string

		for j := 0; j < wordCount; j++ {
			word := words[rand.Intn(len(words))]
			selectedWords = append(selectedWords, word)
		}

		testSentence := ""
		for k, word := range selectedWords {
			if k > 0 {
				testSentence += " "
			}
			testSentence += word
		}

		result, err := classifier.ClassifyContent(testSentence)
		if err != nil {
			t.Errorf("隨機測試失敗: %v", err)
			continue
		}

		// 這些隨機組合應該大多是L1-L2
		if result.Level <= 2 {
			successCount++
		} else if i < 10 { // 顯示前10個高等級案例
			t.Logf("隨機案例 %d: '%s' → L%d (%s)",
				i+1, testSentence, result.Level, result.Reason)
		}
	}

	accuracy := float64(successCount) / float64(totalTests) * 100
	t.Logf("隨機組合測試: %d/%d (%.1f%%) 被正確分類為低等級",
		successCount, totalTests, accuracy)

	if accuracy < 80 {
		t.Logf("⚠️  隨機組合準確率過低，說明關鍵字過於激進")
	}
}
