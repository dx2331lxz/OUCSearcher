package tools

import (
	"OUCSearcher/database"
	"OUCSearcher/models"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/yanyiwu/gojieba"
	"log"
	"sync"
	"time"
)

// 不需要进行分词的字符列表
var stopWords = map[string]struct{}{
	" ": {}, "，": {}, "。": {}, "！": {}, "？": {}, "、": {}, "；": {}, "：": {},
	"“": {}, "”": {}, "‘": {}, "’": {}, "（": {}, "）": {}, "《": {}, "》": {},
	"--": {}, "……": {}, "-": {}, "～": {}, "·": {}, "•": {}, "｜": {}, "「": {},
	"」": {}, "『": {}, "』": {}, "【": {}, "】": {}, "［": {}, "］": {}, "＜": {},
	"＞": {}, "〈": {}, "〉": {}, "%": {}, "(": {}, ")": {}, "&": {}, "+": {}, "/": {},
}

// 全局单例
var (
	jiebaInstance *gojieba.Jieba
	once          sync.Once
)

// 初始化单例实例
func initJieba() *gojieba.Jieba {
	once.Do(func() {
		jiebaInstance = gojieba.NewJieba()
	})
	return jiebaInstance
}

func fenci(s string) []string {
	// 使用全局单例的 jieba 实例
	x := initJieba()
	use_hmm := true
	return x.CutForSearch(s, use_hmm)
}

// SaveInvertedIndexStringToMysqlTimer 定时执行saveInvertedIndexStringToMysql
//func SaveInvertedIndexStringToMysqlTimer() {
//	// 使用协程执行定时任务
//	c := cron.New(cron.WithSeconds())
//	// 每个10s执行一次
//	c.AddFunc("*/10 * * * * *", func() {
//		err := saveInvertedIndexStringToMysql()
//		if err != nil {
//			log.Println("Error saving inverted index to MySQL:", err)
//		}
//	})
//	c.Start()
//}

// GenerateInvertedIndexAndAddToRedis 生成倒排索引
func GenerateInvertedIndexAndAddToRedis() error {

	var wg sync.WaitGroup
	// 从数据库中取出所有的数据
	for i := 0; i < 256; i++ {
		// 获取TableSuffix
		tableSuffix := fmt.Sprintf("%02x", i)
		// 获取未分词的数据
		pageDics, err := models.GetNUnDicDone(tableSuffix, 10)
		if err != nil {
			log.Println("Error getting pageDics from mysql:", err)
			continue
		}
		// 生成倒排索引
		for _, Dic := range pageDics {
			wg.Add(1)
			go func(Dic models.PageDic) {
				defer wg.Done()
				// 分词
				words := fenci(Dic.Text)
				// 统计词频
				wordCount := countWords(words)
				// 过滤停用词
				wordCount = filterStopWords(wordCount)
				//fmt.Println(wordCount)
				log.Println(wordCount)
				// 生成倒排索引
				for word, count := range wordCount {
					err := addToRedis(word, tableSuffix, Dic.ID, count, len(Dic.Text))
					if err != nil {
						log.Println("Error pushing word to redis:", err)
					}
					// 更新数据库，将dic_done设置为1
					err = updateDicDone(tableSuffix, Dic.ID)
					if err != nil {
						log.Println("Error updating dic_done:", err)
					}
				}
			}(Dic)
		}
	}
	wg.Wait()
	return nil
}

// 将词信息添加到Redis
func addToRedis(word, tableSuffix string, DicID, count, textLength int) error {
	// 使用Redis连接池
	err := database.RDB.RPush(context.Background(), word, fmt.Sprintf("%s,%d,%d,%d", tableSuffix, DicID, count, textLength)).Err()
	if err != nil {
		return fmt.Errorf("Error pushing word to redis: %v", err)
	}
	return nil
}

// 更新数据库中dic_done字段
func updateDicDone(tableSuffix string, DicID int) error {

	_, err := models.UpdateDicDone(tableSuffix, DicID)
	if err != nil {
		return fmt.Errorf("Error updating dic_done: %v", err)
	}
	return nil
}

// 统计分词中每个词在整个文章中出现的次数
func countWords(words []string) map[string]int {
	wordCount := make(map[string]int)
	for _, word := range words {
		if _, ok := wordCount[word]; ok {
			wordCount[word]++
		} else {
			wordCount[word] = 1
		}
	}
	return wordCount
}

