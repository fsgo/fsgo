// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package fsio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/fsgo/fsgo/fssync/fsatomic"
)

var _ io.WriteCloser = (*AsyncWriter)(nil)

// AsyncWriter 异步化的 writer
type AsyncWriter struct {
	Writer io.Writer

	writeStats fsatomic.Value[WriteStatus]

	buffers chan *bytes.Buffer
	pool    *sync.Pool

	done     chan bool
	ChanSize int

	once sync.Once

	closed     atomic.Bool
	NeedStatus bool
	initMux    sync.Mutex
}

var errClosed = errors.New("already closed")

// Write 异步写
func (aw *AsyncWriter) Write(p []byte) (n int, err error) {
	if aw.closed.Load() {
		return 0, errClosed
	}
	aw.once.Do(aw.init)
	bf := aw.pool.Get().(*bytes.Buffer)
	n, err = bf.Write(p)
	select {
	case <-aw.done:
		return 0, errClosed
	case aw.buffers <- bf:
		return n, err
	}
}

func (aw *AsyncWriter) init() {
	aw.initMux.Lock()
	defer aw.initMux.Unlock()

	aw.done = make(chan bool)
	aw.pool = &sync.Pool{
		New: func() any {
			return &bytes.Buffer{}
		},
	}
	aw.buffers = make(chan *bytes.Buffer, aw.ChanSize)
	go func() {
		defer func() {
			_ = recover()
		}()
		for !aw.closed.Load() {
			aw.doLoop()
		}
	}()
}

func (aw *AsyncWriter) doLoop() {
	defer func() {
		if re := recover(); re != nil {
			err := fmt.Errorf("writer  panic %v", re)
			s := WriteStatus{
				Err: err,
			}
			aw.writeStats.Store(s)
		}
	}()

	for {
		select {
		case <-aw.done:
			return
		case b := <-aw.buffers:
			data := b.Bytes()
			bf := make([]byte, 0, len(data))
			bf = append(bf, data...)
			n, err := aw.Writer.Write(bf)
			b.Reset()
			aw.pool.Put(b)
			if aw.NeedStatus {
				s := WriteStatus{
					Wrote: n,
					Err:   err,
				}
				aw.writeStats.Store(s)
			}
		}
	}
}

// LastWriteStatus 返回的是异步写的最新一次的状态
func (aw *AsyncWriter) LastWriteStatus() WriteStatus {
	return aw.writeStats.Load()
}

// Close 关闭
func (aw *AsyncWriter) Close() error {
	if aw.closed.CompareAndSwap(false, true) {
		aw.initMux.Lock()
		defer aw.initMux.Unlock()
		if aw.done != nil {
			close(aw.done)
		}
	}
	return nil
}
