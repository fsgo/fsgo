// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsdialer

import (
	"context"
	"net"
	"time"

	"github.com/fsgo/fsgo/fsnet"
	"github.com/fsgo/fsgo/fsnet/fsconn"
	"github.com/fsgo/fsgo/fsnet/fsresolver"
	"github.com/fsgo/fsgo/fsnet/internal"
	"github.com/fsgo/fsgo/internal/xctx"
)

// Dialer dial connWithIt type
type Dialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

// CanInterceptor dialer can RegisterInterceptor
type CanInterceptor interface {
	RegisterInterceptor(its ...*Interceptor)
}

// DialContextFunc Dial func type
type DialContextFunc func(ctx context.Context, network string, address string) (net.Conn, error)

// Default default dialer
var Default Dialer = &Simple{}

// DialContext dial default
var DialContext = func(ctx context.Context, network string, address string) (net.Conn, error) {
	return Default.DialContext(ctx, network, address)
}

// Simple dialer
type Simple struct {
	// Invoker 可选，底层拨号器
	Invoker Dialer

	// Resolver 可选，dns 解析
	Resolver fsresolver.Resolver

	// Interceptors 可选，拦截器列表
	Interceptors []*Interceptor

	// Timeout 可选，超时时间
	Timeout time.Duration
}

var _ CanInterceptor = (*Simple)(nil)

// RegisterInterceptor register Interceptor
func (d *Simple) RegisterInterceptor(its ...*Interceptor) {
	d.Interceptors = append(d.Interceptors, its...)
}

// DialContext dial with Context
func (d *Simple) DialContext(ctx context.Context, network string, address string) (conn net.Conn, err error) {
	if d.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.Timeout)
		defer cancel()
	}
	its := d.getInterceptors(ctx)

	dialIdx := -1
	afterIdx := -1
	for i := 0; i < len(its); i++ {
		if its[i].BeforeDialContext != nil {
			ctx, network, address = its[i].BeforeDialContext(ctx, network, address)
		}
		if dialIdx == -1 && its[i].DialContext != nil {
			dialIdx = i
		}
		if afterIdx == -1 && its[i].AfterDialContext != nil {
			afterIdx = i
		}
	}
	if dialIdx == -1 {
		conn, err = d.stdDial(ctx, network, address)
	} else {
		conn, err = its.CallDialContext(ctx, network, address, d.stdDial, dialIdx)
	}

	if afterIdx != -1 {
		for i := afterIdx; i < len(its); i++ {
			if its[i].AfterDialContext != nil {
				conn, err = its[i].AfterDialContext(ctx, network, address, conn, err)
			}
		}
	}

	if err != nil {
		return nil, err
	}
	return fsconn.NewWithContext(ctx, conn), nil
}

func splitHostPort(hostPort string) (host string, port string, err error) {
	host, port, err = net.SplitHostPort(hostPort)
	if err != nil {
		return "", "", err
	}

	if len(host) == 0 {
		return "", "", &net.AddrError{
			Err:  "empty host",
			Addr: hostPort,
		}
	}

	return host, port, nil
}

func (d *Simple) stdDial(ctx context.Context, network string, address string) (net.Conn, error) {
	nt := fsnet.Network(network).Resolver()
	if nt.IsIP() {
		host, port, err := splitHostPort(address)
		if err != nil {
			return nil, err
		}

		ip, _ := internal.ParseIPZone(host)
		if ip != nil {
			return d.dial(ctx, network, address)
		}

		ips, err := fsresolver.LookupIP(ctx, string(nt), host)
		if err != nil {
			return nil, err
		}
		// 在超时允许的范围内，将所有 ip 都尝试一遍
		for _, ip := range ips {
			ad := net.JoinHostPort(ip.String(), port)
			conn, err := d.dial(ctx, network, ad)
			if err == nil || ctx.Err() != nil {
				return conn, ctx.Err()
			}
		}
		return nil, err
	}
	return d.dial(ctx, network, address)
}

func (d *Simple) dial(ctx context.Context, network, address string) (net.Conn, error) {
	return d.getSTDDialer().DialContext(ctx, network, address)
}

var zeroDialer = &net.Dialer{}

func (d *Simple) getSTDDialer() Dialer {
	if d.Invoker != nil {
		return d.Invoker
	}
	return zeroDialer
}

func (d *Simple) getInterceptors(ctx context.Context) interceptors {
	ctxIts := InterceptorsFromContext(ctx)
	return xctx.ValuesMerge(ctxIts, d.Interceptors)
}
