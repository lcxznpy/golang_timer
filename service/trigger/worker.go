package trigger

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"xtimer/common/conf"
	"xtimer/common/model/vo"
	"xtimer/common/utils"
	"xtimer/pkg/concurrency"
	"xtimer/pkg/pool"
	"xtimer/pkg/redis"
	"xtimer/service/executor"
)

type taskService interface {
	GetTasksByTime(ctx context.Context, key string, bucket int, start, end time.Time) ([]*vo.Task, error)
}

type confProvider interface {
	Get() *conf.TriggerAppConf
}

type Worker struct {
	task         taskService
	confProvider confProvider
	pool         pool.WorkerPool
	executor     *executor.Worker
	lockService  *redis.Client
}

func NewWorker(executor *executor.Worker, task *TaskService, lockService *redis.Client, confProvider *conf.TriggerAppConfProvider) *Worker {
	return &Worker{
		executor:     executor,
		task:         task,
		lockService:  lockService,
		pool:         pool.NewGoWorkerPool(confProvider.Get().WorkersNum),
		confProvider: confProvider,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.executor.Start(ctx)
}

// 触发器工作
func (w *Worker) Work(ctx context.Context, minuteBucketKey string, ack func()) error {
	startTime, err := getStartMinute(minuteBucketKey)
	if err != nil {
		return err
	}

	conf := w.confProvider.Get()
	// 定时器每1s响一次
	ticker := time.NewTicker(time.Duration(conf.ZRangeGapSeconds) * time.Second)
	defer ticker.Stop()

	endTime := startTime.Add(time.Minute)

	//获取接收信息的goroutine
	notifier := concurrency.NewSafeChan(int(time.Minute/(time.Duration(conf.ZRangeGapSeconds)*time.Second)) + 1)
	defer notifier.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := w.handleBatch(ctx, minuteBucketKey, startTime, startTime.Add(time.Duration(conf.ZRangeGapSeconds)*time.Second)); err != nil {
			notifier.Put(err)
		}
	}()
}

func (w *Worker) handleBatch(ctx context.Context, key string, start, end time.Time) error {
	// 获取分到的桶id
	bucket, err := getBucket(key)
	if err != nil {
		return err
	}

	// 从redis中获取位于start到end区间内的任务切片
	tasks, err := w.task.GetTasksByTime(ctx, key, bucket, start, end)
	if err != nil {
		return err
	}
	timerIDs := make([]uint, 0, len(tasks))
	for _, task := range tasks {
		timerIDs = append(timerIDs, task.TimerID)
	}
	//不断从trigge的协程池中获取goroutine 执行任务
	for _,task := range tasks{
		task := task
		if err := w.pool.Submit(func() {
			if err := w.executor.Work(ctx,)
		})
	}
}

func getBucket(slice string) (int, error) {
	timeBucket := strings.Split(slice, "_")
	if len(timeBucket) != 2 {
		return -1, fmt.Errorf("invalid format of msg key: %s", slice)
	}
	return strconv.Atoi(timeBucket[1])
}

// 触发器 : 获取当前触发器所要 选择的时间区间 开始点
func getStartMinute(slice string) (time.Time, error) {
	timeBucket := strings.Split(slice, "_")
	if len(timeBucket) != 2 {
		return time.Time{}, fmt.Errorf("invalid format of msg key: %s", slice)
	}
	return utils.GetStartMinute(timeBucket[0])
}
