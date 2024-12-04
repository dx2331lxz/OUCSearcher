# OUCSearcher 海底捞
## 项目简介
本项目为针对高校中国海洋大学的自建搜索引擎，README.md文档待完善
## 爬取逻辑
### 爬取
1.	从mysql获取urls：
	将 MySQL 中存储的待爬取 URL 拉取到 Redis的名称为urls的列表中： Redis 的 LPUSH 或 RPUSH 将 URL 插入 Redis 列表，供爬虫消费。
    同时将url加入到all_urls这个set中，防止爬取时重复添加url到mysql中
2.	爬取：
	GetUrlFromRedis：从redis中取出url
	IsUrlVisited：判断是否已经爬取过如果没有则进行爬取
	解析url以及对应爬取的页面，存入数据库，对于页面中出现的url，如果不在redis的名为all_urls的set表中则添加到mysql中，并且添加到set表，如果在set表中则跳过
### 倒排索引
1. 	生成倒排索引，将倒排索引添加到redis中
2. 	定期从redis中随机取出一部分使用事务追加到mysql的对应词的位置
### SEO
使用BM25算法对搜索词和文章进行相关性排序


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
## 分词
使用一张大表index存储分词定时迁移到分表index_00到index_ff中

为什么不直接存储到分表中呢？
因为每个分词的值都是由不同的页面共同影响，如果一个页面中的词发生改变，例如删除了一个词，那么这个结果是没有办法反映到分表中的，所以需要一个大表来存储所有的分词，然后定时迁移到分表中
## 刷新分词状态
每24小时检测一次redis中的分词个数，如果分词个数小于1000则将分词库中的分词删除，并且将数据库中爬取状态为1的url的分词状态置0
对于为什么是小于1000，而不是在等于零的时候进行分词状态的刷新呢，因为爬取的页面时时刻刻在增加，会造成分词的增加和减少之间的一种平衡，所以设置一个阈值，当分词个数小于1000时，进行刷新，确保更新的及时性

## 迁移分词
使用定时器进行检测，如果已经分词的页面总数占已经爬取的页面总数的百分比大于0.7，我认为已经分词的页面已经足够多了，所以将已经分词的页面迁移到分表的mysql中，将对应的词直接替换，确保分词不会因为更新而重复

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
