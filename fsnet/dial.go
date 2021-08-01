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

var _ DialerCanHook = (*Dialer)(nil)

// Dialer dialer
type Dialer struct {
	Timeout   time.Duration
	StdDialer DialerType
	Hooks     []*DialerHook
}

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
	return hook.HookDialContext(ctx, network, address, d.stdDial, len(hook)-1)
}

func (d *Dialer) stdDial(ctx context.Context, network string, address string) (net.Conn, error) {
	nt := Network(network).Resolver()
	if nt.IsIP() {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ip, err := lookupOneIP(ctx, nt.String(), host)
		if err != nil {
			return nil, err
		}
		address = net.JoinHostPort(ip.String(), port)
	}
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
	ctxHookMapper := dialerHooksFormContext(ctx)
	if ctxHookMapper == nil || len(ctxHookMapper.hooks) == 0 {
		return d.Hooks
	}
	if len(d.Hooks) == 0 {
		return nil
	}
	return append(d.Hooks, ctxHookMapper.hooks...)
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
	dh := dialerHooksFormContext(ctx)
	if dh == nil {
		dh = &dialerHookMapper{}
		ctx = context.WithValue(ctx, ctxKeyDialerHook, dh)
	}
	dh.Register(hooks...)
	return ctx
}

func dialerHooksFormContext(ctx context.Context) *dialerHookMapper {
	val := ctx.Value(ctxKeyDialerHook)
	if val == nil {
		return nil
	}
	return val.(*dialerHookMapper)
}
