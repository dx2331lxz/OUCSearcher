// Package tools 将index表中的数据迁移到index_%s表中
package tools

import (
	"OUCSearcher/database"
	"OUCSearcher/models"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
)

// Index2IndexsTimer 定时执行Index2Indexs
func Index2IndexsTimer() {

	c := cron.New(cron.WithSeconds())
	// 半天执行一次
	c.AddFunc("0 0 0,12 * * *", func() {
		dicDonePercent, err := models.GetDicDonePercent()
		if err != nil {
			log.Println("Error getting dic done percent:", err)
			return
		}
		if dicDonePercent > 0.7 {
			err := Index2Indexs()
			if err != nil {
				log.Println("Error migrating index to indexs:", err)
				return
			}
		}
	})
	Index2Indexs()

	c.Start()
}

func Index2Indexs() error {
	// 查询所有的索引数据
	sqlString := "SELECT name, index_string FROM `index`"
	rows, err := database.DB.Query(sqlString)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var name, indexString string
		err = rows.Scan(&name, &indexString)
		if err != nil {
			return err
		}
		fmt.Println("Migrating index:", name)
	}
	return nil
}

//func InsertOrUpdateIndex(name string, indexString string) error {
//	// 获取表名
//	tableName, err := models.GetIndexTableName(name)
//	if err != nil {
//		return err
//	}
//
//	// 拼接 SQL 语句
//	sqlSelect := fmt.Sprintf("SELECT index_string FROM %s WHERE name = ?", tableName)
//	sqlUpdate := fmt.Sprintf("UPDATE %s SET index_string = ? WHERE name = ?", tableName)
//	sqlInsert := fmt.Sprintf("INSERT INTO %s (name, index_string) VALUES (?, ?)", tableName)
//
//	// 开始事务
//	tx, err := database.DB.Begin()
//	if err != nil {
//		return err
//	}
//	defer func() {
//		if p := recover(); p != nil {
//			tx.Rollback()
//			panic(p) // 重新抛出 panic
//		} else if err != nil {
//			tx.Rollback() // 遇到错误回滚事务
//		} else {
//			err = tx.Commit() // 没有错误提交事务
//		}
//	}()
//
//	// 查询 name 是否已存在
//	var existingIndexString string
//	err = tx.QueryRow(sqlSelect, name).Scan(&existingIndexString)
//	if err == sql.ErrNoRows {
//		// 如果没有找到记录，则插入新的记录
//		_, err = tx.Exec(sqlInsert, name, indexString)
//		fmt.Println("Inserting index:", name)
//	} else if err != nil {
//		return err
//	} else {
//		// 如果记录已经存在，则更新 index_string
//		_, err = tx.Exec(sqlUpdate, existingIndexString+"-"+indexString, name)
//		fmt.Println("Updating index:", name)
//	}
//	return err
//}
