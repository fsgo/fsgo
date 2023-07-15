// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"

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
	Write(ctx context.Context, resp *Response, payload <-chan *Payload) error
}

var _ ResponseWriter = (*respWriter)(nil)

func newResponseWriter(queue *bufferQueue) *respWriter {
	return &respWriter{
		queue: queue,
	}
}

type respWriter struct {
	queue *bufferQueue
}

func (rw *respWriter) Write(ctx context.Context, resp *Response, payloads <-chan *Payload) error {
	if payloads != nil {
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
	if err = h.Write(bp); err != nil {
		return err
	}
	_, err = bp.Write(bf)
	if err != nil {
		return err
	}
	if err = rw.queue.sendReader(bp); err != nil {
		return err
	}

	if payloads == nil {
		return nil
	}
	pw := newPayloadWriter(resp.GetRequestID(), rw.queue)
	return pw.writeChan(ctx, payloads)
}

func WriteResponseProto(ctx context.Context, w ResponseWriter, resp *Response, body ...proto.Message) error {
	ch, err := toProtoPayloadChan(resp.GetRequestID(), body...)
	if err != nil {
		return err
	}
	return w.Write(ctx, resp, ch)
}

func WriteResponseJSON(ctx context.Context, w ResponseWriter, resp *Response, body ...any) error {
	ch, err := toJSONPayloadChan(resp.GetRequestID(), body...)
	if err != nil {
		return err
	}
	return w.Write(ctx, resp, ch)
}

func WritResponseBytes(ctx context.Context, w ResponseWriter, resp *Response, body ...[]byte) error {
	ch, err := toBytesPayloadChan(resp.GetRequestID(), body...)
	if err != nil {
		return err
	}
	return w.Write(ctx, resp, ch)
}

type ResponseReader interface {
	Response() (*Response, <-chan *Payload, error)
}

func newRespReader() *respReader {
	return &respReader{
		responses: make(chan *Response, 1),
		payloads:  make(chan payloadChan, 1),
		errors:    make(chan error, 1),
	}
}

var _ ResponseReader = (*respReader)(nil)

type respReader struct {
	responses chan *Response
	payloads  chan payloadChan
	errors    chan error
}

func (r *respReader) closeWithError(err error) {
	close(r.responses)
	close(r.payloads)
	r.errors <- err
	close(r.errors)
}

func (r *respReader) Response() (*Response, <-chan *Payload, error) {
	return <-r.responses, <-r.payloads, <-r.errors
}

func ReadResponseProto[T proto.Message](ctx context.Context, r ResponseReader, data T) (*Response, T, error) {
	resp, bodyChan, err := r.Response()
	if err != nil {
		return nil, data, err
	}
	d, err1 := ReadPayloadProto(ctx, bodyChan, data)
	return resp, d, err1
}

func ReadResponseJSON[T any](ctx context.Context, r ResponseReader, data T) (*Response, T, error) {
	resp, bodyChan, err := r.Response()
	if err != nil {
		return nil, data, err
	}
	d, err1 := ReadPayloadJSON(ctx, bodyChan, data)
	return resp, d, err1
}

func ReadResponseBytes(ctx context.Context, r ResponseReader) (*Response, []byte, error) {
	resp, bodyChan, err := r.Response()
	if err != nil {
		return nil, nil, err
	}
	d, err1 := ReadPayloadBytes(ctx, bodyChan)
	return resp, d, err1
}
