// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsresolver

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/fsgo/fsgo/fssync"
)

// Interceptor  Resolver Interceptor
type Interceptor struct {
	LookupIP func(ctx context.Context, network, host string, invoker LookupIPFunc) ([]net.IP, error)

	BeforeLookupIP func(ctx context.Context, network, host string) (c context.Context, n, h string)
	AfterLookupIP  func(ctx context.Context, network, host string, ips []net.IP, err error) ([]net.IP, error)

	LookupIPAddr func(ctx context.Context, host string, invoker LookupIPAddrFunc) ([]net.IPAddr, error)

	BeforeLookupIPAddr func(ctx context.Context, host string) (ctxNew context.Context, hostNew string)
	AfterLookupIPAddr  func(ctx context.Context, host string, addrs []net.IPAddr, err error) ([]net.IPAddr, error)
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

// CacheInterceptor 总是缓存查询结果
type CacheInterceptor struct {
	ipCache           fssync.Map[string, []net.IP]
	ipAddrCache       fssync.Map[string, []net.IPAddr]
	cntFBLookupIP     atomic.Int64 // 查询失败，使用 cache 的计数
	cntFBLookupIPAddr atomic.Int64 // 查询失败，使用 cache 的计数
	disabled          atomic.Bool
}

func (ci *CacheInterceptor) Interceptor() *Interceptor {
	return &Interceptor{
		AfterLookupIP:     ci.afterLookupIP,
		AfterLookupIPAddr: ci.afterLookupIPAddr,
	}
}

// SetEnable 设置状态，默认是 enable 的
func (ci *CacheInterceptor) SetEnable(enable bool) {
	ci.disabled.Store(enable)
}

func (ci *CacheInterceptor) afterLookupIP(ctx context.Context, network, host string, ips []net.IP, err error) ([]net.IP, error) {
	if ci.disabled.Load() {
		return ips, err
	}
	key := network + "#" + host
	if err == nil {
		ci.ipCache.Store(key, ips)
		return ips, nil
	}
	cv, ok := ci.ipCache.Load(key)
	if ok {
		ci.cntFBLookupIP.Add(1)
		return cv, nil
	}
	return ips, err
}

func (ci *CacheInterceptor) afterLookupIPAddr(ctx context.Context, host string, addrs []net.IPAddr, err error) ([]net.IPAddr, error) {
	if ci.disabled.Load() {
		return addrs, err
	}
	if err == nil {
		ci.ipAddrCache.Store(host, addrs)
		return addrs, nil
	}
	cv, ok := ci.ipAddrCache.Load(host)
	if ok {
		ci.cntFBLookupIPAddr.Add(1)
		return cv, nil
	}
	return addrs, err
}

func (ci *CacheInterceptor) Stats() map[string]int64 {
	return map[string]int64{
		"LookupIPFallback":     ci.cntFBLookupIP.Load(),
		"LookupIPAddrFallback": ci.cntFBLookupIPAddr.Load(),
	}
}
