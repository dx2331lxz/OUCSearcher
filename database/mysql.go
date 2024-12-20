package database

import (
	"OUCSearcher/config"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var DB *sql.DB

// Initialize connects to the MySQL database
func Initialize(cfg *config.Config) {
	var err error
	dsn := cfg.DSN()
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	// 设置连接池大小
	DB.SetMaxOpenConns(110) // 最大打开连接数
	DB.SetMaxIdleConns(10)  // 最大空闲连接数
	// 测试数据库连接
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Could not ping the database: %v", err)
	}

	fmt.Println("Connected to MySQL database!")
}

// Close closes the database connection
func Close() {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			return
		}
	}
}
