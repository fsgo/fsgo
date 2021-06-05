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
	actionStart  = "start"
	actionReload = "reload"
	actionStop   = "stop"

	actionSubStart = "sub_process_start"

	actionSubProcessExit = "sub_process_exit"
)

const envActionKey = "fsgo_fsnet_grace_action"

// Option  grace 的配置选项
type Option struct {
	StopTimeout int
	StatusDir   string
	Keep        bool

	// 检查版本的间隔时间，默认为 5 秒
	CheckInterval int
}

// Parser 参数解析、检查
func (c *Option) Parser() error {
	if c.StatusDir == "" {
		return fmt.Errorf("empty StatusDir")
	}
	return nil
}

func (c *Option) GetStopTimeout() time.Duration {
	if c.StopTimeout < 1 {
		return 10 * time.Second
	}
	return time.Duration(c.StopTimeout) * time.Millisecond
}

func (c *Option) GetMainPIDPath() string {
	return filepath.Join(c.StatusDir, "main.pid")
}

func (c *Option) GetCheckInterval() time.Duration {
	if c.CheckInterval > 0 {
		return time.Duration(c.CheckInterval) * time.Second
	}
	return 5 * time.Second
}

// Grace 安全的 stop、reload
type Grace struct {
	Option *Option

	// Logger logger，若为空，
	// 将使用默认的使用标准库的log.Writer作为输出
	Logger *log.Logger

	groups map[string]*Worker
}

func (g *Grace) Register(name string, gg *Worker) error {
	g.init()

	_, has := g.groups[name]
	if has {
		return fmt.Errorf("group=%q already exists", name)
	}
	gg.main = g
	g.groups[name] = gg
	return nil
}

func (g *Grace) init() {
	if g.Logger == nil {
		g.Logger = log.New(log.Writer(), log.Prefix(), log.Flags())
	}
	if err := g.Option.Parser(); err != nil {
		panic(err.Error())
	}
	if g.groups == nil {
		g.groups = make(map[string]*Worker)
	}
}

func (g *Grace) logit(msgs ...interface{}) {
	msg := fmt.Sprintf("[grace][main] pid=%d %s", os.Getpid(), fmt.Sprint(msgs...))
	_ = g.Logger.Output(1, msg)
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

func (g *Grace) fireSignal(sig os.Signal) error {
	p, err := g.mainProcess()
	if err != nil {
		return nil
	}
	return p.Signal(sig)
}

// actionReceiveStop 发送 sotp 信号
func (g *Grace) actionReceiveStop() error {
	p, err := g.mainProcess()
	if err != nil {
		return err
	}

	// 给主进程发送 退出信号
	if err := p.Signal(syscall.SIGQUIT); err != nil {
		return err
	}

	g.logit("waiting main_process mainStop, main pid= ", p.Pid)

	ctx, cancel := context.WithTimeout(context.Background(), g.Option.GetStopTimeout()*2)
	defer cancel()

	// 等待主进程程序
	for {
		select {
		case <-ctx.Done():
			return p.Kill()
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
	if err := ioutil.WriteFile(pidPath, []byte(pidStr), 0644); err != nil {
		return err
	}
	return nil
}

func (g *Grace) actionMainStart(ctx context.Context) error {
	defer func() {
		os.Remove(g.Option.GetMainPIDPath())
	}()

	if e := g.writePIDFile(); e != nil {
		return e
	}

	if len(g.groups) == 0 {
		return errors.New("no groups to start")
	}
	go g.watchMainPid()

	return g.mainStart(ctx)
}

func (g *Grace) watchMainPid() {
	pidPath := g.Option.GetMainPIDPath()
	info, _ := os.Stat(pidPath)
	last := info.ModTime()
	tk := time.NewTicker(1 * time.Second)

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
	for _, w := range g.groups {
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
		return w.mainStart(ctx)
	})
}

func (g *Grace) keepSubProcess(ctx context.Context) (err error) {
	return g.workersDo(func(w *Worker) error {
		return w.mainReload(ctx)
	})
}

func (g *Grace) startWorkerProcess(ctx context.Context) error {
	return g.workersDo(func(w *Worker) error {
		return w.subProcessStart(ctx)
	})
}
