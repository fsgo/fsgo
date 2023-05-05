// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsdialer

import (
	"context"
	"net"

	"github.com/fsgo/fsgo/fsnet/fsconn"
	"github.com/fsgo/fsgo/internal/xctx"
)

// Interceptor  dialer interceptor
type Interceptor struct {
	DialContext func(ctx context.Context, network string, address string, invoker DialContextFunc) (net.Conn, error)

	BeforeDialContext func(ctx context.Context, net string, addr string) (c context.Context, n string, a string)

	AfterDialContext func(ctx context.Context, net string, addr string, conn net.Conn, err error) (net.Conn, error)
}

type interceptors []*Interceptor

// CallDialContext 执行 its
func (dhs interceptors) CallDialContext(ctx context.Context, network, address string, invoker DialContextFunc, idx int) (conn net.Conn, err error) {
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

// ContextWithInterceptor set dialer Interceptor to context
// these interceptors will exec before Simple.Interceptors
func ContextWithInterceptor(ctx context.Context, its ...*Interceptor) context.Context {
	return xctx.WithValues(ctx, ctxKeyInterceptor, its...)
}

// InterceptorsFromContext get DialerInterceptors from contexts
func InterceptorsFromContext(ctx context.Context) []*Interceptor {
	return xctx.Values[ctxKey, *Interceptor](ctx, ctxKeyInterceptor)
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
		AfterDialContext: func(ctx context.Context, net string, addr string, conn net.Conn, err error) (net.Conn, error) {
			if err != nil || len(its) == 0 {
				return conn, err
			}
			return fsconn.WithInterceptor(conn, its...), nil
		},
	}
}
