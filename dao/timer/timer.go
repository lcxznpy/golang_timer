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
