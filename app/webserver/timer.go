package webserver

import (
	"context"
	"xtimer/common/model/vo"
	service "xtimer/service/webserver"
)

type timerService interface {
	CreateTimer(ctx context.Context, timer *vo.Timer) (uint, error)
	DeleteTimer(ctx context.Context, app string, id uint) error
	UpdateTimer(ctx context.Context, timer *vo.Timer) error
	GetTimer(ctx context.Context, id uint) (*vo.Timer, error)
	EnableTimer(ctx context.Context, app string, id uint) error
	UnableTimer(ctx context.Context, app string, id uint) error
	GetAppTimers(ctx context.Context, req *vo.GetAppTimersReq) ([]*vo.Timer, int64, error)
	GetTimersByName(ctx context.Context, req *vo.GetTimersByNameReq) ([]*vo.Timer, int64, error)
}

type TimerApp struct {
	service timerService
}

func NewTimerApp(service *service.TimerService) *TimerApp {
	return &TimerApp{service: service}
}
