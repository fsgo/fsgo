// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/10

package grace

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	actionStart = "start"

	actionReload = "reload"

	actionStop = "stop"

	actionSubStart = "sub_process_start"

	// 一个特殊的 event，当子进程退出了，需要创建新进程的时候，发送此信号
	actionKeepSubProcess = "keep_sub_process"
)

const envActionKey = "fsgo_grace_action"

// Option  grace 的配置选项
type Option struct {
	StatusDir string

	LogDir string

	// StopTimeout 子进程优雅退出的超时时间
	StopTimeout time.Duration

	// 检查版本的间隔时间，默认为 5 秒
	CheckInterval time.Duration

	// StartWait 可选，启动新进程后，老进程退出前的等待时间,默认为 3 秒
	StartWait time.Duration

	// Keep 是否保持子进程存活
	// 若为 true，当子进程不存在时，将自动拉起
	Keep bool
}

// Parser 参数解析、检查
func (c *Option) Parser() error {
	if c.StatusDir == "" {
		return errors.New("empty StatusDir")
	}
	return nil
}

// GetStopTimeout 获取停止超时时间
func (c *Option) GetStopTimeout() time.Duration {
	if c.StopTimeout < 1 {
		return 10 * time.Second
	}
	return c.StopTimeout
}

func (c *Option) GetStartWait() time.Duration {
	if c.StartWait > 0 {
		return c.StartWait
	}
	return 5 * time.Second
}

// GetMainPIDPath 获取主程序的 PID 文件路径
func (c *Option) GetMainPIDPath() string {
	return filepath.Join(c.StatusDir, "main.pid")
}

// GetCheckInterval 获取检查的时间间隔
func (c *Option) GetCheckInterval() time.Duration {
	if c.CheckInterval > 0 {
		return c.CheckInterval
	}
	return 5 * time.Second
}

// Grace 安全的 stop、reload
type Grace struct {
	Option *Option

	// Logger logger，若为空，
	// 将使用默认的使用标准库的log.Writer作为输出
	Logger *log.Logger

	workers map[string]*Worker
}

// Register 注册一个新的 worker
func (g *Grace) Register(name string, gg *Worker) error {
	g.init()
	_, has := g.workers[name]
	if has {
		return fmt.Errorf("worker=%q already exists", name)
	}
	gg.main = g
	g.workers[name] = gg
	return nil
}

// MustRegister 注册，若失败会 panic
func (g *Grace) MustRegister(name string, gg *Worker) {
	err := g.Register(name, gg)
	if err != nil {
		panic("register " + name + " failed, err=" + err.Error())
	}
}

func (g *Grace) init() {
	if g.Logger == nil {
		g.Logger = log.Default()
	}

	if err := g.Option.Parser(); err != nil {
		panic(err.Error())
	}
	if g.workers == nil {
		g.workers = make(map[string]*Worker)
	}
}

func (g *Grace) logit(msgs ...any) {
	msg := fmt.Sprintf("[grace][main] %s", fmt.Sprint(msgs...))
	_ = g.Logger.Output(2, msg)
}

// Start 开始服务，阻塞、同步的
func (g *Grace) Start(ctx context.Context) (err error) {
	startTime := time.Now()
	defer func() {
		g.logit("grace.Start() exit, err=", err, ", cost=", time.Since(startTime))
	}()
	g.init()
	action := actionStart
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	if do := os.Getenv(envActionKey); do != "" {
		action = do
	}

	switch action {
	case actionStart,
		actionReload,
		actionStop,
		actionSubStart:
	default:
		// 可能是其他的 参数，如 -conf app.toml
		action = actionStart
	}

	g.logit("action=", action)

	switch action {
	case actionStart:
		return g.actionMainStart(ctx)
	case actionReload: // 给主进程发送信号
		return g.fireSignal(syscall.SIGUSR2)
	case actionStop:
		return g.actionReceiveStop()
	case actionSubStart: // 子进程:启动
		return g.startWorkerProcess(context.Background())
	default:
		return fmt.Errorf("not support action %q", action)
	}
}

