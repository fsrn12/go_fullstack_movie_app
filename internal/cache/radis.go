package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// NewRedisClient creates and returns a new Redis client
func NewRedisClient(redisURL string, logger logging.Logger) (*redis.Client, error) {
	metaData := common.Envelop{}
	if redisURL == "" {
		metaData["error"] = "REDIS_URL is not configured in environment"
		appErr := apperror.ErrRedisURLNotSet(nil, logger, metaData)
		appErr.LogError()
		return nil, appErr
	}

	// Parse the Redis URL to extract options
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		metaData["error"] = "failed to parse redis url"
		appErr := apperror.ErrInternalServer(err, logger, metaData)
		appErr.LogError()
		return nil, appErr
	}

	// Set up TLS for Upstash
	opt.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	client := redis.NewClient(opt)

	// Connection test
	if err := client.Ping(ctx).Err(); err != nil {
		metaData["error"] = "failed to ping redis"
		appErr := apperror.ErrInternalServer(err, logger, metaData)
		appErr.LogError()
		return nil, err
	}

	logger.Info("✈️ Successfully connected to Redis")
	return client, nil
}

// GetCache retieves cached data from Redis
func GetCache(redisClient *redis.Client, key string, dest any) bool {
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false
		}
		return false
	}

	err = json.Unmarshal([]byte(val), dest)
	return err == nil
}

// SetCache sets data to Redis cache
func SetCache(redisClient *redis.Client, key string, data any, expiration time.Duration) error {
	serializedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to serialized data for redis cache")
	}

	// Set the cache with the specified expiration time
	err = redisClient.Set(ctx, key, serializedData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set data in redis cache")
	}

	return nil
}
