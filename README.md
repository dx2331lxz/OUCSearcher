## 爬取逻辑
1. 从数据库中获取前10个链接
2. 从指定的 URL 获取 HTML，标记为已爬取
3. 解析 HTML，提取链接
4. 将链接保存到数据库
5. 重复步骤 2-4，直到没有新链接

## 创建数据库
在 GORM 中，默认情况下，表名是根据模型结构体的复数形式生成的。
对于 Page 结构体，GORM 会自动将其表名设置为 pages 而不是 page。这是因为 GORM 使用了复数化规则来生成表名。

在结构体上使用 gorm:"tableName:page" 标签来显式指定表名,通过覆写 TableName 方法来指定表名:
``` go
func (Page) TableName() string {
	return "page"
}
```