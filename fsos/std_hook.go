// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/10

// +build !windows

package fsos

import (
	"syscall"
)

func HookStderr(f HasFd) error {
	return syscall.Dup2(int(f.Fd()), Stderr)
}

func HookStdout(f HasFd) error {
	return syscall.Dup2(int(f.Fd()), Stderr)
}
