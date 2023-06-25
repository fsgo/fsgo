// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
)

func NewRequest(method string) *Request {
	return &Request{
		ID:     globalRequestID.Add(1),
		Method: method,
	}
}

var globalRequestID atomic.Uint64

type RequestWriter interface {
	Write(ctx context.Context, req *Request, body ...proto.Message) (ResponseReader, error)
	WriteChan(ctx context.Context, req *Request, pl <-chan io.Reader) (ResponseReader, error)
}

var _ RequestWriter = (*reqWriter)(nil)

type reqWriter struct {
	queue             *bufferQueue
	newResponseReader func(req *Request) ResponseReader
}

func (rw *reqWriter) Write(ctx context.Context, req *Request, body ...proto.Message) (ResponseReader, error) {
	ch, err := msgChan[proto.Message](body...)
	if err != nil {
		return nil, err
	}
	return rw.WriteChan(ctx, req, ch)
}

func (rw *reqWriter) WriteChan(ctx context.Context, req *Request, pl <-chan io.Reader) (ResponseReader, error) {
	if pl != nil {
		req.HasPayload = true
	}
	bf, err1 := proto.Marshal(req)
	if err1 != nil {
		return nil, err1
	}
	h := Header{
		Type:   HeaderTypeRequest,
		Length: uint32(len(bf)),
	}

	bp := bytesPool.Get()
	if err2 := h.Write(bp); err2 != nil {
		return nil, err2
	}
	_, err3 := bp.Write(bf)
	if err3 != nil {
		return nil, err3
	}
	rw.queue.send(bp)

	reader := rw.newResponseReader(req)

	if pl != nil {
		pw := newPayloadWriter(req.GetID(), rw.queue)
		if err4 := pw.writeChan(ctx, pl); err4 != nil {
			return reader, err4
		}
	}

	return reader, nil
}

type RequestReader interface {
	Request() (*Request, <-chan *Payload)
}

func newReqReader(req *Request) *reqReader {
	var pl chan *Payload
	if req.GetHasPayload() {
		pl = make(chan *Payload, 128)
	} else {
		pl = payloadChanEmpty
	}
	return &reqReader{
		req: req,
		pl:  pl,
	}
}

type reqReader struct {
	req       *Request
	pl        chan *Payload
	closeOnce sync.Once
}

func (r *reqReader) Request() (*Request, <-chan *Payload) {
	return r.req, r.pl
}

func (r *reqReader) sendPayload(pl *Payload) {
	r.pl <- pl
}

func (r *reqReader) Close() {
	r.closeOnce.Do(func() {
		if r.req.GetHasPayload() {
			close(r.pl)
		}
	})
}

func msgChan[T proto.Message](items ...T) (<-chan io.Reader, error) {
	if len(items) == 0 {
		return nil, nil
	}
	ch := make(chan io.Reader, len(items))
	for i := 0; i < len(items); i++ {
		bf, err := proto.Marshal(items[i])
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		ch <- bytes.NewBuffer(bf)
	}
	close(ch)
	return ch, nil
}

func QuickWriteRequest[T proto.Message](ctx context.Context, w RequestWriter, req *Request, data ...T) (ResponseReader, error) {
	bc, err := msgChan[T](data...)
	if err != nil {
		return nil, err
	}
	if bc != nil {
		req.HasPayload = true
	}
	return w.WriteChan(ctx, req, bc)
}

func QuickReadRequest[T proto.Message](r RequestReader, data T) (*Request, T, error) {
	req, body := r.Request()
	if body == nil {
		return nil, data, ErrNoPayload
	}
	pl, ok := <-body
	if !ok {
		return nil, data, ErrNoPayload
	}
	if pl.Err != nil {
		return nil, data, pl.Err
	}
	err := proto.Unmarshal(pl.Data, data)
	return req, data, err
}
