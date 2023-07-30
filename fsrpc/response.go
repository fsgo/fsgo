// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"errors"
	"sync/atomic"

	"google.golang.org/protobuf/proto"

	"github.com/fsgo/fsgo/fssync/fsatomic"
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

var respReaderID atomic.Int32

func newRespReader() *respReader {
	return &respReader{
		responses:  make(chan *Response, 1),
		payloads:   make(payloadChan, 32),
		closedChan: make(chan struct{}),
		id:         respReaderID.Add(1),
	}
}

var _ ResponseReader = (*respReader)(nil)

type respReader struct {
	responses     chan *Response // 只允许接收一个
	payloads      payloadChan    // 允许接收 0-n 个
	closedErr     fsatomic.Error
	closedChan    chan struct{}
	closed        fsatomic.Once
	readResponse  fsatomic.Once
	payloadClosed fsatomic.Once
	id            int32
}

func (rd *respReader) receiveResponseOnce(resp *Response) error {
	select {
	case rd.responses <- resp:
		return nil
	case <-rd.closedChan:
		return stringError("response reader already closed")
	}
}

func (rd *respReader) receivePayload(p *Payload) error {
	if rd.payloadClosed.Done() {
		return stringError("should no more payload")
	}
	select {
	case rd.payloads <- p:
		return nil
	case <-rd.closedChan:
		return stringError("response reader already closed")
	}
}

func (rd *respReader) closeWithError(err error) {
	if !rd.closed.DoOnce() {
		return
	}
	rd.closedErr.Store(err)
	close(rd.closedChan)
	if rd.payloadClosed.DoOnce() {
		close(rd.payloads)
	}
}

func (rd *respReader) readFinish() {
	rd.closeWithError(stringError("response reader finished"))
}

func (rd *respReader) Response() (s *Response, p <-chan *Payload, e error) {
	if !rd.readResponse.DoOnce() {
		return nil, nil, errors.New("cannot Response twice")
	}
	select {
	case s = <-rd.responses:
		return s, rd.payloads, nil
	case <-rd.closedChan:
		return s, rd.payloads, rd.closedErr.Load()
	}
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
