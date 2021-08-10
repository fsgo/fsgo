// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/10

// +build !windows

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
	syscall.Dup2(int(w.Fd()), fd)
	go io.Copy(to, r)
	return nil
}
