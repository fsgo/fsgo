// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsresolver

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"
)

// Resolver resolver type
type Resolver interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// LookupIPer has  LookupIP func
type LookupIPer interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
}

// LookupIPerGroup LookupIPer slice
type LookupIPerGroup []LookupIPer

// LookupIP Lookup IP
func (hs LookupIPerGroup) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	var ret []net.IP
	var err error
	for i := 0; i < len(hs); i++ {
		ret, err = hs[i].LookupIP(ctx, network, host)
		if err == nil {
			return ret, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not found by %d func", len(hs))
}

// CanIntercept 支持注册 Interceptor
type CanIntercept interface {
	RegisterInterceptor(its ...*Interceptor)
}

// LookupIPFunc lookupIP func type
type LookupIPFunc func(ctx context.Context, network, host string) ([]net.IP, error)

// LookupIPAddrFunc LookupIPAddr func type
type LookupIPAddrFunc func(ctx context.Context, host string) ([]net.IPAddr, error)

// Default default Resolver, result has 3 min cache
//
//	Environment Variables 'FSGO_RESOLVER_EXP' can set the default cache lifetime
//	eg: export FSGO_RESOLVER_EXP="10m" set cache lifetime as 10 minute
var Default Resolver = &Cached{
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

// LookupIP Default.LookupIP
var LookupIP = func(ctx context.Context, network, host string) ([]net.IP, error) {
	return Default.LookupIP(ctx, network, host)
}

// LookupIPAddr Default.LookupIPAddr
var LookupIPAddr = func(ctx context.Context, host string) ([]net.IPAddr, error) {
	return Default.LookupIPAddr(ctx, host)
}
