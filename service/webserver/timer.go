package webserver

import (
	"context"
	"errors"
	"fmt"
	"time"
	"xtimer/common/conf"
	"xtimer/common/consts"
	"xtimer/common/model/po"
	"xtimer/common/model/vo"
	"xtimer/common/utils"
	taskdao "xtimer/dao/task"
	timerdao "xtimer/dao/timer"
	"xtimer/pkg/cron"
	"xtimer/pkg/log"
	"xtimer/pkg/mysql"
	"xtimer/pkg/redis"
)

const defaultEnableGapSeconds = 3

type TimerService struct {
	dao                 timerDAO
	confProvider        confProvider
	migrateConfProvider *conf.MigratorAppConfProvider
	cronParser          cronParser
	taskCache           taskCache
	lockService         *redis.Client
}

func NewTimerService(dao *timerdao.TimerDAO, taskCache *taskdao.TaskCache, lockService *redis.Client,
	confProvider *conf.WebServerAppConfProvider, migrateConfProvider *conf.MigratorAppConfProvider, parser *cron.CronParser) *TimerService {
	return &TimerService{
		dao:                 dao,
		confProvider:        confProvider,
		migrateConfProvider: migrateConfProvider,
		taskCache:           taskCache,
		cronParser:          parser,
		lockService:         lockService,
	}
}

type timerDAO interface {
	CreateTimer(ctx context.Context, timer *po.Timer) (uint, error)
	DeleteTimer(ctx context.Context, id uint) error
	UpdateTimer(ctx context.Context, timer *po.Timer) error
	GetTimer(ctx context.Context, opts ...timerdao.Option) (*po.Timer, error)
	BatchCreateRecords(ctx context.Context, tasks []*po.Task) error
	DoWithLock(ctx context.Context, id uint, do func(ctx context.Context, dao *timerdao.TimerDAO, timer *po.Timer) error) error
	GetTimers(ctx context.Context, opts ...timerdao.Option) ([]*po.Timer, error)
	Count(ctx context.Context, opts ...timerdao.Option) (int64, error)
}

// service 层的创建定时器业务处理
func (t *TimerService) CreateTimer(ctx context.Context, timer *vo.Timer) (uint, error) {
	lock := t.lockService.GetDistributionLock(utils.GetCreateLockKey(timer.App))
	if err := lock.Lock(ctx, defaultEnableGapSeconds); err != nil {
		return 0, err
		//errors.New("创建/删除过于频繁，拿不到分布式锁，请稍后再试")
	}
	// 校验提交的cron表达式
	if !t.cronParser.IsValidCronExpr(timer.Cron) {
		return 0, fmt.Errorf("非法的cron表达式:%s", timer.Cron)
	}
	// 封装 可存入数据库的 样式
	pTimer, err := timer.ToPO()
	if err != nil {
		return 0, err
	}
	return t.dao.CreateTimer(ctx, pTimer)
}

func (t *TimerService) DeleteTimer(ctx context.Context, app string, id uint) error {
	lock := t.lockService.GetDistributionLock(utils.GetCreateLockKey(app))
	if err := lock.Lock(ctx, defaultEnableGapSeconds); err != nil {
		return errors.New("创建/删除操作过于频繁，请稍后再试！")
	}
	return t.dao.DeleteTimer(ctx, id)
}

func (t *TimerService) UpdateTimer(ctx context.Context, timer *vo.Timer) error {
	pTimer, err := timer.ToPO()
	if err != nil {
		return err
	}
	return t.dao.UpdateTimer(ctx, pTimer)
}

func (t *TimerService) GetTimer(ctx context.Context, id uint) (*vo.Timer, error) {
	pTimer, err := t.dao.GetTimer(ctx, timerdao.WithID(id))
	if err != nil {
		return nil, err
	}
	return vo.NewTimer(pTimer)
}

func (t *TimerService) BatchCreateRecords(ctx context.Context, tasks []*po.Task) error {
	//TODO implement me
	panic("implement me")
}

func (t *TimerService) GetAppTimers(ctx context.Context, req *vo.GetAppTimersReq) ([]*vo.Timer, int64, error) {

}

// 激活定时器
func (t *TimerService) EnableTimer(ctx context.Context, app string, id uint) error {
	// 通过锁来 限制 激活和使定时器失效的次数
	lock := t.lockService.GetDistributionLock(utils.GetEnableLockKey(app))
	if err := lock.Lock(ctx, defaultEnableGapSeconds); err != nil {
		return errors.New("激活/使失效操作过于频发，稍后再试")
	}
	do := func(ctx context.Context, dao *timerdao.TimerDAO, timer *po.Timer) error {
		// 检查当前定时器状态
		if timer.Status != consts.Unabled.ToInt() {
			return fmt.Errorf("not unabled status,enable failed,timer id : %d", id)
		}
		//取得两次批量迁移的执行时间
		//
		start := time.Now()
		end := utils.GetForwardTwoMigrateStepEnd(start, 2*time.Duration(t.migrateConfProvider.Get().MigrateStepMinutes)*time.Minute)
		executeTimes, err := t.cronParser.NextsBefore(timer.Cron, end)
		if err != nil {
			log.ErrorContextf(ctx, "get executeTimes failed err:%v", err)
			return err
		}
		//根据执行时间创建任务 加入数据库
		tasks := timer.BatchTasksFromTimer(executeTimes)
		// 基于 timer_id + run_timer 唯一键，保证任务不被重复插入 ，不是重复插入错误就返回err
		if err := dao.BatchCreateRecords(ctx, tasks); err != nil && !mysql.IsDuplicateEntryErr(err) {
			return err
		}
		// 加入redis跳表
		if err := t.taskCache.BatchCreateTasks(ctx, tasks, start, end); err != nil {
			return err
		}
		timer.Status = consts.Unabled.ToInt()
		return dao.UpdateTimer(ctx, timer)

	}
	return t.dao.DoWithLock(ctx, id, do)
}

// 使定时器失效
func (t *TimerService) UnableTimer(ctx context.Context, app string, id uint) error {
	// 通过锁来 限制 激活和使定时器失效的次数
	lock := t.lockService.GetDistributionLock(utils.GetEnableLockKey(app))
	if err := lock.Lock(ctx, defaultEnableGapSeconds); err != nil {
		return errors.New("激活/使失效操作过于频发，稍后再试")
	}

	do := func(ctx context.Context, dao *timerdao.TimerDAO, timer *po.Timer) error {
		if timer.Status != consts.Enabled.ToInt() {
			return fmt.Errorf("not enabled status, unable failed, timer id:%d", id)
		}
		timer.Status = consts.Unabled.ToInt()
		return dao.UpdateTimer(ctx, timer)
	}
	return t.dao.DoWithLock(ctx, id, do)
}

func (t *TimerService) Count(ctx context.Context, opts ...timerdao.Option) (int64, error) {
	//TODO implement me
	panic("implement me")
}

type confProvider interface {
	Get() *conf.WebServerAppConf
}

type taskCache interface {
	BatchCreateTasks(ctx context.Context, tasks []*po.Task, start, end time.Time) error
}

type cronParser interface {
	NextsBefore(cron string, end time.Time) ([]time.Time, error)
	IsValidCronExpr(cron string) bool
}
