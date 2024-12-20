package main

import (
	"OUCSearcher/config"
	"OUCSearcher/database"
	"OUCSearcher/models"
	_ "OUCSearcher/routers" // 引入路由
	"OUCSearcher/tools"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/robfig/cron/v3"
	"golang.org/x/net/html"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const NumberOfCrawl = 2000

// init 初始化数据库连接
func init() {
	cfg := config.NewConfig()
	database.Initialize(cfg)
	database.InitializeRedis(cfg)
}

// 数据库迁移
func migrate() {
	// 迁移数据库
	// 打开数据库连接
	cfg := config.NewConfig()
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	// 创建 256 张表
	//for i := 0; i < 256; i++ {
	//	tableName := fmt.Sprintf("page_%02x", i)
	//	// 自动迁移
	//	err = db.Table(tableName).AutoMigrate(&models.Page{})
	//	if err != nil {
	//		log.Fatal("failed to migrate database:", err)
	//	} else {
	//		log.Printf("Database %s migrated successfully!\n", tableName)
	//	}
	//}
	//db.AutoMigrate(&models.IndexTableStatus{})

	for i := 0; i < 256; i++ {
		tableName := fmt.Sprintf("index1_%02x", i)
		// 自动迁移
		err = db.Table(tableName).AutoMigrate(&models.Index{})
		if err != nil {
			log.Fatal("failed to migrate database:", err)
		} else {
			log.Printf("Database %s migrated successfully!\n", tableName)
		}
	}
	for i := 0; i < 256; i++ {
		tableName := fmt.Sprintf("index_%02x", i)
		// 自动迁移
		err = db.Table(tableName).AutoMigrate(&models.Index{})
		if err != nil {
			log.Fatal("failed to migrate database:", err)
		} else {
			log.Printf("Database %s migrated successfully!\n", tableName)
		}
	}
}

// Fetch downloads the webpage and returns its HTML content
func Fetch(url string) (*html.Node, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置自定义的 Header，比如 User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Mobile Safari/537.36 OUCSpider/1.0")

	// 使用 http.Client 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 检查返回状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch: %s, status: %d", url, resp.StatusCode)
	}

	// 解析HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %s", url)
	}
	return doc, nil
}

// RenderHTML 将 *html.Node 转换为字符串形式的 HTML
func RenderHTML(n *html.Node) string {
	var b strings.Builder
	err := html.Render(&b, n)
	if err != nil {
		return ""
	}
	return b.String()
}

// worker 用于下载网页并提取链接
func worker(url string, wg *sync.WaitGroup) error {
	defer wg.Done()
	//fmt.Println("Fetching:", url)
	doc, err := Fetch(url)
	if err != nil {
		//log.Println("Error fetching:", url, err)
		return err
	}
	page := &models.Page{Url: url}
	// 解析当前url
	// 解析url
	urlInfo := tools.ParseURL(url)
	page.Host = urlInfo.Host
	page.Scheme = urlInfo.Scheme
	page.Domain1 = urlInfo.Domain1
	page.Domain2 = urlInfo.Domain2
	page.Path = urlInfo.Path
	page.Query = urlInfo.Query
	//fmt.Println(page)
	// 解析标题
	//fmt.Println(tools.ExtractTitle(doc))
	page.Title = tools.ExtractTitle(doc)
	// 解析文本
	//fmt.Println(tools.ExtractText(doc))
	page.Text = tools.ExtractText(doc)
	// 在正文后面添加标题，便于搜索标题
	page.Text = page.Text + " " + page.Title + " " + page.Title
	// CrawTime    time.Time `gorm:"default:'2001-01-01 00:00:01'"`
	page.CrawTime = time.Now()
	page.CrawDone = 1

	_, err = page.Update()

	if err != nil {
		log.Println("Error updating or creating page:", url, err)
		return err
	}

	err = models.AddUrlToVisitedSet(url)
	if err != nil {
		return err
	}

	// 提取链接
	//fmt.Println(pId)
	links := tools.ExtractLinks(doc, url)
	links = tools.FilterUrl(links)
	for _, link := range links {
		isIn, err := models.IsUrlInAllUrls(link)
		if err != nil {
			log.Println("Error checking if url is in all urls:", link, err)
			return err
		}
		if !isIn {
			err = models.AddUrlToAllUrlSet(link)
			if err != nil {
				return err
			}
			chilePage := &models.Page{Url: link, CrawTime: time.Now()}
			_, err := chilePage.Create()
			if err != nil {
				log.Println("Error updating or creating page:", link, err)
				return err
			} else {
				log.Println("添加链接：", link)
			}

		} else {
			//log.Println("URL already in all urls:", link)
		}
	}

	return nil
}

