// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"io"
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
)

func NewResponse(requestID uint64, code int64, msg string) *Response {
	return &Response{
		RequestID: requestID,
		Code:      code,
		Message:   msg,
	}
}

type ResponseWriter interface {
	WriteResponse(resp *Response) error
	WritePayload(b []byte, more bool) error
}

var _ ResponseWriter = (*respWriter)(nil)

func newResponseWriter(queue *bufferQueue) *respWriter {
	return &respWriter{
		queue: queue,
	}
}

type respWriter struct {
	queue       *bufferQueue
	wroteHeader atomic.Bool
	responseID  uint64
	index       atomic.Uint32
}

func (rw *respWriter) WriteResponse(meta *Response) error {
	bf, err := proto.Marshal(meta)
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
	rw.queue.send(bp)
	if err == nil {
		rw.wroteHeader.Store(true)
	}
	return err
}

func (rw *respWriter) WritePayload(b []byte, more bool) error {
	if !rw.wroteHeader.Load() {
		return ErrMissWriteMeta
	}

	meta := &Payload{
		Index:  rw.index.Add(1),
		RID:    rw.responseID,
		More:   more,
		Length: uint32(len(b)),
	}

	bf, err := proto.Marshal(meta)
	if err != nil {
		return err
	}
	h := Header{
		Type:   HeaderTypePayload,
		Length: uint32(len(bf)),
	}

	bp := bytesPool.Get()
	if err1 := h.Write(bp); err1 != nil {
		return err1
	}
	_, err2 := bp.Write(bf)
	if err2 != nil {
		return err2
	}
	_, err3 := bp.Write(b)
	if err3 == nil {
		rw.queue.send(bp)
	}
	return err3
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
