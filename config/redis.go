package config

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() error {
	runtimeConfig := RuntimeConfig()

	if !runtimeConfig.RedisEnabled {
		RedisClient = nil
		log.Println("Redis is disabled; skipping initialization")
		return nil
	}

	if runtimeConfig.RedisAddr == "" {
		return fmt.Errorf("REDIS_ADDR is required when REDIS_ENABLED is true")
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:      runtimeConfig.RedisAddr,
		Password:  runtimeConfig.RedisPassword,
		DB:        runtimeConfig.RedisDB,
		TLSConfig: RedisTLSConfig(),
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}
