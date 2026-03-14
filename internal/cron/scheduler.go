package cron

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task 定时任务定义
type Task struct {
	Name    string
	Ticker  time.Duration
	Handler func()
}

// Scheduler 定时任务调度器
type Scheduler struct {
	tasks []*Task
	wg    sync.WaitGroup
	stop  chan struct{}
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
	return &Scheduler{
		stop: make(chan struct{}),
	}
}

// Register 注册定时任务
func (s *Scheduler) Register(task *Task) {
	s.tasks = append(s.tasks, task)
}

// Start 启动所有定时任务（后台运行）
func (s *Scheduler) Start() {
	for _, task := range s.tasks {
		s.wg.Add(1)
		go s.runTask(task)
	}
	fmt.Printf("[Scheduler] started %d tasks\n", len(s.tasks))
}

// Stop 停止所有定时任务
func (s *Scheduler) Stop() {
	close(s.stop)
	s.wg.Wait()
	fmt.Println("[Scheduler] all tasks stopped")
}

func (s *Scheduler) runTask(task *Task) {
	defer s.wg.Done()

	ticker := time.NewTicker(task.Ticker)
	defer ticker.Stop()

	var running int32

	fmt.Printf("[Scheduler] task [%s] started, interval: %s\n", task.Name, task.Ticker)

	for {
		select {
		case <-s.stop:
			fmt.Printf("[Scheduler] task [%s] stopping\n", task.Name)
			return
		case <-ticker.C:
			if !atomic.CompareAndSwapInt32(&running, 0, 1) {
				fmt.Printf("[Scheduler] task [%s] still running, skip this tick\n", task.Name)
				continue
			}
			go func() {
				defer atomic.StoreInt32(&running, 0)
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("[Scheduler] task [%s] panic: %v\n", task.Name, r)
					}
				}()
				task.Handler()
			}()
		}
	}
}
