// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsnet

import (
	"context"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/fsgo/fscache"
	"github.com/fsgo/fscache/lrucache"

	"github.com/fsgo/fsgo/fsnet/internal"
)

// Resolver resolver type
type Resolver interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// HasLookupIP has  LookupIP func
type HasLookupIP interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
}

// ResolverCanIntercept 支持注册 ResolverInterceptor
type ResolverCanIntercept interface {
	RegisterInterceptor(its ...*ResolverInterceptor)
}

// LookupIPFunc lookupIP func type
type LookupIPFunc func(ctx context.Context, network, host string) ([]net.IP, error)

// LookupIPAddrFunc LookupIPAddr func type
type LookupIPAddrFunc func(ctx context.Context, host string) ([]net.IPAddr, error)

// ResolverCached Resolver with Cache
type ResolverCached struct {
	// Expiration cache Expiration
	// <=0 means disabled
	Expiration time.Duration

	StdResolver Resolver

	// Interceptors 可选，拦截器，先注册的后执行
	Interceptors []*ResolverInterceptor

	caches map[string]fscache.SCache
	mux    sync.Mutex
}

// LookupIP Lookup IP
func (r *ResolverCached) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return resolverInterceptors(r.Interceptors).CallLookupIP(ctx, network, host, r.lookupIP, 0)
}

func (r *ResolverCached) lookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	if ip, _ := internal.ParseIPZone(host); ip != nil {
		return []net.IP{ip}, nil
	}
	result, err := r.withCache(ctx, "LookupIP", network+host, func() (interface{}, error) {
		ret, err := r.getStdResolver().LookupIP(ctx, network, host)
		return ret, err
	})
	if err != nil {
		return nil, err
	}
	return result.([]net.IP), nil
}

// LookupIPAddr Lookup IPAddr
func (r *ResolverCached) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return resolverInterceptors(r.Interceptors).CallLookupIPAddr(ctx, host, r.lookupIPAddr, 0)
}

func (r *ResolverCached) lookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	if ip, zone := internal.ParseIPZone(host); ip != nil {
		return []net.IPAddr{
			{
				IP:   ip,
				Zone: zone,
			},
		}, nil
	}
	result, err := r.withCache(ctx, "LookupIPAddr", host, func() (interface{}, error) {
		ret, err := r.getStdResolver().LookupIPAddr(ctx, host)
		return ret, err
	})
	if err != nil {
		return nil, err
	}
	return result.([]net.IPAddr), nil
}

func (r *ResolverCached) getStdResolver() Resolver {
	if r.StdResolver != nil {
		return r.StdResolver
	}
	return net.DefaultResolver
}

func (r *ResolverCached) withCache(ctx context.Context, key string, cacheKey interface{}, fn func() (interface{}, error)) (interface{}, error) {
	if r.Expiration <= 0 {
		data, err := fn()
		return data, err
	}
	cache := r.getCache(key)
	cacheData := cache.Get(ctx, cacheKey)
	if cacheData.Err() == nil {
		var data interface{}
		if has, err := cacheData.Value(&data); has && err == nil {
			return data, nil
		}
	}
	data, err := fn()
	if err == nil {
		cache.Set(ctx, cacheKey, data, r.Expiration)
	}
	return data, err
}

func (r *ResolverCached) getCache(key string) fscache.SCache {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.caches == nil {
		r.caches = make(map[string]fscache.SCache)
	}
	if c, has := r.caches[key]; has {
		return c
	}
	c, err := lrucache.NewSCache(&lrucache.Option{Capacity: 128 * 1024})
	if err != nil {
		panic("init cache has error:" + err.Error())
	}
	r.caches[key] = c
	return c
}

var _ ResolverCanIntercept = (*ResolverCached)(nil)

// RegisterInterceptor Register Interceptor
func (r *ResolverCached) RegisterInterceptor(its ...*ResolverInterceptor) {
	r.Interceptors = append(r.Interceptors, its...)
}

// GetInterceptors read Interceptor list
func (r *ResolverCached) GetInterceptors() []*ResolverInterceptor {
	return r.Interceptors
}

