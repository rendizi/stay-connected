package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

var client *redis.Client

func Init() error {
	addr := os.Getenv("REDIS_ADDRESS")
	if addr == "" {
		return errors.New("REDIS_ADDRESS is not set")
	}
	password := os.Getenv("REDIS_PASSWORD")

	client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return nil
}

func StoreInRedis(ctx context.Context, key string, value string, duration time.Duration) error {
	err := client.Set(ctx, key, value, duration).Err()
	if err != nil {
		return fmt.Errorf("failed to store key %s in Redis: %v", key, err)
	}
	return nil
}

func GetFromRedis(ctx context.Context, key string) (string, error) {
	value, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key %s does not exist", key)
	} else if err != nil {
		return "", fmt.Errorf("failed to get key %s from Redis: %v", key, err)
	}
	return value, nil
}
