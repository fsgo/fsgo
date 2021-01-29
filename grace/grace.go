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
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	actionStart  = "start"
	actionReload = "reload"
	actionStop   = "stop"

	actionSubStart = "sub_process_start"
)

const envActionKey = "fsgo_fsnet_grace_action"

// Grace 安全的stop、reload
type Grace struct {
	// PIDFilePath 保存pid的文件路径
	PIDFilePath string

	// SubProcessCommand 子进程的命令，可选
	// 若为空则使用当前进程的命令
	// 如 []string{"./bin/user","-conf","conf/app.toml"}
	SubProcessCommand []string

	// StopTimeout 停止服务的超时时间，默认10s
	StopTimeout time.Duration

	// Keep 是否保持子进程常在，默认 false
	// 若为true，当子进程退出后将立即重新启动新的进程
	Keep bool

	// Log logger，若为空，
	// 将使用默认的使用标准库的log.Writer作为输出
	Log *log.Logger

	processCmd      *exec.Cmd
	subProcessClose context.CancelFunc

	resources []*resourceServer

	mux sync.Mutex
	sub *subProcess

	event   chan string
	stopped bool
}

// Register 注册资源
func (g *Grace) Register(res Resource, c Consumer) error {
	ss := &resourceServer{
		Resource: res,
		Consumer: c,
	}
	if c != nil {
		c.Bind(res)
	}
	g.resources = append(g.resources, ss)
	return nil
}

func (g *Grace) RegisterByDSN(dsn string, c Consumer) error {
	res, err := GenResourceByDSN(dsn)
	if err != nil {
		return err
	}
	return g.Register(res, c)
}

func (g *Grace) init() {
	if g.Log == nil {
		g.Log = log.New(log.Writer(), log.Prefix(), log.Flags())
	}
	if g.StopTimeout < 1 {
		g.StopTimeout = 10 * time.Second
	}

	g.event = make(chan string, 1)

	if g.PIDFilePath == "" {
		panic("PIDFilePath required")
	}

	g.sub = &subProcess{
		resources:       g.resources,
		shutDownTimeout: g.StopTimeout,
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

	action := actionStart
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	if do := os.Getenv(envActionKey); do != "" {
		action = do
	}
	g.logit("action=", action)

	switch action {
	case actionStart:
		return g.actionStart(ctx)
	case actionReload: // 给主进程发送信号
		return g.fireSignal(syscall.SIGUSR2)
	case actionStop:
		return g.fireSignal(syscall.SIGQUIT) // 给主进程发送 退出信号
	case actionSubStart: // 子进程:启动
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
		os.Remove(g.PIDFilePath)
	}()
	pidStr := strconv.Itoa(os.Getpid())
	if err := keepDir(filepath.Dir(g.PIDFilePath)); err != nil {
		return err
	}
	if err := ioutil.WriteFile(g.PIDFilePath, []byte(pidStr), 0644); err != nil {
		return err
	}

	if len(g.resources) == 0 {
		return errors.New("no server to start")
	}
	return g.mainStart(ctx, g.resources)
}

// mainStart 主进程开启开始
func (g *Grace) mainStart(ctx context.Context, servers []*resourceServer) error {
	for _, info := range servers {
		err := info.Resource.Open(ctx)
		g.logit("open resource ", info.Resource.String(), ", error=", err)
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
				g.stopped = true
				g.actionStop(context.Background(), g.processCmd)
				return fmt.Errorf("shutdown by signal(%v)", sig)
			case syscall.SIGUSR2:
				g.keepSubProcess(context.Background())
			}
		case e := <-g.event:
			switch e {
			case actionSubStart:
				if !g.stopped {
					g.keepSubProcess(context.Background())
				}
			}
		}
	}
	return fmt.Errorf("shutdown")
}

// keepSubProcess 主进程-执行reload 动作
//
// 	1. fork 新子进程
// 	2. stop 旧的子进程
func (g *Grace) keepSubProcess(ctx context.Context) (err error) {
	g.logit("keepSubProcess start")
	defer func() {
		g.logit("keepSubProcess finish, error=", err)
	}()

	if err1 := ctx.Err(); err != nil {
		return err1
	}

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

// actionStop 停止服务
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

	ctx, cancel := context.WithTimeout(ctx, g.StopTimeout)
	defer cancel()

	// 等待程序优雅退出
	for {
		select {
		case <-ctx.Done():
			if g.subProcessExited(cmd) {
				return nil
			}
			return cmd.Process.Kill()
		case <-time.After(5 * time.Millisecond):
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
		f, err := s.Resource.File()
		if err != nil {
			return fmt.Errorf("listener[%d].File() has error: %w", idx, err)
		}
		files[idx] = f
	}

	cmdName, args := g.getSubProcessCmds()
	cmd := exec.CommandContext(ctx, cmdName, args...)

	g.logit("fork new sub_process, cmd=", cmd.String())

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
		g.logit("cmd.Wait, error=", errWait, ", duration=", cost)
		if g.Keep {
			g.event <- actionSubStart
		}
	}()

	_ = g.withLock(func() error {
		g.processCmd = cmd
		return nil
	})

	return nil
}

func (g *Grace) getSubProcessCmds() (string, []string) {
	if len(g.SubProcessCommand) > 0 {
		return g.SubProcessCommand[0], g.SubProcessCommand[1:]
	}
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	return os.Args[0], args
}
