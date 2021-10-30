// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsnet

import (
	"context"
	"net"
	"time"
)

// DialerType dial conn type
type DialerType interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

// DialerCanHook dialer can RegisterHook
type DialerCanHook interface {
	RegisterHook(hooks ...*DialerHook)
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
	Timeout   time.Duration
	StdDialer DialerType
	Hooks     []*DialerHook
}

var _ DialerCanHook = (*Dialer)(nil)

// RegisterHook register Hooks
func (d *Dialer) RegisterHook(hooks ...*DialerHook) {
	d.Hooks = append(d.Hooks, hooks...)
}

// DialContext dial with Context
func (d *Dialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	if d.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.Timeout)
		defer cancel()
	}
	hook := d.getHooks(ctx)
	c, err := hook.HookDialContext(ctx, network, address, d.stdDial, len(hook)-1)
	cks := ConnHooksFromContext(ctx)
	if err != nil || len(cks) == 0 {
		return c, err
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

		ip, _ := parseIPZone(host)
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

func (d *Dialer) getHooks(ctx context.Context) dialerHooks {
	ctxHooks := DialerHooksFromContext(ctx)
	if len(ctxHooks) == 0 {
		return d.Hooks
	}
	if len(d.Hooks) == 0 {
		return nil
	}
	return append(d.Hooks, ctxHooks...)
}

// DialerHook  dialer hook
type DialerHook struct {
	DialContext func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error)
}

type dialerHooks []*DialerHook

// HookDialContext 执行 hooks
// 倒序执行
func (dhs dialerHooks) HookDialContext(ctx context.Context, network, address string, fn DialContextFunc, idx int) (net.Conn, error) {
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
	hooks dialerHooks
}

func (dhm *dialerHookMapper) Register(hooks ...*DialerHook) {
	dhm.hooks = append(dhm.hooks, hooks...)
}

// ContextWithDialerHook set dialer hook to context
// these hooks will exec before Dialer.Hooks
func ContextWithDialerHook(ctx context.Context, hooks ...*DialerHook) context.Context {
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

// DialerHooksFromContext get DialerHooks from contexts
func DialerHooksFromContext(ctx context.Context) []*DialerHook {
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

// TryRegisterDialerHook 尝试给 DefaultDialer 注册 DialerHook
// 若注册失败将返回 false
func TryRegisterDialerHook(hooks ...*DialerHook) bool {
	if d, ok := DefaultDialer.(DialerCanHook); ok {
		d.RegisterHook(hooks...)
		return true
	}
	return false
}

// MustRegisterDialerHook 给 DefaultDialer 注册 DialerHook
// 若不支持将 panic
func MustRegisterDialerHook(hooks ...*DialerHook) {
	if !TryRegisterDialerHook(hooks...) {
		panic("DefaultDialer cannot RegisterHook")
	}
}

// NewConnDialerHook 创建一个支持添加 ConnHook 的 DialerHook
// 当想给 DefaultDialer 注册 全局的 ConnHook 的时候，可以使用该方法
func NewConnDialerHook(connHooks ...*ConnHook) *DialerHook {
	return &DialerHook{
		DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
			conn, err = fn(ctx, network, address)
			if err != nil || len(connHooks) == 0 {
				return conn, err
			}
			return NewConn(conn, connHooks...), nil
		},
	}
}
