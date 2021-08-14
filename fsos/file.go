// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/10

package fsos

const (
	Stdout = 1
	Stderr = 2
)

type HasFd interface {
	Fd() uintptr
}
