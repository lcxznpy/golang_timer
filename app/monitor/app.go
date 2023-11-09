package monitor

import (
	"context"
	"sync"
	service "xtimer/service/migrator"
)

// 监管器，检测未成功执行的定时任务和定时器，并响应给pprof检测
type MonitorApp struct {
	sync.Once
	ctx    context.Context
	stop   func()
	worker *service.Worker
}

func NewMonitorApp(worker *service.Worker) *MonitorApp {
	m := MonitorApp{
		worker: worker,
	}
	m.ctx, m.stop = context.WithCancel(context.Background())
	return &m
}

//func (m *MonitorApp) Start() {
//	m.Do(func() {
//		log.InfoContext(m.ctx, "monitor app starting")
//		go m.worker.Start(m.ctx)
//	})
//}

func (m *MonitorApp) Stop() {
	m.stop()
}
