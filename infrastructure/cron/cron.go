package cron

import (
	"sync"

	"github.com/robfig/cron/v3"
)

type CronService interface {
	Start()
	Stop()
	AddFunc(flag string, fn func()) error
	AddJob(flag string, job cron.Job) (cron.EntryID, error)
}

type CronServiceImpl struct {
	cron *cron.Cron
}

var (
	cronService *CronServiceImpl
	once        sync.Once
	taskMutex   sync.RWMutex
	isRunning   = false
)
