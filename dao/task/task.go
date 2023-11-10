package task

import (
	"context"
	"xtimer/common/model/po"
	"xtimer/pkg/mysql"
)

// 负责处理task 定时任务模块的结构体
type TaskDAO struct {
	client *mysql.Client
}

func NewTaskDAO(client *mysql.Client) *TaskDAO {
	return &TaskDAO{
		client: client,
	}
}

// 迁移 : 获取定时任务 准备给redis
func (t *TaskDAO) GetTasks(ctx context.Context, opts ...Option) ([]*po.Task, error) {
	db := t.client.DB.WithContext(ctx)
	for _, opt := range opts {
		db = opt(db)
	}
	var tasks []*po.Task
	return tasks, db.Model(&po.Task{}).Scan(&tasks).Error
}

// 执行器 : 判断任务是否被重复执行
func (t *TaskDAO) GetTask(ctx context.Context, opts ...Option) (*po.Task, error) {
	db := t.client.DB.WithContext(ctx)
	for _, opt := range opts {
		db = opt(db)
	}
	var tasks *po.Task
	return tasks, db.First(&tasks).Error
}

func (t *TaskDAO) UpdateTask(ctx context.Context, task *po.Task) error {
	return t.client.DB.WithContext(ctx).Updates(task).Error
}
