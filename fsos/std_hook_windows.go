// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/10

package fsos

import (
	"io"
	"os"
	"syscall"
)

func StdFileHook(fd int, to io.Writer) error {
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	winHandler := kernel32.NewProc("SetStdHandle")
	var hfd int
	switch fd {
	case Stdout:
		hfd = syscall.STD_OUTPUT_HANDLE
	case Stderr:
		hfd = syscall.STD_ERROR_HANDLE
	}
	v, _, err := winHandler.Call(uintptr(hfd), w.Fd())
	if v == 0 {
		return err
	}
	go io.Copy(to, r)
	return nil
}
