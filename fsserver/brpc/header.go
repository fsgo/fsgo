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

var headerBP = sync.Pool{
	New: func() any {
		bf := &bytes.Buffer{}
		bf.Grow(headerSize)
		return bf
	},
}

const headerSize = 12

type Header struct {
	BodySize uint32
	MetaSize uint32
}

func (h Header) Bytes() []byte {
	bf := make([]byte, headerSize)
	copy(bf, protocol[:])
	binary.BigEndian.PutUint32(bf[4:], h.BodySize)
	binary.BigEndian.PutUint32(bf[8:], h.MetaSize)
	return bf
}

func (h Header) PayloadSize() uint32 {
	return h.BodySize - h.MetaSize
}

func (h Header) WroteTo(w io.Writer) (int64, error) {
	n, err := w.Write(h.Bytes())
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
