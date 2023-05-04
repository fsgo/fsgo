// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsfs

import (
	"log"
	"os"
	"path/filepath"
	"sort"
)

// HasFd 有实现 Fd 方法
type HasFd interface {
	Fd() uintptr
}

// Exists 判断文件/目录是否存在
func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
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
//
//	pattern: eg /home/work/logs/access_log.log.*
//	remaining: 文件保留个数，eq ：24
func CleanFiles(pattern string, remaining int) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(files) <= remaining {
		return nil
	}

	type finfo struct {
		info os.FileInfo
		path string
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
