// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package fsrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
)

type Payload struct {
	Meta *PayloadMeta
	Data io.Reader
}

func (pl *Payload) Bytes() ([]byte, error) {
	return io.ReadAll(pl.Data)
}

func ParserPayload[T any](pl *Payload, data T) (T, error) {
	bf, err := pl.Bytes()
	if err != nil {
		return data, err
	}
	var dataAny any = data

	switch pl.Meta.GetEncodingType() {
	case EncodingType_Bytes:
		switch obj := dataAny.(type) {
		case []byte:
			obj = bf
			_ = obj
			return data, nil
		default:
			return data, fmt.Errorf("data is %T, not []byte", data)
		}
	case EncodingType_Protobuf:
		switch obj := dataAny.(type) {
		case proto.Message:
			err = proto.Unmarshal(bf, obj)
			return data, err
		default:
			return data, fmt.Errorf("data is %T, not proto.Message", data)
		}
	case EncodingType_JSON:
		err = json.Unmarshal(bf, dataAny)
		return data, err
	default:
		return data, fmt.Errorf("%w: %v", ErrInvalidEncodingType, pl.Meta.GetEncodingType())
	}
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

func bytesMarshal(b []byte) ([]byte, error) {
	return b, nil
}

func toJSONPayloadChan(rid uint64, items ...any) (<-chan *Payload, error) {
	return toPayloadChan[any](rid, EncodingType_JSON, json.Marshal, items...)
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
	if err := pw.queue.Err(); err != nil {
		return err
	}
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

func readOnlyOnePayload[T any](ctx context.Context, payloads <-chan *Payload, data T, et EncodingType, dec func(b []byte, m T) error) (T, error) {
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
		if gt := pl.Meta.GetEncodingType(); gt != et {
			return data, fmt.Errorf("%w, want %v,got %v", ErrInvalidEncodingType, et, gt)
		}
		bf, err := pl.Bytes()
		if err != nil {
			return data, err
		}
		err = dec(bf, data)
		return data, err
	}
}

// ReadPayloadProto 用于读取只有一条 proto.Message 的 Payload
func ReadPayloadProto[T proto.Message](ctx context.Context, payloads <-chan *Payload, data T) (T, error) {
	return readOnlyOnePayload[T](ctx, payloads, data, EncodingType_Protobuf, func(b []byte, m T) error {
		return proto.Unmarshal(b, m)
	})
}

// ReadPayloadJSON 用于读取只有一条 JSON 的 Payload
func ReadPayloadJSON[T any](ctx context.Context, payloads <-chan *Payload, data T) (T, error) {
	return readOnlyOnePayload[T](ctx, payloads, data, EncodingType_JSON, func(b []byte, m T) error {
		return json.Unmarshal(b, m)
	})
}

// ReadPayloadBytes 用于读取只有一条 []byte 的 Payload
func ReadPayloadBytes(ctx context.Context, payloads <-chan *Payload) ([]byte, error) {
	var bf []byte
	return readOnlyOnePayload[[]byte](ctx, payloads, bf, EncodingType_JSON, func(b []byte, m []byte) error {
		bf = b
		return nil
	})
}

// PayloadChan 一个可异步发送 Payload 的辅助工具
type PayloadChan[T any] struct {
	// RID Request ID, 必填
	RID uint64

	// EncodingType 数据编码类型,可选
	EncodingType EncodingType

	ch     chan *Payload
	index  atomic.Uint32
	closed atomic.Bool
	once   sync.Once
}

func (pc *PayloadChan[T]) Chan() <-chan *Payload {
	pc.initOnce()
	return pc.ch
}

func (pc *PayloadChan[T]) initOnce() {
	pc.once.Do(func() {
		pc.ch = make(chan *Payload, 128)

		if pc.EncodingType == EncodingType_Unknown {
			var data T
			var obj any = data
			switch obj.(type) {
			case []byte, *bytes.Buffer:
				pc.EncodingType = EncodingType_Bytes
			case proto.Message:
				pc.EncodingType = EncodingType_Protobuf
			default:
				pc.EncodingType = EncodingType_JSON
			}
		}
	})
}

func encodePayload(data any, et EncodingType) (rd io.Reader, length int, err error) {
	switch et {
	case EncodingType_Bytes:
		switch val := data.(type) {
		case []byte:
			rd = bytes.NewBuffer(val)
			length = len(val)
		case *bytes.Buffer:
			rd = val
			length = val.Len()
		case io.Reader:
			bf := &bytes.Buffer{}
			// todo  超大内容拆分
			_, err = bf.ReadFrom(val)
			rd = bf
			length = bf.Len()
		default:
			err = fmt.Errorf("data is %T,not []byte", data)
		}
	case EncodingType_Protobuf:
		if m, ok := data.(proto.Message); ok {
			var bf []byte
			bf, err = proto.Marshal(m)
			if err == nil {
				rd = bytes.NewBuffer(bf)
				length = len(bf)
			}
		} else {
			err = fmt.Errorf("data is %T,not proto.Message", data)
		}
	case EncodingType_JSON:
		var bf []byte
		bf, err = json.Marshal(data)
		if err == nil {
			rd = bytes.NewBuffer(bf)
			length = len(bf)
		}
	default:
		return nil, 0, fmt.Errorf("%w:%v", ErrInvalidEncodingType, et)
	}
	return rd, length, err
}

var errClosedPayloadChan = errors.New("payloadChan already closed")

func (pc *PayloadChan[T]) Write(ctx context.Context, data T, more bool) error {
	if err0 := ctx.Err(); err0 != nil {
		return fmt.Errorf("%w: when PayloadChan.Write", err0)
	}
	if pc.closed.Load() {
		return errClosedPayloadChan
	}

	pc.initOnce()

	if !more {
		if !pc.closed.CompareAndSwap(false, true) {
			return errClosedPayloadChan
		}
		defer close(pc.ch)
	}

	rd, length, err := encodePayload(data, pc.EncodingType)
	if err != nil {
		return err
	}
	pl := &Payload{
		Meta: &PayloadMeta{
			Index:        pc.index.Add(1) - 1,
			RID:          pc.RID,
			More:         more,
			EncodingType: pc.EncodingType,
			Length:       int64(length),
		},
		Data: rd,
	}
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case pc.ch <- pl:
		return nil
	}
}

// RangeParserPayloads 遍历并解析所有的 Payloads
func RangeParserPayloads[T any](ctx context.Context, ps <-chan *Payload, gen func() T, fn func(data T) error) error {
	return RangePayloads(ctx, ps, func(pl *Payload) error {
		data, err := ParserPayload[T](pl, gen())
		if err == nil {
			err = fn(data)
		}
		return err
	})
}

// RangePayloads 遍历所有的 Payloads
func RangePayloads(ctx context.Context, ps <-chan *Payload, fn func(pl *Payload) error) error {
	if ps == nil {
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case pl, ok := <-ps:
			if !ok {
				return nil
			}
			if err := fn(pl); err != nil {
				return err
			}
		}
	}
}

func PayloadsDiscard(ctx context.Context, ps <-chan *Payload) error {
	return RangePayloads(ctx, ps, func(pl *Payload) error {
		_, _ = io.Copy(io.Discard, pl.Data)
		return nil
	})
}
