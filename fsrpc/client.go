// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/fsgo/fsgo/fssync"
	"github.com/fsgo/fsgo/fssync/fsatomic"
)

func NewClientConn(rw io.ReadWriter) *ClientConn {
	cc := &ClientConn{
		readWriter: rw,
	}
	return cc
}

type ClientConn struct {
	readWriter io.ReadWriter
	closed     atomic.Bool

	writeQueue *bufferQueue

	responses fssync.Map[uint64, *respReader]
	payloads  fssync.Map[uint64, payloadChan]

	errors chan error

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
	cc.errors = make(chan error, 2)
	cc.writeQueue = newBufferQueue(1024)
	cc.ctx, cc.ctxCancel = context.WithCancelCause(context.Background())

	storeError := func(err error) {
		if !cc.lastErr.CompareAndSwap(nil, err) {
			return
		}
		_ = cc.closeWithError(err)
		cc.errors <- err
		cc.ctxCancel(err)
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
		<-cc.ctx.Done()
		_ = cc.closeWithError(context.Cause(cc.ctx))
	}()
}

func (cc *ClientConn) readOnePackage(rd io.Reader) error {
	if fn := cc.beforeReadLoop.Load(); fn != nil {
		fn()
	}
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
			return fmt.Errorf("read Payload: %w", err)
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

func (cc *ClientConn) MustOpen(ctx context.Context) RequestWriter {
	s, err := cc.Open(ctx)
	if err != nil {
		panic(err)
	}
	return s
}

func (cc *ClientConn) Open(ctx context.Context) (RequestWriter, error) {
	if cc.closed.Load() {
		return nil, ErrClosed
	}
	cc.initOnce.Do(cc.init)

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(ErrCanceledByDefer)

	go func() {
		select {
		case <-ctx.Done():
		case err, ok := <-cc.errors:
			if ok && err != nil {
				cancel(err)
			}
		case <-cc.ctx.Done():
			cancel(context.Cause(cc.ctx))
		}
	}()

	rw := &reqWriter{
		queue: cc.writeQueue,
		newResReader: func(req *Request) ResponseReader {
			rr := newRespReader()
			cc.responses.Store(req.GetID(), rr)
			return rr
		},
	}
	return rw, nil
}

func (cc *ClientConn) closeWithError(err error) error {
	if !cc.closed.CompareAndSwap(false, true) {
		return nil
	}
	cc.writeQueue.CloseWithErr(err)
	cc.responses.Range(func(key uint64, value *respReader) bool {
		value.closeWithError(err)
		return true
	})
	return nil
}

func (cc *ClientConn) Close() error {
	return cc.closeWithError(errors.New("by Close"))
}
