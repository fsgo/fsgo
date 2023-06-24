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
	WriteRequest(ctx context.Context, req *Request, pl <-chan *bytes.Buffer) (ResponseReader, error)
}

var _ RequestWriter = (*reqWriter)(nil)

type reqWriter struct {
	queue             *bufferQueue
	newResponseReader func(req *Request) ResponseReader
}

func (rw *reqWriter) WriteRequest(ctx context.Context, req *Request, pl <-chan *bytes.Buffer) (ResponseReader, error) {
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
	Request() *Request
	PayloadReader
}

func newReqReader(meta *Request) *reqReader {
	mc := make(chan *Request, 1)
	if meta != nil {
		mc <- meta
	}
	return &reqReader{
		req: mc,
		pl:  make(chan payloadData, 1024),
	}
}

type reqReader struct {
	req       chan *Request
	pl        chan payloadData
	closeOnce sync.Once
}

func (r *reqReader) Request() *Request {
	return <-r.req
}

func (r *reqReader) sendPayload(pl payloadData) {
	r.pl <- pl
}

func (r *reqReader) Payload() (*Payload, []byte, error) {
	pd, ok := <-r.pl
	if ok {
		return pd.Meta, pd.Data, pd.Err
	}
	return nil, nil, io.EOF
}

func (r *reqReader) Close() {
	r.closeOnce.Do(func() {
		close(r.pl)
		close(r.req)
	})
}

func msgChan[T proto.Message](items ...T) (<-chan *bytes.Buffer, error) {
	if len(items) == 0 {
		return nil, nil
	}
	ch := make(chan *bytes.Buffer, len(items))
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
	return w.WriteRequest(ctx, req, bc)
}

func QuickReadRequest[T proto.Message](r RequestReader, data T) (*Request, T, error) {
	req := r.Request()
	_, bf, err := r.Payload()
	if err != nil {
		return nil, data, err
	}
	err = proto.Unmarshal(bf, data)
	return req, data, err
}
