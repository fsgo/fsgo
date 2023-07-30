// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package fsrpc

import (
	"bytes"
	"io"
	"sync"

	"github.com/fsgo/fsgo/fssync/fsatomic"
	"github.com/fsgo/fsgo/internal/xpool"
)

var bytesPool = xpool.NewBytesPool(1024)

func newBufferQueue(size int) *bufferQueue {
	return &bufferQueue{
		queue: make(chan io.Reader, size),
		done:  make(chan struct{}),
	}
}

type bufferQueue struct {
	queue     chan io.Reader
	done      chan struct{}
	closeErr  fsatomic.Error
	closeOnce sync.Once
}

func (sc *bufferQueue) startWrite(w io.Writer) (err error) {
	defer func() {
		sc.CloseWithErr(err)
	}()
	if err = WriteProtocol(w); err != nil {
		return err
	}
	for {
		select {
		case bp := <-sc.queue:
			switch val := bp.(type) {
			case *bytes.Buffer:
				_, err1 := w.Write(val.Bytes())
				if err1 != nil {
					return err1
				}
				bytesPool.Put(val)
			default:
				_, err1 := io.Copy(w, bp)
				if err1 != nil {
					return err1
				}
			}

		case <-sc.done:
			return sc.closeErr.Load()
		}
	}
}

func (sc *bufferQueue) sendReader(b io.Reader) error {
	select {
	case <-sc.done:
		return sc.closeErr.Load()
	case sc.queue <- b:
		return nil
	}
}

func (sc *bufferQueue) CloseWithErr(err error) {
	sc.closeOnce.Do(func() {
		sc.closeErr.Store(err)
		close(sc.done)
	})
}

func (sc *bufferQueue) Err() error {
	return sc.closeErr.Load()
}
