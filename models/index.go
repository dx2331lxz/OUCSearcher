package models

import (
	"OUCSearcher/database"
	"database/sql"
	"fmt"
)

type Index struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"default:null;uniqueIndex:idx_unique_name"`
	IndexString string `gorm:"type:text;default:null"` // 使用 TEXT 类型
}

func (Index) TableName() string {
	return "index"
}

// 使用事务批量存储数据
// SaveMapToDB 使用事务批量将 map[string]string 的数据存储到 MySQL
func SaveMapToDB(data map[string]string) error {
	// 开启事务
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// 使用defer确保在函数退出时提交事务
	defer func() {
		if err != nil {
			tx.Rollback() // 如果发生错误，回滚事务
		} else {
			err = tx.Commit() // 如果没有错误，提交事务
		}
	}()

	// 准备插入或更新的 SQL 语句
	sqlSelect := "SELECT index_string FROM `index` WHERE name = ?"
	sqlUpdate := "UPDATE `index` SET index_string = ? WHERE name = ?"
	sqlInsert := "INSERT INTO `index` (name, index_string) VALUES (?, ?)"

	// 批量插入或更新
	for name, indexStr := range data {
		// 查询 name 是否已存在
		var existingIndexString string
		err := tx.QueryRow(sqlSelect, name).Scan(&existingIndexString)

		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to query for existing record: %v", err)
		}

		if err == sql.ErrNoRows {
			// 如果没有找到记录，则插入新的记录
			_, err = tx.Exec(sqlInsert, name, indexStr)
			if err != nil {
				return fmt.Errorf("failed to insert record: %v", err)
			}
		} else {
			// 如果记录已经存在，则更新 index_string，拼接新的值
			updatedIndexString := existingIndexString + "-" + indexStr
			_, err = tx.Exec(sqlUpdate, updatedIndexString, name)
			if err != nil {
				return fmt.Errorf("failed to update record: %v", err)
			}
		}
	}

	return nil
}

// GetIndexString 通过 name 获取 index_string
func GetIndexString(name string) (string, error) {
	sqlString := "SELECT index_string FROM `index` WHERE name = ?"
	var indexString string
	err := database.DB.QueryRow(sqlString, name).Scan(&indexString)
	if err != nil {
		return "", fmt.Errorf("failed to query index_string: %v", err)
	}
	return indexString, nil
}
