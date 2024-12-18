// Package models 用于切换index数据库，使用数据库表存储当前使用的index
// index_table_status表只允许量条记录
package models

import (
	"OUCSearcher/database"
	"log"
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

func GetCurrentIndexTable(id int) (string, error) {
	var status string
	sqlString := "SELECT current_table FROM index_table_status WHERE id = ?"
	err := database.DB.QueryRow(sqlString, id).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

// SwitchIndexTable 将表中id为1和2的current_table字段值互换
func SwitchIndexTable() error {
	sqlString := `
		UPDATE index_table_status AS it1
		JOIN index_table_status AS it2
		ON it1.id = 1 AND it2.id = 2
		SET it1.current_table = it2.current_table,
			it2.current_table = it1.current_table
	`
	_, err := database.DB.Exec(sqlString)
	if err != nil {
		log.Println("Error switching index table:", err)
		return err
	}

	log.Println("Switched index table successfully!")
	return nil
}
