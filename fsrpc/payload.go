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

func readPayload(rd io.Reader, length int) payloadData {
	meta, err := readMessage(rd, length, &Payload{})
	if err != nil {
		return payloadData{
			Err: err,
		}
	}
	bf := make([]byte, meta.Length)
	_, err = io.ReadFull(rd, bf)
	if err != nil {
		return payloadData{
			Err: fmt.Errorf("read payload failed: %w", err),
		}
	}
	return payloadData{
		Meta: meta,
		Data: bf,
	}
}

type payloadData struct {
	Meta *Payload
	Data []byte
	Err  error
}

type PayloadReader interface {
	Payload() (*Payload, []byte, error)
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

func (pw *plWriter) writeChan(ctx context.Context, pl <-chan *bytes.Buffer) error {
	var last *bytes.Buffer
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

func (pw *plWriter) WritePayload(b *bytes.Buffer, more bool) error {
	meta := &Payload{
		Length: uint32(b.Len()),
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
	_, err4 := bp.Write(b.Bytes())
	if err4 == nil {
		pw.queue.send(bp)
		return nil
	}
	return err4
}
