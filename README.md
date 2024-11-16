## 爬取逻辑
1.	MySQL -> Redis：
•	将 MySQL 中存储的待爬取 URL 拉取到 Redis 中（通过定时任务或实时同步）。
•	通过 Redis 的 LPUSH 或 RPUSH 将 URL 插入 Redis 列表，供爬虫消费。
2.	Redis -> 爬虫：
•	爬虫从 Redis 列表中通过 BRPOP 或 RPOP 取出 URL，进行并发爬取。
•	爬虫可以通过多进程、多线程并发消费 Redis 列表。
3.	缓存结果到 Redis：
•	爬取结果可以存储到 Redis 中（例如使用 SET 存储 HTML 内容），并设置过期时间，避免占用过多内存。
4.	爬虫速率限制：
•	通过 Redis 限制每秒请求的 URL 数量，可以使用 INCRBY 配合过期时间来实现速率控制。

[//]: # (1. 从数据库中获取前10个链接)

[//]: # (2. 从指定的 URL 获取 HTML，标记为已爬取)

[//]: # (3. 解析 HTML，提取链接)

[//]: # (4. 将链接保存到数据库)

[//]: # (5. 重复步骤 2-4，直到没有新链接)

## 创建数据库
在 GORM 中，默认情况下，表名是根据模型结构体的复数形式生成的。
对于 Page 结构体，GORM 会自动将其表名设置为 pages 而不是 page。这是因为 GORM 使用了复数化规则来生成表名。

在结构体上使用 gorm:"tableName:page" 标签来显式指定表名,通过覆写 TableName 方法来指定表名:
``` go
func (Page) TableName() string {
	return "page"
}
```

## 分表
使用main.go中的migrate函数将数据库分为256个表，通过在main函数中调用migrate函数来创建表

## todo
1. 频率控制--使用redis
2. 页面上添加面试题