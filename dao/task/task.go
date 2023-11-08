package task

import "xtimer/pkg/mysql"

// 负责处理task 定时任务模块的结构体
type TaskDAO struct {
	client *mysql.Client
}

func NewTaskDAO(client *mysql.Client) *TaskDAO {
	return &TaskDAO{
		client: client,
	}
}
