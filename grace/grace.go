/*
 * Copyright(C) 2021 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/1/10
 */

package grace

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	ActionStart  = "start"
	ActionReload = "reload"
	ActionStop   = "stop"

	ActionSubStart = "sub_process_start"
)

const envActionKey = "fsgo_fsnet_grace_action"

// Grace 安全的stop、reload
type Grace struct {
	PIDFilePath     string
	ShutdownTimeout time.Duration

	// Log 日志打印
	Log *log.Logger

	processCmd      *exec.Cmd
	subProcessClose context.CancelFunc

	resources []Resource

	mux sync.Mutex
	sub *subProcess
}

func (g *Grace) Register(res Resource) {
	g.resources = append(g.resources, res)
}

func (g *Grace) init() {
	if g.Log == nil {
		g.Log = log.New(log.Writer(), log.Prefix(), log.Flags())
	}
	if g.ShutdownTimeout < 1 {
		g.ShutdownTimeout = 10 * time.Second
	}

	if g.PIDFilePath == "" {
		panic("PIDFilePath required")
	}

	g.sub = &subProcess{
		resources:       g.resources,
		shutDownTimeout: g.ShutdownTimeout,
		Log:             g.Log,
	}
}

func (g *Grace) logit(msgs ...interface{}) {
	msg := fmt.Sprintf("[grace][main_process] pid=%d %s", os.Getpid(), fmt.Sprint(msgs...))
	g.Log.Output(1, msg)
}

// Start 开始服务，阻塞、同步的
func (g *Grace) Start(ctx context.Context) error {
	g.init()

	action := ActionStart
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	if do := os.Getenv(envActionKey); do != "" {
		action = do
	}
	g.logit("action=", action)

	switch action {
	case ActionStart:
		return g.actionStart(ctx)
	case ActionReload: // 给主进程发送信号
		return g.fireSignal(syscall.SIGUSR2)
	case ActionStop:
		return g.fireSignal(syscall.SIGQUIT) // 给主进程发送 退出信号
	case ActionSubStart: // 子进程:启动
		return g.sub.Start(ctx)
	default:
		return fmt.Errorf("not support action %q", action)
	}
}

func (g *Grace) fireSignal(sig os.Signal) error {
	bf, err := ioutil.ReadFile(g.PIDFilePath)
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(string(bf))
	if err != nil {
		return err
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	return p.Signal(sig)
}

func (g *Grace) actionStart(ctx context.Context) error {
	defer func() {
		if g.PIDFilePath != "" {
			os.Remove(g.PIDFilePath)
		}
	}()
	pidStr := strconv.Itoa(os.Getpid())
	if err := ioutil.WriteFile(g.PIDFilePath, []byte(pidStr), 0644); err != nil {
		return err
	}

	if len(g.resources) == 0 {
		return errors.New("no server to start")
	}
	return g.mainStart(ctx, g.resources)
}

// mainStart 主进程开启开始
func (g *Grace) mainStart(ctx context.Context, servers []Resource) error {
	for _, info := range servers {
		err := info.Open(ctx)
		g.logit("open resource ", info.String(), ", error=", err)
		if err != nil {
			return err
		}
	}

	ctxFork, cancel := context.WithCancel(ctx)
	defer cancel()
	g.subProcessClose = cancel

	// 启动一个子进程，用于处理请求
	if err := g.forkAndStart(ctxFork); err != nil {
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
			switch sig {
			case syscall.SIGINT,
				syscall.SIGQUIT,
				syscall.SIGTERM:
				g.actionStop(context.Background(), g.processCmd)
				return fmt.Errorf("shutdown by signal(%v)", sig)
			case syscall.SIGUSR2:
				g.actionReload(context.Background())
			}
		}
	}
	return fmt.Errorf("shutdown")
}

// actionReload 主进程-执行reload 动作
//
// 	1. fork 新子进程
// 	2. stop 旧的子进程
func (g *Grace) actionReload(ctx context.Context) (err error) {
	g.logit("actionReload start")
	defer func() {
		g.logit("actionReload finish, error=", err)
	}()

	lastSubCancel := g.subProcessClose

	g.mux.Lock()
	lastCmd := g.processCmd
	g.mux.Unlock()

	ctxN, cancel := context.WithCancel(ctx)
	g.subProcessClose = cancel
	if errFork := g.forkAndStart(ctxN); errFork != nil {
		return errFork
	}

	// 优雅 关闭老的子进程
	_ = g.actionStop(ctx, lastCmd)

	if lastSubCancel != nil {
		lastSubCancel()
	}
	return nil
}

func (g *Grace) withLock(fn func() error) error {
	g.mux.Lock()
	defer g.mux.Unlock()
	return fn()
}

func (g *Grace) subProcessExited(cmd *exec.Cmd) bool {
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return true
	}
	return false
}

func (g *Grace) actionStop(ctx context.Context, cmd *exec.Cmd) error {
	if cmd == nil {
		return nil
	}

	if cmd == nil {
		return fmt.Errorf("no subprecess need shutdown")
	}

	if g.subProcessExited(cmd) {
		return nil
	}

	// 发送信号给子进程，让其退出
	if err := cmd.Process.Signal(syscall.SIGQUIT); err != nil {
		return err
	}

	if g.subProcessExited(cmd) {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, g.ShutdownTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			if g.subProcessExited(cmd) {
				return nil
			}
			return cmd.Process.Kill()
		case <-time.After(50 * time.Millisecond):
			if g.subProcessExited(cmd) {
				return nil
			}

		}
	}

	return nil
}

func (g *Grace) forkAndStart(ctx context.Context) (ret error) {
	files := make([]*os.File, len(g.resources))
	for idx, s := range g.resources {
		f, err := s.File()
		if err != nil {
			return fmt.Errorf("listener[%d].File() has error: %w", idx, err)
		}
		files[idx] = f
	}
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	cmd := exec.CommandContext(ctx, os.Args[0], args...)

	g.logit("fork new sub_process, cmd=", cmd.String())

	cmd.Env = append(os.Environ(), envActionKey+"="+ActionSubStart)
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
		g.logit("cmd.Wait, error=", errWait, ",cost=", cost)
	}()

	_ = g.withLock(func() error {
		g.processCmd = cmd
		return nil
	})

	return nil
}
