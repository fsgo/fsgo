/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/01/10
 */

package resolver

import (
	"context"
	"net"
	"sync"
	"time"
)

// LookupIPer LookupIP
type LookupIPer interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
}

// LookupIPAddrer LookupIPAddr
type LookupIPAddrer interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// LookupHoster LookupHost
type LookupHoster interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
}

var DefaultCachedResolver = NewCachedResolver(30*time.Second, 5*time.Minute)

func NewCachedResolver(update time.Duration, expire time.Duration) *CachedResolver {
	r := &CachedResolver{
		lookupIPer:     net.DefaultResolver,
		UpdateInterval: update,
		Expiration:     expire,
	}
	r.ipAddrCache = &withCache{
		GetFunc: func(args ...interface{}) (interface{}, error) {
			return r.lookupIPer.LookupIPAddr(args[0].(context.Context), args[1].(string))
		},
		UpdateKeyFunc: func(key interface{}) []interface{} {
			return []interface{}{
				context.Background(),
				key.(string),
			}
		},
		Cache: &memCache{Expiration: expire},
	}
	go r.startUpdateLoop()
	return r
}

// CachedResolver
type CachedResolver struct {
	lookupIPer LookupIPAddrer

	ipAddrCache *withCache

	UpdateInterval time.Duration

	// expire 检查过期的时间间隔
	Expiration time.Duration

	tk *time.Ticker
}

// LookupIP 查询host对应所有ip
func (r *CachedResolver) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	ipAddrs, err := r.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, len(ipAddrs))
	for i, ia := range ipAddrs {
		ips[i] = ia.IP
	}
	return ips, nil
}

func (r *CachedResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	result, err := r.ipAddrCache.Get(ctx, host, host)
	if err != nil {
		return nil, err
	}
	return result.([]net.IPAddr), nil
}

func (r *CachedResolver) startUpdateLoop() {
	r.tk = time.NewTicker(r.UpdateInterval)
	for range r.tk.C {
		r.updateAll(context.Background())
	}
}

func (r *CachedResolver) updateAll(ctx context.Context) {
	r.ipAddrCache.updateAll(ctx)
}

// Close stop update
func (r *CachedResolver) Close() error {
	r.tk.Stop()
	return nil
}

type withCache struct {
	GetFunc       func(args ...interface{}) (interface{}, error)
	UpdateKeyFunc func(key interface{}) []interface{}
	Cache         cache
}

func (wc *withCache) Get(args ...interface{}) (interface{}, error) {
	key := args[len(args)-1] // 最后一个参数是 cache 的 key

	val, err := wc.Cache.Get(key)
	if err != nil {
		return nil, err
	}

	if val != nil {
		return val, nil
	}

	return wc.getAndSetCache(key, args[0:len(args)-1])
}

func (wc *withCache) getAndSetCache(key interface{}, args []interface{}) (interface{}, error) {
	raw, err := wc.GetFunc(args...)
	if err != nil {
		return nil, err
	}
	_ = wc.Cache.Set(key, raw)
	return raw, err
}

func (wc *withCache) updateAll(ctx context.Context) {
	names := wc.Cache.Keys()
	for _, name := range names {
		_, _ = wc.getAndSetCache(name, wc.UpdateKeyFunc(name))
	}
}

type cache interface {
	Keys() []interface{}
	Purge() int
	Set(key interface{}, value interface{}) error
	Get(key interface{}) (interface{}, error)
}

type memCache struct {
	lock  sync.RWMutex
	data  map[interface{}]interface{} // lazy init
	visit map[interface{}]time.Time   // lazy init

	Expiration time.Duration
}

func (mc *memCache) Keys() []interface{} {
	mc.lock.RLock()
	defer mc.lock.RUnlock()
	if mc.data == nil {
		return nil
	}
	ks := make([]interface{}, 0, len(mc.data))
	for k := range mc.data {
		ks = append(ks, k)
	}
	return ks
}

func (mc *memCache) Set(key interface{}, value interface{}) error {
	mc.lock.Lock()
	if mc.data == nil {
		mc.data = make(map[interface{}]interface{})
	}
	mc.data[key] = value
	mc.lock.Unlock()

	return nil
}

func (mc *memCache) Get(key interface{}) (interface{}, error) {
	mc.lock.RLock()
	if mc.data == nil {
		mc.lock.RUnlock()
		return nil, nil
	}
	value := mc.data[key]
	mc.lock.RUnlock()

	mc.updateVisit(key)

	return value, nil
}

func (mc *memCache) updateVisit(key interface{}) {
	mc.lock.Lock()
	if mc.visit == nil {
		mc.visit = make(map[interface{}]time.Time)
	}
	mc.visit[key] = time.Now()
	mc.lock.Unlock()
}

func (mc *memCache) Purge() int {
	if mc.Expiration <= 0 {
		return 0
	}
	var expireKeys []interface{}
	expireTime := time.Now().Add(-1 * mc.Expiration)

	mc.lock.RLock()
	for k, tm := range mc.visit {
		if tm.Before(expireTime) {
			expireKeys = append(expireKeys, k)
		}
	}
	mc.lock.RUnlock()

	for _, host := range expireKeys {
		mc.lock.Lock()
		delete(mc.data, host)
		mc.lock.Unlock()
	}
	return len(expireKeys)
}

var _ cache = (*memCache)(nil)
