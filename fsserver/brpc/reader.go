// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/16

package brpc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"

	"google.golang.org/protobuf/proto"
)

type Reader struct {
	// MaxBodySize 可选，最大 Body 大小，为 0 是会使用 全局变量 MaxBodySize
	MaxBodySize uint32
}

// MaxBodySize 默认的最大 body 大小
var MaxBodySize = uint32(10 << 20)

var ErrBodyTooLarge = errors.New("body size is too large")

func (r *Reader) GetMaxBodySize() uint32 {
	if r.MaxBodySize == 0 {
		return MaxBodySize
	}
	return r.MaxBodySize
}

func (r *Reader) ReadHeader(rd io.Reader) (Header, error) {
	bf := headerBP.Get().(*bytes.Buffer)
	bf.Reset()
	defer headerBP.Put(bf)
	if _, err := io.CopyN(bf, rd, headerSize); err != nil {
		return Header{}, err
	}
	h := bf.Bytes()
	if !bytes.Equal(h[:4], protocol) {
		return Header{}, fmt.Errorf("%w, expect header %q, got %q", ErrInvalidProtocol, protocol, h)
	}
	hd := Header{
		BodySize: binary.BigEndian.Uint32(h[4:8]),
		MetaSize: binary.BigEndian.Uint32(h[8:12]),
	}
	if n := r.GetMaxBodySize(); hd.BodySize > n {
		return hd, fmt.Errorf("%w, got is %d, max allow is %d", ErrBodyTooLarge, hd.BodySize, n)
	}
	return hd, hd.IsValid()
}

var metaBP = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func (r *Reader) ReadMeta(rd io.Reader, metaSize uint32) (*Meta, error) {
	bf := metaBP.Get().(*bytes.Buffer)
	bf.Reset()
	bf.Grow(int(metaSize))
	defer metaBP.Put(bf)

	if _, err := io.CopyN(bf, rd, int64(metaSize)); err != nil {
		return nil, err
	}
	var meta *Meta
	if err := proto.Unmarshal(bf.Bytes(), meta); err != nil {
		return nil, err
	}
	return meta, IsMetaInvalid(meta)
}

func (r *Reader) ReadPackage(rd io.Reader) (Header, *Message, error) {
	h, err := r.ReadHeader(rd)
	if err != nil {
		return h, nil, err
	}
	msg, err := r.ReadMessage(rd, h)
	return h, msg, err
}

func (r *Reader) ReadMessage(rd io.Reader, h Header) (*Message, error) {
	return r.readRR(rd, h)
}

var attachmentNop = io.NopCloser(&bytes.Buffer{})

func (r *Reader) readRR(rd io.Reader, h Header) (*Message, error) {
	meta, err := r.ReadMeta(rd, h.MetaSize)
	if err != nil {
		return nil, err
	}

	if meta.GetRequest() == nil {
		return nil, fmt.Errorf("%w, RequestMeta is nil", ErrInvalidMeta)
	}
	amSize := meta.GetAttachmentSize()
	body := make([]byte, h.PayloadSize()-uint32(amSize))
	if _, err = io.ReadFull(rd, body); err != nil {
		return nil, fmt.Errorf("read payload failed: %w", err)
	}
	ct := meta.GetCompressType()
	if ct > 0 {
		body, err = deCompress(meta.GetCompressType(), body)
		if err != nil {
			return nil, err
		}
	}

	var attachment io.Reader
	if amSize == 0 {
		attachment = attachmentNop
	} else {
		bf := &bytes.Buffer{}
		bf.Grow(int(amSize))
		if _, err = io.CopyN(bf, rd, int64(amSize)); err != nil {
			return nil, err
		}
		attachment = bf
	}

	return &Message{
		Meta:       meta,
		body:       body,
		Attachment: attachment,
	}, nil
}
