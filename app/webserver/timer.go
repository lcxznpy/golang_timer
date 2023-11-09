package webserver

import (
	"context"
	"fmt"
	"net/http"
	"xtimer/common/model/vo"
	service "xtimer/service/webserver"

	"github.com/gin-gonic/gin"
)

type timerService interface {
	CreateTimer(ctx context.Context, timer *vo.Timer) (uint, error)
	//DeleteTimer(ctx context.Context, app string, id uint) error
	//UpdateTimer(ctx context.Context, timer *vo.Timer) error
	//GetTimer(ctx context.Context, id uint) (*vo.Timer, error)
	//EnableTimer(ctx context.Context, app string, id uint) error
	//UnableTimer(ctx context.Context, app string, id uint) error
	//GetAppTimers(ctx context.Context, req *vo.GetAppTimersReq) ([]*vo.Timer, int64, error)
	//GetTimersByName(ctx context.Context, req *vo.GetTimersByNameReq) ([]*vo.Timer, int64, error)
}

type TimerApp struct {
	service timerService
}

func NewTimerApp(service *service.TimerService) *TimerApp {
	return &TimerApp{service: service}
}

// CreateTimer 创建定时器
func (t *TimerApp) CreateTimer(c *gin.Context) {
	// 1. 绑定参数
	var req vo.Timer
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[create timer] bind req failed, err: %v", err)))
		return
	}
	// 2. 业务处理
	id, err := t.service.CreateTimer(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	// 3. 返回成功响应
	c.JSON(http.StatusOK, vo.NewCreateTimerResp(id, vo.NewCodeMsgWithErr(nil)))
}

func (t *TimerApp) DeleteTimer(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (t *TimerApp) UpdateTimer(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (t *TimerApp) GetTimer(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (t *TimerApp) EnableTimer(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (t *TimerApp) UnableTimer(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (t *TimerApp) GetAppTimers(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (t *TimerApp) GetTimersByName(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
