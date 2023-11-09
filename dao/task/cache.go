package task

import (
	"context"
	"fmt"
	"time"
	"xtimer/common/conf"
	"xtimer/common/consts"
	"xtimer/common/model/po"
	"xtimer/common/utils"
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

// 迁移 : mysql 存入 redis
func (t *TaskCache) BatchCreateTasks(ctx context.Context, tasks []*po.Task, start, end time.Time) error {
	if len(tasks) == 0 {
		return nil
	}
	commands := make([]*redis.Command, 0, 2*len(tasks))
	for _, task := range tasks {
		unix := task.RunTimer.UnixMilli()
		tableName := t.GetTableName(task)
		// 根据的zset_name key value 创建记录
		commands = append(commands, redis.NewZAddCommand(tableName, unix, utils.UnionTimerIDUnix(task.TimerID, unix)))
		// 设置zset过期时间 一天
		aliveSeconds := int64(time.Until(task.RunTimer.Add(24*time.Hour)) / time.Second)
		commands = append(commands, redis.NewExpireCommand(tableName, aliveSeconds))
	}
	// 开启事务执行添加操作
	_, err := t.client.Transaction(ctx, commands...)
	return err
}

// 迁移 : 获取表的名称 用于redis建表
func (t *TaskCache) GetTableName(task *po.Task) string {
	// 兜底取值
	maxBucket := t.confProvider.Get().BucketsNum
	return fmt.Sprintf("%s_%d", task.RunTimer.Format(consts.MinuteFormat), int64(task.TimerID)%int64(maxBucket))
}
