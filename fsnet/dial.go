// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsnet

import (
	"context"
	"net"
	"time"

	"github.com/fsgo/fsgo/fsnet/internal"
)

// DialerType dial conn type
type DialerType interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

// DialerCanInterceptor dialer can RegisterInterceptor
type DialerCanInterceptor interface {
	RegisterInterceptor(its ...*DialerInterceptor)
}

// DialContextFunc Dial func type
type DialContextFunc func(ctx context.Context, network string, address string) (net.Conn, error)

// DefaultDialer default dialer
var DefaultDialer DialerType = &Dialer{}

// DialContext dial default
var DialContext = func(ctx context.Context, network string, address string) (net.Conn, error) {
	return DefaultDialer.DialContext(ctx, network, address)
}

// Dialer dialer
type Dialer struct {
	// Timeout 可选，超时时间
	Timeout time.Duration

	// StdDialer 可选，底层拨号器
	StdDialer DialerType

	// Interceptors 可选，拦截器列表,倒序执行
	Interceptors []*DialerInterceptor
}

var _ DialerCanInterceptor = (*Dialer)(nil)

// RegisterInterceptor register Interceptor
func (d *Dialer) RegisterInterceptor(its ...*DialerInterceptor) {
	d.Interceptors = append(d.Interceptors, its...)
}

// DialContext dial with Context
func (d *Dialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	if d.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.Timeout)
		defer cancel()
	}
	its := d.getInterceptors(ctx)
	c, err := its.CallDialContext(ctx, network, address, d.stdDial, len(its)-1)
	if err != nil {
		return nil, err
	}
	cks := ConnInterceptorsFromContext(ctx)
	if len(cks) == 0 {
		return c, nil
	}
	return WrapConn(c, cks...), nil
}

func (d *Dialer) stdDial(ctx context.Context, network string, address string) (conn net.Conn, err error) {
	nt := Network(network).Resolver()
	if nt.IsIP() {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}

		ip, _ := internal.ParseIPZone(host)
		if ip != nil {
			return d.dial(ctx, network, address)
		}

		ips, err := LookupIP(ctx, string(nt), host)
		if err != nil {
			return nil, err
		}
		// 在超时允许的范围内，将所有 ip 都尝试一遍
		for _, ip := range ips {
			ad := net.JoinHostPort(ip.String(), port)
			conn, err = d.dial(ctx, network, ad)
			if err == nil || ctx.Err() != nil {
				return conn, err
			}
		}
		return nil, err

	}
	return d.dial(ctx, network, address)
}

func (d *Dialer) dial(ctx context.Context, network, address string) (net.Conn, error) {
	return d.getSTDDialer().DialContext(ctx, network, address)
}

var zeroDialer = &net.Dialer{}

func (d *Dialer) getSTDDialer() DialerType {
	if d.StdDialer != nil {
		return d.StdDialer
	}
	return zeroDialer
}

func (d *Dialer) getInterceptors(ctx context.Context) dialerInterceptors {
	ctxIts := DialerInterceptorsFromContext(ctx)
	if len(ctxIts) == 0 {
		return d.Interceptors
	}
	if len(d.Interceptors) == 0 {
		return nil
	}
	return append(d.Interceptors, ctxIts...)
}

// DialerInterceptor  dialer interceptor
type DialerInterceptor struct {
	DialContext func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error)
}

type dialerInterceptors []*DialerInterceptor

// CallDialContext 执行 its
// 倒序执行
func (dhs dialerInterceptors) CallDialContext(ctx context.Context, network, address string, fn DialContextFunc, idx int) (net.Conn, error) {
	for ; idx >= 0; idx-- {
		if dhs[idx].DialContext != nil {
			break
		}
	}
	if len(dhs) == 0 || idx < 0 {
		return fn(ctx, network, address)
	}
	return dhs[idx].DialContext(ctx, network, address, func(ctx context.Context, network string, address string) (net.Conn, error) {
		return dhs.CallDialContext(ctx, network, address, fn, idx-1)
	})
}

type dialerInterceptorMapper struct {
	its dialerInterceptors
}

func (dhm *dialerInterceptorMapper) Register(its ...*DialerInterceptor) {
	dhm.its = append(dhm.its, its...)
}

// ContextWithDialerInterceptor set dialer Interceptor to context
// these interceptors will exec before Dialer.Interceptors
func ContextWithDialerInterceptor(ctx context.Context, its ...*DialerInterceptor) context.Context {
	if len(its) == 0 {
		return ctx
	}
	dh := dialerItMapperFormContext(ctx)
	if dh == nil {
		dh = &dialerInterceptorMapper{}
		ctx = context.WithValue(ctx, ctxKeyDialerInterceptor, dh)
	}
	dh.Register(its...)
	return ctx
}

// DialerInterceptorsFromContext get DialerInterceptors from contexts
func DialerInterceptorsFromContext(ctx context.Context) []*DialerInterceptor {
	dhm := dialerItMapperFormContext(ctx)
	if dhm == nil {
		return nil
	}
	return dhm.its
}

func dialerItMapperFormContext(ctx context.Context) *dialerInterceptorMapper {
	val := ctx.Value(ctxKeyDialerInterceptor)
	if val == nil {
		return nil
	}
	return val.(*dialerInterceptorMapper)
}

// TryRegisterDialerInterceptor 尝试给 DefaultDialer 注册 DialerInterceptor
// 若注册失败将返回 false
func TryRegisterDialerInterceptor(its ...*DialerInterceptor) bool {
	if d, ok := DefaultDialer.(DialerCanInterceptor); ok {
		d.RegisterInterceptor(its...)
		return true
	}
	return false
}

// MustRegisterDialerInterceptor 给 DefaultDialer 注册 DialerInterceptor
// 若不支持将 panic
func MustRegisterDialerInterceptor(its ...*DialerInterceptor) {
	if !TryRegisterDialerInterceptor(its...) {
		panic("DefaultDialer cannot Register DialerInterceptor")
	}
}

// NewConnDialerInterceptor 创建一个支持添加 ConnInterceptor 的 DialerInterceptor
// 当想给 DefaultDialer 注册 全局的 ConnInterceptor 的时候，可以使用该方法
func NewConnDialerInterceptor(its ...*ConnInterceptor) *DialerInterceptor {
	return &DialerInterceptor{
		DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
			conn, err = fn(ctx, network, address)
			if err != nil || len(its) == 0 {
				return conn, err
			}
			return WrapConn(conn, its...), nil
		},
	}
}
