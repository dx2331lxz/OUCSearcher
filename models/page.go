package models

import (
	"OUCSearcher/database"
	"OUCSearcher/types"
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type Page struct {
	ID           uint      `gorm:"primaryKey"`                              // primary key
	Url          string    `gorm:"default:null;uniqueIndex:idx_unique_url"` // 网页链接
	Host         string    `gorm:"default:null"`                            // 域名
	CrawDone     int       `gorm:"type:tinyint(1);default:0"`               // 已爬
	DicDone      int       `gorm:"type:tinyint(1);default:0"`               // 已拆分进词典
	CrawTime     time.Time `gorm:"default:'2001-01-01 00:00:01'"`           // 爬取时刻
	OriginTitle  string    `gorm:"default:''"`                              // 上级页面超链接文字
	ReferrerId   uint      `gorm:"default:0"`                               // 上级页面ID
	ReferrerPage string    `gorm:"default:'01,1'"`                          // 上级页面URL
	Scheme       string    `gorm:"default:null"`                            // http/https
	Domain1      string    `gorm:"default:null"`                            // 一级域名后缀
	Domain2      string    `gorm:"default:null"`                            // 二级域名后缀
	Path         string    `gorm:"default:null"`                            // URL 路径
	Query        string    `gorm:"default:null"`                            // URL 查询参数
	Title        string    `gorm:"default:null"`                            // 页面标题
	Text         string    `gorm:"type:LONGTEXT;default:null"`              // 页面文字
	CreatedAt    time.Time // 插入时间
}

type PageDic struct {
	ID   int    `json:"id"`
	Url  string `json:"url"`
	Text string `json:"text"`
}

type SearchResult struct {
	Title       string // 搜索结果标题
	URL         string // 搜索结果链接
	Description string // 搜索结果描述
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
//	// 获取表名
//	tableName, err := GetTableName(url)
//	sqlString := fmt.Sprintf("SELECT id FROM %s WHERE url = ?", tableName)
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

// UpdateDicDone 更新已经分词
func UpdateDicDone(TableSuffix string, id int) (sql.Result, error) {
	sqlString := fmt.Sprintf("UPDATE page_%s SET dic_done = 1 WHERE id = ?", TableSuffix)
	result, err := database.DB.Exec(sqlString, id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetDicDoneAboutCount 获取已经分词的页面总数（估算）
func GetDicDoneAboutCount() (int, error) {
	sqlString := "SELECT COUNT(*) FROM page_0f WHERE dic_done = 1"
	rows, err := database.DB.Query(sqlString)
	if err != nil {
		return 0, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(rows)
	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}
	return count * 256, nil
}

// GetPageDicFromPair 通过事务循环[]pair，从数据库中提取信息存储进[]PageDic中
func GetSearchResultFromPair(data []types.Pair) ([]SearchResult, error) {
	var searchResultList []SearchResult

	// 开启事务
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	// 确保事务在退出时正确提交或回滚
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback() // 遇到 panic 回滚
			panic(p)      // 继续传播 panic
		} else if err != nil {
			tx.Rollback() // 遇到错误回滚
		} else {
			err = tx.Commit() // 成功则提交事务
		}
	}()

	// 循环处理
	for _, pair := range data {
		var searchRes SearchResult

		// 分割 key 获取表后缀和 ID
		pageInfo := strings.Split(pair.Key, ",")
		if len(pageInfo) != 2 {
			return nil, fmt.Errorf("invalid key format: %s", pair.Key)
		}

		tableSuffix := pageInfo[0]
		id, convErr := strconv.Atoi(pageInfo[1])
		if convErr != nil {
			return nil, fmt.Errorf("failed to convert string to int: %v", convErr)
		}

		// 准备查询语句
		sqlSelect := fmt.Sprintf("SELECT url, title, text FROM page_%s WHERE id = ?", tableSuffix)
		stmt, prepErr := tx.Prepare(sqlSelect)
		if prepErr != nil {
			return nil, fmt.Errorf("failed to prepare statement: %v", prepErr)
		}
		defer stmt.Close()

		// 执行查询
		scanErr := stmt.QueryRow(id).Scan(&searchRes.URL, &searchRes.Title, &searchRes.Description)
		if scanErr != nil {
			if scanErr == sql.ErrNoRows {
				// 如果没有找到记录，可以选择跳过或记录日志
				log.Printf("No record found for table: %s, id: %d", tableSuffix, id)
				continue
			}
			return nil, fmt.Errorf("failed to query for record: %v", scanErr)
		}

		// 添加到结果列表
		searchResultList = append(searchResultList, searchRes)
	}

	return searchResultList, nil
}

// SetCrawDoneToZero 将表中爬取时间与当前时间相差超过一天的数据的爬取状态设置为0
func SetCrawDoneToZero() error {
	// 获取表名
	for i := 0; i < 256; i++ {
		tableName := fmt.Sprintf("page_%02x", i)
		sqlString := fmt.Sprintf("UPDATE %s SET craw_done = 0 WHERE DATEDIFF(NOW(), craw_time) > 1", tableName)
		_, err := database.DB.Exec(sqlString)
		if err != nil {
			log.Println("Error setting craw_done to 0:", err)
			return err
		}
	}
	return nil
}

// SetDicDoneToZero 将表中所有数据的分词状态设置为0
func SetDicDoneToZero() error {
	// 获取表名
	for i := 0; i < 256; i++ {
		tableName := fmt.Sprintf("page_%02x", i)
		sqlString := fmt.Sprintf("UPDATE %s SET dic_done = 0", tableName)
		_, err := database.DB.Exec(sqlString)
		if err != nil {
			log.Println("Error setting dic_done to 0:", err)
			return err
		}
	}
	return nil
}

// GetDicDonePercent 已经分词的页面总数占已经爬取的页面总数的百分比
func GetDicDonePercent() (float64, error) {
	dicDoneCountSum := 0
	countSum := 0
	for i := 0; i < 256; i++ {
		tableName := fmt.Sprintf("page_%02x", i)
		sqlString := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE dic_done = 1", tableName)
		rows, err := database.DB.Query(sqlString)
		if err != nil {
			return 0, err
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(rows)
		var dicDoneCount int
		for rows.Next() {
			if err := rows.Scan(&dicDoneCount); err != nil {
				return 0, err
			}
		}
		dicDoneCountSum += dicDoneCount
		sqlString = fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		rows, err = database.DB.Query(sqlString)
		if err != nil {
			return 0, err
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(rows)
		var count int
		for rows.Next() {
			if err := rows.Scan(&count); err != nil {
				return 0, err
			}
		}
		countSum += count
	}
	return float64(dicDoneCountSum) / float64(countSum), nil
}
