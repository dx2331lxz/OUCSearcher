// Package models 用于切换index数据库，使用数据库表存储当前使用的index
// index_table_status表只允许一条记录，id为1，current_table为当前使用的表名
package models

import (
	"OUCSearcher/database"
)

type IndexTableStatus struct {
	ID           int    `gorm:"primaryKey"`                                 // 主键且值约束为1
	CurrentTable string `gorm:"type:varchar(255);not null;default:'index'"` // 当前表名，不能为空，默认值为 'index'
}

func (IndexTableStatus) TableName() string {
	return "index_table_status"
}

// IndexList 常量list存储index类型
var IndexList = []string{"index", "index1"}

func GetCurrentIndexTable() (string, error) {
	var status string
	sqlString := "SELECT current_table FROM index_table_status WHERE id = 1"
	err := database.DB.QueryRow(sqlString).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

func SwitchIndexTable() error {
	currentTable, err := GetCurrentIndexTable()
	if err != nil {
		return err
	}
	nextTable := IndexList[0]
	if currentTable == IndexList[0] {
		nextTable = IndexList[1]
	}
	sqlString := "UPDATE index_table_status SET current_table = ? WHERE id = 1"
	_, err = database.DB.Exec(sqlString, nextTable)
	if err != nil {
		return err
	}
	return nil
}
