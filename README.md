## 爬取逻辑
1.	从mysql获取urls：
	将 MySQL 中存储的待爬取 URL 拉取到 Redis的名称为urls的列表中： Redis 的 LPUSH 或 RPUSH 将 URL 插入 Redis 列表，供爬虫消费。
    同时将url加入到all_urls这个set中，防止爬取时重复添加url到mysql中
2.	爬取： GetUrlFromRedis
3. 


## 创建数据库
在 GORM 中，默认情况下，表名是根据模型结构体的复数形式生成的。
对于 Page 结构体，GORM 会自动将其表名设置为 pageDics 而不是 page。这是因为 GORM 使用了复数化规则来生成表名。

在结构体上使用 gorm:"tableName:page" 标签来显式指定表名,通过覆写 TableName 方法来指定表名:
``` go
func (Page) TableName() string {
	return "page"
}
```

## 分表
使用main.go中的migrate函数将数据库分为256个表，通过在main函数中调用migrate函数来创建表

## 去重
使用redis创建了两个set，一个叫做visited_urls，另一个叫all_urls，visited_urls用来存储已经访问过的url，all_urls用来存储所有的url
visited_urls用来防止之前添加到待爬取list中的url没有被爬取，又被添加到待爬取list中，从而进行重复爬取
all_urls用来防止重复添加和数据库中url重复的url到待爬取list中，确保数据库中存储的url唯一

## 更新分词库

## 交叉编译linux运行
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o oucsearch

## todo
1. 频率控制--使用redis
2. 页面上添加面试题
3. 爬虫超时重试
4. 自动切换内网搜索
5. pagerank算法
6. seo优化
7. 循环爬取
