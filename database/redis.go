package database

import (
	"context"
	"sync"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// GetRedisClient 獲取全局 Redis 客戶端（單例模式）
func GetRedisClient() *redis.Client {
	redisOnce.Do(func() {
		redisClient = initRedisClient()
	})
	return redisClient
}

// initRedisClient 初始化 Redis 客戶端（遵循官方生產環境最佳實踐）
func initRedisClient() *redis.Client {
	opt := &redis.Options{
		// 基本配置
		Addr:     utils.GetEnvWithDefault("REDIS_ADDR", "localhost:6379"),
		Password: utils.GetEnvWithDefault("REDIS_PASSWORD", ""),
		DB:       utils.GetEnvIntWithDefault("REDIS_DB", 0),

		// 連接池配置
		PoolSize:     utils.GetEnvIntWithDefault("REDIS_POOL_SIZE", 10),
		MinIdleConns: utils.GetEnvIntWithDefault("REDIS_MIN_IDLE", 5),
		PoolTimeout:  4 * time.Second,

		// 超時配置（官方推薦）
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,

		// 重試配置（官方推薦）
		MaxRetries:      5,
		MinRetryBackoff: 10 * time.Millisecond,
		MaxRetryBackoff: 100 * time.Millisecond,
	}

	client := redis.NewClient(opt)

	// 健康檢查（官方推薦的 PING 方式）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		utils.Logger.WithError(err).Warning("Redis 連接失敗，降級為記憶體模式")
		client.Close()
		return nil
	}

	utils.Logger.Info("Redis 客戶端初始化成功")
	return client
}

// IsRedisAvailable 檢查 Redis 是否可用（官方推薦的健康檢查方式）
func IsRedisAvailable() bool {
	client := GetRedisClient()
	if client == nil {
		return false
	}

	// 使用官方推薦的 PING 命令進行健康檢查
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		utils.Logger.WithError(err).Debug("Redis 健康檢查失敗")
		return false
	}

	return true
}