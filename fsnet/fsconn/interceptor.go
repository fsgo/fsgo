// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsconn

import (
	"context"
	"net"
	"time"
)

// Interceptor  for net.Conn
type Interceptor struct {
	Read      func(b []byte, invoker func([]byte) (int, error)) (int, error)
	AfterRead func(b []byte, readSize int, err error)

	Write      func(b []byte, invoker func([]byte) (int, error)) (int, error)
	AfterWrite func(b []byte, wroteSize int, err error)

	Close      func(invoker func() error) error
	AfterClose func(err error)

	LocalAddr  func(invoker func() net.Addr) net.Addr
	RemoteAddr func(invoker func() net.Addr) net.Addr

	SetDeadline      func(tm time.Time, invoker func(tm time.Time) error) error
	AfterSetDeadline func(tm time.Time, err error)

	SetReadDeadline      func(tm time.Time, invoker func(tm time.Time) error) error
	AfterSetReadDeadline func(tm time.Time, err error)

	SetWriteDeadline      func(tm time.Time, invoker func(tm time.Time) error) error
	AfterSetWriteDeadline func(tm time.Time, err error)
}

// 先注册的先执行
type interceptors []*Interceptor

func (chs interceptors) CallRead(b []byte, invoker func(b []byte) (int, error), idx int) (n int, err error) {
	if idx == 0 {
		defer func() {
			for i := 0; i < len(chs); i++ {
				if chs[i].AfterRead == nil {
					continue
				}
				chs[i].AfterRead(b, n, err)
			}
		}()
	}
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

func (chs interceptors) CallWrite(b []byte, invoker func(b []byte) (int, error), idx int) (n int, err error) {
	if idx == 0 {
		defer func() {
			for i := 0; i < len(chs); i++ {
				if chs[i].AfterWrite == nil {
					continue
				}
				chs[i].AfterWrite(b, n, err)
			}
		}()
	}
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

func (chs interceptors) CallClose(invoker func() error, idx int) (err error) {
	if idx == 0 {
		defer func() {
			for i := 0; i < len(chs); i++ {
				if chs[i].AfterClose == nil {
					continue
				}
				chs[i].AfterClose(err)
			}
		}()
	}
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

func (chs interceptors) CallLocalAddr(invoker func() net.Addr, idx int) net.Addr {
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

func (chs interceptors) CallRemoteAddr(invoker func() net.Addr, idx int) net.Addr {
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

func (chs interceptors) CallSetDeadline(dl time.Time, invoker func(time.Time) error, idx int) (err error) {
	if idx == 0 {
		defer func() {
			for i := 0; i < len(chs); i++ {
				if chs[i].AfterSetDeadline == nil {
					continue
				}
				chs[i].AfterSetDeadline(dl, err)
			}
		}()
	}
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

func (chs interceptors) CallSetReadDeadline(dl time.Time, invoker func(time.Time) error, idx int) (err error) {
	if idx == 0 {
		defer func() {
			for i := 0; i < len(chs); i++ {
				if chs[i].AfterSetReadDeadline == nil {
					continue
				}
				chs[i].AfterSetReadDeadline(dl, err)
			}
		}()
	}
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

func (chs interceptors) CallSetWriteDeadline(dl time.Time, invoker func(time.Time) error, idx int) (err error) {
	if idx == 0 {
		defer func() {
			for i := 0; i < len(chs); i++ {
				if chs[i].AfterSetWriteDeadline == nil {
					continue
				}
				chs[i].AfterSetWriteDeadline(dl, err)
			}
		}()
	}
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
