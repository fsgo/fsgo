// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/fsgo/fsgo/fssync"
	"github.com/fsgo/fsgo/fssync/fsatomic"
)

func NewClientConn(rw io.ReadWriter) *ClientConn {
	cc := &ClientConn{
		rw: rw,
	}
	cc.ctx, cc.ctxCancel = context.WithCancelCause(context.Background())
	return cc
}

type ClientConn struct {
	rw     io.ReadWriter
	closed atomic.Bool

	writeQueue *bufferQueue

	responses fssync.Map[uint64, *respReader]
	payloads  fssync.Map[uint64, payloadChan]

	rwErr chan error

	beforeReadLoop fsatomic.FuncVoid // 每次循环前执行

	lastErr   fsatomic.Error // 读错误
	ctx       context.Context
	ctxCancel context.CancelCauseFunc

	initOnce sync.Once
}

func (cc *ClientConn) SetBeforeReadLoop(fn func()) {
	cc.beforeReadLoop.Store(fn)
}

func (cc *ClientConn) LastError() error {
	return cc.lastErr.Load()
}

func (cc *ClientConn) init() {
	cc.rwErr = make(chan error, 2)
	cc.writeQueue = newBufferQueue(1024)

	go func() {
		defer cc.Close()

		// rd := io.TeeReader(cc.rw, &fsio.PrintByteWriter{Name: "Client Read"})
		if err := ReadProtocol(cc.rw); err != nil {
			cc.lastErr.Store(err)
			cc.ctxCancel(err)
			cc.rwErr <- err
			return
		}

		for !cc.closed.Load() {
			err := cc.readOnePackage(cc.rw)
			if err != nil {
				cc.lastErr.Store(err)
				cc.ctxCancel(err)
				cc.rwErr <- err
				return
			}
		}
	}()

	go func() {
		// mw := io.MultiWriter(cc.rw, &fsio.PrintByteWriter{Name: "Client Write"})
		if err := cc.writeQueue.startWrite(cc.rw); err != nil {
			cc.lastErr.Store(err)
			cc.ctxCancel(err)
			cc.rwErr <- err
		}
	}()
}

func (cc *ClientConn) readOnePackage(rd io.Reader) error {
	if fn := cc.beforeReadLoop.Load(); fn != nil {
		fn()
	}
	header, err1 := ReadHeader(rd)
	if err1 != nil {
		return err1
	}
	switch header.Type {
	default:
		return fmt.Errorf("%w, got=%d", ErrInvalidHeader, header.Type)
	case HeaderTypeResponse:
		resp, err := readMessage(rd, int(header.Length), &Response{})
		if err != nil {
			return err
		}
		rid := resp.GetRequestID()
		reader, ok := cc.responses.LoadAndDelete(rid)
		if !ok {
			return fmt.Errorf("response reader not found, rid=%d", rid)
		}
		reader.responses <- resp
		reader.errors <- nil
		if resp.GetHasPayload() {
			pl := make(payloadChan, 1)
			reader.payloads <- pl
			cc.payloads.Store(rid, pl)
		} else {
			reader.payloads <- emptyPayloadChan
		}
	case HeaderTypePayload:
		payload, err := readPayload(rd, int(header.Length))
		if err != nil {
			return err
		}
		rid := payload.Meta.RID
		plReader, ok := cc.payloads.Load(rid)
		if !ok {
			return fmt.Errorf("response not found, rid=%d", rid)
		}
		plReader <- payload
		if !payload.Meta.More {
			close(plReader)
			cc.payloads.Delete(rid)
		}
	}
	return nil
}

func (cc *ClientConn) MustOpen(ctx context.Context) *Stream {
	s, err := cc.Open(ctx)
	if err != nil {
		panic(err)
	}
	return s
}

func (cc *ClientConn) Open(ctx context.Context) (*Stream, error) {
	if cc.closed.Load() {
		return nil, ErrClosed
	}
	cc.initOnce.Do(cc.init)

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(ErrCanceledByDefer)

	go func() {
		select {
		case <-ctx.Done():
		case err, ok := <-cc.rwErr:
			if ok && err != nil {
				cancel(err)
			}
		case <-cc.ctx.Done():
			cancel(context.Cause(cc.ctx))
		}
	}()

	rw := &Stream{
		queue: cc.writeQueue,
		newResponseReader: func(req *Request) ResponseReader {
			rr := newRespReader()
			cc.responses.Store(req.GetID(), rr)
			return rr
		},
	}
	return rw, nil
}

func (cc *ClientConn) Close() error {
	if !cc.closed.CompareAndSwap(false, true) {
		return nil
	}
	cc.writeQueue.Close()
	cc.responses.Range(func(key uint64, value *respReader) bool {
		value.sendError(cc.lastErr.Load())
		return true
	})
	return nil
}
