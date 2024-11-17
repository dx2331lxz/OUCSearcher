package models

import (
	"OUCSearcher/database"
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Page struct {
	ID          uint      `gorm:"primaryKey"`                              // primary key
	Url         string    `gorm:"default:null;uniqueIndex:idx_unique_url"` // 网页链接
	Host        string    `gorm:"default:null"`                            // 域名
	CrawDone    int       `gorm:"type:tinyint(1);default:0"`               // 已爬
	DicDone     int       `gorm:"type:tinyint(1);default:0"`               // 已拆分进词典
	CrawTime    time.Time `gorm:"default:'2001-01-01 00:00:01'"`           // 爬取时刻
	OriginTitle string    `gorm:"default:''"`                              // 上级页面超链接文字
	ReferrerId  uint      `gorm:"default:0"`                               // 上级页面ID
	Scheme      string    `gorm:"default:null"`                            // http/https
	Domain1     string    `gorm:"default:null"`                            // 一级域名后缀
	Domain2     string    `gorm:"default:null"`                            // 二级域名后缀
	Path        string    `gorm:"default:null"`                            // URL 路径
	Query       string    `gorm:"default:null"`                            // URL 查询参数
	Title       string    `gorm:"default:null"`                            // 页面标题
	Text        string    `gorm:"type:LONGTEXT;default:null"`              // 页面文字
	CreatedAt   time.Time // 插入时间
}

type PageDic struct {
	ID   uint
	Url  string
	Text string
}

// PageDynamic 定义动态表名
//type PageDynamic struct {
//	Page        // 嵌套 Page
//	TableSuffix string
//}

//// 动态表名方法
//func (p PageDynamic) TableName() string {
//	fmt.Println("page_" + p.TableSuffix)
//	return fmt.Sprintf("page_%s", p.TableSuffix)
//}

// TableName 指定表名
//func (Page) TableName() string {
//	return "page"
//}

// 获取表名
func GetTableName(url string) (string, error) {
	// 计算 MD5 哈希值
	hash := md5.New()
	_, err := hash.Write([]byte(url))
	if err != nil {
		return "", fmt.Errorf("failed to write data to hash: %v", err)
	}

	// 获取哈希值的最后一位（字节）
	hashValue := hash.Sum(nil)

	// 提取最后一个字节
	lastByte := hashValue[len(hashValue)-1]

	// 将最后一个字节转换为小写十六进制字符
	lastHexChar := fmt.Sprintf("%02x", lastByte)

	// 返回分表名称，格式为 page_<lastHexChar>
	return fmt.Sprintf("page_%s", lastHexChar), nil
}

// GetByUrl 通过url获取pages
//func (p *Page) GetByUrl(url string) ([]Page, error) {
//	sqlString := "SELECT * FROM page WHERE url = ?"
//	rows, err := database.DB.Query(sqlString, url)
//	if err != nil {
//		log.Println("Error getting page by url:", url, err)
//		return nil, err
//	}
//	defer func(rows *sql.Rows) {
//		err := rows.Close()
//		if err != nil {
//			fmt.Println(err)
//		}
//	}(rows)
//
//	var pages []Page
//	for rows.Next() {
//		var page Page
//		if err := rows.Scan(&page.ID, &page.Url, &page.Host, &page.CrawDone, &page.DicDone, &page.CrawTime, &page.OriginTitle, &page.ReferrerId, &page.Scheme, &page.Domain1, &page.Domain2, &page.Path, &page.Query, &page.Title, &page.Text, &page.CreatedAt); err != nil {
//			return nil, err
//		}
//		pages = append(pages, page)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, err
//	}
//
//	return pages, nil
//}

// UpdateOrCreateByUrl 通过url更新或创建
//func (p *Page) UpdateOrCreateByUrl() (sql.Result, uint, error) {
//	// 查询是否存在
//	pages, err := p.GetByUrl(p.Url)
//	if err != nil {
//		log.Println("Error getting page by url:", p.Url, err)
//		return nil, 0, err
//	}
//	// 获取pages的个数
//	count := len(pages)
//	if count == 0 {
//		// 不存在则创建
//		create, err := p.Create()
//		if err != nil {
//			log.Println("Error creating page:", p.Url, err)
//			return nil, 0, err
//		}
//		return create, uint(p.ID), nil
//	} else {
//		// 存在则更新
//		p.ID = pages[0].ID
//		update, err := p.Update()
//		if err != nil {
//			log.Println("Error updating page:", p.Url, err)
//			return nil, 0, err
//		}
//		return update, p.ID, nil
//	}
//}

func (p *Page) Create() (sql.Result, error) {
	// 创建
	// 获取表名
	tableName, err := GetTableName(p.Url)
	sqlString := fmt.Sprintf("INSERT INTO %s (url, host, craw_done, dic_done, craw_time, origin_title, referrer_id, scheme, domain1, domain2, path, query, title, text, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", tableName)
	result, err := database.DB.Exec(sqlString, p.Url, p.Host, p.CrawDone, p.DicDone, p.CrawTime, p.OriginTitle, p.ReferrerId, p.Scheme, p.Domain1, p.Domain2, p.Path, p.Query, p.Title, p.Text, time.Now())
	if err != nil {
		log.Println("Error creating page:", p.Url, err)
		return nil, err
	}
	return result, nil
}

// Update 以ID为条件更新
func (p *Page) Update() (sql.Result, error) {
	// 获取表名
	tableName, err := GetTableName(p.Url)
	sqlString := fmt.Sprintf("UPDATE %s SET host = ?, craw_done = ?, dic_done = ?, craw_time = ?, origin_title = ?, referrer_id = ?, scheme = ?, domain1 = ?, domain2 = ?, path = ?, query = ?, title = ?, text = ?, created_at = ? WHERE url = ?", tableName)
	//fmt.Println(sqlString, p.Url, p.Host, p.CrawDone, p.DicDone, p.CrawTime, p.OriginTitle, p.ReferrerId, p.Scheme, p.Domain1, p.Domain2, p.Path, p.Query, p.Title, p.Text, time.Now(), p.ID)
	result, err := database.DB.Exec(sqlString, p.Host, p.CrawDone, p.DicDone, p.CrawTime, p.OriginTitle, p.ReferrerId, p.Scheme, p.Domain1, p.Domain2, p.Path, p.Query, p.Title, p.Text, time.Now(), p.Url)
	if err != nil {
		log.Println("Error updating page:", p.Url, err)
		return nil, err
	}
	// 获取影响的行数
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error getting rows affected:", err)
		return nil, err
	}

	if rowsAffected == 0 {
		log.Println("No rows updated for URL:", p.Url)
	} else {
		//fmt.Println(sqlString, p.Url)
		log.Printf("Successfully updated page: %s", p.Url)
	}

	return result, nil
}

