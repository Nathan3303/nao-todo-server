package cron

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

func GetCronServiceImpl() CronService {
	once.Do(func() {
		cronService = &CronServiceImpl{
			cron: cron.New(),
		}
	})
	return cronService
}

// Start implements CronService.
func (cs *CronServiceImpl) Start() {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	if isRunning {
		fmt.Println("Cron 已经在运行")
		return
	}

	cs.cron.Start()
	isRunning = true
	fmt.Println("Cron 已启动")
}

// Stop implements CronService.
func (cs *CronServiceImpl) Stop() {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	if !isRunning {
		return
	}

	cs.cron.Stop()
	isRunning = false
	fmt.Println("Cron 已停止")
}

// Add implements CronService.
func (cs *CronServiceImpl) AddFunc(flag string, fn func()) error {
	_, err := cs.cron.AddFunc(flag, fn)
	return err
}

// AddJob implements CronService.
func (cs *CronServiceImpl) AddJob(flag string, job cron.Job) (cron.EntryID, error) {
	return cs.cron.AddJob(flag, job)
}
