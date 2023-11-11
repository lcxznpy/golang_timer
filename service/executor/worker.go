package executor

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"strings"
	"time"
	"xtimer/common/consts"
	"xtimer/common/model/vo"
	"xtimer/common/utils"
	taskdao "xtimer/dao/task"
	"xtimer/pkg/bloom"
	"xtimer/pkg/log"
	"xtimer/pkg/promethus"
	"xtimer/pkg/xhttp"
)

type Worker struct {
	timerService *TimerService
	taskDAO      *taskdao.TaskDAO
	httpClient   *xhttp.JSONClient
	bloomFilter  *bloom.Filter
	reporter     *promethus.Reporter
}

func NewWorker(timerService *TimerService, taskDAO *taskdao.TaskDAO, httpClient *xhttp.JSONClient, bloomFilter *bloom.Filter, reporter *promethus.Reporter) *Worker {
	return &Worker{
		timerService: timerService,
		taskDAO:      taskDAO,
		httpClient:   httpClient,
		bloomFilter:  bloomFilter,
		reporter:     reporter,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.timerService.Start(ctx)
}

func (w *Worker) Work(ctx context.Context, timeIDUnixKey string) error {
	// 拿到消息，查询一次完整的 timer 定义
	timerID, unix, err := utils.SplitTimerIDUnix(timeIDUnixKey)
	if err != nil {
		return err
	}

	// 幂等去重，通过该任务的 执行时间点 和  定时器的id与执行时间点拼接的string 找布隆过滤器
	// 如果布隆过滤器中有当前的hash 值，说明该任务可能已经被执行过了，再去数据库里面查查该任务到底有没有被执行
	if exist, err := w.bloomFilter.Exist(ctx, utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), timeIDUnixKey); err != nil {
		log.WarnContextf(ctx, "bloom filter check failed, start to check db, "+
			"bloom key: %s, timerIDUnixKey: %s, err: %v, exist: %t",
			utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))),
			timeIDUnixKey,
			err,
			exist)
		task, err := w.taskDAO.GetTask(ctx, taskdao.WithTimerID(timerID), taskdao.WithRunTimer(time.UnixMilli(unix)))
		if err == nil && task.Status != consts.NotRunned.ToInt() {
			// 重复执行的任务
			log.WarnContextf(ctx, "task is already executed, timerID: %d, exec_time: %v", timerID, task.RunTimer)
			return nil
		}
	}
	//数据库中的定时任务没有被执行，说明可以继续执行任务

	return w.executeAndPostProcess(ctx, timerID, unix)
}

// 执行定时任务
func (w *Worker) executeAndPostProcess(ctx context.Context, timerID uint, unix int64) error {
	// 先从内存中的map找，没有再从数据库找s
	timer, err := w.timerService.GetTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("get timer failed, id: %d, err: %w", timerID, err)
	}
	//从数据库找的定时器 没被激活
	if timer.Status != consts.Enabled {
		log.WarnContextf(ctx, "timer has alread been unabled, timerID: %d", timerID)
		return nil
	}
	execTime := time.Now()
	resp, err := w.execute(ctx, timer)
	return w.postProcess(ctx, resp, err, timer.App, timerID, unix, execTime)
}

// 正式执行定时器任务的请求
func (w *Worker) execute(ctx context.Context, timer *vo.Timer) (map[string]interface{}, error) {
	var (
		resp map[string]interface{}
		err  error
	)
	switch strings.ToUpper(timer.NotifyHTTPParam.Method) {
	case nethttp.MethodGet:
		err = w.httpClient.Get(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, nil, &resp)
	case nethttp.MethodPatch:
		err = w.httpClient.Patch(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, timer.NotifyHTTPParam.Body, &resp)
	case nethttp.MethodDelete:
		err = w.httpClient.Delete(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, timer.NotifyHTTPParam.Body, &resp)
	case nethttp.MethodPost:
		err = w.httpClient.Post(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, timer.NotifyHTTPParam.Body, &resp)
	default:
		err = fmt.Errorf("invalid http method: %s, timer: %s", timer.NotifyHTTPParam.Method, timer.Name)
	}

	return resp, err
}

// 封装响应
func (w *Worker) postProcess(ctx context.Context, resp map[string]interface{}, execErr error, app string,
	timerID uint, unix int64, execTime time.Time) error {
	//开启一个协程提交监管信息
	go w.reportMonitorData(app, unix, execTime)

	// 在bitmap中设置当前任务为执行状态
	if err := w.bloomFilter.Set(ctx, utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))),
		utils.UnionTimerIDUnix(timerID, unix),
		consts.BloomFilterKeyExpireSeconds); err != nil {
		log.ErrorContextf(ctx, "set bloom filter failed, key: %s, err: %v",
			utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), err)
	}
	// todo 布隆过滤器与数据库的一致性问题
	task, err := w.taskDAO.GetTask(ctx, taskdao.WithTimerID(timerID),
		taskdao.WithRunTimer(time.UnixMilli(unix)))
	if err != nil {
		return fmt.Errorf("get task failed, timerID: %d, runTimer: %d, err: %w", timerID, time.UnixMilli(unix), err)
	}
	respBody, _ := json.Marshal(resp)

	task.Output = string(respBody)
	if execErr != nil {
		log.InfoContextf(ctx, "http request error : %v", execErr)
		task.Status = consts.Failed.ToInt()
	} else {
		task.Status = consts.Successed.ToInt()
	}
	return w.taskDAO.UpdateTask(ctx, task)
}

// 上报pprof 监管信息
func (w *Worker) reportMonitorData(app string, expectExecTimeUnix int64, acutalExecTime time.Time) {
	w.reporter.ReportExecRecord(app)
	// 上报毫秒
	w.reporter.ReportTimerDelayRecord(app, float64(acutalExecTime.UnixMilli()-expectExecTimeUnix))
}
