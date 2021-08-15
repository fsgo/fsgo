// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/10

package fsos

import (
	"os"
)

const (
	Stdout = 1
	Stderr = 2
)

type HasFd interface {
	Fd() uintptr
}

// FileExists 判断文件是否存在
func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// KeepDirExists 保持文件夹存在，若不存在则创建
// 若路径为文件，则删除，然后创建文件夹
func KeepDirExists(dir string) error {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0777)
		if err == nil || os.IsExist(err) {
			return nil
		}
		return err
	}
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	if err = os.Remove(dir); err != nil && !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(dir, 0777)
	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}
