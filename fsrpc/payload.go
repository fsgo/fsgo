// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package fsrpc

import (
	"bytes"
	"context"
	"io"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
)

type Payload struct {
	Meta *PayloadMeta
	Data io.Reader
}

func (pl *Payload) ProtoUnmarshal(obj proto.Message) error {
	bf, err := io.ReadAll(pl.Data)
	if err != nil {
		return err
	}
	return proto.Unmarshal(bf, obj)
}

func readPayload(rd io.Reader, length int) (*Payload, error) {
	meta, err := readMessage(rd, length, &PayloadMeta{})
	if err != nil {
		return nil, err
	}
	bf := make([]byte, meta.Length)
	_, err = io.ReadFull(rd, bf)
	if err != nil {
		return nil, err
	}
	return &Payload{
		Meta: meta,
		// Data: io.LimitReader(rd, meta.Length),
		Data: bytes.NewBuffer(bf),
	}, nil
}

type payloadChan chan *Payload

var emptyPayloadChan = make(chan *Payload)

func init() {
	close(emptyPayloadChan)
}

func newPayloadWriter(rid uint64, q *bufferQueue) *payloadWriter {
	return &payloadWriter{
		queue: q,
		RID:   rid,
	}
}

type payloadWriter struct {
	queue *bufferQueue
	RID   uint64
	index atomic.Uint32
}

func (pw *payloadWriter) writeChan(ctx context.Context, readers <-chan io.Reader) error {
	var last io.Reader
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case data, ok := <-readers:
			if !ok && last != nil {
				return pw.WritePayload(last, false)
			}

			if last != nil {
				if err := pw.WritePayload(last, true); err != nil {
					return err
				}
			}
			last = data
		}
	}
}

func (pw *payloadWriter) WritePayload(b io.Reader, more bool) error {
	var bb *bytes.Buffer
	switch val := b.(type) {
	case *bytes.Buffer:
		bb = val
	default:
		bb = &bytes.Buffer{}
		_, err0 := io.Copy(bb, b)
		if err0 != nil {
			return err0
		}
	}
	meta := &PayloadMeta{
		Length: int64(bb.Len()),
		More:   more,
		RID:    pw.RID,
		Index:  pw.index.Add(1) - 1,
	}
	bf1, err1 := proto.Marshal(meta)
	if err1 != nil {
		return err1
	}
	header := Header{
		Type:   HeaderTypePayload,
		Length: uint32(len(bf1)),
	}
	bp := bytesPool.Get()
	if err2 := header.Write(bp); err2 != nil {
		return err2
	}
	if _, err3 := bp.Write(bf1); err3 != nil {
		return err3
	}
	_, err4 := bp.Write(bb.Bytes())
	if err4 == nil {
		return pw.queue.sendReader(bp)
	}
	return err4
}
