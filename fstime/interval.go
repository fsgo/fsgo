// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/31

package fstime

import (
	"sync"
	"sync/atomic"
	"time"
)

// Interval 定时器
type Interval struct {
	concurrency chan struct{}

	tk     *time.Ticker
	closed chan struct{}
	fns    []func()
	// Concurrency 回调任务并发度，当为 0 时，为全并发
	Concurrency int
	mux         sync.RWMutex
	stopped     int32
}

// Start 启动任务
func (it *Interval) Start(d time.Duration) {
	if it.tk != nil {
		panic("Interval already started")
	}
	it.tk = time.NewTicker(d)
	it.closed = make(chan struct{})
	if it.Concurrency > 0 {
		it.concurrency = make(chan struct{}, it.Concurrency)
	}
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
	it.mux.RLock()
	defer it.mux.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(it.fns))
	defer wg.Wait()

	if it.Concurrency < 1 {
		for i := 0; i < len(it.fns); i++ {
			fn := it.fns[i]
			go func() {
				fn()
				wg.Done()
			}()
		}
		return
	}

	for i := 0; i < len(it.fns); i++ {
		it.concurrency <- struct{}{}
		fn := it.fns[i]
		go func() {
			fn()
			<-it.concurrency
			wg.Done()
		}()
	}
}

// Stop 停止运行
func (it *Interval) Stop() {
	if !it.Running() {
		return
	}
	atomic.StoreInt32(&it.stopped, 1)
	it.tk.Stop()
	close(it.closed)
}

// Add 注册回调函数
//
// 应确保函数不会 panic
func (it *Interval) Add(fn func()) {
	it.mux.Lock()
	it.fns = append(it.fns, fn)
	it.mux.Unlock()
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
	return atomic.LoadInt32(&it.stopped) == 0
}

// Done 运行状态
func (it *Interval) Done() <-chan struct{} {
	return it.closed
}
