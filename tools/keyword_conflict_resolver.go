package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ConflictResolution 衝突解決策略
type ConflictResolution struct {
	Keyword     string `json:"keyword"`
	CurrentLevels []int `json:"current_levels"`
	RecommendedLevel int `json:"recommended_level"`
	Reason      string `json:"reason"`
	Action      string `json:"action"` // "keep_in", "remove_from", "make_specific"
}

// ConflictResolver 衝突解決器
type ConflictResolver struct {
	basePath string
	files    []string
}

// 關鍵字上下文規則 - 根據詞彙特性決定最適合的等級
var contextRules = map[string]map[int]string{
	"興奮": {
		1: "日常興奮",
		2: "情感興奮",
		3: "性興奮感",
		4: "激烈性興奮",
	},
	"神魂顛倒": {
		2: "情感迷戀",
		3: "親密迷戀",
		4: "性愛迷戀",
		5: "極度迷戀",
	},
	"胸部": {
		1: "身體部位",
		3: "性感胸部",
		4: "性器官胸部",
	},
	"夜晚": {
		2: "浪漫夜晚",
		3: "親密夜晚",
		4: "激情夜晚",
	},
	"溫暖": {
		1: "日常溫暖",
		2: "情感溫暖",
		3: "身體溫暖",
	},
	"熱烈": {
		2: "熱烈情感",
		3: "熱烈親密",
		4: "熱烈性愛",
	},
	"心跳加速": {
		2: "情感心跳",
		3: "親密心跳",
		4: "性愛心跳",
	},
}

// 等級優先權重 - 某些關鍵字更適合特定等級
var levelPreferences = map[string]int{
	// L1 - 日常安全詞彙
	"溫暖":   1,
	"胸部":   1, // 醫學詞彙
	"夜晚":   1, // 時間詞彙

	// L2 - 情感浪漫
	"癡迷":   2,
	"沉醉":   2,
	"迷戀":   2,

	// L3 - 親密接觸
	"深陷其中": 3,
	"心跳加速": 3,

	// L4 - 明確性行為
	"神魂顛倒": 4,
	"興奮":    4, // 性語境下
	"熱烈":    4,

	// L5 - 極度露骨
	"欲仙欲死": 5,
}

func NewConflictResolver(basePath string) *ConflictResolver {
	files := []string{
		"keyword_classifier_l1.go",
		"keyword_classifier_l2.go",
		"keyword_classifier_l3.go",
		"keyword_classifier_l4.go",
		"keyword_classifier_l5.go",
	}

	return &ConflictResolver{
		basePath: basePath,
		files:    files,
	}
}

func (cr *ConflictResolver) analyzeConflicts() ([]ConflictResolution, error) {
	fmt.Println("🔍 分析關鍵字衝突...")

	// 收集所有關鍵字及其出現的等級
	allKeywords := make(map[string][]int)

	for i, file := range cr.files {
		level := i + 1
		filePath := filepath.Join(cr.basePath, file)

		keywords, err := cr.extractKeywords(filePath)
		if err != nil {
			return nil, fmt.Errorf("讀取檔案 %s 失敗: %v", file, err)
		}

		for _, keyword := range keywords {
			allKeywords[keyword] = append(allKeywords[keyword], level)
		}
	}

	// 找出衝突並生成解決方案
	var resolutions []ConflictResolution

	for keyword, levels := range allKeywords {
		if len(levels) > 1 {
			resolution := cr.resolveConflict(keyword, levels)
			resolutions = append(resolutions, resolution)
		}
	}

	// 按關鍵字排序
	sort.Slice(resolutions, func(i, j int) bool {
		return resolutions[i].Keyword < resolutions[j].Keyword
	})

	return resolutions, nil
}

func (cr *ConflictResolver) resolveConflict(keyword string, levels []int) ConflictResolution {
	// 排序等級
	sort.Ints(levels)

	// 檢查是否有預設偏好
	if preferredLevel, exists := levelPreferences[keyword]; exists {
		return ConflictResolution{
			Keyword:          keyword,
			CurrentLevels:    levels,
			RecommendedLevel: preferredLevel,
			Reason:          fmt.Sprintf("根據詞彙語義，最適合 L%d", preferredLevel),
			Action:          "keep_in",
		}
	}

	// 檢查是否有上下文規則
	if rules, exists := contextRules[keyword]; exists {
		// 選擇最高等級作為推薦
		maxLevel := levels[len(levels)-1]
		if reason, hasRule := rules[maxLevel]; hasRule {
			return ConflictResolution{
				Keyword:          keyword,
				CurrentLevels:    levels,
				RecommendedLevel: maxLevel,
				Reason:          fmt.Sprintf("建議創建特定版本: %s", reason),
				Action:          "make_specific",
			}
		}
	}

	// 預設策略：保留在最高等級
	maxLevel := levels[len(levels)-1]
	return ConflictResolution{
		Keyword:          keyword,
		CurrentLevels:    levels,
		RecommendedLevel: maxLevel,
		Reason:          fmt.Sprintf("保留在最高等級 L%d，從其他等級移除", maxLevel),
		Action:          "keep_in",
	}
}

