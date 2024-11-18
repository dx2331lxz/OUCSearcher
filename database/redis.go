package database

import (
	"OUCSearcher/config"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)

var RDB *redis.Client

func InitializeRedis(cfg *config.Config) {
	RDB = redis.NewClient(&redis.Options{
		Addr:       cfg.RedisHost + ":" + cfg.RedisPort,
		Password:   cfg.RedisPassword,
		DB:         0,
		PoolSize:   1000,
		MaxRetries: 3,
	})
	// 测试 Redis 连接
	pong, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	fmt.Println("Redis connection successful:", pong)
}

func CloseRedis() {
	if RDB != nil {
		err := RDB.Close()
		if err != nil {
			return
		}
	}
}
