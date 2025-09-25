package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// KeywordStats 關鍵字統計結構
type KeywordStats struct {
	Level           int                 `json:"level"`
	File            string              `json:"file"`
	TotalKeywords   int                 `json:"total_keywords"`
	UniqueKeywords  int                 `json:"unique_keywords"`
	Duplicates      []string            `json:"duplicates"`
	Keywords        []string            `json:"keywords"`
	LengthStats     LengthStatistics    `json:"length_stats"`
}

// LengthStatistics 長度統計
type LengthStatistics struct {
	Min     int     `json:"min"`
	Max     int     `json:"max"`
	Average float64 `json:"average"`
	Median  int     `json:"median"`
}

// ConflictInfo 衝突信息
type ConflictInfo struct {
	Keyword  string `json:"keyword"`
	Files    []string `json:"files"`
	Levels   []int  `json:"levels"`
	Count    int    `json:"count"`
}

// AnalysisReport 分析報告
type AnalysisReport struct {
	Timestamp       string              `json:"timestamp"`
	TotalFiles      int                 `json:"total_files"`
	TotalKeywords   int                 `json:"total_keywords"`
	UniqueKeywords  int                 `json:"unique_keywords"`
	DuplicateCount  int                 `json:"duplicate_count"`
	CrossFileConflicts []ConflictInfo   `json:"cross_file_conflicts"`
	LevelStats      []KeywordStats      `json:"level_stats"`
	QualityScore    float64             `json:"quality_score"`
	Recommendations []string            `json:"recommendations"`
}

func main() {
	fmt.Println("🔍 關鍵字分析工具啟動")
	fmt.Println(strings.Repeat("=", 50))

	// 分析關鍵字檔案
	report, err := analyzeKeywordFiles()
	if err != nil {
		fmt.Printf("❌ 分析失敗: %v\n", err)
		os.Exit(1)
	}

	// 輸出結果
	printReport(report)

	// 儲存JSON報告
	saveJSONReport(report)

	fmt.Println("\n✅ 分析完成")
}

func analyzeKeywordFiles() (*AnalysisReport, error) {
	basePath := "services"
	files := []string{
		"keyword_classifier_l1.go",
		"keyword_classifier_l2.go",
		"keyword_classifier_l3.go",
		"keyword_classifier_l4.go",
		"keyword_classifier_l5.go",
	}

	report := &AnalysisReport{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		TotalFiles: len(files),
	}

	allKeywords := make(map[string][]int) // keyword -> levels
	levelStats := make([]KeywordStats, 0, len(files))

	// 分析每個檔案
	for i, file := range files {
		level := i + 1
		filePath := filepath.Join(basePath, file)

		keywords, err := extractKeywords(filePath)
		if err != nil {
			return nil, fmt.Errorf("讀取檔案 %s 失敗: %v", file, err)
		}

		// 統計本檔案
		stats := analyzeFileKeywords(level, file, keywords)
		levelStats = append(levelStats, stats)

		// 記錄到全域統計
		for _, keyword := range keywords {
			allKeywords[keyword] = append(allKeywords[keyword], level)
		}

		report.TotalKeywords += len(keywords)
	}

	// 計算整體統計
	report.UniqueKeywords = len(allKeywords)
	report.LevelStats = levelStats

	// 找出跨檔案衝突
	conflicts := findCrossFileConflicts(allKeywords)
	report.CrossFileConflicts = conflicts
	report.DuplicateCount = len(conflicts)

	// 計算品質分數
	report.QualityScore = calculateQualityScore(report)

	// 生成建議
	report.Recommendations = generateRecommendations(report)

	return report, nil
}

func extractKeywords(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 使用正則表達式提取所有quoted字符串
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	keywords := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			keyword := strings.TrimSpace(match[1])
			if keyword != "" {
				keywords = append(keywords, keyword)
			}
		}
	}

	return keywords, nil
}

func analyzeFileKeywords(level int, file string, keywords []string) KeywordStats {
	// 去重並找出重複項
	uniqueMap := make(map[string]int)
	duplicates := make([]string, 0)

	for _, keyword := range keywords {
		uniqueMap[keyword]++
		if uniqueMap[keyword] == 2 {
			duplicates = append(duplicates, keyword)
		}
	}

	// 計算長度統計
	lengths := make([]int, len(keywords))
	totalLength := 0
	for i, keyword := range keywords {
		lengths[i] = len(keyword)
		totalLength += len(keyword)
	}

	sort.Ints(lengths)

	lengthStats := LengthStatistics{
		Min:     lengths[0],
		Max:     lengths[len(lengths)-1],
		Average: float64(totalLength) / float64(len(keywords)),
		Median:  lengths[len(lengths)/2],
	}

	return KeywordStats{
		Level:          level,
		File:           file,
		TotalKeywords:  len(keywords),
		UniqueKeywords: len(uniqueMap),
		Duplicates:     duplicates,
		Keywords:       keywords,
		LengthStats:    lengthStats,
	}
}

