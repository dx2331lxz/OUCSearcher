package models

import (
	"OUCSearcher/database"
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

// TableName 指定表名
func (Page) TableName() string {
	return "page"
}

// GetByUrl 通过url获取pages
func (p *Page) GetByUrl(url string) ([]Page, error) {
	sqlString := "SELECT * FROM page WHERE url = ?"
	rows, err := database.DB.Query(sqlString, url)
	if err != nil {
		log.Println("Error getting page by url:", url, err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(rows)

	var pages []Page
	for rows.Next() {
		var page Page
		if err := rows.Scan(&page.ID, &page.Url, &page.Host, &page.CrawDone, &page.DicDone, &page.CrawTime, &page.OriginTitle, &page.ReferrerId, &page.Scheme, &page.Domain1, &page.Domain2, &page.Path, &page.Query, &page.Title, &page.Text, &page.CreatedAt); err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

// UpdateOrCreateByUrl 通过url更新或创建
func (p *Page) UpdateOrCreateByUrl() (sql.Result, uint, error) {
	// 查询是否存在
	pages, err := p.GetByUrl(p.Url)
	if err != nil {
		log.Println("Error getting page by url:", p.Url, err)
		return nil, 0, err
	}
	// 获取pages的个数
	count := len(pages)
	if count == 0 {
		// 不存在则创建
		create, err := p.Create()
		if err != nil {
			log.Println("Error creating page:", p.Url, err)
			return nil, 0, err
		}
		return create, uint(p.ID), nil
	} else {
		// 存在则更新
		p.ID = pages[0].ID
		update, err := p.Update()
		if err != nil {
			log.Println("Error updating page:", p.Url, err)
			return nil, 0, err
		}
		return update, p.ID, nil
	}
}

func (p *Page) Create() (sql.Result, error) {
	// 创建
	sqlString := `INSERT INTO page (url, host, craw_done, dic_done, craw_time, origin_title, referrer_id, scheme, domain1, domain2, path, query, title, text, created_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := database.DB.Exec(sqlString, p.Url, p.Host, p.CrawDone, p.DicDone, p.CrawTime, p.OriginTitle, p.ReferrerId, p.Scheme, p.Domain1, p.Domain2, p.Path, p.Query, p.Title, p.Text, time.Now())
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update 以ID为条件更新
func (p *Page) Update() (sql.Result, error) {
	sqlString := `UPDATE page SET url = ?, host = ?, craw_done = ?, dic_done = ?, craw_time = ?, origin_title = ?, referrer_id = ?, scheme = ?, domain1 = ?, domain2 = ?, path = ?, query = ?, title = ?, text = ?, created_at = ? WHERE id = ?`
	result, err := database.DB.Exec(sqlString, p.Url, p.Host, p.CrawDone, p.DicDone, p.CrawTime, p.OriginTitle, p.ReferrerId, p.Scheme, p.Domain1, p.Domain2, p.Path, p.Query, p.Title, p.Text, time.Now(), p.ID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 添加或着跳过
func (p *Page) CreateOrPassByUrl() (sql.Result, error) {
	// 查询是否存在
	pages, err := p.GetByUrl(p.Url)
	if err != nil {
		return nil, err
	}
	// 获取pages的个数
	count := len(pages)
	if count == 0 {
		// 不存在则创建
		return p.Create()
	} else {
		// 存在则跳过
		return nil, nil
	}
}

// GetNUnCrawled 提取N条未爬取的链接
func GetNUnCrawled(n int) ([]string, error) {
	sqlString := `SELECT url FROM page WHERE craw_done = 0 LIMIT ?`
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
