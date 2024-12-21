package tools

import (
	"OUCSearcher/models"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const NumberOfCrawl = 10000

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
	urlInfo := ParseURL(url)
	page.Host = urlInfo.Host
	page.Scheme = urlInfo.Scheme
	page.Domain1 = urlInfo.Domain1
	page.Domain2 = urlInfo.Domain2
	page.Path = urlInfo.Path
	page.Query = urlInfo.Query
	//fmt.Println(page)
	// 解析标题
	//fmt.Println(tools.ExtractTitle(doc))
	page.Title = ExtractTitle(doc)
	// 解析文本
	//fmt.Println(tools.ExtractText(doc))
	page.Text = ExtractText(doc)
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
	links := ExtractLinks(doc, url)
	links = FilterUrl(links)
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
