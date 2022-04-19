// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/20

package fsnet

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ConnStatTrace 用于获取网络状态的拦截器
type ConnStatTrace struct {
	readSize int64
	readCost int64

	writeSize int64
	writeCost int64

	dialCost      int64
	dialTimes     int64
	dialFailTimes int64

	connInterceptor   *ConnInterceptor
	dialerInterceptor *DialerInterceptor

	once sync.Once
}

func (ch *ConnStatTrace) init() {
	ch.connInterceptor = &ConnInterceptor{
		Read: func(b []byte, invoker func([]byte) (int, error)) (n int, err error) {
			start := time.Now()
			defer func() {
				atomic.AddInt64(&ch.readCost, time.Since(start).Nanoseconds())
				atomic.AddInt64(&ch.readSize, int64(n))
			}()
			return invoker(b)
		},

		Write: func(b []byte, invoker func([]byte) (int, error)) (n int, err error) {
			start := time.Now()
			defer func() {
				atomic.AddInt64(&ch.writeCost, time.Since(start).Nanoseconds())
				atomic.AddInt64(&ch.writeSize, int64(n))
			}()
			return invoker(b)
		},
	}

	ch.dialerInterceptor = &DialerInterceptor{
		DialContext: func(ctx context.Context, network string, address string, invoker DialContextFunc) (conn net.Conn, err error) {
			start := time.Now()

			conn, err = invoker(ctx, network, address)

			atomic.AddInt64(&ch.dialCost, time.Since(start).Nanoseconds())
			atomic.AddInt64(&ch.dialTimes, 1)

			if err != nil {
				atomic.AddInt64(&ch.dialFailTimes, 1)
				return nil, err
			}
			return WrapConn(conn, ch.connInterceptor), nil
		},
	}
}

// ConnInterceptor 获取 net.Conn 的状态拦截器
func (ch *ConnStatTrace) ConnInterceptor() *ConnInterceptor {
	ch.once.Do(ch.init)
	return ch.connInterceptor
}

// DialerInterceptor 获取拨号器的拦截器，之后可将其注册到 Dialer
func (ch *ConnStatTrace) DialerInterceptor() *DialerInterceptor {
	ch.once.Do(ch.init)
	return ch.dialerInterceptor
}

// ReadSize 获取累计读到的的字节大小
func (ch *ConnStatTrace) ReadSize() int64 {
	return atomic.LoadInt64(&ch.readSize)
}

// ReadCost 获取累积的读耗时
func (ch *ConnStatTrace) ReadCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.readCost))
}

// WriteSize 获取累计写出的的字节大小
func (ch *ConnStatTrace) WriteSize() int64 {
	return atomic.LoadInt64(&ch.writeSize)
}

// WriteCost 获取累积的写耗时
func (ch *ConnStatTrace) WriteCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.writeCost))
}

// DialCost 获取累积的 Dial 耗时
func (ch *ConnStatTrace) DialCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.dialCost))
}

// DialTimes 获取累积的 Dial 总次数
func (ch *ConnStatTrace) DialTimes() int64 {
	return atomic.LoadInt64(&ch.dialTimes)
}

// DialFailTimes 获取累积的 Dial 失败次数
func (ch *ConnStatTrace) DialFailTimes() int64 {
	return atomic.LoadInt64(&ch.dialFailTimes)
}

// Reset 将所有状态数据重置为 0
func (ch *ConnStatTrace) Reset() {
	atomic.StoreInt64(&ch.dialCost, 0)
	atomic.StoreInt64(&ch.dialTimes, 0)
	atomic.StoreInt64(&ch.dialFailTimes, 0)

	atomic.StoreInt64(&ch.readSize, 0)
	atomic.StoreInt64(&ch.readCost, 0)

	atomic.StoreInt64(&ch.writeSize, 0)
	atomic.StoreInt64(&ch.writeCost, 0)
}

// ConnReadBytesTrace 获取所有通过 Read 方法读取的数据的副本
type ConnReadBytesTrace struct {
	interceptor *ConnInterceptor
	buf         bytes.Buffer
	once        sync.Once
	mux         sync.RWMutex
}

func (ch *ConnReadBytesTrace) init() {
	ch.interceptor = &ConnInterceptor{
		AfterRead: func(b []byte, readSize int, err error) {
			if readSize > 0 {
				ch.mux.Lock()
				ch.buf.Write(b[:readSize])
				ch.mux.Unlock()
			}
		},
	}
}

// ReadBytes Read 方法读取到的数据的副本
func (ch *ConnReadBytesTrace) ReadBytes() []byte {
	ch.mux.RLock()
	defer ch.mux.RUnlock()
	return append([]byte(nil), ch.buf.Bytes()...)
}

// ConnInterceptor 获取 ConnInterceptor 实例
func (ch *ConnReadBytesTrace) ConnInterceptor() *ConnInterceptor {
	ch.once.Do(ch.init)
	return ch.interceptor
}

// Reset 重置 buffer
func (ch *ConnReadBytesTrace) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}

// ConnWriteBytesTrace 获取所有通过 Write 方法写出的数据的副本
type ConnWriteBytesTrace struct {
	interceptor *ConnInterceptor
	buf         bytes.Buffer
	once        sync.Once
	mux         sync.RWMutex
}

func (ch *ConnWriteBytesTrace) init() {
	ch.interceptor = &ConnInterceptor{
		AfterWrite: func(b []byte, wroteSize int, err error) {
			if wroteSize > 0 {
				ch.mux.Lock()
				ch.buf.Write(b[:wroteSize])
				ch.mux.Unlock()
			}
		},
	}
}

// WriteBytes Write 方法写出的数据的副本
func (ch *ConnWriteBytesTrace) WriteBytes() []byte {
	ch.mux.RLock()
	defer ch.mux.RUnlock()
	return append([]byte(nil), ch.buf.Bytes()...)
}

// ConnInterceptor 获取 ConnInterceptor 实例
func (ch *ConnWriteBytesTrace) ConnInterceptor() *ConnInterceptor {
	ch.once.Do(ch.init)
	return ch.interceptor
}

// Reset 重置 buffer
func (ch *ConnWriteBytesTrace) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}

// ConnCopy 实现对网络连接读写数据的复制
type ConnCopy struct {
	interceptor *ConnInterceptor
	once        sync.Once

	// ReadTo 将 Read 到的数据写入此处,比如 os.Stdout
	ReadTo io.Writer

	// WriterTo 将 Writer 的数据写入此处，比如 os.Stdout
	WriterTo io.Writer
}

func (cc *ConnCopy) init() {
	cc.interceptor = &ConnInterceptor{
		AfterRead: func(b []byte, readSize int, err error) {
			if readSize > 0 && cc.ReadTo != nil {
				_, _ = cc.ReadTo.Write(b[:readSize])
			}
		},
		AfterWrite: func(b []byte, wroteSize int, err error) {
			if wroteSize > 0 && cc.ReadTo != nil {
				_, _ = cc.WriterTo.Write(b[:wroteSize])
			}
		},
	}
}

// ConnInterceptor 获取 ConnInterceptor 实例
func (cc *ConnCopy) ConnInterceptor() *ConnInterceptor {
	cc.once.Do(cc.init)
	return cc.interceptor
}
