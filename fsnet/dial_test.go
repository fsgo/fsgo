// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsnet

import (
	"context"
	"fmt"
	"net"
	"testing"
)

func TestDialer_DialContext(t *testing.T) {
	t.Run("default no hooks", func(t *testing.T) {
		wantErr := fmt.Errorf("err must")
		d := &Dialer{
			StdDialer: &testDialer{
				retConn: nil,
				retErr:  wantErr,
			},
		}
		_, err := d.DialContext(context.Background(), "tcp", "127.0.0.1:80")
		if err != wantErr {
			t.Fatalf("not eq, got=%v want=%v", err, wantErr)
		}
	})

	t.Run("with many hooks", func(t *testing.T) {
		wantErr := fmt.Errorf("err must")
		var num int32
		checkNum := func(want int32) {
			if num != want {
				t.Fatalf("got num=%d want=%d", num, want)
			}
			num++
		}
		d := &Dialer{
			StdDialer: &testDialer{
				retConn: nil,
				retErr:  wantErr,
			},
			Hooks: []*DialerHook{
				{
					DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
						checkNum(4)
						return fn(ctx, network, address)
					},
				},
				{
					DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
						checkNum(3)
						return fn(ctx, network, address)
					},
				},
			},
		}
		ctx := context.Background()
		ctx = ContextWithDialerHook(ctx, &DialerHook{
			DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
				checkNum(2)
				return fn(ctx, network, address)
			},
		}, &DialerHook{
			DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
				checkNum(1)
				return fn(ctx, network, address)
			},
		})
		ctx = ContextWithDialerHook(ctx, &DialerHook{
			DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
				checkNum(0)
				return fn(ctx, network, address)
			},
		})
		_, err := d.DialContext(ctx, "tcp", "127.0.0.1:80")
		if err != wantErr {
			t.Fatalf("not eq")
		}
	})
}

var _ DialerType = (*testDialer)(nil)

type testDialer struct {
	retConn net.Conn
	retErr  error
}

func (t *testDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	return t.retConn, t.retErr
}

func Test_dialerHooks_HookDialContext(t *testing.T) {
	td := &testDialer{
		retErr: fmt.Errorf("mustErr"),
	}
	t.Run("zero dhs", func(t *testing.T) {
		var dhs dialerHooks
		_, err := dhs.HookDialContext(context.Background(), "tcp", "127.0.0.1:80", td.DialContext, -1)
		if err != td.retErr {
			t.Fatalf("not eq")
		}
	})

	t.Run("one dhs", func(t *testing.T) {
		var num int32
		checkNum := func(want int32) {
			if num != want {
				t.Fatalf("got num=%d want=%d", num, want)
			}
			num++
		}
		dhs := dialerHooks{
			{
				DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
					checkNum(0)
					return fn(ctx, network, address)
				},
			},
		}
		_, err := dhs.HookDialContext(context.Background(), "tcp", "127.0.0.1:80", td.DialContext, len(dhs)-1)
		if err != td.retErr {
			t.Fatalf("not eq")
		}
		checkNum(1)
	})
	t.Run("tow dhs", func(t *testing.T) {
		var num int32
		checkNum := func(want int32) {
			if num != want {
				t.Fatalf("got num=%d want=%d", num, want)
			}
			num++
		}
		dhs := dialerHooks{
			{
				DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
					checkNum(1)
					return fn(ctx, network, address)
				},
			},
			{
				DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
					checkNum(0)
					return fn(ctx, network, address)
				},
			},
		}
		_, err := dhs.HookDialContext(context.Background(), "tcp", "127.0.0.1:80", td.DialContext, len(dhs)-1)
		if err != td.retErr {
			t.Fatalf("not eq")
		}
		checkNum(2)
	})
}
