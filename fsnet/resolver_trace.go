// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/12/5

package fsnet

import (
	"context"
	"log"
	"net"
	"time"
)

// PrintResolverLog print Resolver log interceptor
var PrintResolverLog = &ResolverInterceptor{
	LookupIP: func(ctx context.Context, network, host string, invoker LookupIPFunc) ([]net.IP, error) {
		start := time.Now()
		ret, err := invoker(ctx, network, host)
		cost := time.Since(start)
		log.Printf("LookupIP(%q,%q)=(%v,%v) cost=%s\n", network, host, ret, err, cost)
		return ret, err
	},
	LookupIPAddr: func(ctx context.Context, host string, invoker LookupIPAddrFunc) ([]net.IPAddr, error) {
		start := time.Now()
		ret, err := invoker(ctx, host)
		cost := time.Since(start)
		log.Printf("LookupIPAddr(%q)=(%v,%v) cost=%s\n", host, ret, err, cost)
		return ret, err
	},
}
