// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
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
	Write(ctx context.Context, req *Request, pl <-chan *Payload) (ResponseReader, error)
}

var _ RequestWriter = (*reqWriter)(nil)

type reqWriter struct {
	queue        *bufferQueue // 用于发送消息的队列
	newResReader func(req *Request) ResponseReader
}

func (rw *reqWriter) Write(ctx context.Context, req *Request, payloads <-chan *Payload) (ResponseReader, error) {
	if err := rw.queue.Err(); err != nil {
		return nil, err
	}
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

func WriteRequestProto(ctx context.Context, w RequestWriter, req *Request, payload ...proto.Message) (ResponseReader, error) {
	ch, err := toProtoPayloadChan(req.GetID(), payload...)
	if err != nil {
		return nil, err
	}
	return w.Write(ctx, req, ch)
}

func WriteRequestBytes(ctx context.Context, w RequestWriter, req *Request, payload ...[]byte) (ResponseReader, error) {
	ch, err := toBytesPayloadChan(req.GetID(), payload...)
	if err != nil {
		return nil, err
	}
	return w.Write(ctx, req, ch)
}

func WriteRequestJSON(ctx context.Context, w RequestWriter, req *Request, payload ...any) (ResponseReader, error) {
	ch, err := toJSONPayloadChan(req.GetID(), payload...)
	if err != nil {
		return nil, err
	}
	return w.Write(ctx, req, ch)
}

type RequestReader interface {
	Request() (*Request, <-chan *Payload)
}

var _ RequestReader = (*reqReader)(nil)

func newRequestReader() *reqReader {
	return &reqReader{
		requests: make(chan *Request, 1),
		payloads: make(chan payloadChan, 1),
	}
}

type reqReader struct {
	requests chan *Request
	payloads chan payloadChan
}

func (r *reqReader) Request() (*Request, <-chan *Payload) {
	return <-r.requests, <-r.payloads
}

func ReadRequestProto[T proto.Message](ctx context.Context, r RequestReader, data T) (*Request, T, error) {
	req, bodyChan := r.Request()
	d, err := ReadPayloadProto(ctx, bodyChan, data)
	return req, d, err
}

func ReadRequestJSON[T any](ctx context.Context, r RequestReader, data T) (*Request, T, error) {
	req, bodyChan := r.Request()
	d, err := ReadPayloadJSON(ctx, bodyChan, data)
	return req, d, err
}

func ReadRequestBytes(ctx context.Context, r RequestReader) (*Request, []byte, error) {
	req, bodyChan := r.Request()
	d, err := ReadPayloadBytes(ctx, bodyChan)
	return req, d, err
}