// crawl 用于爬取网页
func crawl() {
	var wg sync.WaitGroup

	for i := 0; i < NumberOfCrawl; i++ {
		wg.Add(1)
		go func() {
			// 获取url
			url, err := models.GetUrlFromRedis()
			if err != nil {
				log.Println("Error getting url from redis:", err)
				return
			}
			if url == "" {
				log.Println("No URL fetched, skipping...")
				return // 空值直接跳过
			}
			isVisited, err := models.IsUrlVisited(url)
			if err != nil {
				log.Println("Error getting url from redis:", err)
				return
			}
			if isVisited {
				//log.Println("URL has been visited, skipping...")
				return
			}

			err = worker(url, &wg)
			if err != nil {
				log.Println("Error fetching:", url, err)
			}
		}()
	}
	wg.Wait()
}

// CrawlTimer 定时启动crawl
func CrawlTimer() {
	c := cron.New(cron.WithSeconds())
	// 每个10s执行一次
	c.AddFunc("*/20 * * * * *", crawl)
	c.Start()
}

//// 定时更新page的爬取状态
//func updateCrawDoneTimer() {
//	c := cron.New(cron.WithSeconds())
//	// 每天执行一次
//	c.AddFunc("0 0 0 * * *", func() {
//		err := models.DeleteAllVisitedUrls()
//		if err != nil {
//			log.Println("Error deleting all visited urls:", err)
//			return
//		}
//		err = models.SetCrawDoneToZero()
//		if err != nil {
//			log.Println("Error setting craw_done to zero:", err)
//			return
//		}
//	})
//	c.Start()
//}

func updateDicDone(cronJob *tools.CronJob) {
	// 停止GenerateInvertedIndexAndAddToRedisTimer定时器
	cronJob.StopTask("GenerateInvertedIndexAndAddToRedis")
	fmt.Println("停止GenerateInvertedIndexAndAddToRedis定时器")
	// 停止SaveInvertedIndexStringToMysql定时器
	cronJob.StopTask("SaveInvertedIndexStringToMysql")
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

// 更新页面的分词状态SetDicDoneToZero
func updateDicDoneTimer(cronJob *tools.CronJob) {
	c := cron.New(cron.WithSeconds())
	// 每天执行一次
	c.AddFunc("0 0 0 * * *", func() {
		// 如果redis中的列表数量小于1000则更新
		listKeysCount, err := models.GetListKeysCount()
		if err != nil {
			log.Println("Error getting list keys count:", err)
			return
		}
		if listKeysCount < 1000 {
			updateDicDone(cronJob)
		}
	})

	c.Start()
}

func main() {
	currentTime := time.Now().Format("2006-01-02") // 格式化为 YYYY-MM-DD
	logFileName := currentTime + ".log"
	// 创建一个日志文件
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// 设置日志输出到文件
	log.SetOutput(file)
	// 设置日志格式，记录文件名和行号
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	serviceLogFile := "log/" + time.Now().Format("2006-01-02") + ".log"
	logs.SetLogger(logs.AdapterFile, `{
        "filename": "`+serviceLogFile+`",
        "daily": false,
        "maxlines": 1000,
        "maxsize": 0,
        "level": 7, 
        "perm": "0660"
    }`)
	logs.Info(serviceLogFile)
	logs.SetLogFuncCallDepth(3)
	logs.SetLogFuncCall(true) // 记录文件名和行号

	// 迁移数据库
	//migrate()

	// 创建定时器
	cronJob := tools.NewCronJob()

	// 启动redis从mysql获取urls
	models.GetUrlsFromMysqlTimer()

	// 开始爬取，定时爬取，每隔一段时间爬取一次
	CrawlTimer()

	// 启动定时任务，生成倒排索引并且将结果添加到redis中
	//tools.GenerateInvertedIndexAndAddToRedisTimer()
	cronJob.StartTask("GenerateInvertedIndexAndAddToRedis")

	// 启动定时任务，将倒排索引存入mysql
	//tools.SaveInvertedIndexStringToMysqlTimer()
	cronJob.StartTask("SaveInvertedIndexStringToMysql")

	////启动定时任务，更新爬取状态
	cronJob.StartTask("UpdateCrawDone")
	//// 启动定时任务，更新分词状态
	updateDicDoneTimer(cronJob)

	beego.Run()
	database.Close()
	database.CloseRedis()
}

//// 启动定时任务，将索引迁移到索引表
//tools.Index2IndexsTimer()
