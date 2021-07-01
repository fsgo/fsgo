// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/31

package grace

import (
	"fmt"
)

// Config 配置文件的结构体
type Config struct {
	// StatusDir 状态数据文件目录，如 主进程的 pid 文件都存放在这里
	StatusDir string

	// StopTimeout 优雅关闭的最长时间，毫秒，若不填写使用默认值  10000
	StopTimeout int

	// Keep 是否保持子进程一直存在
	Keep bool

	// Workers 工作进程配置
	Workers map[string]*ConfigWorker

	// CheckInterval 检查版本的间隔时间，默认为 5 秒
	CheckInterval int
}

func (c *Config) Parser() error {
	if len(c.Workers) == 0 {
		return fmt.Errorf("empty Workers")
	}
	if c.StatusDir == "" {
		return fmt.Errorf("empty StatusDir")
	}
	if len(c.Workers) == 0 {
		return fmt.Errorf("empty Workers")
	}
	for _, w := range c.Workers {
		if e := w.Parser(); e != nil {
			return e
		}
	}
	return nil
}

func (c *Config) ToOption() *Option {
	return &Option{
		StopTimeout:   c.GetStopTimeout(),
		StatusDir:     c.StatusDir,
		Keep:          c.Keep,
		CheckInterval: c.CheckInterval,
	}
}

func (c *Config) GetStopTimeout() int {
	if c.StopTimeout < 1 {
		return 10 * 1000
	}
	return c.StopTimeout
}

type ConfigWorker struct {
	// Listen 监听的资源，如 "tcp@127.0.0.1:8909",
	Listen []string

	// Cmd 工作进程的 cmd
	Cmd string

	// CmdArgs 工作进程 cmd 的其他参数
	CmdArgs []string

	// StopTimeout StopTimeout 优雅关闭的最长时间，毫秒，若不填写，则使用全局 Config 的
	StopTimeout int

	// Watches 用于监听版本变化情况的文件列表
	Watches []string
}

func (c *ConfigWorker) Parser() error {
	return nil
}

func (c *ConfigWorker) ToWorkerOption() *WorkerOption {
	return &WorkerOption{
		Cmd:         c.Cmd,
		CmdArgs:     c.CmdArgs,
		StopTimeout: c.StopTimeout,
		// VersionFile: c.VersionFile,
		Watches: c.Watches,
	}
}
