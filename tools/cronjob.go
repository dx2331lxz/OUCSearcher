package tools

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"sync"
)

// 创建定时器

var CronJobSub *CronJob

func init() {
	CronJobSub = NewCronJob()
}

type JobMutex struct {
	mu             sync.Mutex
	taskInProgress bool
}

type CronJob struct {
	c         *cron.Cron              // cron 调度器
	jobIDs    map[string]cron.EntryID // 存储每个定时任务的 jobID
	mu        sync.Mutex              // 锁，确保线程安全
	isRunning bool                    // 是否正在运行
	jobMutex  map[string]*JobMutex
}

// GetTaskMap 定义一个单独的函数，避免直接使用 TaskMap
func GetTaskMap() map[string]func() error {
	return map[string]func() error{
		"GenerateInvertedIndexAndAddToRedis": GenerateInvertedIndexAndAddToRedis, // 生成倒排索引并存储到 Redis
		"SaveInvertedIndexStringToMysql":     SaveInvertedIndexStringToMysql,     // 将倒排索引字符串存储到 MySQL
		"UpdateCrawDone":                     UpdateCrawDone,                     // 更新爬取状态
		"Crawl":                              Crawl,                              // 爬取网页
		"GetUrlsFromMysqlJob":                GetUrlsFromMysqlJob,                // 从 MySQL 中获取 URL
		"UpdateDicDoneJob":                   UpdateDicDoneJob,                   // 更新分词状态
	}
}

var TaskCronExprMap = map[string]string{
	"GenerateInvertedIndexAndAddToRedis": "*/60 * * * * *",
	"SaveInvertedIndexStringToMysql":     "*/10 * * * * *",
	"UpdateCrawDone":                     "0 0 0 */7 * *",
	"Crawl":                              "*/5 * * * * *",
	"GetUrlsFromMysqlJob":                "*/240 * * * * *",
	"UpdateDicDoneJob":                   "0 0 0 */7 * *",
}

// NewJobMutex 构造函数，设置默认值
func NewJobMutex() *JobMutex {
	return &JobMutex{
		mu:             sync.Mutex{}, // 显式设置默认值
		taskInProgress: false,        // 显式设置默认值
	}
}

// NewCronJob 创建并返回一个新的 CronJob 实例
func NewCronJob() *CronJob {
	// 循环遍历 TaskMap，创建 JobMutex 实例
	TaskMap := GetTaskMap()
	jobMutex := make(map[string]*JobMutex)
	for taskName := range TaskMap {
		jobMutex[taskName] = NewJobMutex()
	}

	return &CronJob{
		c:        cron.New(cron.WithSeconds()),
		jobIDs:   make(map[string]cron.EntryID),
		jobMutex: jobMutex,
	}
}

// Start 启动所有定时任务
func (job *CronJob) Start() {
	TaskMap := GetTaskMap()
	job.mu.Lock()
	defer job.mu.Unlock()

	// 添加定时任务
	for taskName, taskFunc := range TaskMap {
		jobID, err := job.c.AddFunc(TaskCronExprMap[taskName], func() {
			if job.jobMutex[taskName].taskInProgress {
				log.Printf("Task %s is already running, skipping...\n", taskName)
				return
			}
			job.jobMutex[taskName].mu.Lock()
			job.jobMutex[taskName].taskInProgress = true
			defer func() {
				job.jobMutex[taskName].taskInProgress = false
				job.jobMutex[taskName].mu.Unlock()
			}()
			err := taskFunc()
			if err != nil {
				log.Printf("Error running task %s: %v\n", taskName, err)
			}
		})
		if err != nil {
			log.Printf("Error adding task %s: %v\n", taskName, err)
			continue
		}
		job.jobIDs[taskName] = jobID
	}

	// 启动调度器
	job.c.Start()
	job.isRunning = true

	log.Println("Cron job started.")
}

// Stop 停止所有定时任务
func (job *CronJob) Stop() {
	job.mu.Lock()
	defer job.mu.Unlock()

	if !job.isRunning {
		log.Println("Cron job is not running.")
		return
	}

	// 停止调度器
	job.c.Stop()
	job.isRunning = false

	log.Println("Cron job stopped.")
}

// StopTask 停止指定的定时任务
func (job *CronJob) StopTask(taskName string) {
	job.mu.Lock()
	defer job.mu.Unlock()

	if !job.isRunning {
		log.Println("Cron job is not running.")
		return
	}

	if jobID, exists := job.jobIDs[taskName]; exists {
		// 停止并移除指定的任务
		job.c.Remove(jobID)
		delete(job.jobIDs, taskName)
		log.Printf("%s stopped.\n", taskName)
	}

	// 如果没有任务了，可以停止整个调度器
	if len(job.jobIDs) == 0 {
		job.c.Stop()
		job.isRunning = false
		log.Println("All tasks stopped.")
	}
}

// StartTask 启动某个任务
func (job *CronJob) StartTask(taskName string) {
	TaskMap := GetTaskMap()
	job.mu.Lock()
	defer job.mu.Unlock()

	if taskFunc, exists := TaskMap[taskName]; exists {
		jobID, err := job.c.AddFunc(TaskCronExprMap[taskName], func() {
			if job.jobMutex[taskName].taskInProgress {
				fmt.Printf("Task %s is already running, skipping...\n", taskName)
				//log.Printf("Task %s is already running, skipping...\n", taskName)
				return
			}
			job.jobMutex[taskName].mu.Lock()
			job.jobMutex[taskName].taskInProgress = true
			defer func() {
				job.jobMutex[taskName].taskInProgress = false
				job.jobMutex[taskName].mu.Unlock()
			}()
			err := taskFunc()
			if err != nil {
				log.Printf("Error running task %s: %v\n", taskName, err)
			}
		})
		if err != nil {
			log.Printf("Error adding task %s: %v\n", taskName, err)
			return
		}
		job.jobIDs[taskName] = jobID
		job.c.Start()
		job.isRunning = true
		log.Printf("%s started.\n", taskName)
	} else {
		log.Printf("Task %s not found.\n", taskName)
	}

}
