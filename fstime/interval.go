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
	once    sync.Once
}

func (it *Interval) initOnce() {
	it.once.Do(func() {
		it.closed = make(chan struct{})
	})
}

// Start 启动任务，每间隔固定时长 d，就会触发一次任务
func (it *Interval) Start(d time.Duration) {
	if it.tk != nil {
		panic("Interval already started")
	}
	it.initOnce()
	it.tk = time.NewTicker(d)
	go it.goStart()
}

func (it *Interval) goStart() {
	it.runFns()
	for it.Running() {
		select {
		case <-it.closed:
			return
		case <-it.tk.C:
			it.runFns()
		}
	}
}

func (it *Interval) runFns() {
	allFns := it.fns.Load()

	var wg sync.WaitGroup
	wg.Add(len(allFns))
	defer wg.Wait()

	if it.Concurrency < 1 {
		for i := 0; i < len(allFns); i++ {
			fn := allFns[i]
			go func() {
				defer func() {
					wg.Done()
					_ = recover()
				}()
				if it.stopped.Load() {
					return
				}
				fn()
			}()
		}
		return
	}

	limiter := make(chan struct{}, it.Concurrency)

	for i := 0; i < len(allFns); i++ {
		limiter <- struct{}{}
		fn := allFns[i]
		go func() {
			defer func() {
				wg.Done()
				<-limiter
				_ = recover()
			}()
			if it.stopped.Load() {
				return
			}
			fn()
		}()
	}
}

// Stop 停止运行
func (it *Interval) Stop() {
	if !it.stopped.CompareAndSwap(false, true) {
		return
	}
	it.tk.Stop()
	close(it.closed)
}

// Add 注册回调函数
//
//	 应确保函数不会 panic。若 fn panic，会自动 recover 同时将 panic 信息丢弃。
//		默认情况下，若 fn 运行时间 > 调度时间间隔，同一个 fn 在同一时间会有多个运行实例
func (it *Interval) Add(fn func()) {
	it.fns.Add(fn)
}

// Reset 重置时间,应该先使用 Start 启动任务
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
	it.initOnce()
	return it.closed
}