// 过滤字典中的停用词
func filterStopWords(wordsMap map[string]int) map[string]int {
	for stopWord := range stopWords {
		delete(wordsMap, stopWord)
	}
	return wordsMap
}

// 从redis中整合倒排索引字符串
func integrateInvertedIndexString() map[string]string {
	// 从redis随机取出1个不为urls的key
	key, err := database.RDB.RandomKey(context.Background()).Result()
	if err != nil || key == "urls" || key == "visited_urls" || key == "all_urls" {
		return nil
	}
	// 取出key对应的value
	values, err := database.RDB.LRange(context.Background(), key, 0, -1).Result()
	if err != nil || len(values) < 2 {
		return nil
	}
	// 整合倒排索引字符串
	indexString := ""
	for _, value := range values {
		if indexString == "" {
			indexString = value
		} else {
			indexString += "-" + value
		}
	}
	// 删除key
	err = database.RDB.Del(context.Background(), key).Err()
	if err != nil {
		return nil
	}
	return map[string]string{key: indexString}
}

// 获取整合的倒排索引字符串
func getIntegrateInvertedIndexString(n int) map[string]string {
	indexStrings := make(map[string]string)
	channel := make(chan map[string]string, n)

	var wg sync.WaitGroup

	// 合并部分结果
	go func() {
		for partIndex := range channel {
			for key, value := range partIndex {
				if _, ok := indexStrings[key]; !ok {
					indexStrings[key] = value
					continue
				}
				indexStrings[key] += "-" + value
			}
		}
	}()

	// 启动多个goroutine来计算部分结果
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			indexString := integrateInvertedIndexString()
			if indexString != nil {
				channel <- indexString
			}
		}()
	}

	wg.Wait()
	close(channel)
	return indexStrings
}

// SaveInvertedIndexStringToMysql 使用mysql事务将倒排索引字符串存入数据库
func SaveInvertedIndexStringToMysql() error {
	// 打印当前时间，并且附带信息：开始执行 SaveInvertedIndexStringToMysql
	currentTime := time.Now()
	fmt.Printf("当前时间: %s - 开始执行 SaveInvertedIndexStringToMysql\n", currentTime.Format("2006-01-02 15:04:05"))
	// 获取一百个词的倒排索引
	indexStrings := getIntegrateInvertedIndexString(1000)
	// 查看 indexStrings 是否为空
	if len(indexStrings) == 0 {
		return nil
	}

	// 初始化一个 map 来存储倒排索引
	var indexStringsMap = make(map[string]map[string]string)
	for key, value := range indexStrings {
		hash := md5.New()
		_, err := hash.Write([]byte(key))
		if err != nil {
			return fmt.Errorf("failed to write data to hash: %v", err)
		}
		// 获取哈希值的最后一位（字节）
		hashValue := hash.Sum(nil)

		// 提取最后一个字节
		lastByte := hashValue[len(hashValue)-1]

		// 将最后一个字节转换为小写十六进制字符
		lastHexChar := fmt.Sprintf("%02x", lastByte)

		// 如果 indexStringsMap[lastHexChar] 是 nil，初始化它
		if indexStringsMap[lastHexChar] == nil {
			indexStringsMap[lastHexChar] = make(map[string]string)
		}

		// 现在可以安全地将 key 和 value 插入到 map 中
		indexStringsMap[lastHexChar][key] = value
	}

	// 使用 WaitGroup 进行并发插入
	var wg sync.WaitGroup
	for tableSuffix, indexStrings := range indexStringsMap {
		wg.Add(1)
		go func(indexStrings map[string]string, tableSuffix string) {
			defer wg.Done()
			fmt.Println("begin to save inverted index to mysql table:", tableSuffix)
			err := models.SaveMapToTable(indexStrings, tableSuffix)
			if err != nil {
				log.Println("Error saving inverted index to MySQL:", err)
			}
			fmt.Println("save inverted index to mysql table successful:", tableSuffix)
		}(indexStrings, tableSuffix)
	}
	wg.Wait()
	fmt.Println("all inverted index saved to mysql")
	// 打印当前时间，并且附带信息：结束执行 SaveInvertedIndexStringToMysql
	currentTime = time.Now()
	fmt.Printf("当前时间: %s - 结束执行 SaveInvertedIndexStringToMysql\n", currentTime.Format("2006-01-02 15:04:05"))
	return nil
}
