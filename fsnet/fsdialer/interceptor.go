// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsdialer

import (
	"context"
	"net"

	"github.com/fsgo/fsgo/fsnet/fsconn"
)

// Interceptor  dialer interceptor
type Interceptor struct {
	DialContext func(ctx context.Context, network string, address string, invoker DialContextFunc) (conn net.Conn, err error)

	BeforeDialContext func(ctx context.Context, network string, address string) (ctxNew context.Context, networkNew string, addressNew string)

	AfterDialContext func(ctx context.Context, network string, address string, conn net.Conn, err error)
}

type interceptors []*Interceptor

// CallDialContext 执行 its
func (dhs interceptors) CallDialContext(ctx context.Context, network, address string, invoker DialContextFunc, idx int) (conn net.Conn, err error) {
	if idx == 0 {
		for i := 0; i < len(dhs); i++ {
			if dhs[i].BeforeDialContext == nil {
				continue
			}
			ctx, network, address = dhs[i].BeforeDialContext(ctx, network, address)
		}

		defer func() {
			for i := 0; i < len(dhs); i++ {
				if dhs[i].AfterDialContext == nil {
					continue
				}
				dhs[i].AfterDialContext(ctx, network, address, conn, err)
			}
		}()
	}
	for ; idx < len(dhs); idx++ {
		if dhs[idx].DialContext != nil {
			break
		}
	}
	if len(dhs) == 0 || idx >= len(dhs) {
		return invoker(ctx, network, address)
	}
	return dhs[idx].DialContext(ctx, network, address, func(ctx context.Context, network string, address string) (net.Conn, error) {
		return dhs.CallDialContext(ctx, network, address, invoker, idx+1)
	})
}

type ctxKey struct{}

var ctxKeyInterceptor = ctxKey{}

type dialerItCtx struct {
	Ctx context.Context
	Its []*Interceptor
}

func (dc *dialerItCtx) All() []*Interceptor {
	var pits []*Interceptor
	if pic, ok := dc.Ctx.Value(ctxKeyInterceptor).(*dialerItCtx); ok {
		pits = pic.All()
	}
	if len(pits) == 0 {
		return dc.Its
	} else if len(dc.Its) == 0 {
		return pits
	}
	return append(pits, dc.Its...)
}

// ContextWithInterceptor set dialer Interceptor to context
// these interceptors will exec before Simple.Interceptors
func ContextWithInterceptor(ctx context.Context, its ...*Interceptor) context.Context {
	if len(its) == 0 {
		return ctx
	}
	val := &dialerItCtx{
		Ctx: ctx,
		Its: its,
	}
	return context.WithValue(ctx, ctxKeyInterceptor, val)
}

// InterceptorsFromContext get DialerInterceptors from contexts
func InterceptorsFromContext(ctx context.Context) []*Interceptor {
	if val, ok := ctx.Value(ctxKeyInterceptor).(*dialerItCtx); ok {
		return val.All()
	}
	return nil
}

// TryRegisterInterceptor 尝试给 Default 注册 Interceptor
// 若注册失败将返回 false
func TryRegisterInterceptor(its ...*Interceptor) bool {
	if d, ok := Default.(CanInterceptor); ok {
		d.RegisterInterceptor(its...)
		return true
	}
	return false
}

// MustRegisterInterceptor 给 Default 注册 Interceptor
// 若不支持将 panic
func MustRegisterInterceptor(its ...*Interceptor) {
	if !TryRegisterInterceptor(its...) {
		panic("Default cannot Register Interceptor")
	}
}

// TransConnInterceptor 创建一个支持添加 Interceptor 的 Interceptor
// 当想给 Default 注册 全局的 Interceptor 的时候，可以使用该方法
func TransConnInterceptor(its ...*fsconn.Interceptor) *Interceptor {
	return &Interceptor{
		DialContext: func(ctx context.Context, network string, address string, invoker DialContextFunc) (conn net.Conn, err error) {
			conn, err = invoker(ctx, network, address)
			if err != nil || len(its) == 0 {
				return conn, err
			}
			return fsconn.WithInterceptor(conn, its...), nil
		},
	}
}
