// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/31

package grace

import (
	"context"
	"syscall"
	"time"
)

//
// https://man7.org/linux/man-pages/man2/kill.2.html
// If sig is 0, then no signal is sent, but existence and permission
// checks are still performed; this can be used to check for the
// existence of a process ID or process worker ID that the caller is
// permitted to signal.
func pidExists(pid int) bool {
	if pid == 0 {
		return false
	}
	return syscall.Kill(pid, syscall.Signal(0)) == nil
}

// stopCmd 停止指定 cmd
func stopCmd(ctx context.Context, pid int) error {
	if pid == 0 {
		return nil
	}
	if !pidExists(pid) {
		return nil
	}

	// 发送信号给子进程，让其退出
	if err := syscall.Kill(-pid, syscall.SIGQUIT); err != nil {
		return err
	}

	if !pidExists(pid) {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 等待程序优雅退出
	for {
		select {
		case <-ctx.Done():
			if !pidExists(pid) {
				return nil
			}
			return syscall.Kill(-pid, syscall.SIGKILL)
		case <-time.After(5 * time.Millisecond):
			if !pidExists(pid) {
				return nil
			}

		}
	}
}
