// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package fsio

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

var _ io.WriteCloser = (*AsyncWriter)(nil)

// AsyncWriter 异步化的 writer
type AsyncWriter struct {
	io.Writer

	Size int

	buffers chan *bytes.Buffer
	once    sync.Once
	pool    *sync.Pool

	status    int32
	writeErr  atomic.Value
	closeDone chan bool
}

const (
	asyncWriterRunning int32 = 1
	asyncWriterClosed  int32 = 2
)

var errClosed = errors.New("already closed")

// Write 异步写
func (aw *AsyncWriter) Write(p []byte) (n int, err error) {
	aw.once.Do(aw.init)
	if atomic.LoadInt32(&aw.status) != asyncWriterRunning {
		return 0, errClosed
	}
	bf := aw.pool.Get().(*bytes.Buffer)
	n, err = bf.Write(p)
	aw.buffers <- bf
	return n, err
}

func (aw *AsyncWriter) init() {
	atomic.StoreInt32(&aw.status, asyncWriterRunning)
	aw.pool = &sync.Pool{
		New: func() any {
			return &bytes.Buffer{}
		},
	}
	aw.buffers = make(chan *bytes.Buffer, aw.Size)
	aw.closeDone = make(chan bool)
	go func() {
		for b := range aw.buffers {
			n, err := aw.Writer.Write(b.Bytes())
			s := WriteStatus{
				Wrote: n,
				Err:   err,
			}
			aw.writeErr.Store(s)
			b.Reset()
			aw.pool.Put(b)
		}
		close(aw.closeDone)
	}()

}

// LastWriteStatus 返回的是异步写的最新一次的状态
func (aw *AsyncWriter) LastWriteStatus() WriteStatus {
	val := aw.writeErr.Load()
	if val == nil {
		return WriteStatus{}
	}
	return val.(WriteStatus)
}

// Close 关闭
func (aw *AsyncWriter) Close() error {
	switch atomic.LoadInt32(&aw.status) {
	case asyncWriterRunning:
		close(aw.buffers)
		atomic.StoreInt32(&aw.status, asyncWriterClosed)
		<-aw.closeDone
	}
	return nil
}
