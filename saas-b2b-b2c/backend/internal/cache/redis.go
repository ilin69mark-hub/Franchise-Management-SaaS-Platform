package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func ConnectRedis() (*redis.Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	Client = client
	return client, nil
}

func GetRedis() *redis.Client {
	return Client
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("redis client not initialized")
	}
	return Client.Get(ctx, key).Result()
}

func Delete(ctx context.Context, key string) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.Del(ctx, key).Err()
}

func SetWithIncrement(ctx context.Context, key string, expiration time.Duration, limit int) (int, error) {
	if Client == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	pipe := Client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	count := int(incr.Val())
	return count, nil
}

func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}
	return Client.SetNX(ctx, key, value, expiration).Result()
}

func RevokeRefreshToken(ctx context.Context, tokenID string) error {
	if Client == nil {
		return nil
	}
	key := "revoked_token:" + tokenID
	return Client.Set(ctx, key, "1", 7*24*time.Hour).Err()
}

func IsTokenRevoked(ctx context.Context, tokenID string) bool {
	if Client == nil {
		return false
	}
	key := "revoked_token:" + tokenID
	exists, err := Client.Exists(ctx, key).Result()
	return err == nil && exists > 0
}

func CacheUserSession(ctx context.Context, userID string, data interface{}, expiration time.Duration) error {
	if Client == nil {
		return nil
	}
	key := "session:" + userID
	return Client.Set(ctx, key, data, expiration).Err()
}

func GetUserSession(ctx context.Context, userID string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("redis client not initialized")
	}
	key := "session:" + userID
	return Client.Get(ctx, key).Result()
}

func InvalidateUserSession(ctx context.Context, userID string) error {
	if Client == nil {
		return nil
	}
	key := "session:" + userID
	return Client.Del(ctx, key).Err()
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}