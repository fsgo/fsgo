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
)

// Notes:
// 1. 12-byte header [PRPC][body_size][meta_size]
// 2. body_size and meta_size are in network byte order
// 3. Use service->full_name() + method_name to specify the method to call
// 4. `attachment_size' is set iff request/response has attachment
// 5. Not supported: chunk_info
//
// see https://github.com/apache/incubator-brpc/blob/master/src/brpc/policy/baidu_rpc_protocol.cpp

var protocol = []byte{'P', 'R', 'P', 'C'}

var (
	ErrInvalidProtocol = errors.New("invalid protocol")
	ErrInvalidHeader   = errors.New("invalid header")
	ErrInvalidMeta     = errors.New("invalid meta")
)

const headerSize = 12

type Header struct {
	BodySize uint32
	MetaSize uint32
}

func (h Header) Bytes() []byte {
	var bf [headerSize]byte
	h.toBytes(bf[:])
	return bf[:]
}

func (h Header) toBytes(bf []byte) {
	// copy(bf[0:], protocol[:])
	binary.BigEndian.PutUint32(bf[4:], h.BodySize)
	binary.BigEndian.PutUint32(bf[8:], h.MetaSize)
}

func (h Header) PayloadSize() uint32 {
	return h.BodySize - h.MetaSize
}

var headerWritePool = &sync.Pool{
	New: func() any {
		b := make([]byte, headerSize)
		copy(b[0:], protocol[:])
		return &b
	},
}

func (h Header) WroteTo(w io.Writer) (int64, error) {
	b := make([]byte, headerSize)
	copy(b[0:], protocol)
	binary.BigEndian.PutUint32(b[4:], h.BodySize)
	binary.BigEndian.PutUint32(b[8:], h.MetaSize)
	n, err := w.Write(b)
	return int64(n), err
}

func (h Header) IsValid() error {
	if h.BodySize < h.MetaSize {
		return fmt.Errorf("%w, meta_size=%d is bigger than body_size=%d",
			ErrInvalidHeader, h.BodySize, h.MetaSize)
	}
	return nil
}

func IsMetaInvalid(meta *Meta) error {
	if n := meta.GetAttachmentSize(); n < 0 {
		return fmt.Errorf("%w, attachment_size=%d expect >=0", ErrInvalidMeta, n)
	}
	return nil
}

var headerReadPool = &sync.Pool{
	New: func() any {
		b := make([]byte, headerSize)
		return &b
	},
}

func ReadHeader(rd io.Reader) (Header, error) {
	bf := headerReadPool.Get().(*[]byte)
	h := *bf
	defer headerReadPool.Put(bf)

	if _, err := io.ReadFull(rd, h); err != nil {
		return Header{}, err
	}
	if !bytes.Equal(h[:4], protocol) {
		return Header{}, fmt.Errorf("%w, expect header %q, got %q", ErrInvalidProtocol, protocol, h)
	}
	hd := Header{
		BodySize: binary.BigEndian.Uint32(h[4:8]),
		MetaSize: binary.BigEndian.Uint32(h[8:12]),
	}
	return hd, nil
}
