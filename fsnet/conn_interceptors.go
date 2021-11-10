// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/20

package fsnet

import (
	"bytes"
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// NewConnStatInterceptor create instance
func NewConnStatInterceptor() *ConnStatInterceptor {
	return &ConnStatInterceptor{}
}

// ConnStatInterceptor 用于获取网络状态的 Hook
type ConnStatInterceptor struct {
	readSize int64
	readCost int64

	writeSize int64
	writeCost int64

	dialCost      int64
	dialTimes     int64
	dialFailTimes int64

	connHook *ConnInterceptor
	dialHook *DialerInterceptor

	once sync.Once
}

func (ch *ConnStatInterceptor) init() {
	ch.connHook = &ConnInterceptor{
		Read: func(b []byte, raw func([]byte) (int, error)) (n int, err error) {
			start := time.Now()
			defer func() {
				atomic.AddInt64(&ch.readCost, time.Since(start).Nanoseconds())
				atomic.AddInt64(&ch.readSize, int64(n))
			}()
			return raw(b)
		},

		Write: func(b []byte, raw func([]byte) (int, error)) (n int, err error) {
			start := time.Now()
			defer func() {
				atomic.AddInt64(&ch.writeCost, time.Since(start).Nanoseconds())
				atomic.AddInt64(&ch.writeSize, int64(n))
			}()
			return raw(b)
		},
	}

	ch.dialHook = &DialerInterceptor{
		DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
			start := time.Now()

			conn, err = fn(ctx, network, address)

			atomic.AddInt64(&ch.dialCost, time.Since(start).Nanoseconds())
			atomic.AddInt64(&ch.dialTimes, 1)

			if err != nil {
				atomic.AddInt64(&ch.dialFailTimes, 1)
				return nil, err
			}
			return WrapConn(conn, ch.connHook), nil
		},
	}
}

// ConnInterceptor 获取 net.Conn 的状态拦截器
func (ch *ConnStatInterceptor) ConnInterceptor() *ConnInterceptor {
	ch.once.Do(ch.init)
	return ch.connHook
}

// DialerInterceptor 获取拨号器的拦截器，之后可将其注册到 Dialer
func (ch *ConnStatInterceptor) DialerInterceptor() *DialerInterceptor {
	ch.once.Do(ch.init)
	return ch.dialHook
}

// ReadSize 获取累计读到的的字节大小
func (ch *ConnStatInterceptor) ReadSize() int64 {
	return atomic.LoadInt64(&ch.readSize)
}

// ReadCost 获取累积的读耗时
func (ch *ConnStatInterceptor) ReadCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.readCost))
}

// WriteSize 获取累计写出的的字节大小
func (ch *ConnStatInterceptor) WriteSize() int64 {
	return atomic.LoadInt64(&ch.writeSize)
}

// WriteCost 获取累积的写耗时
func (ch *ConnStatInterceptor) WriteCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.writeCost))
}

// DialCost 获取累积的 Dial 耗时
func (ch *ConnStatInterceptor) DialCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.dialCost))
}

// DialTimes 获取累积的 Dial 总次数
func (ch *ConnStatInterceptor) DialTimes() int64 {
	return atomic.LoadInt64(&ch.dialTimes)
}

// DialFailTimes 获取累积的 Dial 失败次数
func (ch *ConnStatInterceptor) DialFailTimes() int64 {
	return atomic.LoadInt64(&ch.dialFailTimes)
}

// Reset 将所有状态数据重置为 0
func (ch *ConnStatInterceptor) Reset() {
	atomic.StoreInt64(&ch.dialCost, 0)
	atomic.StoreInt64(&ch.dialTimes, 0)
	atomic.StoreInt64(&ch.dialFailTimes, 0)

	atomic.StoreInt64(&ch.readSize, 0)
	atomic.StoreInt64(&ch.readCost, 0)

	atomic.StoreInt64(&ch.writeSize, 0)
	atomic.StoreInt64(&ch.writeCost, 0)
}

// NewConnReadBytesHook create ConnReadBytesHook ins
func NewConnReadBytesHook() *ConnReadBytesHook {
	return &ConnReadBytesHook{}
}

// ConnReadBytesHook 获取所有通过 Read 方法读取的数据的副本
type ConnReadBytesHook struct {
	connHook *ConnInterceptor
	buf      bytes.Buffer
	once     sync.Once
	mux      sync.RWMutex
}

func (ch *ConnReadBytesHook) init() {
	ch.connHook = &ConnInterceptor{
		Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
			n, err := raw(b)
			if n > 0 {
				ch.mux.Lock()
				ch.buf.Write(b[:n])
				ch.mux.Unlock()
			}
			return n, err
		},
	}
}

// ReadBytes Read 方法读取到的数据的副本
func (ch *ConnReadBytesHook) ReadBytes() []byte {
	ch.mux.RLock()
	defer ch.mux.RUnlock()
	return ch.buf.Bytes()
}

// ConnInterceptor 获取 ConnInterceptor 实例
func (ch *ConnReadBytesHook) ConnInterceptor() *ConnInterceptor {
	ch.once.Do(ch.init)
	return ch.connHook
}

// Reset 重置 buffer
func (ch *ConnReadBytesHook) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}

// NewConnWriteBytesInterceptor create ConnWriteBytesInterceptor
func NewConnWriteBytesInterceptor() *ConnWriteBytesInterceptor {
	return &ConnWriteBytesInterceptor{}
}

// ConnWriteBytesInterceptor 获取所有通过 Write 方法写出的数据的副本
type ConnWriteBytesInterceptor struct {
	connHook *ConnInterceptor
	buf      bytes.Buffer
	once     sync.Once
	mux      sync.RWMutex
}

func (ch *ConnWriteBytesInterceptor) init() {
	ch.connHook = &ConnInterceptor{
		Write: func(b []byte, raw func([]byte) (int, error)) (int, error) {
			n, err := raw(b)
			if n > 0 {
				ch.mux.Lock()
				ch.buf.Write(b[:n])
				ch.mux.Unlock()
			}
			return n, err
		},
	}
}

// WriteBytes Write 方法写出的数据的副本
func (ch *ConnWriteBytesInterceptor) WriteBytes() []byte {
	ch.mux.RLock()
	defer ch.mux.RUnlock()
	return ch.buf.Bytes()
}

// ConnInterceptor 获取 ConnInterceptor 实例
func (ch *ConnWriteBytesInterceptor) ConnInterceptor() *ConnInterceptor {
	ch.once.Do(ch.init)
	return ch.connHook
}

// Reset 重置 buffer
func (ch *ConnWriteBytesInterceptor) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}
