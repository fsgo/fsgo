// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsgo/fsgo/fsfn"
	"github.com/fsgo/fsgo/fssync"
	"github.com/fsgo/fsgo/fssync/fsatomic"
)

func NewClient(rw io.ReadWriter) *Client {
	cc := &Client{
		readWriter: rw,
	}
	return cc
}

type Client struct {
	readWriter io.ReadWriter
	closed     atomic.Bool

	writeQueue *bufferQueue

	respReaders fssync.Map[uint64, *respReader]

	beforeReadLoop fsatomic.FuncVoid // 每次循环前执行

	lastErr fsatomic.Error

	initOnce sync.Once

	onClose fssync.Slice[func()]
}

func (cc *Client) SetBeforeReadLoop(fn func()) {
	cc.beforeReadLoop.Store(fn)
}

func (cc *Client) OnClose(fn func()) {
	cc.onClose.Add(fn)
}

func (cc *Client) LastError() error {
	return cc.lastErr.Load()
}

func (cc *Client) init() {
	cc.writeQueue = newBufferQueue(1024)
	ctx, cancel := context.WithCancelCause(context.Background())

	var errOnce atomic.Bool
	storeError := func(err error) {
		if !errOnce.CompareAndSwap(false, true) {
			return
		}
		cc.lastErr.Store(err)
		_ = cc.closeWithError(err)
		cancel(err)
	}

	go func() {
		if err := ReadProtocol(cc.readWriter); err != nil {
			storeError(err)
			return
		}

		for !cc.closed.Load() {
			err := cc.readOnePackage(cc.readWriter)
			if err != nil {
				storeError(err)
				return
			}
		}
	}()

	go func() {
		err := cc.writeQueue.startWrite(cc.readWriter)
		if err != nil {
			storeError(err)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = cc.closeWithError(context.Cause(ctx))
	}()
}

func (cc *Client) readOnePackage(rd io.Reader) error {
	fsfn.RunVoid(cc.beforeReadLoop.Load())

	header, err1 := ReadHeader(rd)
	if err1 != nil {
		return fmt.Errorf("read Header: %w", err1)
	}

	switch header.Type {
	default:
		return fmt.Errorf("%w, got=%d", ErrInvalidHeader, header.Type)
	case HeaderTypeResponse:
		resp, err := readProtoMessage(rd, int(header.Length), &Response{})
		if err != nil {
			return fmt.Errorf("read Response: %w", err)
		}
		rid := resp.GetRequestID()
		reader, ok := cc.respReaders.Load(rid)
		if !ok {
			return fmt.Errorf("response reader not found, rid=%d", rid)
		}
		if err = reader.receiveResponseOnce(resp); err != nil {
			return err
		}
		if !resp.HasPayload {
			reader.readFinish()
			cc.respReaders.Delete(rid)
		}
	case HeaderTypePayload:
		payload, err := readPayload(rd, int(header.Length))
		if err != nil {
			return fmt.Errorf("read Payload: %w", err)
		}
		rid := payload.Meta.RID
		reader, ok := cc.respReaders.Load(rid)
		if !ok {
			return fmt.Errorf("response not found, rid=%d", rid)
		}
		if err = reader.receivePayload(payload); err != nil {
			return err
		}
		if !payload.Meta.More {
			reader.readFinish()
			cc.respReaders.Delete(rid)
		}
	}
	return nil
}

func (cc *Client) OpenStream() RequestWriter {
	cc.initOnce.Do(cc.init)
	rw := &reqWriter{
		queue: cc.writeQueue,
		newResReader: func(req *Request) ResponseReader {
			rr := newRespReader()
			cc.respReaders.Store(req.GetID(), rr)
			return rr
		},
	}
	return rw
}

func (cc *Client) closeWithError(err error) error {
	if !cc.closed.CompareAndSwap(false, true) {
		return nil
	}
	cc.writeQueue.CloseWithErr(err)
	cc.respReaders.Range(func(_ uint64, value *respReader) bool {
		value.closeWithError(err)
		return true
	})
	fsfn.RunVoids(cc.onClose.Load())
	return nil
}

func (cc *Client) Close() error {
	return cc.closeWithError(stringError("by Close"))
}

func DialTimeout(network string, addr string, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	nc := NewClient(conn)
	nc.OnClose(func() {
		_ = conn.Close()
	})
	return nc, nil
}
