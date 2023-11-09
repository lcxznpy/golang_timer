package executor

import (
	"context"
	"time"
	"xtimer/common/consts"
	"xtimer/common/utils"
	taskdao "xtimer/dao/task"
	"xtimer/pkg/bloom"
	"xtimer/pkg/log"
	"xtimer/pkg/promethus"
	"xtimer/pkg/xhttp"
)

type Worker struct {
	timerService *TimerService
	taskDAO      *taskdao.TaskDAO
	httpClient   *xhttp.JSONClient
	bloomFilter  *bloom.Filter
	reporter     *promethus.Reporter
}

func NewWorker(timerService *TimerService, taskDAO *taskdao.TaskDAO, httpClient *xhttp.JSONClient, bloomFilter *bloom.Filter, reporter *promethus.Reporter) *Worker {
	return &Worker{
		timerService: timerService,
		taskDAO:      taskDAO,
		httpClient:   httpClient,
		bloomFilter:  bloomFilter,
		reporter:     reporter,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.timerService.Start(ctx)
}

func (w *Worker) Work(ctx context.Context, timeIDUnixKey string) error {
	// 拿到消息，查询一次完整的 timer 定义
	timerID, unix, err := utils.SplitTimerIDUnix(timeIDUnixKey)
	if err != nil {
		return err
	}

	// 幂等去重，通过该任务的 执行时间点 和  定时器的id与执行时间点拼接的string 找布隆过滤器
	// 如果布隆过滤器中有当前的hash 值，说明该任务可能已经被执行过了，再去数据库里面查查该任务到底有没有被执行
	if exist, err := w.bloomFilter.Exist(ctx, utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), timeIDUnixKey); err != nil {
		log.WarnContextf(ctx, "bloom filter check failed, start to check db, "+
			"bloom key: %s, timerIDUnixKey: %s, err: %v, exist: %t",
			utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))),
			timeIDUnixKey,
			err,
			exist)
		task, err := w.taskDAO.GetTask(ctx, taskdao.WithTimerID(timerID), taskdao.WithRunTimer(time.UnixMilli(unix)))
		if err == nil && task.Status != consts.NotRunned.ToInt() {
			// 重复执行的任务
			log.WarnContextf(ctx, "task is already executed, timerID: %d, exec_time: %v", timerID, task.RunTimer)
			return nil
		}
	}
	//数据库中的定时任务没有被执行，说明可以继续执行任务

	return w.executeAndPostProcess(ctx, timerID, unix)
}

// 执行定时任务
func (w *Worker) executeAndPostProcess(ctx context.Context, timerID uint, unix int64) {
	从数据库中 看这个timer还在不在
	timer,err := w.timerService.
}
