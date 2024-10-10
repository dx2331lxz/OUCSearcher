package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

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

// ExtractLinks extracts and returns all the links from a webpage
func ExtractLinks(doc *html.Node, baseURL string) []string {
	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link := attr.Val
					// Handle relative URLs
					if strings.HasPrefix(link, "/") {
						link = baseURL + link
					}
					links = append(links, link)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return links
}

func worker(url string, wg *sync.WaitGroup, results chan<- []string) {
	defer wg.Done()
	fmt.Println("Fetching:", url)

	doc, err := Fetch(url)
	if err != nil {
		log.Println("Error fetching:", url, err)
		return
	}

	links := ExtractLinks(doc, url)
	results <- links
}

func main() {
	startURL := "https://www.ouc.edu.cn/"
	//doc, err := Fetch(url)
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	return
	//}
	//fmt.Println("Successfully fetched and parsed the document:", doc)
	//
	//// 渲染并打印 HTML 内容
	//htmlContent := RenderHTML(doc)
	//fmt.Println(htmlContent)

	var wg sync.WaitGroup
	results := make(chan []string)

	wg.Add(1)
	go worker(startURL, &wg, results)

	go func() {
		wg.Wait()
		close(results)
	}()

	for links := range results {
		fmt.Println("Found links:", links)
	}
}

//
