// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"fmt"
	"io"
	"log"

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

type (
	ResponseWriter interface {
		ResponseProtoWriter
		ResponseChanWriter
	}

	ResponseProtoWriter interface {
		Write(ctx context.Context, resp *Response, body ...proto.Message) error
	}

	ResponseChanWriter interface {
		WriteChan(ctx context.Context, resp *Response, payload <-chan io.Reader) error
	}
)

var _ ResponseProtoWriter = (*respWriter)(nil)
var _ ResponseChanWriter = (*respWriter)(nil)

func newResponseWriter(queue *bufferQueue) *respWriter {
	return &respWriter{
		queue: queue,
	}
}

type respWriter struct {
	queue *bufferQueue
}

func (rw *respWriter) Write(ctx context.Context, resp *Response, body ...proto.Message) error {
	ch, err := msgChan[proto.Message](body...)
	if err != nil {
		return err
	}
	return rw.WriteChan(ctx, resp, ch)
}

func (rw *respWriter) WriteChan(ctx context.Context, resp *Response, payloads <-chan io.Reader) error {
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
	log.Println("respReader) closeWithError:", err)
}

func (r *respReader) Response() (*Response, <-chan *Payload, error) {
	return <-r.responses, <-r.payloads, <-r.errors
}

func ReadProtoResponse[T proto.Message](ctx context.Context, r ResponseReader, data T) (*Response, T, error) {
	resp, bodyChan, err := r.Response()
	if err != nil {
		return nil, data, err
	}
	if bodyChan == nil {
		return nil, data, fmt.Errorf("read response: %w, bodyChan is nil", ErrNoPayload)
	}
	select {
	case <-ctx.Done():
		return nil, data, context.Cause(ctx)
	case pl, ok := <-bodyChan:
		if !ok {
			return nil, data, fmt.Errorf("read response: %w, bodyChan closed, rid=%d", ErrNoPayload, resp.GetRequestID())
		}
		err1 := pl.ProtoUnmarshal(data)
		return resp, data, err1
	}
}
