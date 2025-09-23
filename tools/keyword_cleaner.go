package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// CleanupResult 清理結果
type CleanupResult struct {
	File             string
	OriginalCount    int
	CleanedCount     int
	RemovedDuplicates []string
	RemovedEmpty     int
}

func main() {
	fmt.Println("🧹 關鍵字清理工具啟動")
	fmt.Println("=" + strings.Repeat("=", 49))

	// 檢查是否為自動模式
	autoMode := len(os.Args) > 1 && os.Args[1] == "--auto"

	// 確認用戶意圖
	if !autoMode && !confirmAction() {
		fmt.Println("❌ 操作已取消")
		return
	}

	if autoMode {
		fmt.Println("🤖 自動模式啟動，跳過確認")
	}

	// 清理關鍵字檔案
	results, err := cleanKeywordFiles()
	if err != nil {
		fmt.Printf("❌ 清理失敗: %v\n", err)
		os.Exit(1)
	}

	// 輸出結果
	printCleanupResults(results)

	fmt.Println("\n✅ 清理完成")
}

func confirmAction() bool {
	fmt.Println("⚠️  這個工具將會:")
	fmt.Println("   1. 移除每個檔案內的重複關鍵字")
	fmt.Println("   2. 移除空白關鍵字")
	fmt.Println("   3. 重新格式化關鍵字列表")
	fmt.Println("   4. 備份原檔案為 .backup")
	fmt.Print("\n是否繼續? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}

func cleanKeywordFiles() ([]CleanupResult, error) {
	basePath := "services"
	files := []string{
		"keyword_classifier_l1.go",
		"keyword_classifier_l2.go",
		"keyword_classifier_l3.go",
		"keyword_classifier_l4.go",
		"keyword_classifier_l5.go",
	}

	results := make([]CleanupResult, 0, len(files))

	for _, file := range files {
		filePath := filepath.Join(basePath, file)
		result, err := cleanSingleFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("清理檔案 %s 失敗: %v", file, err)
		}
		results = append(results, result)
	}

	return results, nil
}

func cleanSingleFile(filePath string) (CleanupResult, error) {
	result := CleanupResult{
		File: filepath.Base(filePath),
	}

	// 讀取檔案
	content, err := os.ReadFile(filePath)
	if err != nil {
		return result, err
	}

	// 備份原檔案
	backupPath := filePath + ".backup"
	err = os.WriteFile(backupPath, content, 0644)
	if err != nil {
		return result, fmt.Errorf("備份檔案失敗: %v", err)
	}

	// 解析和清理關鍵字
	cleanedContent, cleanStats := cleanFileContent(string(content))
	result.OriginalCount = cleanStats.originalCount
	result.CleanedCount = cleanStats.cleanedCount
	result.RemovedDuplicates = cleanStats.removedDuplicates
	result.RemovedEmpty = cleanStats.removedEmpty

	// 寫回檔案
	err = os.WriteFile(filePath, []byte(cleanedContent), 0644)
	if err != nil {
		return result, err
	}

	fmt.Printf("✅ %s: %d → %d 關鍵字\n",
		result.File, result.OriginalCount, result.CleanedCount)

	return result, nil
}

type cleanStats struct {
	originalCount     int
	cleanedCount      int
	removedDuplicates []string
	removedEmpty      int
}

func cleanFileContent(content string) (string, cleanStats) {
	stats := cleanStats{}

	// 找到關鍵字數組的開始
	startPattern := regexp.MustCompile(`l\dKeywords := \[\]string\{`)

	lines := strings.Split(content, "\n")
	var keywordStart, keywordEnd int

	// 找到關鍵字區域
	for i, line := range lines {
		if startPattern.MatchString(line) {
			keywordStart = i
		}
		if strings.Contains(line, "// 添加到關鍵字映射中") {
			keywordEnd = i
			break
		}
	}

	// 提取關鍵字部分
	keywordLines := lines[keywordStart+1 : keywordEnd-1]

	// 解析所有關鍵字
	re := regexp.MustCompile(`"([^"]+)"`)
	keywordMap := make(map[string]bool)
	var allKeywords []string

	for _, line := range keywordLines {
		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				keyword := strings.TrimSpace(match[1])
				stats.originalCount++

				if keyword == "" {
					stats.removedEmpty++
					continue
				}

				if keywordMap[keyword] {
					stats.removedDuplicates = append(stats.removedDuplicates, keyword)
					continue
				}

				keywordMap[keyword] = true
				allKeywords = append(allKeywords, keyword)
			}
		}
	}

	stats.cleanedCount = len(allKeywords)

	// 按字母順序排序
	sort.Strings(allKeywords)

	// 重新生成關鍵字部分
	var newKeywordLines []string

	// 保留註釋結構，重新組織關鍵字
	currentSection := ""
	sectionKeywords := make(map[string][]string)
	var sectionOrder []string

	// 按原有的註釋分組重新組織
	for _, line := range keywordLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "//") {
			currentSection = line
			if _, exists := sectionKeywords[currentSection]; !exists {
				sectionOrder = append(sectionOrder, currentSection)
				sectionKeywords[currentSection] = make([]string, 0)
			}
		} else if line != "" && currentSection != "" {
			// 提取該行的關鍵字
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					keyword := strings.TrimSpace(match[1])
					if keyword != "" && keywordMap[keyword] {
						sectionKeywords[currentSection] = append(sectionKeywords[currentSection], keyword)
						delete(keywordMap, keyword) // 標記為已使用
					}
				}
			}
		}
	}

	// 處理沒有分組的關鍵字
	orphanKeywords := make([]string, 0)
	for keyword := range keywordMap {
		orphanKeywords = append(orphanKeywords, keyword)
	}
	if len(orphanKeywords) > 0 {
		sort.Strings(orphanKeywords)
		sectionOrder = append(sectionOrder, "// === 其他關鍵字 ===")
		sectionKeywords["// === 其他關鍵字 ==="] = orphanKeywords
	}

	// 生成新的關鍵字內容
	for _, section := range sectionOrder {
		keywords := sectionKeywords[section]
		if len(keywords) == 0 {
			continue
		}

		newKeywordLines = append(newKeywordLines, "\t\t"+section)

		// 每行8個關鍵字
		for i := 0; i < len(keywords); i += 8 {
			end := i + 8
			if end > len(keywords) {
				end = len(keywords)
			}

			lineKeywords := keywords[i:end]
			quotedKeywords := make([]string, len(lineKeywords))
			for j, kw := range lineKeywords {
				quotedKeywords[j] = fmt.Sprintf(`"%s"`, kw)
			}

			line := "\t\t" + strings.Join(quotedKeywords, ", ")
			if end < len(keywords) {
				line += ","
			} else if section != sectionOrder[len(sectionOrder)-1] {
				line += ","
			} else {
				// 最後一行不加逗號
			}

			newKeywordLines = append(newKeywordLines, line)
		}

		if section != sectionOrder[len(sectionOrder)-1] {
			newKeywordLines = append(newKeywordLines, "")
		}
	}

	// 重新組合完整內容
	var newLines []string
	newLines = append(newLines, lines[:keywordStart+1]...)
	newLines = append(newLines, newKeywordLines...)
	newLines = append(newLines, lines[keywordEnd-1:]...)

	return strings.Join(newLines, "\n"), stats
}

func printCleanupResults(results []CleanupResult) {
	fmt.Println("\n📋 清理結果摘要:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-25s %-8s %-8s %-8s\n", "檔案", "原始", "清理後", "移除")
	fmt.Println(strings.Repeat("-", 60))

	totalOriginal := 0
	totalCleaned := 0
	totalRemoved := 0

	for _, result := range results {
		removed := result.OriginalCount - result.CleanedCount
		fmt.Printf("%-25s %-8d %-8d %-8d\n",
			result.File, result.OriginalCount, result.CleanedCount, removed)

		totalOriginal += result.OriginalCount
		totalCleaned += result.CleanedCount
		totalRemoved += removed
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-25s %-8d %-8d %-8d\n", "總計", totalOriginal, totalCleaned, totalRemoved)

	// 顯示重複關鍵字詳情
	if totalRemoved > 0 {
		fmt.Println("\n🔄 移除的重複關鍵字:")
		fmt.Println(strings.Repeat("-", 60))
		for _, result := range results {
			if len(result.RemovedDuplicates) > 0 {
				fmt.Printf("%s: %v\n", result.File, result.RemovedDuplicates)
			}
		}
	}
}