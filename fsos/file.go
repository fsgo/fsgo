// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/10

package fsos

import (
	"log"
	"os"
	"path/filepath"
	"sort"
)

const (
	// Stdout 标准输出
	Stdout = 1

	// Stderr 标准错误输出
	Stderr = 2
)

// HasFd 有实现 Fd 方法
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

// CleanFiles 按照文件前缀清理文件
// 	pattern: eg /home/work/logs/access_log.log.*
// 	remaining: 文件保留个数，eq ：24
func CleanFiles(pattern string, remaining int) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(files) <= remaining {
		return nil
	}

	type finfo struct {
		path string
		info os.FileInfo
	}

	var infos []*finfo
	for _, p := range files {
		info, err := os.Stat(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			log.Fatalf("[fsgo][cleanFiles] os.Stat(%q) has error:%v\n", p, err)
			continue
		}
		infos = append(infos, &finfo{path: p, info: info})
	}

	if len(infos) <= remaining {
		return nil
	}

	sort.Slice(infos, func(i, j int) bool {
		a := infos[i].info.ModTime()
		b := infos[j].info.ModTime()
		return b.Before(a)
	})

	for i := remaining; i < len(infos); i++ {
		p := infos[i].path
		if err = os.Remove(p); err != nil {
			log.Printf("[fsgo][cleanFiles] os.Remove(%q), err=%v\n", p, err)
		}
	}
	return nil
}