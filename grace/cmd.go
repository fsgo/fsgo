// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/31

package grace

import (
	"context"
	"os/exec"
	"syscall"
	"time"
)

func cmdExited(cmd *exec.Cmd) bool {
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return true
	}
	return false
}

// stopCmd 停止指定 cmd
func stopCmd(ctx context.Context, cmd *exec.Cmd) error {
	if cmd == nil {
		return nil
	}

	if cmdExited(cmd) {
		return nil
	}

	// 发送信号给子进程，让其退出
	if err := cmd.Process.Signal(syscall.SIGQUIT); err != nil {
		return err
	}

	if cmdExited(cmd) {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 等待程序优雅退出
	for {
		select {
		case <-ctx.Done():
			if cmdExited(cmd) {
				return nil
			}
			// return cmd.Process.Kill()
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		case <-time.After(5 * time.Millisecond):
			if cmdExited(cmd) {
				return nil
			}

		}
	}
}
