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
	"github.com/fsgo/fsgo/fssync/fsatomic"
)

// ConnStatTracer 用于获取网络状态的拦截器
type ConnStatTracer struct {
	connInterceptor   *fsconn.Interceptor
	dialerInterceptor *Interceptor

	readSize atomic.Int64
	readCost fsatomic.TimeDuration

	writeSize atomic.Int64
	writeCost fsatomic.TimeDuration

	dialCost      fsatomic.TimeDuration
	dialTimes     atomic.Int64
	dialFailTimes atomic.Int64

	once sync.Once
}

func (ch *ConnStatTracer) init() {
	ch.connInterceptor = &fsconn.Interceptor{
		Read: func(_ fsconn.Info, b []byte, invoker func([]byte) (int, error)) (n int, err error) {
			start := time.Now()
			defer func() {
				ch.readCost.Add(time.Since(start))
				ch.readSize.Add(int64(n))
			}()
			return invoker(b)
		},

		Write: func(_ fsconn.Info, b []byte, invoker func([]byte) (int, error)) (n int, err error) {
			start := time.Now()
			defer func() {
				ch.writeCost.Add(time.Since(start))
				ch.writeSize.Add(int64(n))
			}()
			return invoker(b)
		},
	}

	ch.dialerInterceptor = &Interceptor{
		DialContext: func(ctx context.Context, network string, address string, invoker DialContextFunc) (conn net.Conn, err error) {
			start := time.Now()

			conn, err = invoker(ctx, network, address)

			ch.dialCost.Add(time.Since(start))
			ch.dialTimes.Add(1)

			if err != nil {
				ch.dialFailTimes.Add(1)
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
	return ch.readSize.Load()
}

// ReadCost 获取累积的读耗时
func (ch *ConnStatTracer) ReadCost() time.Duration {
	return ch.readCost.Load()
}

// WriteSize 获取累计写出的的字节大小
func (ch *ConnStatTracer) WriteSize() int64 {
	return ch.writeSize.Load()
}

// WriteCost 获取累积的写耗时
func (ch *ConnStatTracer) WriteCost() time.Duration {
	return ch.writeCost.Load()
}

// DialCost 获取累积的 Dial 耗时
func (ch *ConnStatTracer) DialCost() time.Duration {
	return ch.dialCost.Load()
}

// DialTimes 获取累积的 Dial 总次数
func (ch *ConnStatTracer) DialTimes() int64 {
	return ch.dialTimes.Load()
}

// DialFailTimes 获取累积的 Dial 失败次数
func (ch *ConnStatTracer) DialFailTimes() int64 {
	return ch.dialFailTimes.Load()
}

// Reset 将所有状态数据重置为 0
func (ch *ConnStatTracer) Reset() {
	ch.dialCost.Store(0)
	ch.dialTimes.Store(0)
	ch.dialFailTimes.Store(0)

	ch.readSize.Store(0)
	ch.readCost.Store(0)

	ch.writeSize.Store(0)
	ch.writeCost.Store(0)
}
