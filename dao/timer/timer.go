package timer

import "xtimer/pkg/mysql"

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
