package main

import (
	"OUCSearcher/config"
	"OUCSearcher/database"
	"OUCSearcher/models"
	"OUCSearcher/tools"
	"fmt"
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

const NumberOfCrawl = 1000

// init 初始化数据库连接
func init() {
	cfg := config.NewConfig()
	database.Initialize(cfg)
	database.InitializeRedis(cfg)

	// 迁移数据库
	// 打开数据库连接
	//db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	//if err != nil {
	//	log.Fatal("failed to connect database:", err)
	//}
	//
	//// 自动迁移
	//err = db.AutoMigrate(&models.Page{})
	//if err != nil {
	//	log.Fatal("failed to migrate database:", err)
	//}
	//
	//log.Println("Database migrated successfully!")
}

// 数据库迁移
func migrate() {
	cfg := config.NewConfig()
	database.Initialize(cfg)
	// 迁移数据库
	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	//// 创建 256 张表
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
	//	 自动迁移index表
	err = db.AutoMigrate(&models.Index{})
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
	client := &http.Client{}
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
		log.Println("Error fetching:", url, err)
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

	//rowsAffected, err := res.RowsAffected()
	//
	//if err != nil {
	//	log.Println("Error getting rows affected:", url, err)
	//	return err
	//}
	//if rowsAffected == 0 {
	//	log.Println("No rows affected:", url)
	//	return fmt.Errorf("no rows affected")
	//} else if rowsAffected > 1 {
	//	log.Println("Multiple rows affected:", url)
	//	return fmt.Errorf("有重复的行")
	//} else {
	//	log.Println("One row affected:", url)
	//}

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
			log.Println("URL already in all urls:", link)
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
				log.Println("URL has been visited, skipping...")
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

// 定时启动crawl
func CrawlTimer() {
	c := cron.New(cron.WithSeconds())
	// 每个10s执行一次
	c.AddFunc("*/10 * * * * *", crawl)
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

	// 迁移数据库
	//migrate()

	// 启动redis从mysql获取urls
	//models.GetUrlsFromMysqlTimer()

	// 启动定时任务，生成倒排索引并且将结果添加到redis中
	tools.GenerateInvertedIndexAndAddToRedisTimer()

	// 启动定时任务，将倒排索引存入mysql
	// todo 处理没有数据情况
	//tools.SaveInvertedIndexStringToMysqlTimer()

	// 开始爬取，定时爬取，每隔一段时间爬取一次
	//CrawlTimer()
	select {}
	database.Close()
	database.CloseRedis()
}
