package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/dgraph-io/ristretto"
)

// CharacterCache 高性能角色快取，使用 Ristretto
type CharacterCache struct {
	cache *ristretto.Cache
	store *CharacterStore
}

// NewCharacterCache 創建角色快取服務
func NewCharacterCache(store *CharacterStore) *CharacterCache {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e3,     // 1000 個計數器
		MaxCost:     1 << 20, // 1MB 限制
		BufferItems: 64,      // 64 個緩衝項
		Metrics:     true,    // 啟用統計（必需）
	})
	if err != nil {
		utils.Logger.WithError(err).Fatal("Failed to create cache")
	}

	return &CharacterCache{
		cache: cache,
		store: store,
	}
}

// GetCharacter 獲取角色
func (c *CharacterCache) GetCharacter(ctx context.Context, id string) (*models.Character, error) {
	if id == "" {
		return nil, fmt.Errorf("character ID required")
	}

	// 嘗試快取
	if val, found := c.cache.Get(id); found {
		if char, ok := val.(*models.Character); ok {
			return char, nil
		}
	}

	// 查詢資料庫
	char, err := c.store.GetAggregate(ctx, id)
	if err != nil {
		return nil, err
	}

	// 寫入快取（固定成本 1）
	c.cache.Set(id, char, 1)

	// Ristretto 異步寫入，短暫等待確保寫入
	time.Sleep(10 * time.Millisecond)

	return char, nil
}

// InvalidateCharacter 清除快取
func (c *CharacterCache) InvalidateCharacter(ctx context.Context, id string) {
	c.cache.Del(id)
}

// GetCacheStats 基本統計
func (c *CharacterCache) GetCacheStats() map[string]interface{} {
	metrics := c.cache.Metrics
	hits := metrics.Hits()
	misses := metrics.Misses()
	total := hits + misses

	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(hits) / float64(total)
	}

	return map[string]interface{}{
		"hits":      hits,
		"misses":    misses,
		"hit_ratio": hitRatio,
	}
}

// 全域實例
var (
	characterCacheInstance *CharacterCache
	characterCacheOnce     sync.Once
)

// GetCharacterCache 獲取全域實例
func GetCharacterCache() *CharacterCache {
	characterCacheOnce.Do(func() {
		store := NewCharacterStore()
		characterCacheInstance = NewCharacterCache(store)
	})
	return characterCacheInstance
}
