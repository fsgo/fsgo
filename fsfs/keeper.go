// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsfs

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsgo/fsgo/fssync"
	"github.com/fsgo/fsgo/fstime"
)

// Keeper 保持文件存在
type Keeper struct {
	// FilePath 返回文件地址，必填
	FilePath func() string

	// OpenFile 创建文件的函数，可选
	// 默认为 os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	OpenFile func(fp string) (*os.File, error)

	// CheckInterval 检查间隔，可选
	// 默认为 100ms
	CheckInterval time.Duration

	file  *os.File
	info  os.FileInfo
	timer *fstime.Interval

	beforeChanges fssync.Slice[func(old *os.File)]
	afterChanges  fssync.Slice[func(newFile *os.File)]

	mux     sync.RWMutex
	running bool
}

// Start 开始,非阻塞运行
//
//	与之对应的有 Stop 方法
func (kf *Keeper) Start() error {
	if err := kf.init(); err != nil {
		return err
	}

	if err := kf.checkFile(); err != nil {
		return err
	}
	kf.mux.Lock()
	defer kf.mux.Unlock()
	if kf.running {
		return errors.New("already started")
	}
	kf.running = true
	kf.timer = &fstime.Interval{}
	kf.timer.Add(kf.loop)
	kf.timer.Start(kf.CheckInterval)
	return nil
}

func (kf *Keeper) loop() {
	if err := kf.checkFile(); err != nil {
		log.Println("[fsgo][Keeper][error]", err)
	}
	kf.timer.Reset(kf.CheckInterval)
}

// Stop 停止运行
func (kf *Keeper) Stop() {
	kf.mux.Lock()
	defer kf.mux.Unlock()
	if !kf.running {
		return
	}
	kf.running = false
	kf.timer.Stop()
	_ = kf.file.Close()
}

func (kf *Keeper) init() error {
	if kf.FilePath == nil {
		return errors.New("fn FilePath is nil")
	}
	if kf.CheckInterval <= 0 {
		kf.CheckInterval = 100 * time.Millisecond
	}
	return nil
}

// File 获取文件
func (kf *Keeper) File() *os.File {
	kf.mux.RLock()
	defer kf.mux.RUnlock()
	return kf.file
}

// BeforeChange 注册当文件变化前的回调函数
func (kf *Keeper) BeforeChange(fn func(old *os.File)) {
	kf.beforeChanges.Add(fn)
}

// AfterChange 注册当文件变化后的回调函数
func (kf *Keeper) AfterChange(fn func(newFile *os.File)) {
	kf.afterChanges.Add(fn)
}

func (kf *Keeper) checkFile() error {
	fp := kf.FilePath()

	if len(fp) == 0 {
		return errors.New("empty file path")
	}

	if has, err := kf.exists(fp); has {
		return nil
	} else if err != nil {
		return err
	}

	dir := filepath.Dir(fp)
	if err := KeepDirExists(dir); err != nil {
		return err
	}
	file, err := kf.openFile(fp)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	kf.mux.RLock()
	old := kf.file
	kf.mux.RUnlock()

	if old != nil {
		fns := kf.beforeChanges.Load()
		for _, fn := range fns {
			fn(old)
		}
	}

	kf.mux.Lock()
	kf.file = file
	kf.info = info
	kf.mux.Unlock()

	for _, fn := range kf.afterChanges.Load() {
		fn(file)
	}

	if old != nil {
		return old.Close()
	}
	return nil
}

func (kf *Keeper) openFile(fp string) (*os.File, error) {
	if kf.OpenFile != nil {
		return kf.openFile(fp)
	}
	return os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func (kf *Keeper) exists(fp string) (bool, error) {
	kf.mux.RLock()
	info := kf.info
	kf.mux.RUnlock()

	if info == nil {
		return false, nil
	}
	curInfo, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return os.SameFile(info, curInfo), nil
}
