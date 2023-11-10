package service

import (
	"context"
	"fmt"
	"time"
	mconf "xtimer/common/conf"
	"xtimer/common/consts"
	"xtimer/common/utils"
	taskdao "xtimer/dao/task"
	timerdao "xtimer/dao/timer"
	"xtimer/pkg/cron"
	"xtimer/pkg/log"
	"xtimer/pkg/pool"
	"xtimer/pkg/redis"
)

type Worker struct {
	timerDAO          *timerdao.TimerDAO
	taskDAO           *taskdao.TaskDAO
	taskCache         *taskdao.TaskCache
	cronParser        *cron.CronParser
	lockService       *redis.Client
	appConfigProvider *mconf.MigratorAppConfProvider
	pool              pool.WorkerPool
}

func NewWorker(timerDAO *timerdao.TimerDAO, taskDAO *taskdao.TaskDAO, taskCache *taskdao.TaskCache, lockService *redis.Client,
	cronParser *cron.CronParser, appConfigProvider *mconf.MigratorAppConfProvider) *Worker {
	return &Worker{
		pool:              pool.NewGoWorkerPool(appConfigProvider.Get().WorkersNum),
		timerDAO:          timerDAO,
		taskDAO:           taskDAO,
		taskCache:         taskCache,
		lockService:       lockService,
		cronParser:        cronParser,
		appConfigProvider: appConfigProvider,
	}
}

// migrator 启动
func (w *Worker) Start(ctx context.Context) error {
	conf := w.appConfigProvider.Get()
	//ticker := time.NewTicker(time.Duration(conf.MigrateStepMinutes) * time.Minute)
	//defer ticker.Stop()
	//range ticker.C

	// 每隔一段时间获取一次定时器
	for i := 0; i < 1; i++ {
		log.InfoContext(ctx, "migrator ticking...")
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		// 尝试获取当前时间段的分布式锁
		locker := w.lockService.GetDistributionLock(utils.GetMigratorLockKey(utils.GetStartHour(time.Now())))
		// 20min 分布式锁expire
		if err := locker.Lock(ctx, int64(conf.MigrateTryLockMinutes)*int64(time.Minute/time.Second)); err != nil {
			log.ErrorContextf(ctx, "migrator get lock failed,key:%s, err:%v", utils.GetMigratorLockKey(utils.GetStartHour(time.Now())), err)
			continue
		}
		// 成功获取分布式锁，开始迁移
		if err := w.migrate(ctx); err != nil {
			log.ErrorContextf(ctx, "migrator failed, err:%v", err)
			continue
		}
		// 设置分布式锁的过期时间
		_ = locker.ExpireLock(ctx, int64(conf.MigrateSucessExpireMinutes)*int64(time.Minute/time.Second))
	}
	return nil
}

// 定时从mysql db中获取定时器
func (w *Worker) migrate(ctx context.Context) error {
	// 获取定时器切片
	timers, err := w.timerDAO.GetTimers(ctx, timerdao.WithStatus(int32(consts.Enabled.ToInt())))
	if err != nil {
		return err
	}

	conf := w.appConfigProvider.Get()
	now := time.Now()
	fmt.Println(now, "?????????????????????????????????")
	// 获取定时器 的任务时间区间[start,end]，只有符合定时器的cron匹配且在这个区间内的时间任务才会被迁移
	start := utils.GetStartHour(now.Add(time.Duration(conf.MigrateStepMinutes) * time.Minute))
	end := utils.GetStartHour(now.Add(2 * time.Duration(conf.MigrateStepMinutes) * time.Minute))

	// 迁移可以慢慢来，不着急，根据定时器批量创建定时任务
	for _, timer := range timers {
		//返回start 和end区间内符合cron表达式的日期切片
		nexts, _ := w.cronParser.NextsBetween(timer.Cron, start, end)
		//通过日期切片包装任务切片,并在数据库里创建记录
		if err := w.timerDAO.BatchCreateRecords(ctx, timer.BatchTasksFromTimer(nexts)); err != nil {
			log.ErrorContextf(ctx, "migrator batch create records for timer:%d failed,err:%v", timer.ID, err)
		}
		//time.Sleep(5 * time.Second)
	}
	return w.migrateToCache(ctx, start, end)
}

// 将mysql中的定时任务 迁移至 redis中
func (w *Worker) migrateToCache(ctx context.Context, start, end time.Time) error {
	tasks, err := w.taskDAO.GetTasks(ctx, taskdao.WithStartTime(start), taskdao.WithEndTime(end))
	if err != nil {
		log.ErrorContextf(ctx, "migrator batch get tasks failed,err:%v", err)
		return err
	}
	return w.taskCache.BatchCreateTasks(ctx, tasks, start, end)
}
