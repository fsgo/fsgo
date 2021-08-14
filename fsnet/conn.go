// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/14

package fsnet

import (
	"net"
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

var _ net.Conn = (*conn)(nil)

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
