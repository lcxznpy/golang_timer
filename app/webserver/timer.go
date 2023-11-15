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

// 删除定时器
func (t *TimerApp) DeleteTimer(c *gin.Context) {
	var req vo.TimerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[delete timer] bind req failed, err: %v", err)))
		return
	}
	if err := t.service.DeleteTimer(c.Request.Context(), req.App, req.ID); err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewCodeMsgWithErr(nil))
}

// 不允许修改定时器,如果要修改，请先删除旧的定时器后，再增加新的定时器
func (t *TimerApp) UpdateTimer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"msg": "不允许修改定时器,如果要修改，请先删除旧的定时器后，再增加新的定时器",
	})
}

// 获取定时器信息
func (t *TimerApp) GetTimer(c *gin.Context) {
	var req vo.TimerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[get timer] bind req failed, err: %v", err)))
		return
	}
	timer, err := t.service.GetTimer(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
	}
	c.JSON(http.StatusOK, vo.NewGetTimerResp(timer, vo.NewCodeMsgWithErr(nil)))

}

// 根据app 获取timer
func (t *TimerApp) GetAppTimers(c *gin.Context) {
	var req vo.GetAppTimersReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[get app timers] bind req failed, err: %v", err)))
		return
	}
	timers, total, err := t.service.GetAppTimers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewGetTimersResp(timers, total, vo.NewCodeMsgWithErr(nil)))
}

// 根据 app 指定的name查询定时器
func (t *TimerApp) GetTimersByName(c *gin.Context) {
	var req vo.GetTimersByNameReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[get timers by name] bind req failed, err: %v", err)))
		return
	}
	timers, total, err := t.service.GetTimersByName(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewGetTimersResp(timers, total, vo.NewCodeMsgWithErr(nil)))

}

// 激活定时器
func (t *TimerApp) EnableTimer(c *gin.Context) {
	var req vo.TimerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[enable timer] bind req failed, err: %v", err)))
		return
	}
	if err := t.service.EnableTimer(c.Request.Context(), req.App, req.ID); err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewCodeMsgWithErr(nil))

}

func (t *TimerApp) UnableTimer(c *gin.Context) {
	var req vo.TimerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[enable timer] bind req failed, err:%v", err)))
		return
	}

	if err := t.service.UnableTimer(c.Request.Context(), req.App, req.ID); err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewCodeMsgWithErr(nil))
}
