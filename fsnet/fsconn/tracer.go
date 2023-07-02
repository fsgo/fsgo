// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsconn

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/fsgo/fsgo/fsio"
)

// ReadTracer 获取所有通过 Read 方法读取的数据的副本
type ReadTracer struct {
	interceptor *Interceptor
	buf         bytes.Buffer
	once        sync.Once
	mux         sync.RWMutex
}

func (ch *ReadTracer) init() {
	ch.interceptor = &Interceptor{
		AfterRead: func(info Info, b []byte, readSize int, _ error) {
			if readSize > 0 {
				ch.mux.Lock()
				ch.buf.Write(b[:readSize])
				ch.mux.Unlock()
			}
		},
	}
}

// ReadBytes Read 方法读取到的数据的副本
func (ch *ReadTracer) ReadBytes() []byte {
	ch.mux.RLock()
	defer ch.mux.RUnlock()
	return append([]byte(nil), ch.buf.Bytes()...)
}

// ConnInterceptor 获取 Interceptor 实例
func (ch *ReadTracer) ConnInterceptor() *Interceptor {
	ch.once.Do(ch.init)
	return ch.interceptor
}

// Reset 重置 buffer
func (ch *ReadTracer) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}

// WriteTracer 获取所有通过 Write 方法写出的数据的副本
type WriteTracer struct {
	interceptor *Interceptor
	buf         bytes.Buffer
	once        sync.Once
	mux         sync.RWMutex
}

func (ch *WriteTracer) init() {
	ch.interceptor = &Interceptor{
		AfterWrite: func(info Info, b []byte, wroteSize int, _ error) {
			if wroteSize > 0 {
				ch.mux.Lock()
				ch.buf.Write(b[:wroteSize])
				ch.mux.Unlock()
			}
		},
	}
}

// WriteBytes Write 方法写出的数据的副本
func (ch *WriteTracer) WriteBytes() []byte {
	ch.mux.RLock()
	defer ch.mux.RUnlock()
	return append([]byte(nil), ch.buf.Bytes()...)
}

// Interceptor 获取 Interceptor 实例
func (ch *WriteTracer) Interceptor() *Interceptor {
	ch.once.Do(ch.init)
	return ch.interceptor
}

// Reset 重置 buffer
func (ch *WriteTracer) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}

type PrintByteTracer struct {
	interceptor *Interceptor
	once        sync.Once
	Out         io.Writer
}

func (pb *PrintByteTracer) init() {
	rb := &fsio.PrintByteWriter{
		Name: "Read",
		Out:  pb.Out,
	}
	wb := &fsio.PrintByteWriter{
		Name: "Write",
		Out:  pb.Out,
	}
	pb.interceptor = &Interceptor{
		AfterRead: func(info Info, b []byte, readSize int, err error) {
			meta := info.RemoteAddr().String() + ", err=" + fmt.Sprint(err)
			_, _ = rb.WriteWithMeta(b, meta)
		},
		AfterWrite: func(info Info, b []byte, readSize int, err error) {
			meta := info.RemoteAddr().String() + ", err=" + fmt.Sprint(err)
			_, _ = wb.WriteWithMeta(b, meta)
		},
	}
}

func (pb *PrintByteTracer) Interceptor() *Interceptor {
	pb.once.Do(pb.init)
	return pb.interceptor
}
