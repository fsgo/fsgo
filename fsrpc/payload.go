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

func toPayloadChan[T proto.Message](rid uint64, items ...T) (<-chan *Payload, error) {
	if len(items) == 0 {
		return nil, nil
	}
	ch := make(chan *Payload, len(items))
	for i := 0; i < len(items); i++ {
		bf, err := proto.Marshal(items[i])
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		pl := &Payload{
			Meta: &PayloadMeta{
				Index:  uint32(i),
				Length: int64(len(bf)),
				RID:    rid,
				Type:   1,
				More:   i < len(items)-1,
			},
			Data: bytes.NewBuffer(bf),
		}
		ch <- pl
	}
	close(ch)
	return ch, nil
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

func (pw *payloadWriter) writeChan(ctx context.Context, payloads <-chan *Payload) error {
	var last *Payload
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case data, ok := <-payloads:
			if !ok && last != nil {
				return pw.writePayload(last)
			}
			if last != nil {
				if err := pw.writePayload(last); err != nil {
					return err
				}
			}
			last = data
		}
	}
}

func (pw *payloadWriter) writePayload(pl *Payload) error {
	bf1, err1 := proto.Marshal(pl.Meta)
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
	_, err4 := io.Copy(bp, pl.Data)
	if err4 == nil {
		return pw.queue.sendReader(bp)
	}
	return err4
}
