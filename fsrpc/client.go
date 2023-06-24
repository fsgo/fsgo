// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync/atomic"

	"github.com/fsgo/fsgo/fssync"
	"github.com/fsgo/fsgo/fssync/fsatomic"
)

func NewClientConn(rw io.ReadWriter) *ClientConn {
	cc := &ClientConn{
		rw: rw,
	}
	cc.init()
	return cc
}

type ClientConn struct {
	rw     io.ReadWriter
	closed atomic.Bool

	writeQueue *bufferQueue

	responses fssync.Map[uint64, *respReader]

	rwErr chan error

	beforeReadLoop fsatomic.FuncVoid // 每次循环前执行

	lastErr fsatomic.Error // 读错误
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
		if err := ReadProtocol(cc.rw); err != nil {
			cc.lastErr.Store(err)
			cc.rwErr <- err
			return
		}

		for !cc.closed.Load() {
			if err := cc.readOnePackage(); err != nil {
				cc.lastErr.Store(err)
				cc.rwErr <- err
				return
			}
		}
	}()

	go func() {
		if err := cc.writeQueue.startWrite(cc.rw); err != nil {
			log.Println("writeQueue:", err)
			cc.lastErr.Store(err)
			cc.rwErr <- err
		}
	}()
}

func (cc *ClientConn) readOnePackage() error {
	if fn := cc.beforeReadLoop.Load(); fn != nil {
		fn()
	}
	header, err1 := ReadHeader(cc.rw)
	if err1 != nil {
		return err1
	}
	switch header.Type {
	default:
		return fmt.Errorf("%w, got=%d", ErrInvalidHeaderType, header.Type)
	case HeaderTypeResponse:
		resp, err := readMessage(cc.rw, int(header.Length), &Response{})
		if err != nil {
			return err
		}
		res, ok := cc.responses.Load(resp.GetRequestID())
		if !ok {
			return errors.New("response reader not found")
		}
		res.resp <- resp
	case HeaderTypePayload:
		pl := readPayload(cc.rw, int(header.Length))
		if pl.Err != nil {
			return pl.Err
		}
		responseReader, ok := cc.responses.Load(pl.Meta.RID)
		if !ok {
			return errors.New("response reader not found")
		}
		responseReader.sendPayload(pl)
		if !pl.Meta.More {
			responseReader.Close()
		}
	}
	return nil
}

func (cc *ClientConn) Open(ctx context.Context, hd ClientHandlerFunc) error {
	if cc.closed.Load() {
		return ErrClosed
	}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(ErrCanceledByDefer)

	go func() {
		select {
		case <-ctx.Done():
		case err, ok := <-cc.rwErr:
			if ok && err != nil {
				cancel(err)
			}
		}
	}()

	req := &reqWriter{
		queue: cc.writeQueue,
		newResponseReader: func(req *Request) ResponseReader {
			rr := newRespReader()
			cc.responses.Store(req.GetID(), rr)
			return rr
		},
	}
	return hd(ctx, req)
}

func (cc *ClientConn) Close() error {
	if !cc.closed.CompareAndSwap(false, true) {
		return nil
	}
	cc.writeQueue.Close()
	return nil
}

type ClientHandlerFunc func(ctx context.Context, rw RequestWriter) error
