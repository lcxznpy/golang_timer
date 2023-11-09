package scheduler

import (
	"context"
	"sync"
	"xtimer/common/conf"
	"xtimer/pkg/log"
	service "xtimer/service/migrator"
)

type workerService interface {
	Start(context.Context) error
}

type confProvider interface {
	Get() *conf.SchedulerAppConf
}

type WorkerApp struct {
	sync.Once
	service workerService
	ctx     context.Context
	stop    func()
}

// 调度器模块app 初始化
func NewWorkerApp(service *service.Worker) *WorkerApp {
	w := WorkerApp{
		service: service,
	}
	w.ctx, w.stop = context.WithCancel(context.Background())
	return &w
}

func (w *WorkerApp) Start() {
	w.Do(w.start)
}

func (w *WorkerApp) start() {
	log.InfoContext(w.ctx, "scheduler worker_app starting")
	go func() {
		if err := w.service.Start(w.ctx); err != nil {
			log.ErrorContextf(w.ctx, "worker start failed, err: %v", err)
		}
	}()
}

func (w *WorkerApp) Stop() {
	w.stop()
	log.WarnContext(w.ctx, "scheduler worker_app is stopped")
}
