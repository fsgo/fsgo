// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsnet

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/fsgo/fscache"
	"github.com/fsgo/fscache/lrucache"
)

// Resolver resolver type
type Resolver interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
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

	Hooks []*ResolverHook

	caches map[string]fscache.SCache
	mux    sync.Mutex
}

// LookupIP Lookup IP
func (r *ResolverCached) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return resolverHooks(r.Hooks).HookLookupIP(ctx, network, host, r.lookupIP, len(r.Hooks)-1)
}

func (r *ResolverCached) lookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
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
	return resolverHooks(r.Hooks).HookLookupIPAddr(ctx, host, r.lookupIPAddr, len(r.Hooks)-1)
}

func (r *ResolverCached) lookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
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
	c, err := lrucache.NewSCache(&lrucache.Option{Capacity: 128})
	if err != nil {
		panic("init cache has error:" + err.Error())
	}
	r.caches[key] = c
	return c
}

// RegisterHook Register Hook
func (r *ResolverCached) RegisterHook(hooks ...*ResolverHook) {
	r.Hooks = append(r.Hooks, hooks...)
}

// DefaultResolver default Resolver,result has 1 min cache
var DefaultResolver Resolver = &ResolverCached{
	Expiration: time.Minute,
}

// LookupIP DefaultResolver.LookupIP
var LookupIP = func(ctx context.Context, network, host string) ([]net.IP, error) {
	return DefaultResolver.LookupIP(ctx, network, host)
}

// LookupIPAddr DefaultResolver.LookupIPAddr
var LookupIPAddr = func(ctx context.Context, host string) ([]net.IPAddr, error) {
	return DefaultResolver.LookupIPAddr(ctx, host)
}

// ResolverHook  Resolver Hook
type ResolverHook struct {
	LookupIP     func(ctx context.Context, network, host string, fn LookupIPFunc) ([]net.IP, error)
	LookupIPAddr func(ctx context.Context, host string, fn LookupIPAddrFunc) ([]net.IPAddr, error)
}

// ContextWithResolverHook set Resolver Hook to context
// these hooks will exec before Dialer.Hooks
func ContextWithResolverHook(ctx context.Context, hooks ...*ResolverHook) context.Context {
	if len(hooks) == 0 {
		return ctx
	}
	dhm := resolverHookMapperFormContext(ctx)
	if dhm == nil {
		dhm = &resolverHookMapper{}
		ctx = context.WithValue(ctx, ctxKeyResolverHook, dhm)
	}
	dhm.Register(hooks...)
	return ctx
}

func resolverHookMapperFormContext(ctx context.Context) *resolverHookMapper {
	val := ctx.Value(ctxKeyResolverHook)
	if val == nil {
		return nil
	}
	return val.(*resolverHookMapper)
}

type resolverHookMapper struct {
	hooks resolverHooks
}

func (rhm *resolverHookMapper) Register(hooks ...*ResolverHook) {
	rhm.hooks = append(rhm.hooks, hooks...)
}

type resolverHooks []*ResolverHook

func (rhs resolverHooks) HookLookupIP(ctx context.Context, network, host string, fn LookupIPFunc, idx int) ([]net.IP, error) {
	for ; idx >= 0; idx-- {
		if rhs[idx].LookupIP != nil {
			break
		}
	}
	if len(rhs) == 0 || idx < 0 {
		return fn(ctx, network, host)
	}

	return rhs[idx].LookupIP(ctx, network, host, func(ctx context.Context, network string, host string) ([]net.IP, error) {
		return rhs.HookLookupIP(ctx, network, host, fn, idx-1)
	})
}

func (rhs resolverHooks) HookLookupIPAddr(ctx context.Context, host string, fn LookupIPAddrFunc, idx int) ([]net.IPAddr, error) {
	for ; idx >= 0; idx-- {
		if rhs[idx].LookupIPAddr != nil {
			break
		}
	}

	if len(rhs) == 0 || idx < 0 {
		return fn(ctx, host)
	}
	return rhs[idx].LookupIPAddr(ctx, host, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return rhs.HookLookupIPAddr(ctx, host, fn, idx-1)
	})
}
