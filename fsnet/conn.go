// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/14

package fsnet

import (
	"context"
	"net"
	"time"
)

var _ net.Conn = (*conn)(nil)

// WrapConn wrap conn with ConnInterceptor
// its 将倒序执行：后注册的先执行
func WrapConn(c net.Conn, its ...*ConnInterceptor) net.Conn {
	if rc, ok := c.(*conn); ok {
		rc.its = append(rc.its, its...)
		return rc
	}

	nc := &conn{
		raw: c,
		its: its,
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
	raw net.Conn
	its connInterceptors
}

func (c *conn) Raw() net.Conn {
	return c.raw
}

func (c *conn) Read(b []byte) (n int, err error) {
	return c.its.CallRead(b, c.raw.Read, len(c.its)-1)
}

func (c *conn) Write(b []byte) (n int, err error) {
	return c.its.CallWrite(b, c.raw.Write, len(c.its)-1)
}

func (c *conn) Close() error {
	return c.its.CallClose(c.raw.Close, len(c.its)-1)
}

func (c *conn) LocalAddr() net.Addr {
	return c.its.CallLocalAddr(c.raw.LocalAddr, len(c.its)-1)
}

func (c *conn) RemoteAddr() net.Addr {
	return c.its.CallRemoteAddr(c.raw.RemoteAddr, len(c.its)-1)
}

func (c *conn) SetDeadline(t time.Time) error {
	return c.its.CallSetDeadline(t, c.raw.SetDeadline, len(c.its)-1)
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return c.its.CallSetReadDeadline(t, c.raw.SetReadDeadline, len(c.its)-1)
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.its.CallSetWriteDeadline(t, c.raw.SetWriteDeadline, len(c.its)-1)
}

// ConnInterceptor conn Interceptor
type ConnInterceptor struct {
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
type connInterceptors []*ConnInterceptor

func (chs connInterceptors) CallRead(b []byte, raw func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx >= 0; idx-- {
		if chs[idx].Read != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(b)
	}
	return chs[idx].Read(b, func(b []byte) (int, error) {
		return chs.CallRead(b, raw, idx-1)
	})
}

func (chs connInterceptors) CallWrite(b []byte, raw func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx >= 0; idx-- {
		if chs[idx].Write != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(b)
	}
	return chs[idx].Write(b, func(b []byte) (int, error) {
		return chs.CallWrite(b, raw, idx-1)
	})
}

func (chs connInterceptors) CallClose(raw func() error, idx int) error {
	for ; idx >= 0; idx-- {
		if chs[idx].Close != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw()
	}
	return chs[idx].Close(func() error {
		return chs.CallClose(raw, idx-1)
	})
}

func (chs connInterceptors) CallLocalAddr(raw func() net.Addr, idx int) net.Addr {
	for ; idx >= 0; idx-- {
		if chs[idx].LocalAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw()
	}
	return chs[idx].LocalAddr(func() net.Addr {
		return chs.CallLocalAddr(raw, idx-1)
	})
}

func (chs connInterceptors) CallRemoteAddr(raw func() net.Addr, idx int) net.Addr {
	for ; idx >= 0; idx-- {
		if chs[idx].RemoteAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw()
	}
	return chs[idx].RemoteAddr(func() net.Addr {
		return chs.CallRemoteAddr(raw, idx-1)
	})
}

func (chs connInterceptors) CallSetDeadline(dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx >= 0; idx-- {
		if chs[idx].SetDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(dl)
	}
	return chs[idx].SetDeadline(func(dl time.Time) error {
		return chs.CallSetDeadline(dl, raw, idx-1)
	})
}

func (chs connInterceptors) CallSetReadDeadline(dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx >= 0; idx-- {
		if chs[idx].SetReadDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(dl)
	}
	return chs[idx].SetReadDeadline(func(dl time.Time) error {
		return chs.CallSetReadDeadline(dl, raw, idx-1)
	})
}

func (chs connInterceptors) CallSetWriteDeadline(dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx >= 0; idx-- {
		if chs[idx].SetWriteDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx < 0 {
		return raw(dl)
	}
	return chs[idx].SetWriteDeadline(func(dl time.Time) error {
		return chs.CallSetWriteDeadline(dl, raw, idx-1)
	})
}

// ContextWithConnInterceptor set conn interceptor to context
func ContextWithConnInterceptor(ctx context.Context, hooks ...*ConnInterceptor) context.Context {
	if len(hooks) == 0 {
		return ctx
	}
	dh := connHookMapperFormContext(ctx)
	if dh == nil {
		dh = &connHookMapper{}
		ctx = context.WithValue(ctx, ctxKeyConnHook, dh)
	}
	dh.Register(hooks...)
	return ctx
}

// ConnInterceptorsFromContext get conn its from context
func ConnInterceptorsFromContext(ctx context.Context) []*ConnInterceptor {
	chm := connHookMapperFormContext(ctx)
	if chm == nil {
		return nil
	}
	return chm.hooks
}

func connHookMapperFormContext(ctx context.Context) *connHookMapper {
	val := ctx.Value(ctxKeyConnHook)
	if val == nil {
		return nil
	}
	return val.(*connHookMapper)
}

type connHookMapper struct {
	hooks connInterceptors
}

func (chm *connHookMapper) Register(hooks ...*ConnInterceptor) {
	chm.hooks = append(chm.hooks, hooks...)
}