// DefaultResolver default Resolver, result has 3 min cache
// 	Environment Variables 'FSGO_RESOLVER_EXP' can set the default cache lifetime
// 	eg: export FSGO_RESOLVER_EXP="10m" set cache lifetime as 10 minute
var DefaultResolver Resolver = &ResolverCached{
	Expiration: defaultResolverExpiration(),
}

func defaultResolverExpiration() time.Duration {
	val := os.Getenv("FSGO_RESOLVER_EXP")
	ts, _ := time.ParseDuration(val)
	if ts > time.Second {
		return ts
	}
	return 3 * time.Minute
}

// LookupIP DefaultResolver.LookupIP
var LookupIP = func(ctx context.Context, network, host string) ([]net.IP, error) {
	return DefaultResolver.LookupIP(ctx, network, host)
}

// LookupIPAddr DefaultResolver.LookupIPAddr
var LookupIPAddr = func(ctx context.Context, host string) ([]net.IPAddr, error) {
	return DefaultResolver.LookupIPAddr(ctx, host)
}

// ResolverInterceptor  Resolver Interceptor
type ResolverInterceptor struct {
	LookupIP func(ctx context.Context, network, host string, fn LookupIPFunc) ([]net.IP, error)

	LookupIPAddr func(ctx context.Context, host string, fn LookupIPAddrFunc) ([]net.IPAddr, error)
}

// ContextWithResolverInterceptor set Resolver Interceptor to context
// these interceptors will exec before Dialer.Interceptors
func ContextWithResolverInterceptor(ctx context.Context, its ...*ResolverInterceptor) context.Context {
	if len(its) == 0 {
		return ctx
	}
	dhm := resolverInterceptorMapperFormContext(ctx)
	if dhm == nil {
		dhm = &resolverInterceptorMapper{}
		ctx = context.WithValue(ctx, ctxKeyResolverInterceptor, dhm)
	}
	dhm.Register(its...)
	return ctx
}

func resolverInterceptorMapperFormContext(ctx context.Context) *resolverInterceptorMapper {
	val := ctx.Value(ctxKeyResolverInterceptor)
	if val == nil {
		return nil
	}
	return val.(*resolverInterceptorMapper)
}

type resolverInterceptorMapper struct {
	its resolverInterceptors
}

func (rhm *resolverInterceptorMapper) Register(its ...*ResolverInterceptor) {
	rhm.its = append(rhm.its, its...)
}

type resolverInterceptors []*ResolverInterceptor

func (rhs resolverInterceptors) CallLookupIP(ctx context.Context, network, host string, fn LookupIPFunc, idx int) ([]net.IP, error) {
	for ; idx < len(rhs); idx++ {
		if rhs[idx].LookupIP != nil {
			break
		}
	}
	if len(rhs) == 0 || idx >= len(rhs) {
		return fn(ctx, network, host)
	}

	return rhs[idx].LookupIP(ctx, network, host, func(ctx context.Context, network string, host string) ([]net.IP, error) {
		return rhs.CallLookupIP(ctx, network, host, fn, idx+1)
	})
}

func (rhs resolverInterceptors) CallLookupIPAddr(ctx context.Context, host string, fn LookupIPAddrFunc, idx int) ([]net.IPAddr, error) {
	for ; idx < len(rhs); idx++ {
		if rhs[idx].LookupIPAddr != nil {
			break
		}
	}

	if len(rhs) == 0 || idx >= len(rhs) {
		return fn(ctx, host)
	}
	return rhs[idx].LookupIPAddr(ctx, host, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return rhs.CallLookupIPAddr(ctx, host, fn, idx+1)
	})
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TryRegisterResolverInterceptor 尝试给 DefaultResolver 注册 ResolverInterceptor
// 若注册失败将返回 false
func TryRegisterResolverInterceptor(its ...*ResolverInterceptor) bool {
	if d, ok := DefaultResolver.(ResolverCanIntercept); ok {
		d.RegisterInterceptor(its...)
		return true
	}
	return false
}

// MustRegisterResolverInterceptor 给 DefaultDialer 注册 DialerInterceptor
// 若不支持将 panic
func MustRegisterResolverInterceptor(its ...*ResolverInterceptor) {
	if !TryRegisterResolverInterceptor(its...) {
		panic("DefaultResolver cannot Register Interceptor")
	}
}
