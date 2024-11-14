package tools

import "strings"

// 定义允许存储的url
var allowedHosts = []string{
	"ouc.edu.cn",
}

// FilterUrl 判断url是否允许存储
func FilterUrl(urls []string) []string {
	var resUrls []string

	for _, url := range urls {
		if url == "" {
			continue
		}
		// 判断是否包含http或者https
		if !strings.Contains(url, "http") {
			continue
		}

		// 判断是否包含允许存储的url
		for _, host := range allowedHosts {
			if strings.Contains(url, host) {
				resUrls = append(resUrls, url)
			}
		}
	}
	return resUrls
}
