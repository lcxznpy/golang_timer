package webserver

import (
	"context"
	"xtimer/common/model/vo"
	service "xtimer/service/webserver"
)

type taskService interface {
	GetTask(ctx context.Context, id uint) (*vo.Task, error)
	GetTasks(ctx context.Context, req *vo.GetTasksReq) ([]*vo.Task, int64, error)
}

type TaskApp struct {
	service taskService
}

func NewTaskApp(service *service.TaskService) *TaskApp {
	return &TaskApp{
		service: service,
	}
}
