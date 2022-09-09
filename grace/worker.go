// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/31

package grace

import (
	"context"
	"encoding/json"
	"errors"
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
		option: cfg,
		event:  make(chan string, 1),
	}

	w.sub = &subProcess{
		worker: w,
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

	// 子进程上次退出时间
	lastExit time.Time

	// 用于控制 cmd 子进程的 ctx
	cmdCtx context.Context
	stdout io.Writer

	stderr io.Writer

	// 创建当前 cmd 是对应的 cancel
	cmdClose context.CancelFunc

	sub *subProcess

	event chan string

	main   *Grace
	option *WorkerConfig

	resources []*resourceAndConsumer

	pid int // cmd 对应的 pid

	nextListenDSNIndex int

	mux sync.Mutex

	// 是否正在加载进程
	isReloading bool
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

// start 有主进程调用，worker 开始运行
func (w *Worker) start(ctx context.Context) error {
	start := time.Now()
	w.logit("worker starting ...")
	defer func() {
		dur := time.Since(start)
		w.logit("worker stopped, start_at= ", start.String(), ", duration=", dur.String())
	}()

	var cmdCancel context.CancelFunc
	w.cmdCtx, cmdCancel = context.WithCancel(context.Background())
	defer cmdCancel()

	ctxWatch, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()
	go w.watch(ctxWatch)

	// 启动一个子进程，用于处理请求
	err := w.forkAndStart(w.cmdCtx)
	w.logit("first forkAndStart sub process: ", err)
	if err != nil {
		if IsSubProcess() {
			w.logit("start sub process failed")
			return err
		}
		// 非独立子进程方式，可以忽略错误，以方便后续解除问题后，自动恢复
		w.logit("start sub process failed, it will retry later")
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2, syscall.SIGQUIT)

	var stopped bool

	// hold on
	for {
		select {
		case <-ctx.Done():
			w.logit("receive ctx.Done: ", ctx.Err())
			watchCancel()
			stopped = true
			_ = w.stop(context.Background())
			cmdCancel() // 在 stop 之后，让 cmd 尽量完成优雅退出
			return ctx.Err()

		case sig := <-ch:
			w.logit("receive signal: ", sig)
			switch sig {
			case syscall.SIGINT,
				syscall.SIGQUIT,
				syscall.SIGTERM:
				stopped = true
				watchCancel()
				_ = w.stop(context.Background())
				cmdCancel() // 在 stop 之后，让 cmd 尽量完成优雅退出
				return fmt.Errorf("shutdown by signal(%v)", sig)
			case syscall.SIGUSR2:
				_ = w.reload(w.cmdCtx)
			}

		case e := <-w.event:
			w.logit("receive event:", e)
			switch e {
			case actionKeepSubProcess:
				if !stopped {
					_ = w.keepPrecess(w.cmdCtx)
				}
			}
		}
	}
}

type watchStats struct {
	LastSuc    time.Time
	LastFail   time.Time
	CheckTimes uint64 // 检查总次数
	FailTimes  uint64 // 失败总次数
	SucTimes   uint64 //
}

func (rs *watchStats) String() string {
	bf, _ := json.Marshal(rs)
	return string(bf)
}

func (w *Worker) watch(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	oldVersion := w.option.version()

	dur := w.main.Option.GetCheckInterval()
	tk := time.NewTimer(dur)
	defer tk.Stop()
	st := &watchStats{}

	doCheck := func() bool {
		if err := ctx.Err(); err != nil {
			w.logit("[watch] exit by ctx.Err():", err)
			return false
		}
		st.CheckTimes++

		w.mux.Lock()
		isReloading := w.isReloading
		w.mux.Unlock()

		if isReloading {
			w.logit("[watch] skipped by isReloading")
			return true
		}

		var err error

		newVersion := w.option.version()
		change := oldVersion != newVersion
		pid := w.getLastPID()

		exists := pidExists(pid)

		if !exists || change {
			w.logit("[watch] reload it start...", "pid=", pid, ", exists=", exists, ", version_change=", change, ", ", st.String())
			err = w.reload(w.cmdCtx)
			w.logit("[watch] reload it finish, err=", err)
			if err == nil {
				oldVersion = newVersion
				st.SucTimes++
				st.LastSuc = time.Now()
			} else {
				st.FailTimes++
				st.LastFail = time.Now()
			}
		} else {
			w.logit("[watch] not change, pid=", pid, ", version=", newVersion)
		}
		return true
	}
	for range tk.C {
		if !doCheck() {
			return
		}
		tk.Reset(dur)
	}
}

func (w *Worker) subProcessStart(ctx context.Context) error {
	return w.sub.Start(ctx)
}

func (w *Worker) logit(msgs ...any) {
	w.logitDepth(3, msgs...)
}

func (w *Worker) logitDepth(depth int, msgs ...any) {
	msg := fmt.Sprintf("[grace][worker][%s] %s", w.option.Cmd, fmt.Sprint(msgs...))
	_ = w.main.Logger.Output(depth, msg)
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

	// 解析用户自定义的环境变量文件
	var userEnv []string
	if envFile := w.option.getEnvFilePath(); len(envFile) > 0 {
		var errParser error
		userEnv, errParser = envfile.ParserFile(ctx, envFile)
		w.logit(fmt.Sprintf("parserEvnFile(%q)", w.option.EnvFile), ", gotEnv=", userEnv, ", err=", errParser)
		if errParser != nil {
			return fmt.Errorf("parserEvnFile(%q) failed %w", w.option.EnvFile, errParser)
		}
	}

	envs := append(os.Environ(), userEnv...)
	envs = append(envs, envActionKey+"="+actionSubStart)

	ctx, cancel := context.WithCancel(ctx)
	cmdName, args := w.option.getWorkerCmd()
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = w.option.HomeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	w.logit("fork new sub_process, work_dir=", cmd.Dir, ", cmd=", cmd.String())

	cmd.Env = envs
	// cmd.Stdout = os.Stdout
	cmd.Stdout = w.stdout
	// cmd.Stderr = os.Stderr
	cmd.Stderr = w.stderr
	cmd.ExtraFiles = files
	err := cmd.Start()
	w.logit("cmd.Start, err=", err)
	if err != nil {
		cancel()
		return err
	}

	_ = w.withLock(func() error {
		w.pid = cmd.Process.Pid
		w.cmdClose = cancel
		return nil
	})

	go func() {
		start := time.Now()
		logFields := make(map[string]any)
		errWait := cmd.Wait()
		if cmd.Process != nil {
			logFields["pid"] = cmd.Process.Pid
		}
		cost := time.Since(start)

		w.logit("sub process exit, error=", errWait, ", duration=", cost, ", sub_process_info=", logFields)
		w.event <- actionKeepSubProcess
	}()
	return nil
}

func (w *Worker) withLock(fn func() error) error {
	w.mux.Lock()
	defer w.mux.Unlock()
	return fn()
}

func (w *Worker) subProcessExists() bool {
	pid := w.getLastPID()
	return pidExists(pid)
}

// keepPrecess 检查检查是否存在
func (w *Worker) keepPrecess(ctx context.Context) (err error) {
	w.mux.Lock()
	lastExit := w.lastExit
	isReloading := w.isReloading
	w.mux.Unlock()

	if isReloading {
		w.logit("[keepPrecess] skipped by isReloading")
		return nil
	}

	pid := w.getLastPID()
	if w.subProcessExists() {
		w.logit("[keepPrecess] work process exists, pid=", pid)
		return nil
	}

	w.logit("[keepPrecess] work process not exists, will reload it, pid=", pid)

	// 避免子进程有异常时， 不停重启服务导致 CPU 消耗特别高
	if !lastExit.IsZero() && time.Since(lastExit) < time.Second {
		time.Sleep(time.Second)
	}

	w.mux.Lock()
	w.lastExit = time.Now()
	w.mux.Unlock()

	// 若进程不存在，则执行 reload
	return w.reload(ctx)
}

func (w *Worker) getLastPID() int {
	w.mux.Lock()
	defer w.mux.Unlock()
	return w.pid
}

// reload 执行 reload 动作
// 这个方法都是由 master 进程来调用的
//
//  1. fork 新子进程
//  2. stop 旧的子进程
func (w *Worker) reload(ctx context.Context) (err error) {
	// -----------------------------------------------------------------
	// 添加状态判断，避免多种条件在同时触发 reload
	w.mux.Lock()
	isReloading := w.isReloading
	if !isReloading {
		w.isReloading = true
	}
	w.mux.Unlock()
	if isReloading {
		return errors.New("already in reloading, cannot reload it")
	}

	defer func() {
		w.mux.Lock()
		w.isReloading = false
		w.mux.Unlock()
	}()
	// -----------------------------------------------------------------

	w.logit("start reloading  ...")
	defer func() {
		w.logit("reload finish, error=", err)
	}()

	if err1 := ctx.Err(); err != nil {
		return err1
	}

	lastCmdCancel := w.cmdClose

	lastPID := w.getLastPID()

	// 启动新进程
	if errFork := w.forkAndStart(ctx); errFork != nil {
		return errFork
	}
	newPID := w.getLastPID()

	checkNewPID := func() error {
		if w.subProcessExists() {
			return nil
		}
		w.withLock(func() error {
			w.pid = lastPID
			w.cmdClose = lastCmdCancel
			return nil
		})
		errCheck := fmt.Errorf("new process pid=%d not exists, restore pid=%d", newPID, lastPID)
		w.logit(errCheck.Error())
		return errCheck
	}

	// 启动后的检查时间，只有超过此时间，检查新进程没有问题，才能继续
	tm := time.NewTimer(w.getStartWait())
	defer tm.Stop()

	// 每间隔 0.5 秒检查一次新的进程是否存在，若不存在则退出 reload
	tk := time.NewTicker(500 * time.Millisecond)
	defer tk.Stop()

	for {
		select {
		case <-tm.C:
			break
		case <-tk.C:
			if errCheck := checkNewPID(); errCheck != nil {
				return errCheck
			}
		}
	}

	if errCheck := checkNewPID(); errCheck != nil {
		return errCheck
	}

	// 优雅关闭老的子进程
	err = w.stopCmd(ctx, lastPID)
	w.logit("stop pid=", lastPID, ", err=", err)

	if lastCmdCancel != nil {
		lastCmdCancel()
	}
	return nil
}

// stopCmd 停止指定的cmd
func (w *Worker) stopCmd(ctx context.Context, pid int) error {
	if pid == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, w.getStopTimeout())
	defer cancel()
	return stopCmd(ctx, pid)
}

func (w *Worker) getStopTimeout() time.Duration {
	if t := w.option.getStopTimeout(); t > 0 {
		return t
	}
	return w.main.Option.GetStopTimeout()
}

func (w *Worker) getStartWait() time.Duration {
	if t := w.option.getStartWait(); t > 0 {
		return t
	}
	return w.main.Option.GetStartWait()
}

func (w *Worker) stop(ctx context.Context) error {
	return w.stopCmd(ctx, w.getLastPID())
}

// Resource 将配置的 Listen 的第 index 个 元素解析为可传递使用的 Resource
//
//	如 配置的 "tcp@127.0.0.1:8080" 会解析为 listenDSN
//	若解析失败，panic
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
//
//	若解析失败，panic
func (w *Worker) NextResource() Resource {
	res := w.Resource(w.nextListenDSNIndex)
	w.nextListenDSNIndex++
	return res
}
