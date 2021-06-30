// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/6/30

package tasks

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// TaskHelper task 的一些辅助功能
type TaskHelper struct {
}

// ConfDir 返回任务的配置目录
func (h *TaskHelper) ConfDir() string {
	return ""
}

// ReadConf 从 task 自己的目录里解析配置
func (h *TaskHelper) ReadConf(name string, value interface{}) error {
	fpath := filepath.Join(h.ConfDir(), name)
	bf, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}
	// todo 支持其他类型的文件
	return json.Unmarshal(bf, value)
}

func (h *TaskHelper) DataDir() string {
	return ""
}

func (h *TaskHelper) TmpDir() string {
	return ""
}
