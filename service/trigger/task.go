package trigger

import (
	"context"
	"time"
	"xtimer/common/conf"
	"xtimer/common/consts"
	"xtimer/common/model/po"
	"xtimer/common/model/vo"
	dao "xtimer/dao/task"
)

type TaskService struct {
	confPrivder *conf.SchedulerAppConfProvider
	cache       *dao.TaskCache
	dao         taskDAO
}

func NewTaskService(dao *dao.TaskDAO, cache *dao.TaskCache, confPrivder *conf.SchedulerAppConfProvider) *TaskService {
	return &TaskService{
		confPrivder: confPrivder,
		dao:         dao,
		cache:       cache,
	}
}

type taskDAO interface {
	GetTasks(ctx context.Context, opts ...dao.Option) ([]*po.Task, error)
}

func (t *TaskService) GetTasksByTime(ctx context.Context, key string, bucket int, start, end time.Time) ([]*vo.Task, error) {
	// 先走缓存
	if tasks, err := t.cache.GetTasksByTime(ctx, key, start.UnixMilli(), end.UnixMilli()); err == nil && len(tasks) > 0 {
		return vo.NewTasks(tasks), nil
	}

	// 缓存没有，再走db
	tasks, err := t.dao.GetTasks(ctx, dao.WithStartTime(start), dao.WithEndTime(end), dao.WithStatus(int32(consts.NotRunned.ToInt())))
	if err != nil {
		return nil, err
	}
	maxBucket := t.confPrivder.Get().BucketsNum
	var validTask []*po.Task
	for _, task := range tasks {
		if task.TimerID%uint(maxBucket) != uint(bucket) {
			continue
		}
		validTask = append(validTask, task)
	}
	return vo.NewTasks(validTask), nil
}
