package webserver

import (
	"context"
	"fmt"
	"time"
	"xtimer/common/conf"
	"xtimer/common/model/po"
	"xtimer/common/model/vo"
	"xtimer/common/utils"
	taskdao "xtimer/dao/task"
	timerdao "xtimer/dao/timer"
	"xtimer/pkg/cron"
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
	//DeleteTimer(ctx context.Context, id uint) error
	//UpdateTimer(ctx context.Context, timer *po.Timer) error
	//GetTimer(ctx context.Context, opts ...timerdao.Option) (*po.Timer, error)
	//BatchCreateRecords(ctx context.Context, tasks []*po.Task) error
	//DoWithLock(ctx context.Context, id uint, do func(ctx context.Context, dao *timerdao.TimerDAO, timer *po.Timer) error) error
	//GetTimers(ctx context.Context, opts ...timerdao.Option) ([]*po.Timer, error)
	//Count(ctx context.Context, opts ...timerdao.Option) (int64, error)
}

type confProvider interface {
	Get() *conf.WebServerAppConf
}

type taskCache interface {
	//BatchCreateTasks(ctx context.Context, tasks []*po.Task, start, end time.Time) error
}

type cronParser interface {
	NextsBefore(cron string, end time.Time) ([]time.Time, error)
	IsValidCronExpr(cron string) bool
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
