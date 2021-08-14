// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/14

package fsnet

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var _ net.Conn = (*conn)(nil)

// NewConn wrap conn with hooks
// hooks 将倒序执行：后注册的先执行
func NewConn(c net.Conn, hooks ...*ConnHook) net.Conn {
	nc := &conn{
		raw:   c,
		hooks: hooks,
	}
	return nc
}

// HasRawConn 有原始的 net.Conn
type HasRawConn interface {
	Raw() net.Conn
}

// OriginConn 获取最底层的 net.Conn
func OriginConn(conn net.Conn) net.Conn {
	for {
		c, ok := conn.(HasRawConn)
		if ok {
			conn = c.Raw()
		} else {
			return conn
		}
	}
}

var _ net.Conn = (*conn)(nil)
var _ HasRawConn = (*conn)(nil)

type conn struct {
	raw   net.Conn
	hooks connHooks
}

func (c *conn) Raw() net.Conn {
	return c.raw
}

func (c *conn) Read(b []byte) (n int, err error) {
	return c.hooks.HookRead(b, c.raw.Read, len(c.hooks)-1)
}

func (c *conn) Write(b []byte) (n int, err error) {
	return c.hooks.HookWrite(b, c.raw.Write, len(c.hooks)-1)
}

func (c *conn) Close() error {
	return c.hooks.HookClose(c.raw.Close, len(c.hooks)-1)
}

func (c *conn) LocalAddr() net.Addr {
	return c.hooks.HookLocalAddr(c.raw.LocalAddr, len(c.hooks)-1)
}

func (c *conn) RemoteAddr() net.Addr {
	return c.hooks.HookRemoteAddr(c.raw.RemoteAddr, len(c.hooks)-1)
}

func (c *conn) SetDeadline(t time.Time) error {
	return c.hooks.HookSetDeadline(t, c.raw.SetDeadline, len(c.hooks)-1)
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return c.hooks.HookSetReadDeadline(t, c.raw.SetReadDeadline, len(c.hooks)-1)
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.hooks.HookSetWriteDeadline(t, c.raw.SetWriteDeadline, len(c.hooks)-1)
}

// ConnHook conn hook
type ConnHook struct {
	Read             func(b []byte, raw func([]byte) (int, error)) (int, error)
	Write            func(b []byte, raw func([]byte) (int, error)) (int, error)
	Close            func(raw func() error) error
	LocalAddr        func(raw func() net.Addr) net.Addr
	RemoteAddr       func(raw func() net.Addr) net.Addr
	SetDeadline      func(raw func(t time.Time) error) error
	SetReadDeadline  func(raw func(t time.Time) error) error
	SetWriteDeadline func(raw func(t time.Time) error) error
}

//
// 倒序执行,后注册的先执行
type connHooks []*ConnHook

func (chs connHooks) HookRead(b []byte, raw func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx >= 0; idx-- {
		if chs[idx].Read != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(b)
	}
	return chs[idx].Read(b, func(b []byte) (int, error) {
		return chs.HookRead(b, raw, idx-1)
	})
}

func (chs connHooks) HookWrite(b []byte, raw func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx >= 0; idx-- {
		if chs[idx].Write != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(b)
	}
	return chs[idx].Write(b, func(b []byte) (int, error) {
		return chs.HookWrite(b, raw, idx-1)
	})
}

func (chs connHooks) HookClose(raw func() error, idx int) error {
	for ; idx >= 0; idx-- {
		if chs[idx].Close != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw()
	}
	return chs[idx].Close(func() error {
		return chs.HookClose(raw, idx-1)
	})
}

func (chs connHooks) HookLocalAddr(raw func() net.Addr, idx int) net.Addr {
	for ; idx >= 0; idx-- {
		if chs[idx].LocalAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw()
	}
	return chs[idx].LocalAddr(func() net.Addr {
		return chs.HookLocalAddr(raw, idx-1)
	})
}

func (chs connHooks) HookRemoteAddr(raw func() net.Addr, idx int) net.Addr {
	for ; idx >= 0; idx-- {
		if chs[idx].RemoteAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw()
	}
	return chs[idx].RemoteAddr(func() net.Addr {
		return chs.HookRemoteAddr(raw, idx-1)
	})
}

func (chs connHooks) HookSetDeadline(dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx >= 0; idx-- {
		if chs[idx].SetDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(dl)
	}
	return chs[idx].SetDeadline(func(dl time.Time) error {
		return chs.HookSetDeadline(dl, raw, idx-1)
	})
}

func (chs connHooks) HookSetReadDeadline(dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx >= 0; idx-- {
		if chs[idx].SetReadDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(dl)
	}
	return chs[idx].SetReadDeadline(func(dl time.Time) error {
		return chs.HookSetReadDeadline(dl, raw, idx-1)
	})
}

func (chs connHooks) HookSetWriteDeadline(dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx >= 0; idx-- {
		if chs[idx].SetWriteDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(dl)
	}
	return chs[idx].SetWriteDeadline(func(dl time.Time) error {
		return chs.HookSetWriteDeadline(dl, raw, idx-1)
	})
}

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

	dialCost int64

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
			defer func() {
				atomic.AddInt64(&ch.dialCost, time.Since(start).Nanoseconds())
			}()
			conn, err = fn(ctx, network, address)
			if err != nil {
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

// Reset 将所有状态数据重置为 0
func (ch *ConnStatHook) Reset() {
	atomic.StoreInt64(&ch.dialCost, 0)
	atomic.StoreInt64(&ch.readSize, 0)
	atomic.StoreInt64(&ch.readCost, 0)
	atomic.StoreInt64(&ch.writeSize, 0)
	atomic.StoreInt64(&ch.writeCost, 0)
}
