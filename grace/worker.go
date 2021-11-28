// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/31

package grace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/fsgo/fsgo/grace/internal/envfile"
)

// NewWorker 创建一个新的 worker
func NewWorker(cfg *WorkerConfig) *Worker {
	if cfg == nil {
		cfg = &WorkerConfig{}
	}
	w := &Worker{
		option:  cfg,
		stopped: false,
		event:   make(chan string, 1),
	}

	w.sub = &subProcess{
		group: w,
	}

	stderr := &fsfs.Rotator{
		Path:    filepath.Join(cfg.LogDir, "stderr.log"),
		ExtRule: "1hour",
	}
	_ = stderr.Init()
	w.stderr = io.MultiWriter(os.Stderr, stderr)

	stdout := &fsfs.Rotator{
		Path:    filepath.Join(cfg.LogDir, "stdout.log"),
		ExtRule: "1hour",
	}
	_ = stdout.Init()
	w.stdout = io.MultiWriter(os.Stdout, stdout)

	return w
}

type resourceAndConsumer struct {
	Resource Resource
	Consumer Consumer
}

// Worker 工作进程的逻辑
type Worker struct {
	main      *Grace
	option    *WorkerConfig
	cmd       *exec.Cmd
	closeFunc context.CancelFunc

	resources []*resourceAndConsumer
	mux       sync.Mutex
	sub       *subProcess
	stopped   bool

	event chan string

	// 子进程上次退出时间
	lastExit time.Time

	stderr io.Writer
	stdout io.Writer

	nextListenDSNIndex int
}

// Register 注册新的消费者
func (w *Worker) Register(c Consumer, res Resource) error {
	return w.register(c, res)
}

// MustRegister 注册，若失败会 panic
func (w *Worker) MustRegister(c Consumer, res Resource) {
	err := w.Register(c, res)
	if err != nil {
		panic("register failed: " + err.Error())
	}
}

// RegisterServer 注册/绑定一个 server
func (w *Worker) RegisterServer(ser Server, res Resource) error {
	c := NewServerConsumer(ser, res)
	return w.Register(c, res)
}

// MustRegisterServer 注册一个 server，若失败会 panic
func (w *Worker) MustRegisterServer(ser Server, res Resource) {
	c := NewServerConsumer(ser, res)
	w.MustRegister(c, res)
}

// register 注册资源
func (w *Worker) register(c Consumer, res Resource) error {
	ss := &resourceAndConsumer{
		Consumer: c,
		Resource: res,
	}
	w.resources = append(w.resources, ss)
	return nil
}

// mainStart 主进程开启开始
func (w *Worker) mainStart(ctx context.Context) error {
	go w.watchChange()

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
			w.stopped = true
			_ = w.stop(context.Background())
			return ctx.Err()
		case sig := <-ch:
			w.logit(fmt.Sprintf("receive signal(%v)", sig))
			switch sig {
			case syscall.SIGINT,
				syscall.SIGQUIT,
				syscall.SIGTERM:
				w.stopped = true
				_ = w.stop(context.Background())
				return fmt.Errorf("shutdown by signal(%v)", sig)
			case syscall.SIGUSR2:
				_ = w.mainReload(context.Background())
			}
		case e := <-w.event:
			switch e {
			case actionSubProcessExit:
				if !w.stopped {
					_ = w.keepPrecess(context.Background())
				}
			}
		}
	}
}

type watchReloadStats struct {
	CheckTimes uint64
	FailTimes  uint64
	SucTimes   uint64
	LastSuc    time.Time
	LastFail   time.Time
}

func (rs *watchReloadStats) String() string {
	bf, _ := json.Marshal(rs)
	return string(bf)
}

func (w *Worker) watchChange() {
	oldVersion := w.option.version()

	dur := w.main.Option.GetCheckInterval()
	tk := time.NewTimer(dur)

	st := &watchReloadStats{}

	defer tk.Stop()
	for range tk.C {
		st.CheckTimes++
		newVersion := w.option.version()
		change := oldVersion != newVersion
		// w.logit("watchChange ",change,oldVersion,newVersion)
		if change {
			w.logit("version changed, reload it start...", st.String())
			err := w.mainReload(context.Background())
			w.logit("version changed, reload it finish, err=", err)
			if err == nil {
				oldVersion = newVersion
				st.SucTimes++
				st.LastSuc = time.Now()
			} else {
				st.FailTimes++
				st.LastFail = time.Now()
			}
		}
		tk.Reset(dur)
	}
}

