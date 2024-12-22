package tools

import (
	"OUCSearcher/models"
	"fmt"
	"log"
	"sync"
)

func UpdateCrawDone() error {
	err := models.DeleteAllVisitedUrls()
	if err != nil {
		log.Println("Error deleting all visited urls:", err)
		return err
	}
	err = models.SetCrawDoneToZero()
	if err != nil {
		log.Println("Error setting craw_done to zero:", err)
		return err
	}
	return nil
}

func Crawl() error {
	log.Println("Crawling...")
	crawl()
	log.Println("Crawling done.")
	return nil
}

func GetUrlsFromMysqlJob() error {
	var wg sync.WaitGroup
	count, err := models.GetUrlsCount()
	log.Println("Urls count:", count)
	if err != nil {
		log.Println("Error getting urls count:", err)
		return err
	}
	if count < 10000 {
		for i := 0; i < 256; i++ {
			wg.Add(1)
			go models.GetUrlsFromMysql(i, &wg)
		}
	}
	wg.Wait()
	return nil
}

func UpdateDicDone() {
	// 停止GenerateInvertedIndexAndAddToRedisTimer定时器
	CronJobSub.StopTask("GenerateInvertedIndexAndAddToRedis")
	fmt.Println("停止GenerateInvertedIndexAndAddToRedis定时器")
	// 停止SaveInvertedIndexStringToMysql定时器
	CronJobSub.StopTask("SaveInvertedIndexStringToMysql")
	fmt.Println("停止SaveInvertedIndexStringToMysql定时器")
	// 转换分词表
	err := models.SwitchIndexTable()
	if err != nil {
		log.Println("Error switching index table:", err)
		return
	}
	fmt.Println("转换分词表")

	err = models.SetDicDoneToZero()
	if err != nil {
		log.Println("Error setting dic_done to zero:", err)
		return
	}
	fmt.Println("设置dic_done为0")
	// 清空redis中的列表
	err = models.DeleteListKeysExcludingUrls()
	if err != nil {
		log.Println("Error deleting list keys excluding urls:", err)
		return
	}
	fmt.Println("清空redis中的列表")
	// 置空分词表
	err = models.ClearIndexString()
	if err != nil {
		log.Println("Error clearing index string:", err)
		return
	}
	fmt.Println("置空分词表")

}

// UpdateDicDoneJob 更新页面的分词状态SetDicDoneToZero
func UpdateDicDoneJob() error {
	// 如果redis中的列表数量小于1000则更新
	listKeysCount, err := models.GetListKeysCount()
	if err != nil {
		log.Println("Error getting list keys count:", err)
		return err
	}
	if listKeysCount < 1000 {
		UpdateDicDone()
		// 重新开始GenerateInvertedIndexAndAddToRedisTimer定时器
		CronJobSub.StartTask("GenerateInvertedIndexAndAddToRedis")
		fmt.Println("重新开始GenerateInvertedIndexAndAddToRedis定时器")
		// 重新开始SaveInvertedIndexStringToMysql定时器
		CronJobSub.StartTask("SaveInvertedIndexStringToMysql")
	}
	return nil
}
