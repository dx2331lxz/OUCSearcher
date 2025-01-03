// Package tools Description: 提取 HTML 文档中的文本、链接和标题，对文档内容进行处理
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

// 全局编译正则表达式
var whitespaceAndLetterRegexp = regexp.MustCompile(`[a-zA-Z\s\p{Zs}]+`)

// StringStrip 替换输入字符串中的空白字符、零宽空格和英文字符为连字符 "-"
func StringStrip(input string) string {
	if input == "" {
		return ""
	}
	// 替换空白字符、零宽空格和英文字符为连字符
	result := whitespaceAndLetterRegexp.ReplaceAllString(input, "-")
	// 去除前导和尾随的连字符
	result = strings.Trim(result, "-")
	// 将连续的连字符合并为一个
	result = strings.Join(strings.FieldsFunc(result, func(r rune) bool {
		return r == '-'
	}), "-")
	return result
}

// ExtractTitle 解析html的title
func ExtractTitle(doc *html.Node) string {
	// 定义一个递归函数，直接返回 title
	var findTitle func(*html.Node) (string, bool)
	findTitle = func(n *html.Node) (string, bool) {
		// 检查当前节点是否为 <title>
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			return n.FirstChild.Data, true
		}
		// 遍历子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if title, found := findTitle(c); found {
				return title, true
			}
		}
		return "", false
	}

	// 调用递归函数，忽略 bool 值
	title, _ := findTitle(doc)
	if title == "" {
		return "未知标题"
	}
	return title
}

// ExtractText 从 HTML 文档中提取所有文本
//
//	func ExtractText(doc *html.Node) string {
//		var b strings.Builder
//		var f func(*html.Node)
//		f = func(n *html.Node) {
//			// 跳过 <script> 标签及其内容
//			if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
//				return
//			}
//
//			if n.Type == html.TextNode {
//				b.WriteString(n.Data)
//				b.WriteString(" ")
//			}
//			for c := n.FirstChild; c != nil; c = c.NextSibling {
//				f(c)
//			}
//		}
//		f(doc)
//		s := removeBase64Images(b.String())
//		s = StringStrip(s)
//		return s
//	}
func ExtractText(doc *html.Node) string {
	var b strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		// 跳过 <script> 和 <style> 标签及其内容
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}

		// 跳过 <div> 中 class 或 id 包含 "header" 的内容
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if (attr.Key == "class" || attr.Key == "id") && strings.Contains(attr.Val, "header") {
					return
				}
			}
		}

		// 提取文本节点内容
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteString(" ")
		}

		// 遍历子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	// 处理提取的字符串，移除多余内容
	s := removeBase64Images(b.String())
	s = StringStrip(s)
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
