package tools

import (
	"OUCSearcher/database"
	"OUCSearcher/models"
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/yanyiwu/gojieba"
	"log"
	"strconv"
	"sync"
)

// 不需要进行分词的字符列表
var stopWords = []string{
	" ", "，", "。", "！", "？", "、", "；", "：", "“", "”", "‘", "’", "（", "）", "《", "》", "--", "……", "-", "～", "·", "•", "｜", "「", "」", "『", "』", "【", "】", "［", "］", "＜", "＞", "〈", "〉",
}

func fenci(s string) []string {
	var words []string
	use_hmm := true
	x := gojieba.NewJieba()
	defer x.Free()
	words = x.CutForSearch(s, use_hmm)
	return words
}

// 定时执行generateInvertedIndexAndAddToRedis
func GenerateInvertedIndexAndAddToRedisTimer() {
	// 使用协程执行定时任务
	c := cron.New(cron.WithSeconds())
	// 每个10s执行一次
	c.AddFunc("*/10 * * * * *", func() {
		generateInvertedIndexAndAddToRedis()
	})
	c.Start()
}

// 定时执行saveInvertedIndexStringToMysql
func SaveInvertedIndexStringToMysqlTimer() {
	// 使用协程执行定时任务
	c := cron.New(cron.WithSeconds())
	// 每个10s执行一次
	c.AddFunc("*/120 * * * * *", func() {
		err := saveInvertedIndexStringToMysql()
		if err != nil {
			return
		}
	})
	c.Start()
}

// 生成倒排索引
// 分表的顺序，例如 0f 转为十进制为 15 strconv.Itoa(i) + "," + // pages.id 该 URL 的主键 ID strconv.Itoa(int(pages.ID)) + "," + // 词频：这个词在该 HTML 中出现的次数 strconv.Itoa(v.count) + "," + // 该 HTML 的总长度，BM25 算法需要 strconv.Itoa(textLength) + ","  + // 不同 page 之间的间隔符 "-"
// 生成倒排索引并且将结果添加到redis中
func generateInvertedIndexAndAddToRedis() {
	var wg sync.WaitGroup
	// 从数据库中取出所有的数据
	for i := 0; i < 256; i++ {
		// 获取TableSuffix
		tableSuffix := fmt.Sprintf("%02x", i)
		// 获取未分词的数据
		pageDics, err := models.GetNUnDicDone(tableSuffix, 5)
		if err != nil {
			log.Println("Error getting pageDics from mysql:", err)
			return
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
				fmt.Println(wordCount)
				// 生成倒排索引
				for word, count := range wordCount {
					// 判断word是否存在，存在则将新的数据添加到后面，不存在则直接添加
					err := database.RDB.RPush(context.Background(), word, tableSuffix+","+strconv.Itoa(int(Dic.ID))+","+strconv.Itoa(count)+","+strconv.Itoa(len(Dic.Text))).Err()
					if err != nil {
						log.Println("Error pushing word to redis:", err)
					}
					//	 更新数据库，将dic_done设置为1
					_, err = models.UpdateDicDone(tableSuffix, Dic.ID)
					if err != nil {
						log.Println("Error updating dic_done:", err)
						return
					}
				}
			}(Dic)
		}
	}
	wg.Wait()
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
	for _, stopWord := range stopWords {
		delete(wordsMap, stopWord)
	}
	return wordsMap
}

// 从redis中整合倒排索引字符串
func integrateInvertedIndexString() map[string]string {
	// 从redis随机取出1个不为urls的key
	key, err := database.RDB.RandomKey(context.Background()).Result()
	if err != nil {
		return nil
	}
	// 过滤掉无关的key
	if key == "urls" || key == "visited_urls" || key == "all_urls" {
		return nil
	}
	// 取出key对应的value
	values, err := database.RDB.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil
	}
	// 判断是否个数大于2
	if len(values) < 2 {
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
	// 返回整合后的倒排索引字符串
	return map[string]string{key: indexString}
}

func getIntegrateInvertedIndexString(n int) map[string]string {
	indexStrings := make(map[string]string)
	// 使用 channel 传递每个 goroutine 计算的部分结果
	channel := make(chan map[string]string, n)

	// 等待组
	var wg sync.WaitGroup

	// 在主 goroutine 中合并部分结果
	go func() {
		for partIndex := range channel {
			for key, value := range partIndex {
				if _, ok := indexStrings[key]; ok {
					indexStrings[key] += "-" + value
				} else {
					indexStrings[key] = value
				}
			}
		}
	}()

	// 启动多个 goroutine 并将每个结果发送到 channel
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 获取每个 goroutine 的局部结果
			indexString := integrateInvertedIndexString()
			if indexString != nil {
				// 将结果通过 channel 传回主线程
				channel <- indexString
			}
		}()
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	// 关闭 channel，确保没有更多的发送操作
	close(channel)

	// 返回合并后的结果
	return indexStrings
}

// 使用mysql事务将倒排索引字符串存入数据库
func saveInvertedIndexStringToMysql() error {
	indexStrings := getIntegrateInvertedIndexString(10)
	err := models.SaveMapToDB(indexStrings)
	if err != nil {
		return err
	}
	return nil
}
