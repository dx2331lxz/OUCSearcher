package tools

import (
	"github.com/robfig/cron/v3"
	"log"
	"sync"
)

type CronJob struct {
	c         *cron.Cron              // cron 调度器
	jobIDs    map[string]cron.EntryID // 存储每个定时任务的 jobID
	mu        sync.Mutex              // 锁，确保线程安全
	isRunning bool                    // 是否正在运行
	taskMap   map[string]func() error // 任务名与任务函数的映射
}

var TaskMap = map[string]func() error{
	"GenerateInvertedIndexAndAddToRedis": GenerateInvertedIndexAndAddToRedis,
	"SaveInvertedIndexStringToMysql":     SaveInvertedIndexStringToMysql,
}

var TaskCronExprMap = map[string]string{
	"GenerateInvertedIndexAndAddToRedis": "*/120 * * * * *",
	"SaveInvertedIndexStringToMysql":     "*/10 * * * * *",
}

// NewCronJob 创建并返回一个新的 CronJob 实例
func NewCronJob() *CronJob {
	return &CronJob{
		c:      cron.New(cron.WithSeconds()),
		jobIDs: make(map[string]cron.EntryID),
	}
}

// Start 启动所有定时任务
func (job *CronJob) Start() {
	job.mu.Lock()
	defer job.mu.Unlock()

	// 添加定时任务
	for taskName, taskFunc := range TaskMap {
		jobID, err := job.c.AddFunc(TaskCronExprMap[taskName], func() {
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
	job.mu.Lock()
	defer job.mu.Unlock()

	if taskFunc, exists := TaskMap[taskName]; exists {
		jobID, err := job.c.AddFunc(TaskCronExprMap[taskName], func() {
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
