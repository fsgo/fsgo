// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsfs

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsgo/fsgo/fsio"
)

var _ io.WriteCloser = (*Rotator)(nil)

// Rotator 文件具备自动切割的功能
type Rotator struct {
	// Path 文件名
	Path string

	// ExtRule 文件后缀生成的规则，可选
	// 目前支持：1hour、1day、no(默认)
	ExtRule string

	// ExtFunc 文件后缀生成的自定义函数,可选
	// 优先使用 ExtRule
	ExtFunc func() string

	// MaxFiles 最多保留文件数，默认为 24
	MaxFiles int

	// MaxDelay 最大延迟时间,可选，默认为 100ms.
	// 影响文件状态、buffer.
	// 如文件被删除了，则最大间隔 MaxDelay 时长会检查到
	MaxDelay time.Duration

	// NewWriter 对文件的封装
	NewWriter func(w io.Writer) fsio.ResetWriter

	filePathFn func() string

	kp        *Keeper
	mux       sync.RWMutex
	onceSetup sync.Once
	onceInit  sync.Once

	writer fsio.ResetWriter
	timer  time.Timer
}

// Init 初始化
func (f *Rotator) Init() error {
	return f.initOnce()
}

func (f *Rotator) initOnce() error {
	var err error
	f.onceInit.Do(func() {
		err = f.setupOnce()
		if err == nil {
			err = f.kp.Start()
		}
	})
	return err
}

func (f *Rotator) mustInit() {
	if err := f.initOnce(); err != nil {
		panic(err)
	}
}

func (f *Rotator) mustSetup() {
	if err := f.setupOnce(); err != nil {
		panic(err)
	}
}

func (f *Rotator) setupOnce() error {
	var err error
	f.onceSetup.Do(func() {
		err = f.setup()
	})
	return err
}

// setup 初始化
func (f *Rotator) setup() error {
	if err := f.setFilePathFn(); err != nil {
		return err
	}

	if f.NewWriter != nil {
		f.writer = f.NewWriter(io.Discard)
	} else {
		f.writer = fsio.NewResetWriter(io.Discard)
	}

	f.setupKeepFile()
	f.setupClean()

	return nil
}

func (f *Rotator) setupKeepFile() {
	maxDelay := f.MaxDelay
	if maxDelay < time.Microsecond {
		maxDelay = 100 * time.Millisecond
	}

	f.kp = &Keeper{
		FilePath:      f.filePathFn,
		CheckInterval: maxDelay,
	}

	f.kp.AfterChange(f.onFileChange)
}

func (f *Rotator) onFileChange(file *os.File) {
	_ = fsio.TryFlush(f.writer)
	f.writer.Reset(file)
}

func (f *Rotator) setupClean() {
	if f.MaxFiles < 0 {
		return
	}

	num := f.MaxFiles
	if num == 0 {
		num = 24
	}
	f.kp.AfterChange(func(_ *os.File) {
		err := CleanFiles(f.Path+"*", num)
		if err != nil {
			log.Println("[Rotator][CleanFiles][error]", err)
		}
	})
}

func (f *Rotator) setFilePathFn() error {
	if f.Path == "" {
		return fmt.Errorf("f.Path is empty")
	}

	if f.ExtRule != "" {
		if fn, has := rotateExtRules[f.ExtRule]; has {
			f.filePathFn = func() string {
				return f.Path + fn()
			}
			return nil
		}
		return fmt.Errorf("extRule=%q not support", f.ExtRule)
	}

	if f.ExtFunc != nil {
		f.filePathFn = func() string {
			return f.Path + f.ExtFunc()
		}
		return nil
	}

	f.filePathFn = func() string {
		return f.Path
	}

	return nil
}

// Write 写入
func (f *Rotator) Write(p []byte) (n int, err error) {
	if err := f.initOnce(); err != nil {
		return 0, err
	}
	return f.writer.Write(p)
}

// File 获取当前的文件
func (f *Rotator) File() *os.File {
	f.mustInit()
	return f.kp.File()
}

// AfterChange 注册当文件变化时的回调函数
func (f *Rotator) AfterChange(fn func(f *os.File)) {
	f.mustSetup()
	f.kp.AfterChange(fn)
}

// Close 关闭文件
func (f *Rotator) Close() error {
	if f.kp == nil {
		return nil
	}
	_ = fsio.TryFlush(f.writer)
	return f.kp.Stop()
}

var rotateExtRules = map[string]func() string{
	"no": func() string {
		return ""
	},
	"1year": func() string {
		return "." + time.Now().Format("2006")
	},
	"1month": func() string {
		return "." + time.Now().Format("200601")
	},
	"1day": func() string {
		return "." + time.Now().Format("20060102")
	},
	"1hour": func() string {
		return "." + time.Now().Format("2006010215")
	},
	"1minute": func() string {
		return "." + time.Now().Format("200601021504")
	},
	"1second": func() string {
		return "." + time.Now().Format("20060102150405")
	},
}
