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

// NewConnStatTrace create new ConnStatTrace instance
func NewConnStatTrace() *ConnStatTrace {
	return &ConnStatTrace{}
}

// ConnStatTrace 用于获取网络状态的拦截器
type ConnStatTrace struct {
	readSize int64
	readCost int64

	writeSize int64
	writeCost int64

	dialCost      int64
	dialTimes     int64
	dialFailTimes int64

	connIt   *ConnInterceptor
	dialerIt *DialerInterceptor

	once sync.Once
}

func (ch *ConnStatTrace) init() {
	ch.connIt = &ConnInterceptor{
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

	ch.dialerIt = &DialerInterceptor{
		DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
			start := time.Now()

			conn, err = fn(ctx, network, address)

			atomic.AddInt64(&ch.dialCost, time.Since(start).Nanoseconds())
			atomic.AddInt64(&ch.dialTimes, 1)

			if err != nil {
				atomic.AddInt64(&ch.dialFailTimes, 1)
				return nil, err
			}
			return WrapConn(conn, ch.connIt), nil
		},
	}
}

// ConnInterceptor 获取 net.Conn 的状态拦截器
func (ch *ConnStatTrace) ConnInterceptor() *ConnInterceptor {
	ch.once.Do(ch.init)
	return ch.connIt
}

// DialerInterceptor 获取拨号器的拦截器，之后可将其注册到 Dialer
func (ch *ConnStatTrace) DialerInterceptor() *DialerInterceptor {
	ch.once.Do(ch.init)
	return ch.dialerIt
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

// NewConnReadBytesTrace create ConnReadBytesTrace instance
func NewConnReadBytesTrace() *ConnReadBytesTrace {
	return &ConnReadBytesTrace{}
}

// ConnReadBytesTrace 获取所有通过 Read 方法读取的数据的副本
type ConnReadBytesTrace struct {
	connHook *ConnInterceptor
	buf      bytes.Buffer
	once     sync.Once
	mux      sync.RWMutex
}

func (ch *ConnReadBytesTrace) init() {
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
func (ch *ConnReadBytesTrace) ReadBytes() []byte {
	ch.mux.RLock()
	defer ch.mux.RUnlock()
	return ch.buf.Bytes()
}

// ConnInterceptor 获取 ConnInterceptor 实例
func (ch *ConnReadBytesTrace) ConnInterceptor() *ConnInterceptor {
	ch.once.Do(ch.init)
	return ch.connHook
}

// Reset 重置 buffer
func (ch *ConnReadBytesTrace) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}

// NewConnWriteBytesTrace create ConnWriteBytesTrace instance
func NewConnWriteBytesTrace() *ConnWriteBytesTrace {
	return &ConnWriteBytesTrace{}
}

// ConnWriteBytesTrace 获取所有通过 Write 方法写出的数据的副本
type ConnWriteBytesTrace struct {
	connHook *ConnInterceptor
	buf      bytes.Buffer
	once     sync.Once
	mux      sync.RWMutex
}

func (ch *ConnWriteBytesTrace) init() {
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
func (ch *ConnWriteBytesTrace) WriteBytes() []byte {
	ch.mux.RLock()
	defer ch.mux.RUnlock()
	return ch.buf.Bytes()
}

// ConnInterceptor 获取 ConnInterceptor 实例
func (ch *ConnWriteBytesTrace) ConnInterceptor() *ConnInterceptor {
	ch.once.Do(ch.init)
	return ch.connHook
}

// Reset 重置 buffer
func (ch *ConnWriteBytesTrace) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}
