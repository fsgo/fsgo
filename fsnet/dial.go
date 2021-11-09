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

// DialerCanInterceptor dialer can RegisterHook
type DialerCanInterceptor interface {
	RegisterInterceptor(hooks ...*DialerInterceptor)
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
func (d *Dialer) RegisterInterceptor(hooks ...*DialerInterceptor) {
	d.Interceptors = append(d.Interceptors, hooks...)
}

// DialContext dial with Context
func (d *Dialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	if d.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.Timeout)
		defer cancel()
	}
	hook := d.getInterceptors(ctx)
	c, err := hook.HookDialContext(ctx, network, address, d.stdDial, len(hook)-1)
	if err != nil {
		return nil, err
	}
	cks := ConnInterceptorsFromContext(ctx)
	if len(cks) == 0 {
		return c, nil
	}
	return NewConn(c, cks...), nil
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
	ctxHooks := DialerInterceptorsFromContext(ctx)
	if len(ctxHooks) == 0 {
		return d.Interceptors
	}
	if len(d.Interceptors) == 0 {
		return nil
	}
	return append(d.Interceptors, ctxHooks...)
}

// DialerInterceptor  dialer interceptor
type DialerInterceptor struct {
	DialContext func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error)
}

type dialerInterceptors []*DialerInterceptor

// HookDialContext 执行 hooks
// 倒序执行
func (dhs dialerInterceptors) HookDialContext(ctx context.Context, network, address string, fn DialContextFunc, idx int) (net.Conn, error) {
	for ; idx >= 0; idx-- {
		if dhs[idx].DialContext != nil {
			break
		}
	}
	if len(dhs) == 0 || idx < 0 {
		return fn(ctx, network, address)
	}
	return dhs[idx].DialContext(ctx, network, address, func(ctx context.Context, network string, address string) (net.Conn, error) {
		return dhs.HookDialContext(ctx, network, address, fn, idx-1)
	})
}

type dialerHookMapper struct {
	hooks dialerInterceptors
}

func (dhm *dialerHookMapper) Register(hooks ...*DialerInterceptor) {
	dhm.hooks = append(dhm.hooks, hooks...)
}

// ContextWithDialerHook set dialer hook to context
// these hooks will exec before Dialer.Interceptors
func ContextWithDialerHook(ctx context.Context, hooks ...*DialerInterceptor) context.Context {
	if len(hooks) == 0 {
		return ctx
	}
	dh := dialerHookMapperFormContext(ctx)
	if dh == nil {
		dh = &dialerHookMapper{}
		ctx = context.WithValue(ctx, ctxKeyDialerHook, dh)
	}
	dh.Register(hooks...)
	return ctx
}

// DialerInterceptorsFromContext get DialerHooks from contexts
func DialerInterceptorsFromContext(ctx context.Context) []*DialerInterceptor {
	dhm := dialerHookMapperFormContext(ctx)
	if dhm == nil {
		return nil
	}
	return dhm.hooks
}

func dialerHookMapperFormContext(ctx context.Context) *dialerHookMapper {
	val := ctx.Value(ctxKeyDialerHook)
	if val == nil {
		return nil
	}
	return val.(*dialerHookMapper)
}

// TryRegisterDialerInterceptor 尝试给 DefaultDialer 注册 DialerInterceptor
// 若注册失败将返回 false
func TryRegisterDialerInterceptor(hooks ...*DialerInterceptor) bool {
	if d, ok := DefaultDialer.(DialerCanInterceptor); ok {
		d.RegisterInterceptor(hooks...)
		return true
	}
	return false
}

// MustRegisterDialerHook 给 DefaultDialer 注册 DialerInterceptor
// 若不支持将 panic
func MustRegisterDialerHook(hooks ...*DialerInterceptor) {
	if !TryRegisterDialerInterceptor(hooks...) {
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
			return NewConn(conn, its...), nil
		},
	}
}
