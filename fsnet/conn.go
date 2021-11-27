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
		nc.allIts = append(globalConnInterceptors, nc.args...)
		return nc
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
	Read             func(b []byte, raw func([]byte) (int, error)) (int, error)
	Write            func(b []byte, raw func([]byte) (int, error)) (int, error)
	Close            func(raw func() error) error
	LocalAddr        func(raw func() net.Addr) net.Addr
	RemoteAddr       func(raw func() net.Addr) net.Addr
	SetDeadline      func(tm time.Time, raw func(tm time.Time) error) error
	SetReadDeadline  func(tm time.Time, raw func(tm time.Time) error) error
	SetWriteDeadline func(tm time.Time, raw func(tm time.Time) error) error
}

// 先注册的先执行
type connInterceptors []*ConnInterceptor

func (chs connInterceptors) CallRead(b []byte, raw func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Read != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(b)
	}
	return chs[idx].Read(b, func(b []byte) (int, error) {
		return chs.CallRead(b, raw, idx+1)
	})
}

func (chs connInterceptors) CallWrite(b []byte, raw func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Write != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(b)
	}
	return chs[idx].Write(b, func(b []byte) (int, error) {
		return chs.CallWrite(b, raw, idx+1)
	})
}

func (chs connInterceptors) CallClose(raw func() error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].Close != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw()
	}
	return chs[idx].Close(func() error {
		return chs.CallClose(raw, idx+1)
	})
}

func (chs connInterceptors) CallLocalAddr(raw func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].LocalAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw()
	}
	return chs[idx].LocalAddr(func() net.Addr {
		return chs.CallLocalAddr(raw, idx+1)
	})
}

func (chs connInterceptors) CallRemoteAddr(raw func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].RemoteAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw()
	}
	return chs[idx].RemoteAddr(func() net.Addr {
		return chs.CallRemoteAddr(raw, idx+1)
	})
}

func (chs connInterceptors) CallSetDeadline(dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(dl)
	}
	return chs[idx].SetDeadline(dl, func(dl time.Time) error {
		return chs.CallSetDeadline(dl, raw, idx+1)
	})
}

func (chs connInterceptors) CallSetReadDeadline(dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetReadDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(dl)
	}
	return chs[idx].SetReadDeadline(dl, func(dl time.Time) error {
		return chs.CallSetReadDeadline(dl, raw, idx+1)
	})
}

func (chs connInterceptors) CallSetWriteDeadline(dl time.Time, raw func(time.Time) error, idx int) error {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetWriteDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return raw(dl)
	}
	return chs[idx].SetWriteDeadline(dl, func(dl time.Time) error {
		return chs.CallSetWriteDeadline(dl, raw, idx+1)
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
