package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/clarencetw/thewavess-ai-core/utils"
)

type corpusDataEntry struct {
	ID      string   `json:"id"`
	Level   int      `json:"level"`
	Tags    []string `json:"tags"`
	Locale  string   `json:"locale"`
	Text    string   `json:"text"`
	Reason  string   `json:"reason"`
	Version string   `json:"version,omitempty"`
}

type embeddingEntry struct {
	ID        string    `json:"id"`
	Embedding []float64 `json:"embedding"`
	Version   string    `json:"version"`
}

func main() {
	utils.LoadEnv()

	dataPath := utils.GetEnvWithDefault("NSFW_CORPUS_DATA_PATH", "configs/nsfw/corpus.json")
	embeddingPath := utils.GetEnvWithDefault("NSFW_CORPUS_EMBEDDING_PATH", "configs/nsfw/embeddings.json")
	locale := utils.GetEnvWithDefault("NSFW_RAG_LOCALE", "zh-Hant")

	log.Printf("🔍 檢查 NSFW 語料庫 embedding 向量狀態")
	log.Printf("📝 數據檔案: %s", dataPath)
	log.Printf("🧠 向量檔案: %s", embeddingPath)
	log.Printf("🌐 語系過濾: %s", locale)

	// 讀取數據檔案
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		log.Fatalf("❌ 讀取數據檔案失敗: %v", err)
	}

	var corpusData []corpusDataEntry
	if err := json.Unmarshal(dataBytes, &corpusData); err != nil {
		log.Fatalf("❌ 解析數據檔案失敗: %v", err)
	}

	// 讀取向量檔案
	embeddingMap := make(map[string]embeddingEntry)
	if embeddingBytes, err := os.ReadFile(embeddingPath); err == nil {
		var embeddings []embeddingEntry
		if err := json.Unmarshal(embeddingBytes, &embeddings); err == nil {
			for _, emb := range embeddings {
				embeddingMap[emb.ID] = emb
			}
		} else {
			log.Printf("⚠️ 向量檔案格式錯誤: %v", err)
		}
	} else {
		log.Printf("⚠️ 向量檔案不存在: %v", err)
	}

	total := 0
	withEmbedding := 0
	withoutEmbedding := []string{}
	versionStats := make(map[string]int)

	for _, data := range corpusData {
		// 語系過濾
		if data.Locale != "" && locale != "" && data.Locale != locale {
			continue
		}

		total++
		if emb, exists := embeddingMap[data.ID]; exists && len(emb.Embedding) > 0 {
			withEmbedding++
			versionStats[emb.Version]++
		} else {
			withoutEmbedding = append(withoutEmbedding, data.ID)
		}
	}

	log.Printf("📊 語料庫狀態統計:")
	log.Printf("  總條目數: %d", total)
	log.Printf("  已有向量: %d", withEmbedding)
	log.Printf("  缺少向量: %d", total-withEmbedding)

	if len(versionStats) > 0 {
		log.Printf("📅 版本分佈:")
		for version, count := range versionStats {
			log.Printf("  %s: %d 個條目", version, count)
		}
	}

	if len(withoutEmbedding) > 0 {
		log.Printf("⚠️ 缺少 embedding 的條目:")
		for _, id := range withoutEmbedding {
			log.Printf("  - %s", id)
		}
		log.Printf("💡 請執行 'make nsfw-embeddings' 計算缺少的向量")
	} else {
		log.Printf("✅ 所有向量已就緒，可以正常使用")
	}
}