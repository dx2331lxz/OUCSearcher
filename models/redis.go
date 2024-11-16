package models

import (
	"OUCSearcher/database"
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"hash/fnv"
	"log"
	"sync"
)

// GetUrlsFromMysql 从mysql中取出1000条url
func GetUrlsFromMysql(n int, wg *sync.WaitGroup) {
	defer wg.Done()
	// 从mysql中取出1000条url
	// 获取TableSuffix
	tableSuffix := fmt.Sprintf("%02x", n)
	// 获取未爬取的url
	urls, err := GetNUnCrawled(tableSuffix, 1000)
	if err != nil {
		return
	}
	// 存入redis
	for _, url := range urls {
		err := database.RDB.LPush(context.Background(), "urls", url).Err()
		if err != nil {
			log.Println("Error pushing url to redis:", err)
		}
		//	 添加到all_urls集合
		err = AddUrlToAllUrlSet(url)
		if err != nil {
			log.Println("Error adding url to all_urls:", err)
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
	url, err := database.RDB.BRPop(context.Background(), 20, "urls").Result()
	if err != nil {
		return "", err
	}
	// 确保返回结果长度正确
	if len(url) != 2 {
		return "", fmt.Errorf("unexpected response from BRPop: %v", url)
	}
	return url[1], nil
}

// AddUrlToVisitedSet AddUrlToSet 添加 URL 到去重集合
func AddUrlToVisitedSet(url string) error {
	ctx := context.Background()
	// 计算 URL 哈希值
	h := fnv.New32a()
	_, err := h.Write([]byte(url))
	if err != nil {
		// 错误处理：计算哈希值失败
		return fmt.Errorf("failed to calculate hash for URL %s: %v", url, err)
	}
	hashedUrl := fmt.Sprintf("%d", h.Sum32()) // 将哈希值转换为字符串形式

	// 将哈希值添加到 Redis 的 set 集合中
	err = database.RDB.SAdd(ctx, "visited_urls", hashedUrl).Err()
	if err != nil {
		// 错误处理：Redis 操作失败
		return fmt.Errorf("failed to add hashed URL to Redis set: %v", err)
	}

	log.Printf("Added hashed URL to set: %s (original URL: %s)", hashedUrl, url)
	return nil
}

// IsUrlVisited 检查 URL 是否已存在
func IsUrlVisited(url string) (bool, error) {
	// 计算 URL 哈希值
	h := fnv.New32a()
	_, err := h.Write([]byte(url))
	if err != nil {
		// 错误处理：计算哈希值失败
		log.Printf("failed to calculate hash for URL %s: %v\n", url, err)
	}
	hashedUrl := fmt.Sprintf("%d", h.Sum32()) // 将哈希值转换为字符串形式
	return database.RDB.SIsMember(context.Background(), "visited_urls", hashedUrl).Result()
}

// AddUrlToAllUrlSet 添加所有url的哈希值到add_urls
func AddUrlToAllUrlSet(url string) error {
	ctx := context.Background()
	// 计算 URL 哈希值
	h := fnv.New32a()
	_, err := h.Write([]byte(url))
	if err != nil {
		// 错误处理：计算哈希值失败
		return fmt.Errorf("failed to calculate hash for URL %s: %v", url, err)
	}
	hashedUrl := fmt.Sprintf("%d", h.Sum32()) // 将哈希值转换为字符串形式

	// 将哈希值添加到 Redis 的 set 集合中
	err = database.RDB.SAdd(ctx, "all_urls", hashedUrl).Err()
	if err != nil {
		// 错误处理：Redis 操作失败
		return fmt.Errorf("failed to add hashed URL to Redis set: %v", err)
	}

	log.Printf("Added hashed URL to set: %s (original URL: %s)", hashedUrl, url)
	return nil
}

// 判断是否在all_urls中
func IsUrlInAllUrls(url string) (bool, error) {
	// 计算 URL 哈希值
	h := fnv.New32a()
	_, err := h.Write([]byte(url))
	if err != nil {
		// 错误处理：计算哈希值失败
		log.Printf("failed to calculate hash for URL %s: %v\n", url, err)
	}
	hashedUrl := fmt.Sprintf("%d", h.Sum32()) // 将哈希值转换为字符串形式
	return database.RDB.SIsMember(context.Background(), "all_urls", hashedUrl).Result()
}