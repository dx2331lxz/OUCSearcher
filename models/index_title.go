// Package models Description: page中title的倒排索引表
package models

type TitleIndex struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"default:null;uniqueIndex:idx_unique_name"`
	IndexString string `gorm:"type:text;default:null"` // 使用 TEXT 类型
}
