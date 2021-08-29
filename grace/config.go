// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/31

package grace

import (
	"fmt"
	"path/filepath"

	"github.com/fsgo/fsconf"
)

// Config 配置文件的结构体
type Config struct {
	// StatusDir 状态数据文件目录，如 主进程的 pid 文件都存放在这里
	StatusDir string

	// LogDir 日志文件目录，可选
	// 每个子进程一个子目录
	LogDir string

	// StopTimeout 优雅关闭的最长时间，毫秒，若不填写使用默认值  10000
	StopTimeout int

	// Keep 是否保持子进程一直存在
	Keep bool

	// Workers 工作进程配置
	Workers map[string]*WorkerConfig

	// CheckInterval 检查版本的间隔时间，默认为 5 秒
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
	if len(c.Workers) == 0 {
		return fmt.Errorf("empty Workers")
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
