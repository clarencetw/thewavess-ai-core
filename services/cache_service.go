package services

import (
	"context"
	"fmt"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/dgraph-io/ristretto"
)

// RelationshipCache 關係狀態快取服務，使用 Ristretto
type RelationshipCache struct {
	cache *ristretto.Cache
}

// NewRelationshipCache 創建關係狀態快取服務
func NewRelationshipCache() *RelationshipCache {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // 10000 個計數器（關係數據較多）
		MaxCost:     2 << 20, // 2MB 限制
		BufferItems: 64,      // 64 個緩衝項
		Metrics:     true,    // 啟用統計
	})
	if err != nil {
		utils.Logger.WithError(err).Fatal("Failed to create relationship cache")
	}

	return &RelationshipCache{
		cache: cache,
	}
}

// GetRelationship 獲取關係狀態快取
func (r *RelationshipCache) GetRelationship(ctx context.Context, userID, characterID, chatID string) (*db.RelationshipDB, error) {
	key := fmt.Sprintf("relationship:%s:%s:%s", userID, characterID, chatID)

	if val, found := r.cache.Get(key); found {
		if relationship, ok := val.(*db.RelationshipDB); ok {
			return relationship, nil
		}
	}

	return nil, nil // 快取未命中
}

// SetRelationship 設置關係狀態快取
func (r *RelationshipCache) SetRelationship(ctx context.Context, userID, characterID, chatID string, relationship *db.RelationshipDB, ttl time.Duration) error {
	key := fmt.Sprintf("relationship:%s:%s:%s", userID, characterID, chatID)

	// 使用 TTL 版本寫入快取
	if r.cache.SetWithTTL(key, relationship, 1, ttl) {
		// 短暫等待確保 Ristretto 異步寫入完成
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	return fmt.Errorf("cache set rejected")
}

// DeleteRelationship 刪除關係狀態快取
func (r *RelationshipCache) DeleteRelationship(ctx context.Context, userID, characterID, chatID string) error {
	key := fmt.Sprintf("relationship:%s:%s:%s", userID, characterID, chatID)
	r.cache.Del(key)
	return nil
}

// GetCacheStats 獲取關係快取統計
func (r *RelationshipCache) GetCacheStats() map[string]interface{} {
	metrics := r.cache.Metrics
	hits := metrics.Hits()
	misses := metrics.Misses()
	total := hits + misses

	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(hits) / float64(total)
	}

	return map[string]interface{}{
		"type":      "relationship_cache",
		"hits":      hits,
		"misses":    misses,
		"hit_ratio": hitRatio,
	}
}