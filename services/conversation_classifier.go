package services

import "strings"

// ConversationClassifier 對話功能分類器
// 負責識別對話中的特殊意圖和狀態變化
type ConversationClassifier struct {
	exitKeywords     []string
	topicKeywords    []string
	implicitKeywords []string
}

// NewConversationClassifier 創建對話分類器
func NewConversationClassifier() *ConversationClassifier {
	return &ConversationClassifier{
		exitKeywords:     loadExitKeywords(),
		topicKeywords:    loadTopicKeywords(),
		implicitKeywords: loadImplicitKeywords(),
	}
}

// IsExitSignal 檢測退出信號
func (c *ConversationClassifier) IsExitSignal(msg string) bool {
	msg = strings.ToLower(strings.TrimSpace(msg))
	for _, keyword := range c.exitKeywords {
		if strings.Contains(msg, keyword) {
			return true
		}
	}
	return false
}

// IsTopicChange 檢測話題轉換
func (c *ConversationClassifier) IsTopicChange(msg string) bool {
	msg = strings.ToLower(strings.TrimSpace(msg))
	for _, keyword := range c.topicKeywords {
		if strings.Contains(msg, keyword) {
			return true
		}
	}
	return false
}

// IsPotentialImplicitContent 檢測疑似隱晦表達
func (c *ConversationClassifier) IsPotentialImplicitContent(msg string) bool {
	msg = strings.ToLower(strings.TrimSpace(msg))

	// 檢查是否包含多個隱晦模式
	matchCount := 0
	for _, pattern := range c.implicitKeywords {
		if strings.Contains(msg, pattern) {
			matchCount++
		}
	}

	// 如果匹配2個以上隱晦模式，且消息較短（可能是試探性的），認為是隱晦內容
	return matchCount >= 2 || (matchCount >= 1 && len(msg) < 20)
}

// loadExitKeywords 載入退出信號關鍵字
func loadExitKeywords() []string {
	return []string{
		// 明確拒絕
		"不要", "停止", "算了", "不用", "不想", "不願意", "不感興趣", "不需要",
		"住手", "夠了", "適可而止", "到此為止", "就這樣", "別再", "不再",

		// 話題轉換
		"聊別的", "換話題", "換個話題", "說點別的", "討論別的", "談點別的",
		"改變話題", "轉移話題", "另說別的", "聊其他", "說其他", "講別的",

		// 狀態退出
		"累了", "疲憊", "睏了", "想睡", "睡覺", "休息", "放鬆", "暫停",
		"結束", "結束了", "好了", "差不多了", "可以了", "夠了",

		// 委婉表達
		"下次再說", "以後再聊", "改天再談", "另外時間", "換個時間",
		"不太合適", "時機不對", "不是時候", "環境不對", "場合不對",

		// 心情轉變
		"沒心情", "心情不好", "狀態不對", "感覺不對", "不在狀態",
		"情緒不佳", "心情低落", "沒興致", "提不起興趣",
	}
}

// loadTopicKeywords 載入話題轉換關鍵字
func loadTopicKeywords() []string {
	return []string{
		// 話題引導詞
		"話題", "聊聊", "說說", "討論", "談談", "聊天", "交流", "分享",
		"講講", "說說看", "聊一聊", "談一談", "講一講", "來說說",

		// 工作職場
		"工作", "上班", "職場", "同事", "老闆", "公司", "會議", "專案",
		"薪水", "升職", "加班", "出差", "面試", "履歷", "職業", "事業",

		// 日常生活
		"天氣", "氣候", "下雨", "晴天", "颱風", "溫度", "季節", "春夏秋冬",
		"吃飯", "料理", "餐廳", "美食", "購物", "逛街", "家庭", "家人",

		// 娛樂文化
		"新聞", "時事", "政治", "經濟", "社會", "國際", "台灣", "中國",
		"電影", "戲劇", "演員", "導演", "院線", "Netflix", "Disney",
		"音樂", "歌曲", "歌手", "演唱會", "MV", "專輯", "流行", "古典",

		// 學習教育
		"讀書", "學習", "考試", "學校", "大學", "研究所", "課程", "老師",
		"學生", "同學", "作業", "論文", "升學", "教育", "知識", "技能",

		// 健康運動
		"健康", "運動", "健身", "跑步", "游泳", "瑜伽", "醫生", "看病",
		"減肥", "保養", "營養", "睡眠", "壓力", "身體", "心理",

		// 科技數位
		"科技", "電腦", "手機", "APP", "軟體", "網路", "AI", "程式",
		"遊戲", "電玩", "社群", "Facebook", "Instagram", "Line", "YouTube",

		// 旅遊休閒
		"旅遊", "旅行", "出國", "度假", "景點", "飯店", "機票", "簽證",
		"拍照", "攝影", "興趣", "嗜好", "收集", "運動", "戶外",

		// 投資理財
		"投資", "理財", "股票", "基金", "房地產", "銀行", "貸款", "保險",
		"經濟", "市場", "金融", "消費", "省錢", "賺錢", "財務",
	}
}

// loadImplicitKeywords 載入隱晦表達關鍵字
func loadImplicitKeywords() []string {
	return []string{
		// 委婉表達
		"想要那個", "做那檔事", "你懂的", "你知道", "那件事", "那種事",
		"某些事", "特別的事", "不可描述", "羞羞的事", "私密的事",

		// 暗示性疑問
		"可以嗎", "願意嗎", "想不想", "要不要", "好不好", "如何", "怎麼樣",
		"有沒有興趣", "有想法嗎", "考慮一下", "試試看", "體驗一下",

		// 時間地點暗示
		"私下", "單獨", "兩個人", "沒人的時候", "安靜的地方", "私密空間",
		"房間裡", "床上", "晚上", "深夜", "週末", "假期",

		// 情緒狀態暗示
		"興奮", "刺激", "緊張", "心跳", "害羞", "臉紅", "不好意思",
		"期待", "渴望", "想念", "思念", "掛念", "在意",

		// 身體相關（模糊表達）
		"身體", "肌膚", "溫度", "觸碰", "接觸", "靠近", "貼近",
		"懷抱", "擁抱", "依偎", "親近", "溫柔", "輕撫",

		// 感官表達
		"感覺", "感受", "體驗", "嘗試", "探索", "發現", "享受",
		"滿足", "舒服", "放鬆", "釋放", "自由", "解脫",
	}
}
