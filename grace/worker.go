/*
 * Copyright(C) 2021 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/1/31
 */

package grace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type WorkerOption struct {
	Cmd         string
	CmdArgs     []string
	StopTimeout int
}

func (c *WorkerOption) getWorkerCmd() (string, []string) {
	if len(c.Cmd) > 0 {
		return c.Cmd, c.CmdArgs
	}
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	return os.Args[0], args
}

func NewWorker(c *WorkerOption) *Worker {
	if c == nil {
		c = &WorkerOption{}
	}
	g := &Worker{
		option:  c,
		stopped: false,
		event:   make(chan string, 1),
	}

	g.sub = &subProcess{
		group: g,
	}

	return g
}

// Worker 工作进程的逻辑
type Worker struct {
	main      *Grace
	option    *WorkerOption
	cmd       *exec.Cmd
	closeFunc context.CancelFunc

	resources []*resourceServer
	mux       sync.Mutex
	sub       *subProcess
	stopped   bool

	event chan string
}

func (w *Worker) Register(dsn string, c Consumer) error {
	res, err := GenResourceByDSN(dsn)
	if err != nil {
		return err
	}
	return w.register(res, c)
}

// register 注册资源
func (w *Worker) register(res Resource, c Consumer) error {
	ss := &resourceServer{
		Resource: res,
		Consumer: c,
	}
	if c != nil {
		c.Bind(res)
	}
	w.resources = append(w.resources, ss)
	return nil
}

// mainStart 主进程开启开始
func (w *Worker) mainStart(ctx context.Context) error {
	for _, info := range w.resources {
		err := info.Resource.Open(ctx)
		w.logit("open resource ", info.Resource.String(), ", error=", err)
		if err != nil {
			return err
		}
	}

	ctxFork, cancel := context.WithCancel(ctx)
	defer cancel()
	w.closeFunc = cancel

	// 启动一个子进程，用于处理请求
	if err := w.forkAndStart(ctxFork); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2, syscall.SIGQUIT)

	// hold on
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-ch:
			w.logit(fmt.Sprintf("receive signal(%v)", sig))
			switch sig {
			case syscall.SIGINT,
				syscall.SIGQUIT,
				syscall.SIGTERM:
				w.stopped = true
				w.stop(context.Background())
				return fmt.Errorf("shutdown by signal(%v)", sig)
			case syscall.SIGUSR2:
				w.mainReload(context.Background())
			}
		case e := <-w.event:
			switch e {
			case actionSubProcessExit:
				if !w.stopped {
					w.keepPrecess(context.Background())
				}
			}
		}
	}
	return fmt.Errorf("shutdown")
}

func (w *Worker) subProcessStart(ctx context.Context) error {
	return w.sub.Start(ctx)
}

func (w *Worker) logit(msgs ...interface{}) {
	msg := fmt.Sprintf("[grace][main] pid=%d %s", os.Getpid(), fmt.Sprint(msgs...))
	w.main.Logger.Output(1, msg)
}

func (w *Worker) forkAndStart(ctx context.Context) (ret error) {
	files := make([]*os.File, len(w.resources))
	for idx, s := range w.resources {
		f, err := s.Resource.File()
		if err != nil {
			return fmt.Errorf("listener[%d].File() has error: %w", idx, err)
		}
		files[idx] = f
	}

	cmdName, args := w.option.getWorkerCmd()
	cmd := exec.CommandContext(ctx, cmdName, args...)

	w.logit("fork new sub_process, cmd=", cmd.String())

	cmd.Env = append(os.Environ(), envActionKey+"="+actionSubStart)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	err := cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		start := time.Now()
		errWait := cmd.Wait()
		cost := time.Since(start)
		w.logit("cmd.Wait, error=", errWait, ", duration=", cost)
		w.event <- actionSubProcessExit
	}()

	_ = w.withLock(func() error {
		w.cmd = cmd
		return nil
	})

	return nil
}

func (w *Worker) withLock(fn func() error) error {
	w.mux.Lock()
	defer w.mux.Unlock()
	return fn()
}

// keepPrecess 检查检查是否存在
func (w *Worker) keepPrecess(ctx context.Context) (err error) {
	w.mux.Lock()
	lastCmd := w.cmd
	w.mux.Unlock()
	if !cmdExited(lastCmd) {
		return nil
	}
	// 若进程不存在，则执行reload
	return w.mainReload(ctx)
}

// mainReload 主进程-执行 reload 动作
//
// 	1. fork 新子进程
// 	2. stop 旧的子进程
func (w *Worker) mainReload(ctx context.Context) (err error) {
	w.logit("mainReload start")
	defer func() {
		w.logit("mainReload finish, error=", err)
	}()

	if err1 := ctx.Err(); err != nil {
		return err1
	}

	lastSubCancel := w.closeFunc

	w.mux.Lock()
	lastCmd := w.cmd
	w.mux.Unlock()

	ctxN, cancel := context.WithCancel(ctx)
	w.closeFunc = cancel

	// 启动新进程
	if errFork := w.forkAndStart(ctxN); errFork != nil {
		return errFork
	}

	// 优雅关闭老的子进程
	_ = w.stopCmd(ctx, lastCmd)

	if lastSubCancel != nil {
		lastSubCancel()
	}
	return nil
}

// stopCmd 停止指定的cmd
func (w *Worker) stopCmd(ctx context.Context, cmd *exec.Cmd) error {
	if cmd == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, w.getStopTimeout())
	defer cancel()
	return stopCmd(ctx, cmd)
}

func (w *Worker) getStopTimeout() time.Duration {
	if w.option.StopTimeout > 0 {
		return time.Duration(w.option.StopTimeout) * time.Millisecond
	}
	return w.main.Option.GetStopTimeout()
}

func (w *Worker) stop(ctx context.Context) error {
	return nil
}
