package timer

import (
	"context"
	"xtimer/common/model/po"
	"xtimer/pkg/mysql"
)

// 负责处理timer模块 结构体
type TimerDAO struct {
	client *mysql.Client
}

// 构造函数
func NewTimerDAO(client *mysql.Client) *TimerDAO {
	return &TimerDAO{
		client: client,
	}
}

// 创建定时器
func (t *TimerDAO) CreateTimer(ctx context.Context, timer *po.Timer) (uint, error) {
	return timer.ID, t.client.DB.WithContext(ctx).Create(timer).Error
}

// 迁移 : 从数据库中获取 可执行的 定时器切片
func (t *TimerDAO) GetTimers(ctx context.Context, opts ...Option) ([]*po.Timer, error) {
	db := t.client.DB.WithContext(ctx).Model(&po.Timer{})
	for _, opt := range opts {
		db = opt(db)
	}
	var timers []*po.Timer
	return timers, db.Scan(&timers).Error
}

// 迁移 : 批量创建定时任务
func (t *TimerDAO) BatchCreateRecords(ctx context.Context, tasks []*po.Task) error {
	return t.client.DB.Model(&po.Task{}).WithContext(ctx).CreateInBatches(tasks, len(tasks)).Error
}