func (cr *ConflictResolver) extractKeywords(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var keywords []string
	scanner := bufio.NewScanner(file)
	inKeywordsArray := false

	// 正則表達式匹配關鍵字
	keywordRegex := regexp.MustCompile(`"([^"]+)"`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 檢查是否進入關鍵字陣列
		if strings.Contains(line, "Keywords := []string{") {
			inKeywordsArray = true
			continue
		}

		// 檢查是否離開關鍵字陣列
		if inKeywordsArray && strings.Contains(line, "}") && !strings.Contains(line, `"`) {
			break
		}

		// 提取關鍵字
		if inKeywordsArray {
			matches := keywordRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					keyword := strings.TrimSpace(match[1])
					if keyword != "" && !strings.HasPrefix(keyword, "//") {
						keywords = append(keywords, keyword)
					}
				}
			}
		}
	}

	return keywords, scanner.Err()
}

func (cr *ConflictResolver) applyResolutions(resolutions []ConflictResolution) error {
	fmt.Printf("🔧 套用 %d 個衝突解決方案...\n", len(resolutions))

	// 按動作分組處理
	keepActions := make(map[int][]string)      // level -> keywords to keep
	removeActions := make(map[int][]string)    // level -> keywords to remove
	specificActions := make(map[string]ConflictResolution) // keyword -> resolution

	for _, resolution := range resolutions {
		switch resolution.Action {
		case "keep_in":
			// 保留在推薦等級，從其他等級移除
			keepActions[resolution.RecommendedLevel] = append(
				keepActions[resolution.RecommendedLevel],
				resolution.Keyword,
			)
			for _, level := range resolution.CurrentLevels {
				if level != resolution.RecommendedLevel {
					removeActions[level] = append(removeActions[level], resolution.Keyword)
				}
			}

		case "make_specific":
			specificActions[resolution.Keyword] = resolution
		}
	}

	// 執行移除動作
	for level, keywords := range removeActions {
		if len(keywords) > 0 {
			err := cr.removeKeywordsFromFile(level, keywords)
			if err != nil {
				return fmt.Errorf("從 L%d 移除關鍵字失敗: %v", level, err)
			}
			fmt.Printf("✅ 從 L%d 移除 %d 個重複關鍵字\n", level, len(keywords))
		}
	}

	// 執行特定化動作
	for keyword, resolution := range specificActions {
		err := cr.makeKeywordSpecific(keyword, resolution)
		if err != nil {
			return fmt.Errorf("特定化關鍵字 %s 失敗: %v", keyword, err)
		}
	}

	return nil
}

