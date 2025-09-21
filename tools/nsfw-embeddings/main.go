package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// Types are defined in nsfw_types.go

func main() {
	utils.LoadEnv()
	utils.InitLogger()

	dataPath := utils.GetEnvWithDefault("NSFW_CORPUS_DATA_PATH", "configs/nsfw/corpus.json")
	embeddingPath := utils.GetEnvWithDefault("NSFW_CORPUS_EMBEDDING_PATH", "configs/nsfw/embeddings.json")
	locale := utils.GetEnvWithDefault("NSFW_RAG_LOCALE", "zh-Hant")

	log.Printf("🎯 開始預計算 NSFW 語料庫 embedding 向量")
	log.Printf("📝 數據檔案: %s", dataPath)
	log.Printf("🧠 向量檔案: %s", embeddingPath)
	log.Printf("🌐 語系過濾: %s", locale)

	// 讀取數據檔案
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		log.Fatalf("讀取數據檔案失敗: %v", err)
	}

	var corpusData []corpusDataEntry
	if err := json.Unmarshal(dataBytes, &corpusData); err != nil {
		log.Fatalf("解析數據檔案失敗: %v", err)
	}

	// 讀取現有向量檔案（如果存在）
	embeddingMap := make(map[string]embeddingEntry)
	if embeddingBytes, err := os.ReadFile(embeddingPath); err == nil {
		var existingEmbeddings []embeddingEntry
		if err := json.Unmarshal(embeddingBytes, &existingEmbeddings); err == nil {
			for _, emb := range existingEmbeddings {
				embeddingMap[emb.ID] = emb
			}
			log.Printf("📊 載入 %d 個現有向量", len(embeddingMap))
		} else {
			log.Printf("⚠️ 向量檔案格式錯誤，將重新計算所有向量: %v", err)
		}
	} else {
		log.Printf("💡 向量檔案不存在，將計算所有向量")
	}

	log.Printf("📊 發現 %d 個語料條目", len(corpusData))

	// 初始化 embedding 客戶端
	embedClient, err := services.NewOpenAIEmbeddingClient()
	if err != nil {
		log.Fatalf("初始化 OpenAI embedding 客戶端失敗: %v", err)
	}

	// 預計算 embedding 向量
	var newEmbeddings []embeddingEntry
	updatedCount := 0
	totalCost := 0.0
	today := time.Now().Format("2006-01-02")

	for _, data := range corpusData {
		// 語系過濾
		if data.Locale != "" && locale != "" && data.Locale != locale {
			log.Printf("⏭️  跳過 %s (語系: %s)", data.ID, data.Locale)
			continue
		}

		// 檢查是否已有 embedding 且版本最新
		if existing, exists := embeddingMap[data.ID]; exists && existing.Version == today {
			log.Printf("✅ 跳過 %s (已有 embedding 且版本最新)", data.ID)
			newEmbeddings = append(newEmbeddings, existing)
			continue
		}

		// 計算 embedding
		log.Printf("🔄 計算 %s embedding...", data.ID)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// 正規化文本
		normalized := normalize(data.Text)
		vector, err := embedClient.EmbedText(ctx, normalized)
		cancel()

		if err != nil {
			log.Printf("❌ 計算 %s embedding 失敗: %v", data.ID, err)
			// 如果有舊向量，保留舊向量
			if existing, exists := embeddingMap[data.ID]; exists {
				log.Printf("🔄 保留 %s 的舊向量", data.ID)
				newEmbeddings = append(newEmbeddings, existing)
			}
			continue
		}

		// 轉換為 float64
		embedding := make([]float64, len(vector))
		for j, v := range vector {
			embedding[j] = float64(v)
		}

		// 創建新的 embedding 條目
		newEntry := embeddingEntry{
			ID:        data.ID,
			Embedding: embedding,
			Version:   today,
		}
		newEmbeddings = append(newEmbeddings, newEntry)

		updatedCount++
		// 估算成本 (text-embedding-3-small: $0.00002/1K tokens)
		estimatedTokens := len(normalized) / 4 // 粗略估算
		cost := float64(estimatedTokens) * 0.00002 / 1000
		totalCost += cost

		log.Printf("✅ 完成 %s (向量維度: %d, 預估成本: $%.6f)",
			data.ID, len(embedding), cost)
	}

	// 寫入向量檔案
	if updatedCount > 0 {
		// 建立備份
		backupPath := embeddingPath + ".backup"
		if _, err := os.Stat(embeddingPath); err == nil {
			if err := os.Rename(embeddingPath, backupPath); err != nil {
				log.Printf("⚠️ 無法建立備份: %v", err)
			}
		}

		// 寫入更新後的向量檔案
		embeddingData, err := json.MarshalIndent(newEmbeddings, "", "  ")
		if err != nil {
			log.Fatalf("序列化向量資料失敗: %v", err)
		}

		if err := os.WriteFile(embeddingPath, embeddingData, 0644); err != nil {
			// 恢復備份
			if _, backupErr := os.Stat(backupPath); backupErr == nil {
				if restoreErr := os.Rename(backupPath, embeddingPath); restoreErr != nil {
					log.Fatalf("寫入失敗且無法恢復備份: %v, %v", err, restoreErr)
				}
			}
			log.Fatalf("寫入向量檔案失敗: %v", err)
		}

		// 刪除備份
		if _, err := os.Stat(backupPath); err == nil {
			if err := os.Remove(backupPath); err != nil {
				log.Printf("⚠️ 刪除備份檔案失敗: %v", err)
			}
		}

		log.Printf("🎉 成功更新 %d 個 embedding 向量", updatedCount)
		log.Printf("💰 總預估成本: $%.6f", totalCost)
		log.Printf("📁 已更新: %s (%.1f KB)", embeddingPath, float64(len(embeddingData))/1024)
	} else {
		log.Printf("✨ 所有語料條目已有 embedding 向量且版本最新，無需更新")
	}
}

// normalize 文本正規化（與主程式保持一致）
func normalize(s string) string {
	return s // 簡化版本，可以後續擴展
}
