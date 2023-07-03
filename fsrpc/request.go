// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"fmt"
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

type (
	RequestProtoWriter interface {
		Write(ctx context.Context, req *Request, body ...proto.Message) (ResponseReader, error)
	}

	RequestChanWriter interface {
		WriteChan(ctx context.Context, req *Request, pl <-chan *Payload) (ResponseReader, error)
	}
)

var _ RequestProtoWriter = (*Stream)(nil)
var _ RequestChanWriter = (*Stream)(nil)

type Stream struct {
	queue        *bufferQueue
	newResReader func(req *Request) ResponseReader
}

func (rw *Stream) Write(ctx context.Context, req *Request, payload ...proto.Message) (ResponseReader, error) {
	ch, err := toPayloadChan[proto.Message](req.GetID(), payload...)
	if err != nil {
		return nil, err
	}
	return rw.WriteChan(ctx, req, ch)
}

func (rw *Stream) WriteChan(ctx context.Context, req *Request, payloads <-chan *Payload) (ResponseReader, error) {
	if payloads != nil {
		req.HasPayload = true
	}
	reqBf, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	h := Header{
		Type:   HeaderTypeRequest,
		Length: uint32(len(reqBf)),
	}

	bp := bytesPool.Get()
	if err = h.Write(bp); err != nil {
		return nil, err
	}
	_, err = bp.Write(reqBf)
	if err != nil {
		return nil, err
	}

	if err = rw.queue.sendReader(bp); err != nil {
		return nil, err
	}

	reader := rw.newResReader(req)

	if payloads != nil {
		pw := newPayloadWriter(req.GetID(), rw.queue)
		err4 := pw.writeChan(ctx, payloads)
		return reader, err4
	}

	return reader, nil
}

type RequestReader interface {
	Request() (*Request, <-chan *Payload)
}

var _ RequestReader = (*requestReader)(nil)

func newRequestReader() *requestReader {
	return &requestReader{
		requests: make(chan *Request, 1),
		payloads: make(chan payloadChan, 1),
	}
}

type requestReader struct {
	requests chan *Request
	payloads chan payloadChan
}

func (r *requestReader) Request() (*Request, <-chan *Payload) {
	return <-r.requests, <-r.payloads
}

func ReadProtoRequest[T proto.Message](ctx context.Context, r RequestReader, data T) (*Request, T, error) {
	req, bodyChan := r.Request()
	if bodyChan == nil {
		return nil, data, fmt.Errorf("read request: %w, bodyChan is nil", ErrNoPayload)
	}
	select {
	case <-ctx.Done():
		return nil, data, context.Cause(ctx)
	case pl, ok := <-bodyChan:
		if !ok {
			return nil, data, fmt.Errorf("read request: %w, bodyChan closed,rid=%d", ErrNoPayload, req.GetID())
		}
		err := pl.ProtoUnmarshal(data)
		return req, data, err
	}
}
