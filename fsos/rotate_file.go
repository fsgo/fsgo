// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/15

package fsos

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var _ io.WriteCloser = (*RotateFile)(nil)

// RotateFile 文件具备自动切割的功能
type RotateFile struct {
	// Path 文件名
	Path string

	// ExtRule 文件后缀生成的规则
	// 目前支持：1hour、1day
	ExtRule string

	// ExtFunc 文件后缀生成的自定义函数
	// 优先使用 ExtRule
	ExtFunc func() string

	// MaxFiles 最多保留文件数，默认为 24
	MaxFiles int

	file            *os.File
	bfWrite         *bufio.Writer
	currentFileInfo *os.FileInfo

	mux          sync.RWMutex
	tk           *time.Ticker
	once         sync.Once
	onFileChange []func(f *os.File)
}

// Init 初始化
func (f *RotateFile) Init() error {
	var err error
	f.once.Do(func() {
		err = f.init()
	})
	return err
}

func (f *RotateFile) initOnce() {
	f.once.Do(func() {
		if err := f.init(); err != nil {
			panic(err)
		}
	})
}

func (f *RotateFile) init() error {
	if f.Path == "" {
		fmt.Errorf(" Path is empty")
	}
	if err := f.setFile(); err != nil {
		return err
	}

	if f.MaxFiles == 0 {
		f.MaxFiles = 24
	}
	if f.MaxFiles > 0 {
		f.OnFileChange(f.cleanFiles)
	}

	f.tk = time.NewTicker(time.Second)
	go f.loop()
	return nil
}

func (f *RotateFile) loop() {
	for range f.tk.C {
		if err := f.setFile(); err != nil {
			log.Println("[RotateFile][setFile][error]", err.Error())
		}
		if err := f.Flush(); err != nil {
			log.Println("[RotateFile][flush][error]", err.Error())
		}
	}
}

var rotateExtRules = map[string]func() string{
	"no": func() string {
		return ""
	},
	"1day": func() string {
		return "." + time.Now().Format("20060102")
	},
	"1hour": func() string {
		return "." + time.Now().Format("2006010215")
	},
}

func (f *RotateFile) ext() (string, error) {
	if f.ExtRule != "" {
		if fn, has := rotateExtRules[f.ExtRule]; has {
			return fn(), nil
		}
		return "", fmt.Errorf("extRule=%q not support", f.ExtRule)
	}
	if f.ExtFunc != nil {
		return f.ExtFunc(), nil
	}
	return "", nil
}

func (f *RotateFile) setFile() error {
	fp := f.Path

	if ext, err := f.ext(); err != nil {
		return err
	} else if ext != "" {
		fp += ext
	}

	if has, err := f.exists(fp); has {
		return nil
	} else if err != nil {
		return err
	}

	dir := filepath.Dir(fp)
	if err := KeepDirExists(dir); err != nil {
		return err
	}

	file, err := os.OpenFile(fp, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	f.mux.Lock()

	if f.bfWrite == nil {
		f.bfWrite = bufio.NewWriter(file)
	} else {
		_ = f.bfWrite.Flush()
		f.bfWrite.Reset(file)
	}
	old := f.file
	f.file = file
	f.currentFileInfo = &info
	f.mux.Unlock()

	for _, fn := range f.onFileChange {
		fn(file)
	}

	if old != nil {
		return old.Close()
	}
	return nil
}

func (f *RotateFile) exists(fp string) (bool, error) {
	f.mux.RLock()
	info := f.currentFileInfo
	f.mux.RUnlock()
	if info == nil {
		return false, nil
	}

	info1, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return os.SameFile(*info, info1), nil
}

// Write 写入
func (f *RotateFile) Write(p []byte) (n int, err error) {
	f.initOnce()

	f.mux.RLock()
	defer f.mux.RUnlock()
	if f.bfWrite == nil {
		return 0, fmt.Errorf(" no file opened")
	}
	return f.bfWrite.Write(p)
}

// Flush 刷新 buff
func (f *RotateFile) Flush() error {
	f.mux.RLock()
	defer f.mux.RUnlock()
	if f.bfWrite == nil {
		return fmt.Errorf("writer not exists")
	}
	return f.bfWrite.Flush()
}

// File 获取当前的文件
func (f *RotateFile) File() *os.File {
	f.initOnce()

	f.mux.RLock()
	defer f.mux.RUnlock()
	return f.file
}

// OnFileChange 注册当文件变化时的回调函数
func (f *RotateFile) OnFileChange(fn func(f *os.File)) {
	f.onFileChange = append(f.onFileChange, fn)
}

// cleanFiles 清理过期文件
func (f *RotateFile) cleanFiles(_ *os.File) {
	pattern := f.Path + "*"
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Printf("[RotateFile][cleanFiles] filepath.Glob(%q) has error: %v\n", pattern, err)
		return
	}

	if len(files) <= f.MaxFiles {
		return
	}

	type finfo struct {
		path string
		info os.FileInfo
	}

	f.mux.RLock()
	curInfo := *f.currentFileInfo
	f.mux.RUnlock()

	var finfos []*finfo
	for _, p := range files {
		info, err := os.Stat(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			log.Fatalf("[RotateFile][cleanFiles] os.Stat(%q) has error:%v\n", p, err)
			continue
		}

		if os.SameFile(info, curInfo) {
			continue
		}
		finfos = append(finfos, &finfo{path: p, info: info})
	}

	if len(finfos) <= f.MaxFiles {
		return
	}

	sort.Slice(finfos, func(i, j int) bool {
		a := finfos[i].info.ModTime()
		b := finfos[j].info.ModTime()
		return b.Before(a)
	})

	for i := f.MaxFiles; i < len(finfos); i++ {
		p := finfos[i].path
		err = os.Remove(p)
		log.Printf("[RotateFile][cleanFiles] os.Remove(%q), err=%v\n", p, err)
	}
}

// Close 关闭文件
func (f *RotateFile) Close() error {
	f.mux.Lock()
	defer f.mux.Unlock()

	if f.tk != nil {
		f.tk.Stop()
	}

	if f.file != nil {
		_ = f.bfWrite.Flush()
		f.bfWrite = nil
		if err := f.file.Close(); err != nil {
			return err
		}
		f.file = nil
	}
	return nil
}
