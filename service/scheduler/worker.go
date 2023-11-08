package scheduler

import (
	"context"
	"xtimer/common/conf"
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
