package main

import (
	"OUCSearcher/config"
	"OUCSearcher/database"
	"OUCSearcher/models"
	"OUCSearcher/tools"
	"fmt"
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

const NumberOfCrawl = 200

// init 初始化数据库连接
func init() {
	cfg := config.NewConfig()
	database.Initialize(cfg)
	// 迁移数据库
	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&models.Page{})
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	log.Println("Database migrated successfully!")
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
	fmt.Println("Fetching:", url)

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

	res, pId, err := page.UpdateOrCreateByUrl()
	if err != nil {
		log.Println("Error updating or creating page:", url, err)
		return err
	}
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Println("Error getting rows affected:", url, err)
		return err
	}
	if rowsAffected == 0 {
		log.Println("No rows affected:", url)
		return fmt.Errorf("no rows affected")
	} else if rowsAffected > 1 {
		log.Println("Multiple rows affected:", url)
		return fmt.Errorf("有重复的行")
	} else {
		log.Println("One row affected:", url)
	}

	// 提取链接
	//fmt.Println(pId)
	links := tools.ExtractLinks(doc, url)
	links = tools.FilterUrl(links)

	for _, link := range links {
		chilePage := &models.Page{Url: link, ReferrerId: pId, CrawTime: time.Now()}
		res, err := chilePage.CreateOrPassByUrl()
		if err != nil {
			log.Println("Error updating or creating page:", link, err)
			return err
		}
		if res == nil {
			log.Println("已经爬取在数据库中：", link)
		} else {
			log.Println("添加链接：", link)
		}
	}

	return nil
}

// 从数据库中取出10条未爬取的链接进行并发爬取
func crawl() int {
	// 取出10条未爬取的链接
	urls, err := models.GetNUnCrawled(NumberOfCrawl)
	if err != nil {
		log.Println("Error getting uncrawled pages:", err)
		return 0
	}
	log.Println("Fetched", len(urls), "uncrawled pages", urls)

	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func() {
			err := worker(url, &wg)
			if err != nil {
				log.Println("Error fetching:", url, err)
			}
		}()
	}
	wg.Wait()
	return len(urls)
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
	// 设置启动url
	//startURL := "https://www.ouc.edu.cn/"
	//var wg sync.WaitGroup
	//wg.Add(1)
	//go func() {
	//	err := worker(startURL, &wg)
	//	if err != nil {
	//		log.Println("Error fetching:", startURL, err)
	//	}
	//}()
	//wg.Wait()

	// 开始爬取，定时爬取，每隔一段时间爬取一次
	for {
		urlNum := crawl()
		if urlNum == 0 {
			log.Println("No new pages to crawl, waiting...")
			break
		} else {
			log.Println("Fetched", urlNum, "new pages, waiting...")
		}
		time.Sleep(1 * time.Second)
	}

	database.Close()
}
