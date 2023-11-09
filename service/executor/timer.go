package executor

import (
	"context"
	"sync"
	"xtimer/common/conf"
	"xtimer/common/model/vo"
	taskdao "xtimer/dao/task"
	timerdao "xtimer/dao/timer"
)

type timerDAO interface {
	//GetTimer(context.Context, ...timerdao.Option) (*po.Timer, error)
	//GetTimers(ctx context.Context, opts ...timerdao.Option) ([]*po.Timer, error)
}

type TimerService struct {
	sync.Once
	confProvider *conf.MigratorAppConfProvider
	ctx          context.Context
	stop         func()
	timers       map[uint]*vo.Timer
	timerDAO     timerDAO
	taskDAO      *taskdao.TaskDAO
}

func NewTimerService(timerDAO *timerdao.TimerDAO, taskDAO *taskdao.TaskDAO, confProvider *conf.MigratorAppConfProvider) *TimerService {
	return &TimerService{
		confProvider: confProvider,
		timers:       make(map[uint]*vo.Timer),
		timerDAO:     timerDAO,
		taskDAO:      taskDAO,
	}
}
