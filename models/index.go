package models

import (
	"OUCSearcher/database"
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
)

type Index struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"default:null;uniqueIndex:idx_unique_name"`
	IndexString string `gorm:"type:text;default:null"` // 使用 TEXT 类型
}

func (Index) TableName() string {
	return "index"
}

func GetIndexTableName(name string, num int) (string, error) {
	// 计算 MD5 哈希值
	hash := md5.New()
	_, err := hash.Write([]byte(name))
	if err != nil {
		return "", fmt.Errorf("failed to write data to hash: %v", err)
	}

	// 获取哈希值的最后一位（字节）
	hashValue := hash.Sum(nil)

	// 提取最后一个字节
	lastByte := hashValue[len(hashValue)-1]

	// 将最后一个字节转换为小写十六进制字符
	lastHexChar := fmt.Sprintf("%02x", lastByte)

	// 返回分表名称，格式为 index_<lastHexChar>
	currentIndexTable, err := GetCurrentIndexTable(num)
	if err != nil {
		return "", fmt.Errorf("failed to get current index table: %v", err)
	}

	return fmt.Sprintf("%s_%s", currentIndexTable, lastHexChar), nil
}

// SaveMapToDB 使用事务批量存储数据
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

	// 批量插入或更新
	for name, indexStr := range data {
		// 获取表名
		tableName, err := GetIndexTableName(name, 1)
		if err != nil {
			log.Println("failed to get table name:", err)
		}
		sqlSelect := "SELECT index_string FROM " + tableName + " WHERE name = ?"
		sqlUpdate := "UPDATE " + tableName + " SET index_string = ? WHERE name = ?"
		sqlInsert := "INSERT INTO " + tableName + " (name, index_string) VALUES (?, ?)"

		// 查询 name 是否已存在
		var existingIndexString string
		err = tx.QueryRow(sqlSelect, name).Scan(&existingIndexString)

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
			var updatedIndexString string

			if existingIndexString == "" {
				existingIndexString = indexStr
			} else {
				updatedIndexString = existingIndexString + "-" + indexStr
			}

			_, err = tx.Exec(sqlUpdate, updatedIndexString, name)
			if err != nil {
				return fmt.Errorf("failed to update record: %v", err)
			}
		}
	}
	log.Println("Successfully saved inverted index to MySQL!")
	return nil
}

// GetIndexString 目前确保在获取index数据的时候使用表2的数据
// GetIndexString 通过 name 获取 index_string
func GetIndexString(name string) (string, error) {
	// 获取表名
	tableName, err := GetIndexTableName(name, 2)
	if err != nil {
		return "", fmt.Errorf("failed to get table name: %v", err)
	}

	sqlString := fmt.Sprintf("SELECT index_string FROM %s WHERE name = ?", tableName)
	var indexString string
	err = database.DB.QueryRow(sqlString, name).Scan(&indexString)
	if err != nil {
		return "", fmt.Errorf("failed to query index_string: %v", err)
	}
	return indexString, nil
}

// ClearIndexString 将当前表中的所有IndexString至为空
func ClearIndexString() error {
	//	 循环256个表
	tableName, err := GetCurrentIndexTable(1)
	for i := 0; i < 256; i++ {
		tableName_ := fmt.Sprintf("%s_%02x", tableName, i)
		if err != nil {
			return fmt.Errorf("failed to get current index table: %v", err)
		}
		sqlString := fmt.Sprintf("UPDATE %s SET index_string = ''", tableName_)
		_, err = database.DB.Exec(sqlString)
		if err != nil {
			return fmt.Errorf("failed to clear index_string: %v", err)
		}
	}
	return nil
}
