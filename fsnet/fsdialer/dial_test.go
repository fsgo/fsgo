// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsdialer

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fsgo/fsgo/fsnet/fsconn"
)

func TestDialer_DialContext(t *testing.T) {
	t.Run("default no its", func(t *testing.T) {
		wantErr := errors.New("err must")
		d := &Simple{
			Invoker: &testDialer{
				retConn: nil,
				retErr:  wantErr,
			},
		}
		_, err := d.DialContext(context.Background(), "tcp", "127.0.0.1:80")
		assert.Equal(t, wantErr, err)
	})

	t.Run("with many its", func(t *testing.T) {
		wantErr := errors.New("err must")
		var num int32
		checkNum := func(want int32) {
			assert.Equal(t, want, num)
			num++
		}
		d := &Simple{
			Invoker: &testDialer{
				retConn: nil,
				retErr:  wantErr,
			},
			Interceptors: []*Interceptor{
				{
					DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
						checkNum(0)
						return fn(ctx, network, address)
					},
				},
				{
					DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
						checkNum(1)
						return fn(ctx, network, address)
					},
				},
			},
		}
		ctx := context.Background()
		ctx = ContextWithInterceptor(ctx, &Interceptor{
			DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
				checkNum(2)
				return fn(ctx, network, address)
			},
		}, &Interceptor{
			DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
				checkNum(3)
				return fn(ctx, network, address)
			},
		})
		ctx = ContextWithInterceptor(ctx, &Interceptor{
			DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
				checkNum(4)
				return fn(ctx, network, address)
			},
		})
		_, err := d.DialContext(ctx, "tcp", "127.0.0.1:80")
		assert.Equal(t, wantErr, err)
	})
}

var _ Dialer = (*testDialer)(nil)

type testDialer struct {
	retConn net.Conn
	retErr  error
}

func (t *testDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	return t.retConn, t.retErr
}

func Test_dialerHooks_HookDialContext(t *testing.T) {
	td := &testDialer{
		retErr: errors.New("mustErr"),
	}
	t.Run("zero dhs", func(t *testing.T) {
		var dhs interceptors
		_, err := dhs.CallDialContext(context.Background(), "tcp", "127.0.0.1:80", td.DialContext, 0)
		assert.Equal(t, td.retErr, err)
	})

	t.Run("one dhs", func(t *testing.T) {
		var num int32
		checkNum := func(want int32) {
			assert.Equal(t, want, num)
			num++
		}
		dhs := interceptors{
			{
				DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
					checkNum(0)
					return fn(ctx, network, address)
				},
			},
		}
		_, err := dhs.CallDialContext(context.Background(), "tcp", "127.0.0.1:80", td.DialContext, 0)
		assert.Equal(t, td.retErr, err)
		checkNum(1)
	})
	t.Run("tow dhs", func(t *testing.T) {
		var num int32
		checkNum := func(want int32) {
			assert.Equal(t, want, num)
			num++
		}
		dhs := interceptors{
			{
				DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
					checkNum(0)
					return fn(ctx, network, address)
				},
			},
			{
				DialContext: func(ctx context.Context, network string, address string, fn DialContextFunc) (conn net.Conn, err error) {
					checkNum(1)
					return fn(ctx, network, address)
				},
			},
		}
		_, err := dhs.CallDialContext(context.Background(), "tcp", "127.0.0.1:80", td.DialContext, 0)
		assert.Equal(t, td.retErr, err)
		checkNum(2)
	})
}

func TestMustRegisterDialerHook(t *testing.T) {
	Default = &Simple{}
	defer func() {
		Default = &Simple{}
	}()
	hk := TransConnInterceptor(&fsconn.Interceptor{})
	MustRegisterInterceptor(hk)
	hks := Default.(*Simple).Interceptors
	assert.Len(t, hks, 1)

	assert.Equal(t, hks[0], hk)
}
