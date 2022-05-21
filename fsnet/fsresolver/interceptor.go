// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsresolver

import (
	"context"
	"math/rand"
	"net"
	"time"
)

// Interceptor  Resolver Interceptor
type Interceptor struct {
	LookupIP func(ctx context.Context, network, host string, invoker LookupIPFunc) ([]net.IP, error)

	BeforeLookupIP func(ctx context.Context, network, host string) (ctxNew context.Context, networkNew, hostNew string)
	AfterLookupIP  func(ctx context.Context, network, host string, ips []net.IP, err error)

	LookupIPAddr func(ctx context.Context, host string, invoker LookupIPAddrFunc) ([]net.IPAddr, error)

	BeforeLookupIPAddr func(ctx context.Context, host string) (ctxNew context.Context, hostNew string)
	AfterLookupIPAddr  func(ctx context.Context, host string, addrs []net.IPAddr, err error)
}

type resolverItCtx struct {
	Ctx context.Context
	Its []*Interceptor
}

type ctxKey struct{}

var ctxKeyInterceptor = ctxKey{}

func (dc *resolverItCtx) All() []*Interceptor {
	var pits []*Interceptor
	if pic, ok := dc.Ctx.Value(ctxKeyInterceptor).(*resolverItCtx); ok {
		pits = pic.All()
	}
	if len(pits) == 0 {
		return dc.Its
	} else if len(dc.Its) == 0 {
		return pits
	}
	return append(pits, dc.Its...)
}

// ContextWithInterceptor set Resolver Interceptor to context
// these interceptors will exec before Dialer.Interceptors
func ContextWithInterceptor(ctx context.Context, its ...*Interceptor) context.Context {
	if len(its) == 0 {
		return ctx
	}
	val := &resolverItCtx{
		Ctx: ctx,
		Its: its,
	}
	return context.WithValue(ctx, ctxKeyInterceptor, val)
}

type interceptors []*Interceptor

func (rhs interceptors) CallLookupIP(ctx context.Context, network, host string, invoker LookupIPFunc,
	idx int) (ips []net.IP, err error) {
	if idx == 0 {
		for i := 0; i < len(rhs); i++ {
			if rhs[i].BeforeLookupIP == nil {
				continue
			}
			ctx, network, host = rhs[i].BeforeLookupIP(ctx, network, host)
		}

		defer func() {
			for i := 0; i < len(rhs); i++ {
				if rhs[i].AfterLookupIP == nil {
					continue
				}
				rhs[i].AfterLookupIP(ctx, network, host, ips, err)
			}
		}()
	}
	for ; idx < len(rhs); idx++ {
		if rhs[idx].LookupIP != nil {
			break
		}
	}
	if len(rhs) == 0 || idx >= len(rhs) {
		return invoker(ctx, network, host)
	}

	return rhs[idx].LookupIP(ctx, network, host, func(ctx context.Context, network string, host string) ([]net.IP, error) {
		return rhs.CallLookupIP(ctx, network, host, invoker, idx+1)
	})
}

func (rhs interceptors) CallLookupIPAddr(ctx context.Context, host string, invoker LookupIPAddrFunc,
	idx int) (addrs []net.IPAddr, err error) {
	if idx == 0 {
		for i := 0; i < len(rhs); i++ {
			if rhs[i].BeforeLookupIPAddr == nil {
				continue
			}
			ctx, host = rhs[i].BeforeLookupIPAddr(ctx, host)
		}

		defer func() {
			for i := 0; i < len(rhs); i++ {
				if rhs[i].AfterLookupIPAddr == nil {
					continue
				}
				rhs[i].AfterLookupIPAddr(ctx, host, addrs, err)
			}
		}()
	}
	for ; idx < len(rhs); idx++ {
		if rhs[idx].LookupIPAddr != nil {
			break
		}
	}

	if len(rhs) == 0 || idx >= len(rhs) {
		return invoker(ctx, host)
	}
	return rhs[idx].LookupIPAddr(ctx, host, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return rhs.CallLookupIPAddr(ctx, host, invoker, idx+1)
	})
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TryRegisterInterceptor 尝试给 Default 注册 Interceptor
// 若注册失败将返回 false
func TryRegisterInterceptor(its ...*Interceptor) bool {
	if d, ok := Default.(CanIntercept); ok {
		d.RegisterInterceptor(its...)
		return true
	}
	return false
}

// MustRegisterInterceptor 给 DefaultDialer 注册 DialerInterceptor
// 若不支持将 panic
func MustRegisterInterceptor(its ...*Interceptor) {
	if !TryRegisterInterceptor(its...) {
		panic("Default cannot Register Interceptor")
	}
}

// ToInterceptor convert Resolver to Interceptor
func ToInterceptor(r Resolver) *Interceptor {
	return &Interceptor{
		LookupIP: func(ctx context.Context, network, host string, invoker LookupIPFunc) ([]net.IP, error) {
			return r.LookupIP(ctx, network, host)
		},
		LookupIPAddr: func(ctx context.Context, host string, invoker LookupIPAddrFunc) ([]net.IPAddr, error) {
			return r.LookupIPAddr(ctx, host)
		},
	}
}
