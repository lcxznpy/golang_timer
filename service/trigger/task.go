package trigger

import (
	"xtimer/common/conf"
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
	//GetTasks(ctx context.Context, opts ...dao.Option) ([]*po.Task, error)
}
