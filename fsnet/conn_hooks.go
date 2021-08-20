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

// NewConnStatHook create instance
func NewConnStatHook() *ConnStatHook {
	return &ConnStatHook{}
}

// ConnStatHook 用于获取网络状态的 Hook
type ConnStatHook struct {
	readSize int64
	readCost int64

	writeSize int64
	writeCost int64

	dialCost      int64
	dialTimes     int64
	dialFailTimes int64

	connHook *ConnHook
	dialHook *DialerHook

	once sync.Once
}

func (ch *ConnStatHook) init() {
	ch.connHook = &ConnHook{
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

	ch.dialHook = &DialerHook{
		DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
			start := time.Now()

			conn, err = fn(ctx, network, address)

			atomic.AddInt64(&ch.dialCost, time.Since(start).Nanoseconds())
			atomic.AddInt64(&ch.dialTimes, 1)

			if err != nil {
				atomic.AddInt64(&ch.dialFailTimes, 1)
				return nil, err
			}
			return NewConn(conn, ch.connHook), nil
		},
	}
}

// ConnHook 获取 net.Conn 的状态 hook
func (ch *ConnStatHook) ConnHook() *ConnHook {
	ch.once.Do(ch.init)
	return ch.connHook
}

// DialerHook 获取拨号器的 Hook，之后可将其注册到 Dialer
func (ch *ConnStatHook) DialerHook() *DialerHook {
	ch.once.Do(ch.init)
	return ch.dialHook
}

// ReadSize 获取累计读到的的字节大小
func (ch *ConnStatHook) ReadSize() int64 {
	return atomic.LoadInt64(&ch.readSize)
}

// ReadCost 获取累积的读耗时
func (ch *ConnStatHook) ReadCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.readCost))
}

// WriteSize 获取累计写出的的字节大小
func (ch *ConnStatHook) WriteSize() int64 {
	return atomic.LoadInt64(&ch.writeSize)
}

// WriteCost 获取累积的写耗时
func (ch *ConnStatHook) WriteCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.writeCost))
}

// DialCost 获取累积的 Dial 耗时
func (ch *ConnStatHook) DialCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.dialCost))
}

// DialTimes 获取累积的 Dial 总次数
func (ch *ConnStatHook) DialTimes() int64 {
	return atomic.LoadInt64(&ch.dialTimes)
}

// DialFailTimes 获取累积的 Dial 失败次数
func (ch *ConnStatHook) DialFailTimes() int64 {
	return atomic.LoadInt64(&ch.dialFailTimes)
}

// Reset 将所有状态数据重置为 0
func (ch *ConnStatHook) Reset() {
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
	connHook *ConnHook
	buf      bytes.Buffer
	once     sync.Once
	mux      sync.RWMutex
}

func (ch *ConnReadBytesHook) init() {
	ch.connHook = &ConnHook{
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

// ConnHook 获取 Hook 实例
func (ch *ConnReadBytesHook) ConnHook() *ConnHook {
	ch.once.Do(ch.init)
	return ch.connHook
}

// Reset 重置 buffer
func (ch *ConnReadBytesHook) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}

// NewConnWriteBytesHook create ConnWriteBytesHook ins
func NewConnWriteBytesHook() *ConnWriteBytesHook {
	return &ConnWriteBytesHook{}
}

// ConnWriteBytesHook 获取所有通过 Write 方法写出的数据的副本
type ConnWriteBytesHook struct {
	connHook *ConnHook
	buf      bytes.Buffer
	once     sync.Once
	mux      sync.RWMutex
}

func (ch *ConnWriteBytesHook) init() {
	ch.connHook = &ConnHook{
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
func (ch *ConnWriteBytesHook) WriteBytes() []byte {
	ch.mux.RLock()
	defer ch.mux.RUnlock()
	return ch.buf.Bytes()
}

// ConnHook 获取 Hook 实例
func (ch *ConnWriteBytesHook) ConnHook() *ConnHook {
	ch.once.Do(ch.init)
	return ch.connHook
}

// Reset 重置 buffer
func (ch *ConnWriteBytesHook) Reset() {
	ch.mux.Lock()
	ch.buf.Reset()
	ch.mux.Unlock()
}
