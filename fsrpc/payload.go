// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package fsrpc

import (
	"bytes"
	"context"
	"encoding/json"
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
	bf, err := pl.Bytes()
	if err != nil {
		return err
	}
	return proto.Unmarshal(bf, obj)
}

func (pl *Payload) JSONUnmarshal(obj any) error {
	bf, err := pl.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(bf, obj)
}

func (pl *Payload) Bytes() ([]byte, error) {
	return io.ReadAll(pl.Data)
}

func readPayload(rd io.Reader, length int) (*Payload, error) {
	meta, err := readProtoMessage(rd, length, &PayloadMeta{})
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

func toProtoPayloadChan(rid uint64, items ...proto.Message) (<-chan *Payload, error) {
	return toPayloadChan[proto.Message](rid, EncodingType_Protobuf, proto.Marshal, items...)
}

func toBytesPayloadChan(rid uint64, items ...[]byte) (<-chan *Payload, error) {
	return toPayloadChan[[]byte](rid, EncodingType_Bytes, bytesMarshal, items...)
}

func toJSONPayloadChan(rid uint64, items ...any) (<-chan *Payload, error) {
	return toPayloadChan[any](rid, EncodingType_JSON, json.Marshal, items...)
}

func bytesMarshal(b []byte) ([]byte, error) {
	return b, nil
}

func toPayloadChan[T any](rid uint64, et EncodingType, enc func(m T) ([]byte, error), items ...T) (<-chan *Payload, error) {
	if len(items) == 0 {
		return nil, nil
	}
	ch := make(chan *Payload, len(items))
	for i := 0; i < len(items); i++ {
		bf, err := enc(items[i])
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		pl := &Payload{
			Meta: &PayloadMeta{
				Index:        uint32(i),
				Length:       int64(len(bf)),
				RID:          rid,
				EncodingType: et,
				More:         i < len(items)-1,
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

func ReadProtoPayload[T proto.Message](ctx context.Context, payloads <-chan *Payload, data T) (T, error) {
	if payloads == nil {
		return data, ErrNoPayload
	}
	select {
	case <-ctx.Done():
		return data, context.Cause(ctx)
	case pl, ok := <-payloads:
		if !ok {
			return data, fmt.Errorf("%w, bodyChan closed", ErrNoPayload)
		}
		if pl.Meta.GetMore() {
			return data, ErrMorePayload
		}
		if pl.Meta.GetEncodingType() != EncodingType_Protobuf {
			return data, fmt.Errorf("%w, %v", ErrInvalidEncodingType, pl.Meta.GetEncodingType())
		}
		err := pl.ProtoUnmarshal(data)
		return data, err
	}
}

func ReadJSONPayload[T any](ctx context.Context, payloads <-chan *Payload, data T) (T, error) {
	if payloads == nil {
		return data, ErrNoPayload
	}
	select {
	case <-ctx.Done():
		return data, context.Cause(ctx)
	case pl, ok := <-payloads:
		if !ok {
			return data, fmt.Errorf("%w, bodyChan closed", ErrNoPayload)
		}
		if pl.Meta.GetMore() {
			return data, ErrMorePayload
		}
		if pl.Meta.GetEncodingType() != EncodingType_JSON {
			return data, fmt.Errorf("%w, %v", ErrInvalidEncodingType, pl.Meta.GetEncodingType())
		}
		err := pl.JSONUnmarshal(data)
		return data, err
	}
}

func ReadBytesPayload(ctx context.Context, payloads <-chan *Payload) ([]byte, error) {
	if payloads == nil {
		return nil, ErrNoPayload
	}
	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	case pl, ok := <-payloads:
		if !ok {
			return nil, fmt.Errorf("%w, bodyChan closed", ErrNoPayload)
		}
		if pl.Meta.GetMore() {
			return nil, ErrMorePayload
		}
		if pl.Meta.GetEncodingType() != EncodingType_Bytes {
			return nil, fmt.Errorf("%w, %v", ErrInvalidEncodingType, pl.Meta.GetEncodingType())
		}
		return pl.Bytes()
	}
}
