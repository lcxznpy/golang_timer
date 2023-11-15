package webserver

import (
	"context"
	"xtimer/common/consts"
	"xtimer/common/model/vo"
	dao "xtimer/dao/task"
)

type TaskService struct {
	dao *dao.TaskDAO
}

func NewTaskService(dao *dao.TaskDAO) *TaskService {
	return &TaskService{dao: dao}
}

func (t *TaskService) GetTasks(ctx context.Context, req *vo.GetTasksReq) ([]*vo.Task, int64, error) {
	total, err := t.dao.Count(ctx, dao.WithTimerID(req.TimerID), dao.WithStatuses([]int32{
		int32(consts.Running),
		int32(consts.Successed),
		int32(consts.Failed),
	}))
	if err != nil {
		return nil, -1, err
	}
	offset, limit := req.Get()
	if total <= int64(offset) {
		return []*vo.Task{}, total, nil
	}
	tasks, err := t.dao.GetTasks(ctx, dao.WithTimerID(req.TimerID), dao.WithPageLimit(offset, limit), dao.WithStatuses([]int32{
		int32(consts.Running),
		int32(consts.Successed),
		int32(consts.Failed),
	}), dao.WithDesc())
	if err != nil {
		return nil, -1, err
	}
	return vo.NewTasks(tasks), total, nil
}
