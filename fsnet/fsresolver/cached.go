// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsresolver

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/fsgo/fscache"
	"github.com/fsgo/fscache/lrucache"

	"github.com/fsgo/fsgo/fsnet/internal"
)

// Cached Resolver with Cache
type Cached struct {
	// Expiration cache Expiration
	// <=0 means disabled
	Expiration time.Duration

	Invoker Resolver

	// Interceptors 可选，拦截器，先注册的后执行
	Interceptors []*Interceptor

	caches map[string]fscache.SCache
	mux    sync.Mutex
}

// LookupIP Lookup IP
func (r *Cached) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return interceptors(r.Interceptors).CallLookupIP(ctx, network, host, r.lookupIP, 0)
}

func (r *Cached) lookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
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
func (r *Cached) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return interceptors(r.Interceptors).CallLookupIPAddr(ctx, host, r.lookupIPAddr, 0)
}

func (r *Cached) lookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
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

func (r *Cached) getStdResolver() Resolver {
	if r.Invoker != nil {
		return r.Invoker
	}
	return net.DefaultResolver
}

func (r *Cached) withCache(ctx context.Context, key string, cacheKey interface{},
	fn func() (interface{}, error)) (interface{}, error) {
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

func (r *Cached) getCache(key string) fscache.SCache {
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

var _ CanIntercept = (*Cached)(nil)

// RegisterInterceptor Register Interceptor
func (r *Cached) RegisterInterceptor(its ...*Interceptor) {
	r.Interceptors = append(r.Interceptors, its...)
}

// GetInterceptors read Interceptor list
func (r *Cached) GetInterceptors() []*Interceptor {
	return r.Interceptors
}

// ExpirationFromEnv parser Expiration from os.env
func (r *Cached) ExpirationFromEnv() time.Duration {
	return defaultResolverExpiration()
}
