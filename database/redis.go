package database

import (
	"OUCSearcher/config"
	"OUCSearcher/models"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
	"sync"

	"log"
)

var RDB *redis.Client

func InitializeRedis(cfg *config.Config) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       0,
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

// GetUrlsFromMysql 从mysql中取出1000条url
func GetUrlsFromMysql(n int, wg *sync.WaitGroup) {
	defer wg.Done()
	// 从mysql中取出1000条url
	// 获取TableSuffix
	tableSuffix := fmt.Sprintf("%02x", n)
	// 获取未爬取的url
	urls, err := models.GetNUnCrawled(tableSuffix, 100)
	if err != nil {
		return
	}
	// 存入redis
	for _, url := range urls {
		err := RDB.LPush(context.Background(), "urls", url).Err()
		if err != nil {
			log.Println("Error pushing url to redis:", err)
		}
	}
}

// GetUrlsFromMysqlTimer 定时任务执行GetUrlsFromMysql
func GetUrlsFromMysqlTimer() {
	// 使用协程执行定时任务
	var wg sync.WaitGroup
	c := cron.New()
	// 每个10s执行一次
	c.AddFunc("*/10 * * * * *", func() {
		for i := 0; i < 256; i++ {
			wg.Add(1)
			go GetUrlsFromMysql(i, &wg)
		}
		wg.Wait()
	})
	c.Start()
}

// GetUrlsFromRedis 从redis中取出n条url
//func GetUrlsFromRedis(n int64) ([]string, error) {
//	// 从redis中取出n条url
//	urls, err := RDB.LRange(context.Background(), "urls", 0, n-1).Result()
//	if err != nil {
//		return nil, err
//	}
//	// 删除取出的url
//	err = RDB.LTrim(context.Background(), "urls", int64(len(urls)), -1).Err()
//	if err != nil {
//		return nil, err
//	}
//	return urls, nil
//}

// GetUrlFromRedis 从redis中阻塞地取出1条url
func GetUrlFromRedis() (string, error) {
	url, err := RDB.BRPop(context.Background(), 20, "urls").Result()
	if err != nil {
		return "", err
	}
	// 确保返回结果长度正确
	if len(url) != 2 {
		return "", fmt.Errorf("unexpected response from BRPop: %v", url)
	}
	return url[1], nil
}

// AddUrlToSet 添加 URL 到去重集合
func AddUrlToSet(url string) error {
	return RDB.SAdd(context.Background(), "visited_urls", url).Err()
}

// IsUrlVisited 检查 URL 是否已存在
func IsUrlVisited(url string) (bool, error) {
	return RDB.SIsMember(context.Background(), "visited_urls", url).Result()
}
