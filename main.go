package main

import (
	"OUCSearcher/config"
	"OUCSearcher/database"
	"OUCSearcher/models"
	_ "OUCSearcher/routers" // 引入路由
	"OUCSearcher/tools"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

const NumberOfCrawl = 2000

// init 初始化数据库连接
func init() {
	cfg := config.NewConfig()
	database.Initialize(cfg)
	database.InitializeRedis(cfg)
}

// 数据库迁移
func migrate() {
	// 迁移数据库
	// 打开数据库连接
	cfg := config.NewConfig()
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	// 创建 256 张表
	//for i := 0; i < 256; i++ {
	//	tableName := fmt.Sprintf("page_%02x", i)
	//	// 自动迁移
	//	err = db.Table(tableName).AutoMigrate(&models.Page{})
	//	if err != nil {
	//		log.Fatal("failed to migrate database:", err)
	//	} else {
	//		log.Printf("Database %s migrated successfully!\n", tableName)
	//	}
	//}
	//db.AutoMigrate(&models.IndexTableStatus{})

	for i := 0; i < 256; i++ {
		tableName := fmt.Sprintf("index1_%02x", i)
		// 自动迁移
		err = db.Table(tableName).AutoMigrate(&models.Index{})
		if err != nil {
			log.Fatal("failed to migrate database:", err)
		} else {
			log.Printf("Database %s migrated successfully!\n", tableName)
		}
	}
	for i := 0; i < 256; i++ {
		tableName := fmt.Sprintf("index_%02x", i)
		// 自动迁移
		err = db.Table(tableName).AutoMigrate(&models.Index{})
		if err != nil {
			log.Fatal("failed to migrate database:", err)
		} else {
			log.Printf("Database %s migrated successfully!\n", tableName)
		}
	}
}

func main() {
	currentTime := time.Now().Format("2006-01-02") // 格式化为 YYYY-MM-DD
	logFileName := currentTime + ".log"
	// 创建一个日志文件
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// 设置日志输出到文件
	log.SetOutput(file)
	// 设置日志格式，记录文件名和行号
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	serviceLogFile := "log/" + time.Now().Format("2006-01-02") + ".log"
	logs.SetLogger(logs.AdapterFile, `{
        "filename": "`+serviceLogFile+`",
        "daily": false,
        "maxlines": 1000,
        "maxsize": 0,
        "level": 7, 
        "perm": "0660"
    }`)
	logs.Info(serviceLogFile)
	logs.SetLogFuncCallDepth(3)
	logs.SetLogFuncCall(true) // 记录文件名和行号

	// 迁移数据库
	//migrate()

	// 启动redis从mysql获取urls
	tools.CronJobSub.StartTask("GetUrlsFromMysqlJob")
	//models.GetUrlsFromMysqlTimer()

	// 开始爬取，定时爬取，每隔一段时间爬取一次
	tools.CronJobSub.StartTask("Crawl")

	// 启动定时任务，生成倒排索引并且将结果添加到redis中
	tools.CronJobSub.StartTask("GenerateInvertedIndexAndAddToRedis")

	// 启动定时任务，将倒排索引存入mysql
	//tools.SaveInvertedIndexStringToMysqlTimer()
	tools.CronJobSub.StartTask("SaveInvertedIndexStringToMysql")

	////启动定时任务，更新爬取状态
	tools.CronJobSub.StartTask("UpdateCrawDone")

	//// 启动定时任务，更新分词状态
	tools.CronJobSub.StartTask("UpdateDicDoneJob")

	beego.Run()
	database.Close()
	database.CloseRedis()
}
