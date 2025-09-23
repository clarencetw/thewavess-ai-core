package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/clarencetw/thewavess-ai-core/database"
	"github.com/clarencetw/thewavess-ai-core/models/db"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/go-redis/redis/v8"
)

// RelationshipCache 簡化的關係狀態快取服務
type RelationshipCache struct {
	client *redis.Client
	prefix string
}

// NewRelationshipCache 創建關係狀態快取服務
func NewRelationshipCache() *RelationshipCache {
	return &RelationshipCache{
		client: database.GetRedisClient(), // nil = 降級模式
		prefix: "thewavess:",
	}
}

// GetRelationship 獲取關係狀態快取
func (r *RelationshipCache) GetRelationship(ctx context.Context, userID, characterID, chatID string) (*db.RelationshipDB, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("%srelationship:%s:%s:%s", r.prefix, userID, characterID, chatID)

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 快取未命中
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var relationship db.RelationshipDB
	if err := json.Unmarshal([]byte(val), &relationship); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	utils.Logger.WithField("cache_key", key).Debug("關係狀態快取命中")
	return &relationship, nil
}

// SetRelationship 設置關係狀態快取
func (r *RelationshipCache) SetRelationship(ctx context.Context, userID, characterID, chatID string, relationship *db.RelationshipDB, ttl time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("%srelationship:%s:%s:%s", r.prefix, userID, characterID, chatID)

	data, err := json.Marshal(relationship)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	utils.Logger.WithFields(map[string]interface{}{
		"cache_key": key,
		"ttl": ttl,
	}).Debug("關係狀態快取已設置")
	return nil
}

// DeleteRelationship 刪除關係狀態快取
func (r *RelationshipCache) DeleteRelationship(ctx context.Context, userID, characterID, chatID string) error {
	if r.client == nil {
		return fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("%srelationship:%s:%s:%s", r.prefix, userID, characterID, chatID)
	return r.client.Del(ctx, key).Err()
}