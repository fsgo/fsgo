// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsdialer

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsgo/fsgo/fsnet/fsconn"
)

// ConnStatTracer 用于获取网络状态的拦截器
type ConnStatTracer struct {
	connInterceptor   *fsconn.Interceptor
	dialerInterceptor *Interceptor

	readSize int64
	readCost int64

	writeSize int64
	writeCost int64

	dialCost      int64
	dialTimes     int64
	dialFailTimes int64

	once sync.Once
}

func (ch *ConnStatTracer) init() {
	ch.connInterceptor = &fsconn.Interceptor{
		Read: func(_ fsconn.Info, b []byte, invoker func([]byte) (int, error)) (n int, err error) {
			start := time.Now()
			defer func() {
				atomic.AddInt64(&ch.readCost, time.Since(start).Nanoseconds())
				atomic.AddInt64(&ch.readSize, int64(n))
			}()
			return invoker(b)
		},

		Write: func(_ fsconn.Info, b []byte, invoker func([]byte) (int, error)) (n int, err error) {
			start := time.Now()
			defer func() {
				atomic.AddInt64(&ch.writeCost, time.Since(start).Nanoseconds())
				atomic.AddInt64(&ch.writeSize, int64(n))
			}()
			return invoker(b)
		},
	}

	ch.dialerInterceptor = &Interceptor{
		DialContext: func(ctx context.Context, network string, address string, invoker DialContextFunc) (conn net.Conn, err error) {
			start := time.Now()

			conn, err = invoker(ctx, network, address)

			atomic.AddInt64(&ch.dialCost, time.Since(start).Nanoseconds())
			atomic.AddInt64(&ch.dialTimes, 1)

			if err != nil {
				atomic.AddInt64(&ch.dialFailTimes, 1)
				return nil, err
			}
			return fsconn.WithInterceptor(conn, ch.connInterceptor), nil
		},
	}
}

// ConnInterceptor 获取 net.Conn 的状态拦截器
func (ch *ConnStatTracer) ConnInterceptor() *fsconn.Interceptor {
	ch.once.Do(ch.init)
	return ch.connInterceptor
}

// DialerInterceptor 获取拨号器的拦截器，之后可将其注册到 Dialer
func (ch *ConnStatTracer) DialerInterceptor() *Interceptor {
	ch.once.Do(ch.init)
	return ch.dialerInterceptor
}

// ReadSize 获取累计读到的的字节大小
func (ch *ConnStatTracer) ReadSize() int64 {
	return atomic.LoadInt64(&ch.readSize)
}

// ReadCost 获取累积的读耗时
func (ch *ConnStatTracer) ReadCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.readCost))
}

// WriteSize 获取累计写出的的字节大小
func (ch *ConnStatTracer) WriteSize() int64 {
	return atomic.LoadInt64(&ch.writeSize)
}

// WriteCost 获取累积的写耗时
func (ch *ConnStatTracer) WriteCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.writeCost))
}

// DialCost 获取累积的 Dial 耗时
func (ch *ConnStatTracer) DialCost() time.Duration {
	return time.Duration(atomic.LoadInt64(&ch.dialCost))
}

// DialTimes 获取累积的 Dial 总次数
func (ch *ConnStatTracer) DialTimes() int64 {
	return atomic.LoadInt64(&ch.dialTimes)
}

// DialFailTimes 获取累积的 Dial 失败次数
func (ch *ConnStatTracer) DialFailTimes() int64 {
	return atomic.LoadInt64(&ch.dialFailTimes)
}

// Reset 将所有状态数据重置为 0
func (ch *ConnStatTracer) Reset() {
	atomic.StoreInt64(&ch.dialCost, 0)
	atomic.StoreInt64(&ch.dialTimes, 0)
	atomic.StoreInt64(&ch.dialFailTimes, 0)

	atomic.StoreInt64(&ch.readSize, 0)
	atomic.StoreInt64(&ch.readCost, 0)

	atomic.StoreInt64(&ch.writeSize, 0)
	atomic.StoreInt64(&ch.writeCost, 0)
}
