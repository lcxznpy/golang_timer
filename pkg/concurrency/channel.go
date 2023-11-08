package concurrency

import (
	"context"
	"sync"
)

type SafeChan struct {
	sync.Once
	ctx   context.Context
	close func()
	ch    chan interface{}
}

// 创建一个安全的goroutine
func NewSafeChan(size int) *SafeChan {
	s := SafeChan{
		ch: make(chan interface{}, size),
	}
	s.ctx, s.close = context.WithCancel(context.Background())
	return &s
}
