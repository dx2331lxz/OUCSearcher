package tools

import (
	"golang.org/x/net/html"
	"net/url"
	"regexp"
	"strings"
)

// ReplaceMultipleSpacesAndNewlines 替换多个空格和换行符为单个空格
func replaceMultipleSpacesAndNewlines(input string) string {
	// 创建正则表达式，匹配一个或多个空格和换行符
	re := regexp.MustCompile(`\s+`)
	// 用单个空格替换所有匹配的部分
	return strings.TrimSpace(re.ReplaceAllString(input, " "))
}

// removeBase64Images 去除 Base64 编码的图片内容
func removeBase64Images(s string) string {
	base64Pattern := `data:image\/[a-zA-Z]+;base64,[^\s]+`
	re := regexp.MustCompile(base64Pattern)
	return re.ReplaceAllString(s, "")
}

// ExtractTitle 解析html的title
func ExtractTitle(doc *html.Node) string {
	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		// 检查是否为 <title> 标签
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			// 提取 <title> 标签中的文本
			title = n.FirstChild.Data
			return
		}
		// 继续遍历子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return title
}

// ExtractText 从 HTML 文档中提取所有文本
func ExtractText(doc *html.Node) string {
	var b strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		// 跳过 <script> 标签及其内容
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}

		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteString(" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	s := replaceMultipleSpacesAndNewlines(b.String())
	s = removeBase64Images(s)
	return s
}

// ExtractLinks 从 HTML 文档中提取所有链接
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

type UrlInfo struct {
	Url     string
	Host    string
	Scheme  string
	Domain1 string
	Domain2 string
	Path    string
	Query   string
}

// ParseURL 解析 URL
func ParseURL(rawURL string) *UrlInfo {
	// 创建 UrlInfo 结构体
	info := &UrlInfo{Url: rawURL}
	// 解析 URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return info
	}
	// 提取主机名
	info.Host = u.Hostname()
	// 提取一级域名
	parts := strings.Split(info.Host, ".")
	if len(parts) > 1 {
		info.Domain1 = parts[len(parts)-1]
	}
	// 提取二级域名
	if len(parts) > 2 {
		info.Domain2 = parts[len(parts)-2]
	}
	// 提取路径
	info.Path = u.Path
	// 提取查询参数
	info.Query = u.RawQuery
	// 提取协议
	info.Scheme = u.Scheme
	return info
}
