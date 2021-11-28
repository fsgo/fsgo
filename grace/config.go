// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/31

package grace

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/fsgo/fsconf"
	"github.com/fsgo/fsenv"
)

// NewSimpleConfig 使用默认配置创建 Config
func NewSimpleConfig() *Config {
	return &Config{
		StatusDir:     filepath.Join(fsenv.AppRootDir(), "var"),
		LogDir:        fsenv.LogRootDir(),
		Keep:          true,
		StopTimeout:   10000, // 10s
		CheckInterval: 5,
	}
}

// Config 配置文件的结构体
type Config struct {
	// StatusDir 必填，状态数据文件目录，如 主进程的 pid 文件都存放在这里
	StatusDir string

	// LogDir 必填，日志文件目录
	// 每个子进程一个子目录
	LogDir string

	// StopTimeout 可选，优雅关闭的最长时间，毫秒，若不填写使用默认值  10000
	StopTimeout int

	// Keep 可选，是否保持子进程一直存在
	Keep bool

	// Workers 可选，工作进程配置
	Workers map[string]*WorkerConfig

	// CheckInterval 可选，检查版本的间隔时间,单位秒，默认为 5 秒
	CheckInterval int
}

// Parser 解析配置
func (c *Config) Parser() error {
	if len(c.Workers) == 0 {
		return fmt.Errorf("empty Workers")
	}
	if c.StatusDir == "" {
		return fmt.Errorf("empty StatusDir")
	}
	if c.LogDir == "" {
		return fmt.Errorf("empty LogDir")
	}

	for name, w := range c.Workers {
		if e := w.Parser(); e != nil {
			return e
		}
		if w.LogDir == "" {
			w.LogDir = filepath.Join(c.LogDir, name)
		}
	}
	return nil
}

// ToOption 转换格式
func (c *Config) ToOption() *Option {
	return &Option{
		StopTimeout:   c.GetStopTimeout(),
		StatusDir:     c.StatusDir,
		LogDir:        c.LogDir,
		Keep:          c.Keep,
		CheckInterval: c.CheckInterval,
	}
}

// GetStopTimeout 获取配置的停止服务的超时时间
func (c *Config) GetStopTimeout() int {
	if c.StopTimeout < 1 {
		return 10 * 1000
	}
	return c.StopTimeout
}

// MustNewWorker 加载配置中指定 name 的 worker
func (c *Config) MustNewWorker(name string) *Worker {
	wc, has := c.Workers[name]
	if !has {
		panic(fmt.Sprintf("worker=%q not exists", name))
	}
	return NewWorker(wc)
}

// NewGrace 通过配置生成 grace server
func (c *Config) NewGrace() *Grace {
	return &Grace{
		Option: c.ToOption(),
	}
}

// LoadConfig 加载主程序的配置文件
func LoadConfig(name string) (*Config, error) {
	var c *Config
	if err := fsconf.Parse(name, &c); err != nil {
		return nil, err
	}
	if err := c.Parser(); err != nil {
		return nil, err
	}
	return c, nil
}

// WorkerConfig worker 的配置
type WorkerConfig struct {
	// Listen 可选，监听的资源，如 "tcp@127.0.0.1:8909",
	Listen []string

	// EnvFile 可选，提前配置 Cmd 的环境变量的文件
	// 在执行 Cmd 前，通过此文件获取env 信息
	// 1.先直接解析该文件，获取 kv，正确的格式为：
	// key1=1
	// key2=2
	// 若文件不是这个格式，则解析失败
	// 2.尝试执行当前文件，若文件输出的  kv 对，则解析成功，否则为失败
	// 若文件行以 # 开头，会当做注释
	// 允许有 0 个 kv 对
	EnvFile string

	// RootDir 可选，执行应用程序的根目录
	// Cmd、Watches 对应的文件路径都是相对于此目录的
	RootDir string

	// LogDir 必填，当前子进程的日志目录
	LogDir string

	// Cmd 必填，工作进程的 cmd
	Cmd string

	// CmdArgs 可选，工作进程 cmd 的其他参数
	CmdArgs []string

	// StopTimeout  优雅关闭的最长时间，毫秒，若不填写，则使用全局 Config 的
	StopTimeout int

	// Watches 可选，用于监听版本变化情况的文件列表
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

func (c *WorkerConfig) getFilePath(p string) string {
	if c.RootDir == "" {
		return p
	}
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(c.RootDir, p)
}

func (c *WorkerConfig) getEnvFilePath() string {
	if c.EnvFile == "" {
		return ""
	}
	return c.getFilePath(c.EnvFile)
}

func (c *WorkerConfig) getWatchFiles() []string {
	var files []string
	fm := map[string]int8{}
	for _, watchPath := range c.Watches {
		pattern := c.getFilePath(watchPath)
		ms, _ := filepath.Glob(pattern)
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

	files := make([]string, 0, 3+len(c.Watches))
	if p := c.getEnvFilePath(); p != "" {
		files = append(files, p)
	}

	files = append(files, cmd)                // 可能是在环境变量里的一些命令
	files = append(files, c.getFilePath(cmd)) // 配置在项目里的文件路径

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
