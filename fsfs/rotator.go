// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsfs

import (
	"errors"
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

	// MaxFiles 最多保留文件数，超过的文件将被清理掉，默认值为 24
	// 若值为 -1，则保留所有文件
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

func (f *Rotator) getMaxDelay() time.Duration {
	if f.MaxDelay < time.Microsecond {
		return 100 * time.Millisecond
	}
	return f.MaxDelay
}

func (f *Rotator) setupKeepFile() {
	f.kp = &Keeper{
		FilePath:      f.filePathFn,
		CheckInterval: f.getMaxDelay(),
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
		return errors.New("f.Path is empty")
	}

	if len(f.ExtRule) > 0 {
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
	if err = f.initOnce(); err != nil {
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
		// eg: .2022040611,.2022040612,.2022040613
		return "." + time.Now().Format("2006010215")
	},
	"1minute": func() string {
		// eg: .202204061100,.202204061101,.202204061102
		return "." + time.Now().Format("200601021504")
	},
	"5minute": func() string {
		// eg: .202204061100,.202204061105,.202204061110
		return nMinuteExt(5)
	},
	"10minute": func() string {
		// eg: .202204061100,.202204061110,.202204061120
		return nMinuteExt(10)
	},
	"15minute": func() string {
		// eg: .202204061100,.202204061115,.202204061130
		return nMinuteExt(15)
	},
	"20minute": func() string {
		// eg: .202204061100,.202204061120,.202204061140
		return nMinuteExt(20)
	},
	"30minute": func() string {
		// eg: .202204061100,.202204061130
		return nMinuteExt(30)
	},
	"1second": func() string {
		return "." + time.Now().Format("20060102150405")
	},
}

func nMinuteExt(n int) string {
	now := time.Now()
	return "." + now.Format("2006010215") + fmt.Sprintf("%02d", now.Minute()/n*n)
}
