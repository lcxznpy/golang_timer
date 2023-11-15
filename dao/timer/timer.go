package timer

import (
	"context"
	"xtimer/common/model/po"
	"xtimer/pkg/log"
	"xtimer/pkg/mysql"

	"gorm.io/gorm"
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

// 从数据库获取一个定时器
func (t *TimerDAO) GetTimer(ctx context.Context, opts ...Option) (*po.Timer, error) {
	db := t.client.DB.WithContext(ctx)
	for _, opt := range opts {
		db = opt(db)
	}
	var timer po.Timer
	return &timer, db.First(&timer).Error
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

// 删除定时器
func (t *TimerDAO) DeleteTimer(ctx context.Context, id uint) error {
	return t.client.DB.WithContext(ctx).Delete(&po.Timer{Model: gorm.Model{ID: id}}).Error
}

func (t *TimerDAO) UpdateTimer(ctx context.Context, timer *po.Timer) error {
	return t.client.DB.WithContext(ctx).Updates(timer).Error
}

// 带锁修改定时器状态，并创建两次迁移时间内的任务
func (t *TimerDAO) DoWithLock(ctx context.Context, id uint, do func(ctx context.Context, dao *TimerDAO, timer *po.Timer) error) error {
	return t.client.Transaction(func(tx *gorm.DB) error {
		defer func() {
			if err := recover(); err != nil {
				tx.Rollback()
				log.ErrorContextf(ctx, "transaction with lock err:%v,timer id:%d", err, id)

			}
		}()
		var timer po.Timer
		if err := tx.Set("gorm:query_option", "FOR UPDATE").WithContext(ctx).First(&timer, id).Error; err != nil {
			return err
		}
		return do(ctx, NewTimerDAO(mysql.NewClient(tx)), &timer)
	})
}
