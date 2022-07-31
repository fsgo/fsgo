// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsconn

import (
	"context"
	"net"
	"time"
)

type Info interface {
	// LocalAddr returns the local network address, if known.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address, if known.
	RemoteAddr() net.Addr
}

// Interceptor  for net.Conn
type Interceptor struct {
	Read      func(info Info, b []byte, invoker func([]byte) (int, error)) (int, error)
	AfterRead func(info Info, b []byte, readSize int, err error)

	Write      func(info Info, b []byte, invoker func([]byte) (int, error)) (int, error)
	AfterWrite func(info Info, b []byte, wroteSize int, err error)

	Close      func(info Info, invoker func() error) error
	AfterClose func(info Info, err error)

	LocalAddr  func(info Info, invoker func() net.Addr) net.Addr
	RemoteAddr func(info Info, invoker func() net.Addr) net.Addr

	SetDeadline      func(info Info, tm time.Time, invoker func(tm time.Time) error) error
	AfterSetDeadline func(info Info, tm time.Time, err error)

	SetReadDeadline      func(info Info, tm time.Time, invoker func(tm time.Time) error) error
	AfterSetReadDeadline func(info Info, tm time.Time, err error)

	SetWriteDeadline      func(info Info, tm time.Time, invoker func(tm time.Time) error) error
	AfterSetWriteDeadline func(info Info, tm time.Time, err error)
}

// 先注册的先执行
type interceptors []*Interceptor

func (chs interceptors) CallRead(info Info, b []byte, invoker func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Read != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(b)
	}

	return chs[idx].Read(info, b, func(b []byte) (int, error) {
		return chs.CallRead(info, b, invoker, idx+1)
	})
}

func (chs interceptors) CallWrite(info Info, b []byte, invoker func(b []byte) (int, error), idx int) (n int, err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Write != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(b)
	}
	return chs[idx].Write(info, b, func(b []byte) (int, error) {
		return chs.CallWrite(info, b, invoker, idx+1)
	})
}

func (chs interceptors) CallClose(info Info, invoker func() error, idx int) (err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].Close != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker()
	}
	return chs[idx].Close(info, func() error {
		return chs.CallClose(info, invoker, idx+1)
	})
}

func (chs interceptors) CallLocalAddr(info Info, invoker func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].LocalAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker()
	}
	return chs[idx].LocalAddr(info, func() net.Addr {
		return chs.CallLocalAddr(info, invoker, idx+1)
	})
}

func (chs interceptors) CallRemoteAddr(info Info, invoker func() net.Addr, idx int) net.Addr {
	for ; idx < len(chs); idx++ {
		if chs[idx].RemoteAddr != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker()
	}
	return chs[idx].RemoteAddr(info, func() net.Addr {
		return chs.CallRemoteAddr(info, invoker, idx+1)
	})
}

func (chs interceptors) CallSetDeadline(info Info, dl time.Time, invoker func(time.Time) error, idx int) (err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(dl)
	}
	return chs[idx].SetDeadline(info, dl, func(dl time.Time) error {
		return chs.CallSetDeadline(info, dl, invoker, idx+1)
	})
}

func (chs interceptors) CallSetReadDeadline(info Info, dl time.Time, invoker func(time.Time) error, idx int) (err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetReadDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(dl)
	}
	return chs[idx].SetReadDeadline(info, dl, func(dl time.Time) error {
		return chs.CallSetReadDeadline(info, dl, invoker, idx+1)
	})
}

func (chs interceptors) CallSetWriteDeadline(info Info, dl time.Time, invoker func(time.Time) error, idx int) (err error) {
	for ; idx < len(chs); idx++ {
		if chs[idx].SetWriteDeadline != nil {
			break
		}
	}
	if len(chs) == 0 || idx >= len(chs) {
		return invoker(dl)
	}
	return chs[idx].SetWriteDeadline(info, dl, func(dl time.Time) error {
		return chs.CallSetWriteDeadline(info, dl, invoker, idx+1)
	})
}

type ctxKey struct{}

var ctxKeyInterceptor = ctxKey{}

// ContextWithInterceptor set connWithIt interceptor to context
func ContextWithInterceptor(ctx context.Context, its ...*Interceptor) context.Context {
	if len(its) == 0 {
		return ctx
	}
	val := &connItCtx{
		Ctx: ctx,
		Its: its,
	}
	return context.WithValue(ctx, ctxKeyInterceptor, val)
}

// InterceptorsFromContext get connWithIt ConnInterceptors from context
func InterceptorsFromContext(ctx context.Context) []*Interceptor {
	if val, ok := ctx.Value(ctxKeyInterceptor).(*connItCtx); ok {
		return val.All()
	}
	return nil
}

type connItCtx struct {
	Ctx context.Context
	Its []*Interceptor
}

func (dc *connItCtx) All() []*Interceptor {
	var pits []*Interceptor
	if pic, ok := dc.Ctx.Value(ctxKeyInterceptor).(*connItCtx); ok {
		pits = pic.All()
	}
	if len(pits) == 0 {
		return dc.Its
	} else if len(dc.Its) == 0 {
		return pits
	}
	return append(pits, dc.Its...)
}

var globalConnIts []*Interceptor

// RegisterInterceptor  注册全局的 Interceptor
// 会在通过 ctx 注册的之前执行
func RegisterInterceptor(its ...*Interceptor) {
	globalConnIts = append(globalConnIts, its...)
}
