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

func NewRequest(method string) *Request {
	return &Request{
		ID:     globalRequestID.Add(1),
		Method: method,
	}
}

var globalRequestID atomic.Uint64

type RequestWriter interface {
	WriteRequest(req *Request) (PayloadWriter, ResponseReader, error)
}

var _ RequestWriter = (*reqWriter)(nil)

type reqWriter struct {
	queue             *bufferQueue
	newResponseReader func(req *Request) ResponseReader
}

func (rw *reqWriter) WriteRequest(req *Request) (PayloadWriter, ResponseReader, error) {
	bf, err1 := proto.Marshal(req)
	if err1 != nil {
		return nil, nil, err1
	}
	h := Header{
		Type:   HeaderTypeRequest,
		Length: uint32(len(bf)),
	}

	bp := bytesPool.Get()
	if err2 := h.Write(bp); err2 != nil {
		return nil, nil, err2
	}
	_, err3 := bp.Write(bf)
	if err3 != nil {
		return nil, nil, err3
	}
	pw := newPayloadWriter(req.GetID(), rw.queue)
	rw.queue.send(bp)
	return pw, rw.newResponseReader(req), nil
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
