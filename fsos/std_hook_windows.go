// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/10

package fsos

import (
	"syscall"
)

// HookStderr hook stderr
// now work well for panic and print、println.
//
// not work for log.x and os.Stderr.
// you should:
//
//	os.Stderr=yourFile
//	log.SetOutput(yourFile)
func HookStderr(f HasFd) error {
	return hookStd(syscall.STD_ERROR_HANDLE, f.Fd())
}

// HookStdout hook stdout
func HookStdout(f HasFd) error {
	return hookStd(syscall.STD_OUTPUT_HANDLE, f.Fd())
}

var (
	kernel32   = syscall.NewLazyDLL("kernel32.dll")
	winHandler = kernel32.NewProc("SetStdHandle")
)

func hookStd(stdHandle int, fd uintptr) error {
	// see https://docs.microsoft.com/en-us/windows/console/setstdhandle
	// BOOL WINAPI SetStdHandle(
	//  _In_ DWORD  nStdHandle,
	//  _In_ HANDLE hHandle
	// );
	// If the function succeeds, the return value is nonzero.
	// If the function fails, the return value is zero
	r0, _, errno := syscall.Syscall(winHandler.Addr(), 2, uintptr(stdHandle), fd, 0)
	if r0 == 0 {
		if errno != 0 {
			return error(errno)
		}
		return syscall.EINVAL
	}
	return nil
}
