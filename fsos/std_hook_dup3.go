// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/23

//go:build (linux && arm64) || (linux && loong64)

package fsos

import (
	"syscall"

	"github.com/fsgo/fsgo/fsfs"
)

// HookStderr 劫持标准错误输出
//
//	当前程序的 stderr 的内容将输出指定的文件
func HookStderr(f fsfs.HasFd) error {
	return syscall.Dup3(int(f.Fd()), Stderr, 0)
}

// HookStdout 劫持标准输出
//
//	当前程序的 stdout 的内容将输出指定的文件
func HookStdout(f fsfs.HasFd) error {
	return syscall.Dup3(int(f.Fd()), Stdout, 0)
}
