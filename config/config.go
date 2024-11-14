// 配置mysql和redis的连接信息
package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	// MySQL 数据库连接信息
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	// Redis 数据库连接信息
	RedisHost     string
	RedisPort     string
	RedisPassword string
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return &Config{
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		RedisHost:  os.Getenv("REDIS_HOST"),
		RedisPort:  os.Getenv("REDIS_PORT"),
	}
}

func (c *Config) DSN() string {
	fmt.Println(c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
