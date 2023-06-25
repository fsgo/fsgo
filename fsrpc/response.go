// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"io"
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
)

func NewResponse(requestID uint64, code ErrCode, msg string) *Response {
	return &Response{
		RequestID: requestID,
		Code:      code,
		Message:   msg,
	}
}

func NewResponseSuccess(requestID uint64) *Response {
	return NewResponse(requestID, ErrCode_Success, "OK")
}

type ResponseWriter interface {
	WriteChan(ctx context.Context, resp *Response, payload <-chan io.Reader) error
	Write(ctx context.Context, resp *Response, body ...proto.Message) error
}

var _ ResponseWriter = (*respWriter)(nil)

func newResponseWriter(queue *bufferQueue) *respWriter {
	return &respWriter{
		queue: queue,
	}
}

type respWriter struct {
	queue     *bufferQueue
	requestID uint64
	index     atomic.Uint32
}

func (rw *respWriter) Write(ctx context.Context, resp *Response, body ...proto.Message) error {
	ch, err := msgChan[proto.Message](body...)
	if err != nil {
		return err
	}
	return rw.WriteChan(ctx, resp, ch)
}

func (rw *respWriter) WriteChan(ctx context.Context, resp *Response, payload <-chan io.Reader) error {
	if payload != nil {
		resp.HasPayload = true
	}
	bf, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	h := Header{
		Type:   HeaderTypeResponse,
		Length: uint32(len(bf)),
	}

	bp := bytesPool.Get()
	if err1 := h.Write(bp); err1 != nil {
		return err1
	}
	_, err = bp.Write(bf)
	if err != nil {
		return err
	}
	rw.queue.send(bp)

	if payload == nil {
		return nil
	}
	pw := newPayloadWriter(rw.requestID, rw.queue)
	return pw.writeChan(ctx, payload)
}

type ResponseReader interface {
	Response() (*Response, <-chan *Payload)
}

func newRespReader() *respReader {
	return &respReader{
		resp: make(chan *Response, 1),
		pl:   make(chan *Payload, 128),
	}
}

type respReader struct {
	resp      chan *Response
	pl        chan *Payload
	closeOnce sync.Once
}

func (r *respReader) Response() (*Response, <-chan *Payload) {
	return <-r.resp, r.pl
}

func (r *respReader) sendPayload(pl *Payload) {
	r.pl <- pl
}

func (r *respReader) Close() {
	r.closeOnce.Do(func() {
		close(r.pl)
		close(r.resp)
	})
}

func QuickWriteResponse[T proto.Message](ctx context.Context, w ResponseWriter, resp *Response, data ...T) error {
	bc, err := msgChan[T](data...)
	if err != nil {
		return err
	}
	if bc != nil {
		resp.HasPayload = true
	}
	return w.WriteChan(ctx, resp, bc)
}

func QuickReadResponse[T proto.Message](r ResponseReader, data T) (*Response, T, error) {
	resp, body := r.Response()
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
	return resp, data, err
}
