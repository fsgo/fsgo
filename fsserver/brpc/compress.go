// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/16

package brpc

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"sync"

	"github.com/golang/snappy"
)

const (
	CompressTypeNo     int32 = 0
	CompressTypeSnappy int32 = 1
	CompressTypeGzip   int32 = 2
)

var ErrCompressUnknown = errors.New("unsupported compress type")

func compress(compressType int32, b []byte) ([]byte, error) {
	switch compressType {
	case CompressTypeNo:
		return b, nil
	case CompressTypeSnappy:
		return snappyCompress(b)
	case CompressTypeGzip:
		return gzipCompress(b)
	}
	return nil, fmt.Errorf("%w, compress_type=%d", ErrCompressUnknown, compressType)
}

func deCompress(compressType int32, b []byte) ([]byte, error) {
	switch compressType {
	case CompressTypeNo:
		return b, nil
	case CompressTypeSnappy:
		return snappyDecompress(b)
	case CompressTypeGzip:
		return gzipDecompress(b)
	}
	return nil, fmt.Errorf("%w, compress_type=%d", ErrCompressUnknown, compressType)
}

var gzipPool = &sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

func gzipCompress(b []byte) ([]byte, error) {
	bf := &bytes.Buffer{}
	bf.Grow(len(b) / 2)
	w := gzipPool.Get().(*gzip.Writer)
	w.Reset(bf)
	defer gzipPool.Put(w)

	if _, err := w.Write(b); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func gzipDecompress(b []byte) ([]byte, error) {
	return nil, nil
}

func snappyCompress(b []byte) ([]byte, error) {
	return snappy.Encode(nil, b), nil
}

func snappyDecompress(b []byte) ([]byte, error) {
	return snappy.Decode(nil, b)
}
