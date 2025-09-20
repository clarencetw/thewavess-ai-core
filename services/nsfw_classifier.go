package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sirupsen/logrus"
)

// NSFWClassifier 透過語意檢索執行 NSFW 分級，決定聊天引擎路由。
//
// 🎯 主要防護機制：
// 1. 等級映射：語料庫維護 L1（安全）到 L5（違規）的門檻。
// 2. 引擎路由：L1-L3 維持在 OpenAI/Mistral，L4-L5 升級至 Grok。
// 3. 中文覆蓋：語料包含 zh-Hant 露骨詞彙，確保中文情境精準度。
// 4. 危險詞彙：藉由 `reason` 標記違法或高風險內容，確保高等級判定成立。
// 5. 診斷追蹤：記錄命中片段 ID，方便分析分級是否精準。
//
// ⚠️ 關鍵：若分級錯誤，會直接影響使用者安全與體驗。
type NSFWClassifier struct {
	embedClient EmbeddingClient
	entries     []ragCorpusEntry
	config      ragConfig
}

type ragConfig struct {
	CorpusPath      string
	Locale          string
	TopK            int
	EmbedTimeout    time.Duration
	LevelThresholds map[int]float64
}

type ragCorpusEntry struct {
	ID        string    `json:"id"`
	Level     int       `json:"level"`
	Tags      []string  `json:"tags"`
	Locale    string    `json:"locale"`
	Text      string    `json:"text"`
	Reason    string    `json:"reason"`
	Version   string    `json:"version,omitempty"`
	Embedding []float64 `json:"embedding,omitempty"`

	vector []float32
}