func (cr *ConflictResolver) removeKeywordsFromFile(level int, keywordsToRemove []string) error {
	filePath := filepath.Join(cr.basePath, fmt.Sprintf("keyword_classifier_l%d.go", level))

	// 讀取檔案
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	removeSet := make(map[string]bool)
	for _, k := range keywordsToRemove {
		removeSet[k] = true
	}

	// 處理每一行
	var newLines []string
	for _, line := range lines {
		// 檢查這行是否包含要移除的關鍵字
		shouldRemoveLine := false
		for keyword := range removeSet {
			if strings.Contains(line, fmt.Sprintf(`"%s"`, keyword)) {
				// 確認這是一個完整的關鍵字匹配，不是部分匹配
				keywordRegex := regexp.MustCompile(`"` + regexp.QuoteMeta(keyword) + `"`)
				if keywordRegex.MatchString(line) {
					shouldRemoveLine = true
					break
				}
			}
		}

		if !shouldRemoveLine {
			newLines = append(newLines, line)
		}
	}

	// 寫回檔案
	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

func (cr *ConflictResolver) makeKeywordSpecific(keyword string, resolution ConflictResolution) error {
	fmt.Printf("🔄 特定化關鍵字: %s\n", keyword)

	rules, exists := contextRules[keyword]
	if !exists {
		return fmt.Errorf("沒有找到關鍵字 %s 的上下文規則", keyword)
	}

	// 為每個等級創建特定版本
	for _, level := range resolution.CurrentLevels {
		if specificVersion, hasRule := rules[level]; hasRule {
			// 替換原關鍵字為特定版本
			err := cr.replaceKeywordInFile(level, keyword, specificVersion)
			if err != nil {
				return fmt.Errorf("在 L%d 替換關鍵字失敗: %v", level, err)
			}
			fmt.Printf("  ✅ L%d: %s -> %s\n", level, keyword, specificVersion)
		} else {
			// 移除沒有特定規則的等級
			err := cr.removeKeywordsFromFile(level, []string{keyword})
			if err != nil {
				return fmt.Errorf("從 L%d 移除關鍵字失敗: %v", level, err)
			}
			fmt.Printf("  ❌ L%d: 移除 %s\n", level, keyword)
		}
	}

	return nil
}

func (cr *ConflictResolver) replaceKeywordInFile(level int, oldKeyword, newKeyword string) error {
	filePath := filepath.Join(cr.basePath, fmt.Sprintf("keyword_classifier_l%d.go", level))

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 精確替換，避免部分匹配
	oldPattern := fmt.Sprintf(`"%s"`, regexp.QuoteMeta(oldKeyword))
	newPattern := fmt.Sprintf(`"%s"`, newKeyword)

	regex := regexp.MustCompile(oldPattern)
	newContent := regex.ReplaceAllString(string(content), newPattern)

	return os.WriteFile(filePath, []byte(newContent), 0644)
}

func (cr *ConflictResolver) generateReport(resolutions []ConflictResolution) {
	filename := fmt.Sprintf("scripts/conflict_resolution_report_%s.txt",
		time.Now().Format("20060102_150405"))

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ 無法創建報告檔案: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "關鍵字衝突解決報告\n")
	fmt.Fprintf(file, "產生時間: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "%s\n\n", strings.Repeat("=", 60))

	fmt.Fprintf(file, "總衝突數: %d\n\n", len(resolutions))

	// 按動作分組統計
	actionStats := make(map[string]int)
	for _, r := range resolutions {
		actionStats[r.Action]++
	}

	fmt.Fprintf(file, "處理統計:\n")
	for action, count := range actionStats {
		fmt.Fprintf(file, "  %s: %d 個\n", action, count)
	}
	fmt.Fprintf(file, "\n")

	// 詳細列表
	fmt.Fprintf(file, "詳細解決方案:\n")
	fmt.Fprintf(file, "%s\n", strings.Repeat("-", 60))

	for i, r := range resolutions {
		fmt.Fprintf(file, "%d. 關鍵字: %s\n", i+1, r.Keyword)
		fmt.Fprintf(file, "   當前等級: %v\n", r.CurrentLevels)
		fmt.Fprintf(file, "   推薦等級: L%d\n", r.RecommendedLevel)
		fmt.Fprintf(file, "   處理方式: %s\n", r.Action)
		fmt.Fprintf(file, "   原因: %s\n", r.Reason)
		fmt.Fprintf(file, "\n")
	}

	fmt.Printf("📄 詳細報告已儲存: %s\n", filename)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方式: go run keyword_conflict_resolver.go <services_path> [--dry-run]")
		fmt.Println("範例: go run keyword_conflict_resolver.go ../services")
		os.Exit(1)
	}

	servicesPath := os.Args[1]
	dryRun := len(os.Args) > 2 && os.Args[2] == "--dry-run"

	resolver := NewConflictResolver(servicesPath)

	// 分析衝突
	resolutions, err := resolver.analyzeConflicts()
	if err != nil {
		fmt.Printf("❌ 分析失敗: %v\n", err)
		os.Exit(1)
	}

	if len(resolutions) == 0 {
		fmt.Println("✅ 沒有發現關鍵字衝突")
		return
	}

	fmt.Printf("📊 發現 %d 個關鍵字衝突\n", len(resolutions))

	// 生成報告
	resolver.generateReport(resolutions)

	if dryRun {
		fmt.Println("🔍 Dry-run 模式，不會修改檔案")

		// 顯示前10個解決方案預覽
		fmt.Println("\n預覽前10個解決方案:")
		fmt.Println(strings.Repeat("-", 50))
		for i, r := range resolutions {
			if i >= 10 {
				break
			}
			fmt.Printf("%d. %s [%v] -> L%d (%s)\n",
				i+1, r.Keyword, r.CurrentLevels, r.RecommendedLevel, r.Action)
		}
		if len(resolutions) > 10 {
			fmt.Printf("... 還有 %d 個 (詳見報告檔案)\n", len(resolutions)-10)
		}
	} else {
		// 套用解決方案
		err = resolver.applyResolutions(resolutions)
		if err != nil {
			fmt.Printf("❌ 套用解決方案失敗: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ 所有衝突已解決")
	}
}