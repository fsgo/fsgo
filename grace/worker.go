// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/31

package grace

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// WorkerConfig worker 的配置
type WorkerConfig struct {
	// Listen 监听的资源，如 "tcp@127.0.0.1:8909",
	Listen []string

	// EnvFile 提前配置 Cmd 的环境变量的文件，可选
	// 在执行 Cmd 前，通过此文件获取env 信息
	// 1.先直接解析该文件，获取 kv，正确的格式为：
	// key1=1
	// key2=2
	// 若文件不是这个格式，则解析失败
	// 2.尝试执行当前文件，若文件输出的  kv 对，则解析成功，否则为失败
	// 若文件行以 # 开头，会当做注释
	// 允许有 0 个 kv 对
	EnvFile string

	// Cmd 工作进程的 cmd
	Cmd string

	// CmdArgs 工作进程 cmd 的其他参数
	CmdArgs []string

	// StopTimeout StopTimeout 优雅关闭的最长时间，毫秒，若不填写，则使用全局 Config 的
	StopTimeout int

	// Watches 用于监听版本变化情况的文件列表
	Watches []string
}

// Parser 解析当前配置
func (c *WorkerConfig) Parser() error {
	return nil
}

// String 格式化输出，打印输出时时候
func (c *WorkerConfig) String() string {
	bf, _ := json.Marshal(c)
	return string(bf)
}

func statVersion(info os.FileInfo) string {
	var bf bytes.Buffer
	bf.WriteString(info.Mode().String())
	bf.WriteString(info.ModTime().String())
	bf.WriteString(strconv.FormatInt(info.Size(), 10))
	return bf.String()
}

func (c *WorkerConfig) getWatchFiles() []string {
	var files []string
	fm := map[string]int8{}
	for _, w := range c.Watches {
		ms, _ := filepath.Glob(w)
		for _, f := range ms {
			if _, has := fm[f]; !has {
				fm[f] = 1
				files = append(files, f)
			}
		}
	}
	sort.SliceStable(files, func(i, j int) bool {
		return files[i] > files[j]
	})
	return files
}

// version 获取当前的版本信息
func (c *WorkerConfig) version() string {
	cmd, _ := c.getWorkerCmd()

	files := make([]string, 2+len(c.Watches))
	files = append(files, c.EnvFile)
	files = append(files, cmd)
	files = append(files, c.getWatchFiles()...)

	var buf bytes.Buffer
	for _, fn := range files {
		info, err := os.Stat(fn)
		if err == nil {
			buf.WriteString(statVersion(info))
		}
	}
	h := md5.New()
	_, _ = h.Write(buf.Bytes())
	return hex.EncodeToString(h.Sum(nil))
}

func (c *WorkerConfig) getWorkerCmd() (string, []string) {
	if len(c.Cmd) > 0 {
		return c.Cmd, c.CmdArgs
	}
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	return os.Args[0], args
}

// NewWorker 创建一个新的 worker
func NewWorker(c *WorkerConfig) *Worker {
	if c == nil {
		c = &WorkerConfig{}
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
	option    *WorkerConfig
	cmd       *exec.Cmd
	closeFunc context.CancelFunc

	resources []*resourceServer
	mux       sync.Mutex
	sub       *subProcess
	stopped   bool

	event chan string

	// 子进程上次退出时间
	lastExit time.Time
}

// Register 注册新的消费者
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
	go w.watchChange()

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
	msg := fmt.Sprintf("[grace][worker] pid=%d %s", os.Getpid(), fmt.Sprint(msgs...))
	_ = w.main.Logger.Output(1, msg)
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

	var userEnv []string
	if w.option.EnvFile != "" {
		var errParser error
		userEnv, errParser = parserEvnFile(ctx, w.option.EnvFile)
		w.logit(fmt.Sprintf("parserEvnFile(%q)", w.option.EnvFile), ", gotEnv=", userEnv, ", err=", errParser)
		if errParser != nil {
			return fmt.Errorf("parserEvnFile(%q) failed %w", w.option.EnvFile, errParser)
		}
	}

	envs := append(os.Environ(), userEnv...)
	envs = append(envs, envActionKey+"="+actionSubStart)

	cmdName, args := w.option.getWorkerCmd()
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	w.logit("fork new sub_process, cmd=", cmd.String())

	cmd.Env = envs
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

		logFiles := make(map[string]interface{})
		if cmd.Process != nil {
			logFiles["pid"] = cmd.Process.Pid
		}
		cmd.ProcessState.SysUsage()

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

var errNotEnvKV = fmt.Errorf("not env k-v pair")

func parserEvnFile(ctx context.Context, fp string) ([]string, error) {
	bf, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	keyReg := regexp.MustCompile("^[a-zA-Z_][a-zA-Z_0-9]*$")

	parserEnv := func(bf []byte) ([]string, error) {
		var result []string
		lines := bytes.Split(bf, []byte("\n"))
		for _, line := range lines {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			// 以 # 开头的是注释
			if bytes.HasPrefix(line, []byte("#")) {
				continue
			}
			// 不包含 = 说明不是 kv 对
			if !bytes.ContainsAny(line, "=") {
				errParser := fmt.Errorf("line(%q) %w", line, errNotEnvKV)
				return nil, errParser
			}
			arr := strings.SplitN(string(line), "=", 2)
			k := strings.TrimSpace(arr[0])
			v := strings.TrimSpace(arr[1])
			if !keyReg.MatchString(k) {
				errParser := fmt.Errorf("line(%q) %w", line, errNotEnvKV)
				return nil, errParser
			}
			result = append(result, k+"="+v)
		}
		return result, nil
	}

	result, errParser := parserEnv(bf)
	if errParser == nil {
		return result, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, fp)
	cmd.Stderr = os.Stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parserEnv(data)
}