// mainProcess 找到主进程
func (g *Grace) mainProcess() (*os.Process, error) {
	bf, err := ioutil.ReadFile(g.Option.GetMainPIDPath())
	if err != nil {
		return nil, err
	}
	pid, err := strconv.Atoi(string(bf))
	if err != nil {
		return nil, err
	}
	return os.FindProcess(pid)
}

// fireSignal 给主进程发送信号
func (g *Grace) fireSignal(sig os.Signal) error {
	p, err := g.mainProcess()
	if err != nil {
		return err
	}
	err = p.Signal(sig)
	g.logit("fireSignal to master, master pid=", p.Pid, ", Signal=", sig, ", err=", err)
	return err
}

// actionReceiveStop 给主进程 发送 stop 信号，让主进程和子进程都退出
func (g *Grace) actionReceiveStop() error {
	p, err := g.mainProcess()
	if err != nil {
		return err
	}

	// 给主进程发送 退出信号
	if err = syscall.Kill(-p.Pid, syscall.SIGQUIT); err != nil {
		return err
	}

	g.logit("waiting main_process mainStop, main pid= ", p.Pid)

	ctx, cancel := context.WithTimeout(context.Background(), g.Option.GetStopTimeout()*2)
	defer cancel()

	// 等待主进程程序
	for {
		select {
		case <-ctx.Done():
			return syscall.Kill(-p.Pid, syscall.SIGKILL)
		case <-time.After(5 * time.Millisecond):
			_, err1 := g.mainProcess()
			if err1 != nil {
				return nil
			}
		}
	}
}

func (g *Grace) writePIDFile() error {
	pidPath := g.Option.GetMainPIDPath()
	pidStr := strconv.Itoa(os.Getpid())
	if err := keepDir(filepath.Dir(pidPath)); err != nil {
		return err
	}
	err := ioutil.WriteFile(pidPath, []byte(pidStr), 0644)
	return err
}

func (g *Grace) actionMainStart(ctx context.Context) error {
	defer func() {
		os.Remove(g.Option.GetMainPIDPath())
	}()

	wd, err := os.Getwd()
	g.logit("[grace][master] working dir=", wd, err)
	if err != nil {
		return err
	}

	if e := g.writePIDFile(); e != nil {
		return e
	}
	if len(g.workers) == 0 {
		return errors.New("no workers to start")
	}
	go g.watchMainPid()

	return g.mainStart(ctx)
}

func (g *Grace) watchMainPid() {
	pidPath := g.Option.GetMainPIDPath()
	info, _ := os.Stat(pidPath)
	last := info.ModTime()

	tk := time.NewTicker(1 * time.Second)
	defer tk.Stop()

	for range tk.C {
		info1, err := os.Stat(pidPath)
		if err != nil {
			g.logit(fmt.Sprintf("read MainPIDPath=%q stat failed, err=%v", pidPath, err))
			if os.IsNotExist(err) {
				err2 := g.writePIDFile()
				g.logit("create MainPIDPath:", err2)
			} else {
				panic("read pid status failed, error=" + err.Error())
			}
			continue
		}

		current := info1.ModTime()
		if !last.Equal(current) {
			last = current
			_ = g.keepSubProcess(context.Background())
		}
	}
}

func (g *Grace) workersDo(fn func(w *Worker) error) error {
	var wg errgroup.Group
	for _, w := range g.workers {
		w := w
		wg.Go(func() error {
			return fn(w)
		})
	}
	return wg.Wait()
}

// mainStart 主进程开启开始
func (g *Grace) mainStart(ctx context.Context) error {
	return g.workersDo(func(w *Worker) error {
		return w.start(ctx)
	})
}

func (g *Grace) keepSubProcess(ctx context.Context) (err error) {
	return g.workersDo(func(w *Worker) error {
		return w.reload(ctx)
	})
}

func (g *Grace) startWorkerProcess(ctx context.Context) error {
	return g.workersDo(func(w *Worker) error {
		return w.subProcessStart(ctx)
	})
}

// IsSubProcess 是否子进程运行模式
func IsSubProcess() bool {
	return len(os.Getenv(envActionKey)) > 0
}
