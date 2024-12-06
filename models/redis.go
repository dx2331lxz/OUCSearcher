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
	urls, err := GetNUnCrawled(tableSuffix, 100)
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
	c := cron.New(cron.WithSeconds())
	// 每个10s执行一次
	c.AddFunc("*/120 * * * * *", func() {
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
	url, err := database.RDB.BRPop(context.Background(), 0, "urls").Result()
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
	result, err := database.RDB.SAdd(ctx, "all_urls", hashedUrl).Result()
	if err != nil {
		// 错误处理：Redis 操作失败
		return fmt.Errorf("failed to add hashed URL to Redis set: %v", err)
	}
	if result == 1 {
		//log.Printf("Added hashed URL to set: %s (original URL: %s)", hashedUrl, url)
	} else {
		//log.Printf("URL already in all urls: %s", url)
	}

	return nil
}

// IsUrlInAllUrls 判断是否在all_urls中
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

// DeleteAllVisitedUrls 删除visited_urls中的所有
func DeleteAllVisitedUrls() error {
	_, err := database.RDB.Del(context.Background(), "visited_urls").Result()
	if err != nil {
		return fmt.Errorf("failed to delete visited_urls: %v", err)
	}
	return nil
}

// GetListKeysCount 获取所有类型为 list 的键的数量
func GetListKeysCount() (int64, error) {
	var cursor uint64
	ctx := context.Background()
	var listKeysCount int64

	// 使用 SCAN 命令遍历所有键
	for {
		// SCAN 命令返回游标和匹配的键
		keys, newCursor, err := database.RDB.Scan(ctx, cursor, "*", 0).Result()
		if err != nil {
			return 0, err
		}

		// 更新游标
		cursor = newCursor

		// 遍历返回的键
		for _, key := range keys {
			// 获取键的类型
			keyType, err := database.RDB.Type(ctx, key).Result()
			if err != nil {
				return 0, err
			}

			// 如果键的类型是 list，则计数
			if keyType == "list" {
				listKeysCount++
			}
		}

		// 如果游标为 0，表示扫描完成
		if cursor == 0 {
			break
		}
	}

	return listKeysCount, nil
}

// DeleteListKeysExcludingUrls 删除所有类型为 list 且键名不为 urls 的键
func DeleteListKeysExcludingUrls() error {
	ctx := context.Background()
	var cursor uint64

	// 使用 SCAN 命令遍历所有的键
	for {
		// SCAN 命令返回游标和匹配的键
		keys, newCursor, err := database.RDB.Scan(ctx, cursor, "*", 0).Result()
		if err != nil {
			return err
		}

		// 更新游标
		cursor = newCursor

		// 遍历所有键
		for _, key := range keys {
			// 排除键名为 "urls" 的键
			if key == "urls" {
				continue
			}

			// 获取键的类型
			keyType, err := database.RDB.Type(ctx, key).Result()
			if err != nil {
				return err
			}

			// 如果键的类型是 list，则删除该键
			if keyType == "list" {
				_, err := database.RDB.Del(ctx, key).Result()
				if err != nil {
					return err
				}
				fmt.Printf("Deleted key: %s\n", key)
			}
		}

		// 如果游标为 0，表示扫描完成
		if cursor == 0 {
			break
		}
	}

	return nil
}
