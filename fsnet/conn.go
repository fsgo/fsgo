// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/14

package fsnet

import (
	"context"
	"net"
	"time"
)

var _ net.Conn = (*connWithIt)(nil)

// WrapConn wrap connWithIt with ConnInterceptor
// its 将倒序执行：后注册的先执行
func WrapConn(c net.Conn, its ...*ConnInterceptor) net.Conn {
	if rc, ok := c.(*connWithIt); ok {
		cn := &connWithIt{
			raw:  rc.raw,
			args: append(rc.args, its...),
		}
		cn.allIts = append(globalConnInterceptors, cn.args...)
		return cn
	}

	nc := &connWithIt{
		raw:    c,
		allIts: append(globalConnInterceptors, its...),
		args:   its,
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

var _ net.Conn = (*connWithIt)(nil)
var _ HasRawConn = (*connWithIt)(nil)

type connWithIt struct {
	raw net.Conn
	// 包好了全局和创建时传入的拦截器
	allIts connInterceptors

	// 创建时传入的拦截器
	args connInterceptors
}

func (c *connWithIt) Raw() net.Conn {
	return c.raw
}

func (c *connWithIt) Read(b []byte) (n int, err error) {
	return c.allIts.CallRead(c.raw, b, c.raw.Read, 0)
}

func (c *connWithIt) Write(b []byte) (n int, err error) {
	return c.allIts.CallWrite(c.raw, b, c.raw.Write, 0)
}

func (c *connWithIt) Close() error {
	return c.allIts.CallClose(c.raw, c.raw.Close, 0)
}

func (c *connWithIt) LocalAddr() net.Addr {
	return c.allIts.CallLocalAddr(c.raw, c.raw.LocalAddr, 0)
}

func (c *connWithIt) RemoteAddr() net.Addr {
	return c.allIts.CallRemoteAddr(c.raw, c.raw.RemoteAddr, 0)
}

func (c *connWithIt) SetDeadline(t time.Time) error {
	return c.allIts.CallSetDeadline(c.raw, t, c.raw.SetDeadline, 0)
}

func (c *connWithIt) SetReadDeadline(t time.Time) error {
	return c.allIts.CallSetReadDeadline(c.raw, t, c.raw.SetReadDeadline, 0)
}

func (c *connWithIt) SetWriteDeadline(t time.Time) error {
	return c.allIts.CallSetWriteDeadline(c.raw, t, c.raw.SetWriteDeadline, 0)
}

// ConnInterceptor connWithIt Interceptor
type ConnInterceptor struct {
	Read             func(c net.Conn, b []byte, raw func([]byte) (int, error)) (int, error)
	Write            func(c net.Conn, b []byte, raw func([]byte) (int, error)) (int, error)
	Close            func(c net.Conn, raw func() error) error
	LocalAddr        func(c net.Conn, raw func() net.Addr) net.Addr
	RemoteAddr       func(c net.Conn, raw func() net.Addr) net.Addr
	SetDeadline      func(c net.Conn, tm time.Time, raw func(tm time.Time) error) error
	SetReadDeadline  func(c net.Conn, tm time.Time, raw func(tm time.Time) error) error
	SetWriteDeadline func(c net.Conn, tm time.Time, raw func(tm time.Time) error) error
}

// 先注册的先执行
type connInterceptors []*ConnInterceptor

func (chs connInterceptors) CallRead(c net.Conn, b []byte, raw func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Read != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(b)
	}
	return chs[idx].Read(c, b, func(b []byte) (int, error) {
		return chs.CallRead(c, b, raw, idx+1)
	})
}

func (chs connInterceptors) CallWrite(c net.Conn, b []byte, raw func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Write != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(b)
	}
	return chs[idx].Write(c, b, func(b []byte) (int, error) {
		return chs.CallWrite(c, b, raw, idx+1)
	})
}

func (chs connInterceptors) CallClose(c net.Conn, raw func() error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].Close != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw()
	}
	return chs[idx].Close(c, func() error {
		return chs.CallClose(c, raw, idx+1)
	})
}

func (chs connInterceptors) CallLocalAddr(c net.Conn, raw func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].LocalAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw()
	}
	return chs[idx].LocalAddr(c, func() net.Addr {
		return chs.CallLocalAddr(c, raw, idx+1)
	})
}

func (chs connInterceptors) CallRemoteAddr(c net.Conn, raw func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].RemoteAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw()
	}
	return chs[idx].RemoteAddr(c, func() net.Addr {
		return chs.CallRemoteAddr(c, raw, idx+1)
	})
}

func (chs connInterceptors) CallSetDeadline(c net.Conn, dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(dl)
	}
	return chs[idx].SetDeadline(c, dl, func(dl time.Time) error {
		return chs.CallSetDeadline(c, dl, raw, idx+1)
	})
}

func (chs connInterceptors) CallSetReadDeadline(c net.Conn, dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetReadDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(dl)
	}
	return chs[idx].SetReadDeadline(c, dl, func(dl time.Time) error {
		return chs.CallSetReadDeadline(c, dl, raw, idx+1)
	})
}

func (chs connInterceptors) CallSetWriteDeadline(c net.Conn, dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetWriteDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(dl)
	}
	return chs[idx].SetWriteDeadline(c, dl, func(dl time.Time) error {
		return chs.CallSetWriteDeadline(c, dl, raw, idx+1)
	})
}

// ContextWithConnInterceptor set connWithIt interceptor to context
func ContextWithConnInterceptor(ctx context.Context, its ...*ConnInterceptor) context.Context {
	if len(its) == 0 {
		return ctx
	}
	dh := connHookMapperFormContext(ctx)
	if dh == nil {
		dh = &connHookMapper{}
		ctx = context.WithValue(ctx, ctxKeyConnInterceptor, dh)
	}
	dh.Register(its...)
	return ctx
}

// ConnInterceptorsFromContext get connWithIt ConnInterceptors from context
func ConnInterceptorsFromContext(ctx context.Context) []*ConnInterceptor {
	chm := connHookMapperFormContext(ctx)
	if chm == nil {
		return nil
	}
	return chm.its
}

func connHookMapperFormContext(ctx context.Context) *connHookMapper {
	val := ctx.Value(ctxKeyConnInterceptor)
	if val == nil {
		return nil
	}
	return val.(*connHookMapper)
}

type connHookMapper struct {
	its connInterceptors
}

func (chm *connHookMapper) Register(its ...*ConnInterceptor) {
	chm.its = append(chm.its, its...)
}

var globalConnInterceptors []*ConnInterceptor

// RegisterConnInterceptor  注册全局的 ConnInterceptor
// 会在通过 ctx 注册的之前执行
func RegisterConnInterceptor(its ...*ConnInterceptor) {
	globalConnInterceptors = append(globalConnInterceptors, its...)
}
