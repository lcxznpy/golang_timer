package pool

import (
	"time"

	"github.com/panjf2000/ants/v2"
)

type WorkerPool interface {
	Submit(func()) error
}

// 协程工作池
type GoWorkerPool struct {
	pool *ants.Pool
}

// 提交任务函数，给协程池pool的worker 放入worker_channel中执行任务
func (g *GoWorkerPool) Submit(f func()) error {
	return g.pool.Submit(f)
}

// 创建协程工作池
func NewGoWorkerPool(size int) *GoWorkerPool {
	pool, err := ants.NewPool(
		size,
		ants.WithExpiryDuration(time.Minute))
	if err != nil {
		panic(err)
	}
	return &GoWorkerPool{pool: pool}
}
