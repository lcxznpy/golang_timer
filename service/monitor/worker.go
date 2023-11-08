package monitor

import (
	taskdao "xtimer/dao/task"
	timerdao "xtimer/dao/timer"
	"xtimer/pkg/promethus"
	"xtimer/pkg/redis"
)

type Worker struct {
	lockService *redis.Client
	taskDAO     *taskdao.TaskDAO
	timerDAO    *timerdao.TimerDAO
	reporter    *promethus.Reporter
}

func NewWorker(taskDAO *taskdao.TaskDAO, timerDAO *timerdao.TimerDAO, lockService *redis.Client, reporter *promethus.Reporter) *Worker {
	return &Worker{
		taskDAO:     taskDAO,
		timerDAO:    timerDAO,
		lockService: lockService,
		reporter:    reporter,
	}
}
