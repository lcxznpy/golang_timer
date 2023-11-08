package migrator

import (
	"context"
	"sync"
	"xtimer/pkg/log"
	service "xtimer/service/migrator"
)

// 迁移app
type MigratorApp struct {
	sync.Once
	ctx    context.Context
	stop   func()
	worker *service.Worker
}

// 迁移app构造函数
func NewMigratorApp(worker *service.Worker) *MigratorApp {
	m := MigratorApp{
		worker: worker,
	}
	m.ctx, m.stop = context.WithCancel(context.Background())
	return &m
}

func (m *MigratorApp) Start() {
	m.Do(func() {
		log.InfoContext(m.ctx, "migrator starting")
		go func() {
			if err := m.worker.start(); err != nil {
				log.ErrorContextf(m.ctx, "start migrator worker failed err : %s", err)
			}
		}()
	})
}

func (m *MigratorApp) Stop() {
	m.stop()
}
