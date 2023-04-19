// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/31

package fstime

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsgo/fsgo/fssync"
)

// Interval 定时器
type Interval struct {
	tk     *time.Ticker
	closed chan struct{}
	fns    fssync.Slice[func()]

	// Concurrency 回调任务并发度，当为 0 时，为全并发
	Concurrency int

	stopped atomic.Bool
}

// Start 启动任务
func (it *Interval) Start(d time.Duration) {
	if it.tk != nil {
		panic("Interval already started")
	}
	it.tk = time.NewTicker(d)
	it.closed = make(chan struct{})
	go it.goStart()
}

func (it *Interval) goStart() {
	it.runFns()
	for {
		select {
		case <-it.closed:
			return
		case <-it.tk.C:
			it.runFns()
		}
	}
}

func (it *Interval) runFns() {
	var wg sync.WaitGroup
	allFns := it.fns.Load()
	wg.Add(len(allFns))
	defer wg.Wait()

	if it.Concurrency < 1 {
		for i := 0; i < len(allFns); i++ {
			fn := allFns[i]
			go func() {
				fn()
				wg.Done()
			}()
		}
		return
	}

	limiter := make(chan struct{}, it.Concurrency)

	for i := 0; i < len(allFns); i++ {
		limiter <- struct{}{}
		fn := allFns[i]
		go func() {
			fn()
			<-limiter
			wg.Done()
		}()
	}
}

// Stop 停止运行
func (it *Interval) Stop() {
	if !it.Running() {
		return
	}
	it.stopped.Store(true)
	it.tk.Stop()
	close(it.closed)
}

// Add 注册回调函数
//
// 应确保函数不会 panic
func (it *Interval) Add(fn func()) {
	it.fns.Add(fn)
}

// Reset 重置时间
func (it *Interval) Reset(d time.Duration) {
	if !it.Running() {
		return
	}
	it.tk.Reset(d)
}

// Running 返回定时器的运行状态
func (it *Interval) Running() bool {
	return !it.stopped.Load()
}

// Done 运行状态
func (it *Interval) Done() <-chan struct{} {
	return it.closed
}