func (w *Worker) subProcessStart(ctx context.Context) error {
	return w.sub.Start(ctx)
}

func (w *Worker) logit(msgs ...interface{}) {
	msg := fmt.Sprintf("[grace][worker] %s", fmt.Sprint(msgs...))
	_ = w.main.Logger.Output(2, msg)
}

func (w *Worker) forkAndStart(ctx context.Context) (ret error) {
	files := make([]*os.File, len(w.resources))
	// 依次获取 *os.File,之后将通过 进程的 ExtraFiles 属性传递给子进程
	for idx, s := range w.resources {
		f, err := s.Resource.File(ctx)
		if err != nil {
			return fmt.Errorf("listener[%d].File() has error: %w", idx, err)
		}
		if f == nil {
			return fmt.Errorf("listener[%d].File(), got nil file", idx)
		}
		w.logit("open resource File ", s.Resource.String(), " success, index=", idx, f, f == nil)
		files[idx] = f
	}

	var userEnv []string
	if envFile := w.option.getEnvFilePath(); envFile != "" {
		var errParser error
		userEnv, errParser = envfile.ParserFile(ctx, envFile)
		w.logit(fmt.Sprintf("parserEvnFile(%q)", w.option.EnvFile), ", gotEnv=", userEnv, ", err=", errParser)
		if errParser != nil {
			return fmt.Errorf("parserEvnFile(%q) failed %w", w.option.EnvFile, errParser)
		}
	}

	envs := append(os.Environ(), userEnv...)
	envs = append(envs, envActionKey+"="+actionSubStart)

	cmdName, args := w.option.getWorkerCmd()
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = w.option.RootDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	w.logit("fork new sub_process, work_dir=", cmd.Dir, ", cmd=", cmd.String())

	cmd.Env = envs
	// cmd.Stdout = os.Stdout
	cmd.Stdout = w.stdout
	// cmd.Stderr = os.Stderr
	cmd.Stderr = w.stderr
	cmd.ExtraFiles = files
	err := cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		start := time.Now()
		errWait := cmd.Wait()
		cost := time.Since(start)

		logFiles := make(map[string]interface{})
		if cmd.Process != nil {
			logFiles["pid"] = cmd.Process.Pid
		}

		w.logit("cmd.Wait, error=", errWait, ", duration=", cost, ", sub_process_info=", logFiles)
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
	lastExit := w.lastExit
	w.mux.Unlock()
	if !cmdExited(lastCmd) {
		return nil
	}

	// 避免子进程 不停重启服务导致 CPU 消耗特别高
	if !lastExit.IsZero() && time.Since(lastExit) < time.Second {
		time.Sleep(time.Second)
	}

	w.mux.Lock()
	w.lastExit = time.Now()
	w.mux.Unlock()

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
	w.mux.Lock()
	lastCmd := w.cmd
	w.mux.Unlock()
	return w.stopCmd(ctx, lastCmd)
}

// Resource 将配置的 Listen 的第 index 个 元素解析为可传递使用的 Resource
// 	如 配置的 "tcp@127.0.0.1:8080" 会解析为 listenDSN
// 	若解析失败，panic
func (w *Worker) Resource(index int) Resource {
	if index < 0 || index >= len(w.option.Listen) {
		panic(fmt.Sprintf("invalid index %d, should in [0,%d]", index, len(w.option.Listen)-1))
	}
	res, err := ParserListenDSN(index, w.option.Listen[index])
	if err != nil {
		panic("parser dsn failed: " + err.Error())
	}
	return res
}

// NextResource 自动解析配置的 Listen 的下一个元素为 Resource
// 	若解析失败，panic
func (w *Worker) NextResource() Resource {
	res := w.Resource(w.nextListenDSNIndex)
	w.nextListenDSNIndex++
	return res
}