// ClassificationResult 保存分級結果與信心指標。
type ClassificationResult struct {
	Level      int     `json:"level"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
	ChunkID    string  `json:"chunk_id"`
}

// NewNSFWClassifier 依設定語料與嵌入服務初始化分類器。
func NewNSFWClassifier() *NSFWClassifier {
	utils.LoadEnv()

	embedClient, err := NewOpenAIEmbeddingClient()
	if err != nil {
		utils.Logger.WithError(err).Fatal("failed to initialize NSFW embedding client")
	}

	config := ragConfig{
		CorpusPath:   utils.GetEnvWithDefault("NSFW_RAG_CORPUS_PATH", "configs/nsfw/rag_corpus.json"),
		Locale:       utils.GetEnvWithDefault("NSFW_RAG_LOCALE", "zh-Hant"),
		TopK:         utils.GetEnvIntWithDefault("NSFW_RAG_TOP_K", 4),
		EmbedTimeout: time.Duration(utils.GetEnvIntWithDefault("NSFW_EMBED_TIMEOUT_MS", 2000)) * time.Millisecond,
		LevelThresholds: parseLevelThresholds(utils.GetEnvWithDefault(
			"NSFW_RAG_LEVEL_THRESHOLDS",
			"5:0.55,4:0.42,3:0.30,2:0.18,1:0.10",
		)),
	}

	if config.TopK < 1 {
		config.TopK = 4
	}
	entries, err := loadRAGCorpus(config.CorpusPath)
	if err != nil {
		utils.Logger.WithError(err).Fatal("failed to load NSFW RAG corpus")
	}

	classifier := &NSFWClassifier{
		embedClient: embedClient,
		entries:     entries,
		config:      config,
	}

	if err := classifier.prepareCorpusVectors(); err != nil {
		utils.Logger.WithError(err).Fatal("NSFW 語料嵌入初始化失敗")
	}

	utils.Logger.WithFields(logrus.Fields{
		"method":        "semantic_rag",
		"entries":       len(entries),
		"corpus_path":   config.CorpusPath,
		"locale":        config.Locale,
		"top_k":         config.TopK,
		"threshold_map": config.LevelThresholds,
		"embedding":     "openai",
	}).Info("NSFW RAG 分級器已初始化")

	return classifier
}

// ClassifyContent 針對輸入內容進行語意比對並輸出 NSFW 等級。
func (c *NSFWClassifier) ClassifyContent(ctx context.Context, message string) (*ClassificationResult, error) {
	if strings.TrimSpace(message) == "" {
		return &ClassificationResult{Level: 1, Confidence: 0.0, Reason: "empty"}, nil
	}

	normalized := c.normalize(message)

	embedCtx, cancel := context.WithTimeout(ctx, c.config.EmbedTimeout)
	defer cancel()

	vector, err := c.embedClient.EmbedText(embedCtx, normalized)
	if err != nil {
		utils.Logger.WithError(err).Error("NSFW RAG 嵌入請求失敗")
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	scored := c.scoreAgainstCorpus(vector)
	if len(scored) == 0 {
		return &ClassificationResult{Level: 1, Confidence: 0.2, Reason: "no_match"}, nil
	}

	topK := scored
	if len(topK) > c.config.TopK {
		topK = topK[:c.config.TopK]
	}

	aggregated := map[int]float64{}
	var best ragScore
	for i, s := range topK {
		aggregated[s.entry.Level] += s.similarity
		if i == 0 {
			best = s
		}
	}

	level := c.resolveLevel(aggregated)
	selected := best
	for _, s := range topK {
		if s.entry.Level == level {
			selected = s
			break
		}
	}
	confidence := math.Min(0.99, selected.similarity)
	reason := selected.entry.Reason
	if reason == "" {
		reason = selected.entry.ID
	}

	utils.Logger.WithFields(logrus.Fields{
		"input_preview":  normalized[:utils.Min(len(normalized), 60)],
		"resolved_level": level,
		"confidence":     confidence,
		"reason":         reason,
		"top_chunk":      selected.entry.ID,
	}).Info("NSFW RAG 分級完成")

	return &ClassificationResult{
		Level:      level,
		Confidence: confidence,
		Reason:     reason,
		ChunkID:    selected.entry.ID,
	}, nil
}

func (c *NSFWClassifier) prepareCorpusVectors() error {
	bg := context.Background()
	for i := range c.entries {
		entry := &c.entries[i]
		if entry.Locale != "" && c.config.Locale != "" && entry.Locale != c.config.Locale {
			// 保留語料供未來語系切換使用，當前跳過嵌入準備。
			continue
		}

		if len(entry.vector) > 0 {
			continue
		}

		if len(entry.Embedding) > 0 {
			entry.vector = float64To32(entry.Embedding)
			continue
		}

		ctx, cancel := context.WithTimeout(bg, c.config.EmbedTimeout)
		vector, err := c.embedClient.EmbedText(ctx, c.normalize(entry.Text))
		cancel()
		if err != nil {
			return fmt.Errorf("embed corpus entry %s: %w", entry.ID, err)
		}
		entry.vector = vector
	}
	return nil
}

func (c *NSFWClassifier) scoreAgainstCorpus(vector []float32) []ragScore {
	scored := make([]ragScore, 0, len(c.entries))
	for i := range c.entries {
		entry := &c.entries[i]
		if entry.vector == nil {
			continue
		}
		if entry.Locale != "" && c.config.Locale != "" && entry.Locale != c.config.Locale {
			continue
		}

		similarity := cosineSimilarity(vector, entry.vector)
		if similarity <= 0 {
			continue
		}
		scored = append(scored, ragScore{entry: entry, similarity: similarity})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].similarity > scored[j].similarity
	})

	return scored
}

func (c *NSFWClassifier) resolveLevel(aggregated map[int]float64) int {
	for level := 5; level >= 1; level-- {
		score := aggregated[level]
		threshold := c.config.LevelThresholds[level]
		if threshold == 0 {
			threshold = defaultThresholds[level]
		}
		if score >= threshold {
			return level
		}
	}
	return 1
}

func (c *NSFWClassifier) normalize(s string) string {
	lowered := strings.ToLower(strings.TrimSpace(s))
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case 0x200B, 0x200C, 0x200D, 0xFEFF:
			return -1
		default:
			return r
		}
	}, lowered)

	replacements := map[string]string{
		"seggs": "sex",
		"s3x":   "sex",
		"s*x":   "sex",
		"pr0n":  "porn",
		"p0rn":  "porn",
	}
	for k, v := range replacements {
		cleaned = strings.ReplaceAll(cleaned, k, v)
	}

	return cleaned
}

type ragScore struct {
	entry      *ragCorpusEntry
	similarity float64
}

func loadRAGCorpus(path string) ([]ragCorpusEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read corpus: %w", err)
	}
	var entries []ragCorpusEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("unmarshal corpus: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("corpus is empty: %s", path)
	}

	return entries, nil
}

func parseLevelThresholds(raw string) map[int]float64 {
	thresholds := make(map[int]float64, len(defaultThresholds))
	for level, value := range defaultThresholds {
		thresholds[level] = value
	}

	parts := strings.Split(raw, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.Split(part, ":")
		if len(kv) != 2 {
			continue
		}
		level := strings.TrimSpace(kv[0])
		threshold := strings.TrimSpace(kv[1])
		lvl, err := strconv.Atoi(level)
		if err != nil {
			continue
		}
		val, err := strconv.ParseFloat(threshold, 64)
		if err != nil {
			continue
		}
		thresholds[lvl] = val
	}

	return thresholds
}

func float64To32(src []float64) []float32 {
	dst := make([]float32, len(src))
	for i, v := range src {
		dst[i] = float32(v)
	}
	return dst
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		av := float64(a[i])
		bv := float64(b[i])
		dot += av * bv
		normA += av * av
		normB += bv * bv
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

var defaultThresholds = map[int]float64{
	5: 0.55,
	4: 0.42,
	3: 0.30,
	2: 0.18,
	1: 0.10,
}
