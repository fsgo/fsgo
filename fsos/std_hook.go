// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/10

//go:build !windows
// +build !windows

package fsos

import (
	"syscall"

	"github.com/fsgo/fsgo/fsfs"
)

// HookStderr 劫持标准错误输出
// 	当前程序的 stderr 的内容将输出指定的文件
func HookStderr(f fsfs.HasFd) error {
	return syscall.Dup2(int(f.Fd()), Stderr)
}

// HookStdout 劫持标准输出
// 	当前程序的 stdout 的内容将输出指定的文件
func HookStdout(f fsfs.HasFd) error {
	return syscall.Dup2(int(f.Fd()), Stdout)
}