// 添加或着跳过
//func (p *Page) CreateOrPassByUrl() (sql.Result, error) {
//	// 查询是否存在
//	pages, err := p.GetByUrl(p.Url)
//	if err != nil {
//		return nil, err
//	}
//	// 获取pages的个数
//	count := len(pages)
//	if count == 0 {
//		// 不存在则创建
//		return p.Create()
//	} else {
//		// 存在则跳过
//		return nil, nil
//	}
//}

// GetNUnCrawled 提取N条未爬取的链接
func GetNUnCrawled(TableSuffix string, n int) ([]string, error) {
	sqlString := fmt.Sprintf("SELECT url FROM page_%s WHERE craw_done = 0 LIMIT ?", TableSuffix)
	rows, err := database.DB.Query(sqlString, n)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(rows)

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

// GetNUnDicDone 提取N条已经爬取但是没有分词的数据
func GetNUnDicDone(TableSuffix string, n int) ([]PageDic, error) {
	sqlString := fmt.Sprintf("SELECT id, url, text FROM page_%s WHERE craw_done = 1 AND dic_done = 0 LIMIT ?", TableSuffix)
	rows, err := database.DB.Query(sqlString, n)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(rows)
	var pageDics []PageDic
	for rows.Next() {
		var pageDic PageDic
		if err := rows.Scan(&pageDic.ID, &pageDic.Url, &pageDic.Text); err != nil {
			return nil, err
		}
		pageDics = append(pageDics, pageDic)
	}
	return pageDics, nil
}
