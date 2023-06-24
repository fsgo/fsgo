// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package fsrpc

import (
	"bytes"
	"io"

	"github.com/fsgo/fsgo/internal/xpool"
)

var bytesPool = xpool.NewBytesPool(1024)

func newBufferQueue(size int) *bufferQueue {
	return &bufferQueue{
		queue: make(chan *bytes.Buffer, size),
		done:  make(chan struct{}),
	}
}

type bufferQueue struct {
	queue chan *bytes.Buffer
	done  chan struct{}
}

func (sc *bufferQueue) startWrite(w io.Writer) error {
	if err := WriteProtocol(w); err != nil {
		return err
	}

	for {
		select {
		case bp := <-sc.queue:
			_, err1 := w.Write(bp.Bytes())
			if err1 != nil {
				return err1
			}
			bytesPool.Put(bp)
		case <-sc.done:
			return nil
		}
	}
	return nil
}

func (sc *bufferQueue) send(b *bytes.Buffer) {
	select {
	case <-sc.done:
	case sc.queue <- b:
	}
}

func (sc *bufferQueue) Close() {
	close(sc.done)
}
