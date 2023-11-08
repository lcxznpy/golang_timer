package task

import (
	"xtimer/common/conf"
	"xtimer/pkg/redis"
)

// 操作定时任务 存入redis的模块
type TaskCache struct {
	client       *redis.Client
	confProvider *conf.SchedulerAppConfProvider
}

// 构造函数
func NewTaskCache(client *redis.Client, confProvider *conf.SchedulerAppConfProvider) *TaskCache {
	return &TaskCache{
		client:       client,
		confProvider: confProvider,
	}
}
