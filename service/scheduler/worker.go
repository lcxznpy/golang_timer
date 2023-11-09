package scheduler

import (
	"context"
	"time"
	"xtimer/common/conf"
	"xtimer/common/utils"
	"xtimer/pkg/log"
	"xtimer/pkg/pool"
	"xtimer/pkg/redis"
	"xtimer/service/trigger"
)

type appConfProvider interface {
	Get() *conf.SchedulerAppConf
}

type lockService interface {
	GetDistributionLock(key string) redis.DistributeLocker
}

type bucketGetter interface {
	Get(ctx context.Context, key string) (string, error)
}

type Worker struct {
	pool            pool.WorkerPool
	appConfProvider appConfProvider
	trigger         *trigger.Worker
	lockService     lockService
	bucketGetter    bucketGetter
	minuteBuckets   map[string]int
}

func NewWorker(trigger *trigger.Worker, redisClient *redis.Client, appConfProvider *conf.SchedulerAppConfProvider) *Worker {
	return &Worker{
		pool:            pool.NewGoWorkerPool(appConfProvider.Get().WorkersNum),
		trigger:         trigger,
		lockService:     redisClient,
		bucketGetter:    redisClient,
		appConfProvider: appConfProvider,
		minuteBuckets:   make(map[string]int),
	}
}

// scheduler调度器 启动
func (w *Worker) Start(ctx context.Context) error {
	//同时启动trigger触发器
	w.trigger.Start(ctx)

	ticker := time.NewTicker(time.Duration(w.appConfProvider.Get().TryLockGapMilliSeconds) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			log.WarnContext(ctx, "scheduler stopped")
			return nil
		default:

		}
		w.handleSlices(ctx)
	}
	return nil
}

// 分桶，按桶的数量开协程分配给trigger
func (w *Worker) handleSlices(ctx context.Context) {
	for i := 0; i < w.getValidBucket(ctx); i++ {
		w.handleSlice(ctx, i)
	}
}

// 禁用动态分桶能力,写死分桶的数量
func (w *Worker) getValidBucket(ctx context.Context) int {
	return w.appConfProvider.Get().BucketsNum
}

func (w *Worker) handleSlice(ctx context.Context, bucketID int) {
	now := time.Now()

	if err := w.pool.Submit(func() {
		// 从协程池拿一个worker 将任务添加进worker chan中 用于后续trigger执行
		w.asyncHandleSlice(ctx, now.Add(-time.Minute), bucketID)
	}); err != nil {
		log.ErrorContextf(ctx, "[handle slice] submit task failed, err: %v", err)
	}
	if err := w.pool.Submit(func() {
		// 从协程池拿一个worker 将任务添加进worker chan中 用于后续trigger执行
		w.asyncHandleSlice(ctx, now, bucketID)
	}); err != nil {
		log.ErrorContextf(ctx, "[handle slice] submit task failed, err: %v", err)
	}
}

// 调用trigger 工作
func (w *Worker) asyncHandleSlice(ctx context.Context, t time.Time, bucketID int) {
	// 尝试获取当前执行时间和桶ID的分布式锁
	locker := w.lockService.GetDistributionLock(utils.GetTimeBucketLockKey(t, bucketID))
	if err := locker.Lock(ctx, int64(w.appConfProvider.Get().TryLockSeconds)); err != nil {
		return
	}
	log.InfoContextf(ctx, "get scheduler lock success, key: %s", utils.GetTimeBucketLockKey(t, bucketID))

	ack := func() {
		if err := locker.ExpireLock(ctx, int64(w.appConfProvider.Get().SuccessExpireSeconds)); err != nil {
			log.ErrorContextf(ctx, "expire lock failed, lock key: %s, err: %v", utils.GetTimeBucketLockKey(t, bucketID), err)
		}
	}
	if err := w.trigger.Work(ctx, utils.GetSliceMsgKey(t, bucketID), ack); err != nil {
		log.ErrorContextf(ctx, "trigger work failed, err: %v", err)
	}

}