func findCrossFileConflicts(allKeywords map[string][]int) []ConflictInfo {
	conflicts := make([]ConflictInfo, 0)

	for keyword, levels := range allKeywords {
		// 去重levels並排序
		uniqueLevels := make(map[int]bool)
		for _, level := range levels {
			uniqueLevels[level] = true
		}

		// 只有當關鍵字出現在多個不同等級時才算衝突
		if len(uniqueLevels) > 1 {
			levelList := make([]int, 0, len(uniqueLevels))
			fileList := make([]string, 0, len(uniqueLevels))

			for level := range uniqueLevels {
				levelList = append(levelList, level)
				fileList = append(fileList, fmt.Sprintf("L%d", level))
			}

			sort.Ints(levelList)

			conflicts = append(conflicts, ConflictInfo{
				Keyword: keyword,
				Files:   fileList,
				Levels:  levelList,
				Count:   len(levels),
			})
		}
	}

	// 按衝突數量排序
	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].Count > conflicts[j].Count
	})

	return conflicts
}

func calculateQualityScore(report *AnalysisReport) float64 {
	score := 100.0

	// 重複關鍵字扣分 (每個重複關鍵字扣0.5分)
	score -= float64(report.DuplicateCount) * 0.5

	// 檔案間不平衡扣分 (限制最大扣分)
	if len(report.LevelStats) > 0 && report.TotalKeywords > 0 {
		avgKeywords := float64(report.TotalKeywords) / float64(len(report.LevelStats))
		totalDeviation := 0.0
		for _, stats := range report.LevelStats {
			deviation := float64(stats.TotalKeywords) - avgKeywords
			if deviation < 0 {
				deviation = -deviation
			}
			totalDeviation += deviation / avgKeywords
		}
		// 限制不平衡扣分最多20分
		imbalancePenalty := totalDeviation * 2
		if imbalancePenalty > 20 {
			imbalancePenalty = 20
		}
		score -= imbalancePenalty
	}

	// 如果沒有重複關鍵字，給予獎勵
	if report.DuplicateCount == 0 {
		score += 5
	}

	// 確保分數在0-100範圍內
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func generateRecommendations(report *AnalysisReport) []string {
	recommendations := make([]string, 0)

	if report.DuplicateCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("發現 %d 個跨檔案重複關鍵字，建議清理重複項", report.DuplicateCount))
	}

	if report.QualityScore < 90 {
		recommendations = append(recommendations, "整體品質分數偏低，建議檢查關鍵字分佈和重複問題")
	}

	// 檢查各檔案關鍵字數量平衡
	if len(report.LevelStats) > 0 {
		maxKeywords := 0
		minKeywords := report.LevelStats[0].TotalKeywords

		for _, stats := range report.LevelStats {
			if stats.TotalKeywords > maxKeywords {
				maxKeywords = stats.TotalKeywords
			}
			if stats.TotalKeywords < minKeywords {
				minKeywords = stats.TotalKeywords
			}
		}

		if maxKeywords > minKeywords*3 {
			recommendations = append(recommendations, "關鍵字分佈不平衡，建議調整各等級的關鍵字數量")
		}
	}

	return recommendations
}

func printReport(report *AnalysisReport) {
	fmt.Printf("📊 關鍵字分析報告 - %s\n", report.Timestamp)
	fmt.Println(strings.Repeat("=", 60))

	// 整體統計
	fmt.Printf("📁 總檔案數: %d\n", report.TotalFiles)
	fmt.Printf("📝 總關鍵字: %d\n", report.TotalKeywords)
	fmt.Printf("🔑 唯一關鍵字: %d\n", report.UniqueKeywords)
	fmt.Printf("🔄 重複關鍵字: %d\n", report.DuplicateCount)
	fmt.Printf("⭐ 品質分數: %.1f/100\n", report.QualityScore)

	// 各檔案統計
	fmt.Println("\n📋 各檔案詳細統計:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-6s %-8s %-8s %-8s %-10s\n", "等級", "總數", "唯一", "重複", "平均長度")
	fmt.Println(strings.Repeat("-", 60))

	for _, stats := range report.LevelStats {
		fmt.Printf("L%-5d %-8d %-8d %-8d %.1f\n",
			stats.Level,
			stats.TotalKeywords,
			stats.UniqueKeywords,
			len(stats.Duplicates),
			stats.LengthStats.Average)
	}

	// 跨檔案衝突 (只顯示前10個)
	if len(report.CrossFileConflicts) > 0 {
		fmt.Println("\n⚠️  跨檔案衝突 (前10個):")
		fmt.Println(strings.Repeat("-", 60))
		count := len(report.CrossFileConflicts)
		if count > 10 {
			count = 10
		}

		for i := 0; i < count; i++ {
			conflict := report.CrossFileConflicts[i]
			fmt.Printf("'%s' 出現在: %v\n", conflict.Keyword, conflict.Files)
		}

		if len(report.CrossFileConflicts) > 10 {
			fmt.Printf("... 還有 %d 個衝突 (詳見JSON報告)\n", len(report.CrossFileConflicts)-10)
		}
	}

	// 建議
	if len(report.Recommendations) > 0 {
		fmt.Println("\n💡 改進建議:")
		fmt.Println(strings.Repeat("-", 60))
		for i, rec := range report.Recommendations {
			fmt.Printf("%d. %s\n", i+1, rec)
		}
	}
}

func saveJSONReport(report *AnalysisReport) {
	filename := fmt.Sprintf("scripts/keyword_analysis_%s.json",
		time.Now().Format("20060102_150405"))

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON序列化失敗: %v\n", err)
		return
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Printf("❌ 儲存JSON報告失敗: %v\n", err)
		return
	}

	fmt.Printf("💾 詳細報告已儲存: %s\n", filename)
}