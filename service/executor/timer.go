package executor

import (
	"context"
	"sync"
	"time"
	"xtimer/common/conf"
	"xtimer/common/consts"
	"xtimer/common/model/po"
	"xtimer/common/model/vo"
	taskdao "xtimer/dao/task"
	timerdao "xtimer/dao/timer"
	"xtimer/pkg/log"
)

type timerDAO interface {
	GetTimer(context.Context, ...timerdao.Option) (*po.Timer, error)
	GetTimers(ctx context.Context, opts ...timerdao.Option) ([]*po.Timer, error)
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

func (t *TimerService) Start(ctx context.Context) {
	t.Do(func() {
		go func() {
			t.ctx, t.stop = context.WithCancel(ctx)

			stepMinutes := t.confProvider.Get().TimerDetailCacheMinutes
			ticker := time.NewTicker(time.Duration(stepMinutes) * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				select {
				case <-t.ctx.Done():
					return
				default:
				}

				go func() {
					start := time.Now()
					t.timers, _ = t.getTimersByTime(ctx, start, start.Add(time.Duration(stepMinutes)*time.Minute))
				}()
			}
		}()
	})
}

// 执行器 : 用于返回数据库中定时器的状态
func (t *TimerService) getTimersByTime(ctx context.Context, start, end time.Time) (map[uint]*vo.Timer, error) {
	tasks, err := t.taskDAO.GetTasks(ctx, taskdao.WithStartTime(start), taskdao.WithEndTime(end))
	if err != nil {
		return nil, err
	}

	timerIDs := getTimerIDs(tasks)
	if len(timerIDs) == 0 {
		return nil, nil
	}
	// 存入内存map中的都是保证已经处于激活状态
	pTimers, err := t.timerDAO.GetTimers(ctx, timerdao.WithIDs(timerIDs), timerdao.WithStatus(int32(consts.Enabled)))
	if err != nil {
		return nil, err
	}
	return getTimersMap(pTimers)
}

func getTimerIDs(tasks []*po.Task) []uint {
	timerIDset := make(map[uint]struct{})
	for _, task := range tasks {
		if _, ok := timerIDset[task.TimerID]; ok {
			continue
		}
		timerIDset[task.TimerID] = struct{}{}
	}
	timerIDs := make([]uint, 0, len(timerIDset))
	for id := range timerIDset {
		timerIDs = append(timerIDs, id)
	}
	return timerIDs
}

func getTimersMap(pTimers []*po.Timer) (map[uint]*vo.Timer, error) {
	vTimers, err := vo.NewTimers(pTimers)
	if err != nil {
		return nil, err
	}

	timers := make(map[uint]*vo.Timer, len(vTimers))
	for _, vTimer := range vTimers {
		timers[vTimer.ID] = vTimer
	}
	return timers, nil
}

// 根据定时器id获取定时器信息
func (t *TimerService) GetTimer(ctx context.Context, id uint) (*vo.Timer, error) {
	// 直接从map中找
	if vTimer, ok := t.timers[id]; ok {
		return vTimer, nil
	}
	log.WarnContextf(ctx, "get timer from local cache failed, timerID: %d", id)
	// 从map找不到，再从数据库找
	timer, err := t.timerDAO.GetTimer(ctx, timerdao.WithID(id))
	if err != nil {
		return nil, err
	}
	return vo.NewTimer(timer)
}
