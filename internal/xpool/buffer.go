// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package xpool

import (
	"bytes"
	"sync"
)

func NewBytesPool(size int) *BytesPool {
	return &BytesPool{
		p: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, 0, size))
			},
		},
	}
}

type BytesPool struct {
	p sync.Pool
}

func (bp *BytesPool) Get() *bytes.Buffer {
	val := bp.p.Get()
	return val.(*bytes.Buffer)
}

func (bp *BytesPool) Put(b *bytes.Buffer) {
	b.Reset()
	bp.p.Put(b)
}
