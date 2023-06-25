// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package fsrpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
)

func readPayload(rd io.Reader, length int) *Payload {
	meta, err := readMessage(rd, length, &PayloadMeta{})
	if err != nil {
		return &Payload{
			Err: err,
		}
	}
	bf := make([]byte, meta.Length)
	_, err = io.ReadFull(rd, bf)
	if err != nil {
		return &Payload{
			Err: fmt.Errorf("read payload failed: %w", err),
		}
	}
	return &Payload{
		Meta: meta,
		Data: bf,
	}
}

var payloadChanEmpty = make(chan *Payload)

func init() {
	close(payloadChanEmpty)
}

type Payload struct {
	Meta *PayloadMeta
	Data []byte
	Err  error
}

type PayloadReader interface {
	Payload() (*PayloadMeta, []byte, error)
}

func newPayloadWriter(rid uint64, q *bufferQueue) *plWriter {
	return &plWriter{
		queue: q,
		RID:   rid,
	}
}

type plWriter struct {
	queue *bufferQueue
	RID   uint64
	index atomic.Uint32
}

func (pw *plWriter) writeChan(ctx context.Context, pl <-chan io.Reader) error {
	var last io.Reader
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data, ok := <-pl:
			if !ok {
				if last != nil {
					pw.WritePayload(last, false)
				}
				return nil
			}

			if last != nil {
				pw.WritePayload(last, true)
			}
			last = data
		}
	}
}

func (pw *plWriter) WritePayload(b io.Reader, more bool) error {
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
		Length: uint32(bb.Len()),
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
		pw.queue.send(bp)
		return nil
	}
	return err4
}
