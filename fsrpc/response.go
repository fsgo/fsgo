// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"bytes"
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
	WriteResponse(ctx context.Context, resp *Response, payload <-chan *bytes.Buffer) error
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

func (rw *respWriter) WriteResponse(ctx context.Context, resp *Response, payload <-chan *bytes.Buffer) error {
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
	Response() *Response
	PayloadReader
}

func newRespReader() *respReader {
	return &respReader{
		resp: make(chan *Response, 1),
		pl:   make(chan payloadData, 1024),
	}
}

type respReader struct {
	resp      chan *Response
	pl        chan payloadData
	closeOnce sync.Once
}

func (r *respReader) Response() *Response {
	return <-r.resp
}

func (r *respReader) sendPayload(pl payloadData) {
	r.pl <- pl
}

func (r *respReader) Payload() (*Payload, []byte, error) {
	pd, ok := <-r.pl
	if ok {
		return pd.Meta, pd.Data, pd.Err
	}
	return nil, nil, io.EOF
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
	return w.WriteResponse(ctx, resp, bc)
}

func QuickReadResponse[T proto.Message](r ResponseReader, data T) (*Response, T, error) {
	resp := r.Response()
	_, bf, err := r.Payload()
	if err != nil {
		return nil, data, err
	}
	err = proto.Unmarshal(bf, data)
	return resp, data, err
}
