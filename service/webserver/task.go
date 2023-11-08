package webserver

import dao "xtimer/dao/task"

type TaskService struct {
	dao *dao.TaskDAO
}

func NewTaskService(dao *dao.TaskDAO) *TaskService {
	return &TaskService{dao: dao}
}
