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
		nc := &connWithIt{
			raw:  rc.raw,
			args: append(rc.args, its...),
		}
		nc.allIts = append(globalConnIts, nc.args...)
		return nc
	}

	nc := &connWithIt{
		raw:    c,
		allIts: append(globalConnIts, its...),
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
	return c.allIts.CallRead(b, c.raw.Read, 0)
}

func (c *connWithIt) Write(b []byte) (n int, err error) {
	return c.allIts.CallWrite(b, c.raw.Write, 0)
}

func (c *connWithIt) Close() error {
	return c.allIts.CallClose(c.raw.Close, 0)
}

func (c *connWithIt) LocalAddr() net.Addr {
	return c.allIts.CallLocalAddr(c.raw.LocalAddr, 0)
}

func (c *connWithIt) RemoteAddr() net.Addr {
	return c.allIts.CallRemoteAddr(c.raw.RemoteAddr, 0)
}

func (c *connWithIt) SetDeadline(t time.Time) error {
	return c.allIts.CallSetDeadline(t, c.raw.SetDeadline, 0)
}

func (c *connWithIt) SetReadDeadline(t time.Time) error {
	return c.allIts.CallSetReadDeadline(t, c.raw.SetReadDeadline, 0)
}

func (c *connWithIt) SetWriteDeadline(t time.Time) error {
	return c.allIts.CallSetWriteDeadline(t, c.raw.SetWriteDeadline, 0)
}

// ConnInterceptor connWithIt Interceptor
type ConnInterceptor struct {
	Read             func(b []byte, invoker func([]byte) (int, error)) (int, error)
	Write            func(b []byte, invoker func([]byte) (int, error)) (int, error)
	Close            func(invoker func() error) error
	LocalAddr        func(invoker func() net.Addr) net.Addr
	RemoteAddr       func(invoker func() net.Addr) net.Addr
	SetDeadline      func(tm time.Time, invoker func(tm time.Time) error) error
	SetReadDeadline  func(tm time.Time, invoker func(tm time.Time) error) error
	SetWriteDeadline func(tm time.Time, invoker func(tm time.Time) error) error
}

// 先注册的先执行
type connInterceptors []*ConnInterceptor

func (chs connInterceptors) CallRead(b []byte, invoker func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Read != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(b)
	}
	return chs[idx].Read(b, func(b []byte) (int, error) {
		return chs.CallRead(b, invoker, idx+1)
	})
}

func (chs connInterceptors) CallWrite(b []byte, invoker func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Write != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(b)
	}
	return chs[idx].Write(b, func(b []byte) (int, error) {
		return chs.CallWrite(b, invoker, idx+1)
	})
}

func (chs connInterceptors) CallClose(invoker func() error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].Close != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker()
	}
	return chs[idx].Close(func() error {
		return chs.CallClose(invoker, idx+1)
	})
}

func (chs connInterceptors) CallLocalAddr(invoker func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].LocalAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker()
	}
	return chs[idx].LocalAddr(func() net.Addr {
		return chs.CallLocalAddr(invoker, idx+1)
	})
}

func (chs connInterceptors) CallRemoteAddr(invoker func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].RemoteAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker()
	}
	return chs[idx].RemoteAddr(func() net.Addr {
		return chs.CallRemoteAddr(invoker, idx+1)
	})
}

func (chs connInterceptors) CallSetDeadline(dl time.Time, invoker func(time.Time) error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(dl)
	}
	return chs[idx].SetDeadline(dl, func(dl time.Time) error {
		return chs.CallSetDeadline(dl, invoker, idx+1)
	})
}

func (chs connInterceptors) CallSetReadDeadline(dl time.Time, invoker func(time.Time) error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetReadDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(dl)
	}
	return chs[idx].SetReadDeadline(dl, func(dl time.Time) error {
		return chs.CallSetReadDeadline(dl, invoker, idx+1)
	})
}

func (chs connInterceptors) CallSetWriteDeadline(dl time.Time, invoker func(time.Time) error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetWriteDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(dl)
	}
	return chs[idx].SetWriteDeadline(dl, func(dl time.Time) error {
		return chs.CallSetWriteDeadline(dl, invoker, idx+1)
	})
}

// ContextWithConnInterceptor set connWithIt interceptor to context
func ContextWithConnInterceptor(ctx context.Context, its ...*ConnInterceptor) context.Context {
	if len(its) == 0 {
		return ctx
	}
	val := &connItCtx{
		Ctx: ctx,
		Its: its,
	}
	return context.WithValue(ctx, ctxKeyConnInterceptor, val)
}

// ConnInterceptorsFromContext get connWithIt ConnInterceptors from context
func ConnInterceptorsFromContext(ctx context.Context) []*ConnInterceptor {
	if val, ok := ctx.Value(ctxKeyConnInterceptor).(*connItCtx); ok {
		return val.All()
	}
	return nil
}

type connItCtx struct {
	Ctx context.Context
	Its []*ConnInterceptor
}

func (dc *connItCtx) All() []*ConnInterceptor {
	var pits []*ConnInterceptor
	if pic, ok := dc.Ctx.Value(ctxKeyConnInterceptor).(*connItCtx); ok {
		pits = pic.All()
	}
	if len(pits) == 0 {
		return dc.Its
	} else if len(dc.Its) == 0 {
		return pits
	}
	return append(pits, dc.Its...)
}

var globalConnIts []*ConnInterceptor

// RegisterConnInterceptor  注册全局的 ConnInterceptor
// 会在通过 ctx 注册的之前执行
func RegisterConnInterceptor(its ...*ConnInterceptor) {
	globalConnIts = append(globalConnIts, its...)
}
