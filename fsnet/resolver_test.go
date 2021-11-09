// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsnet

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

var _ Resolver = (*testResolver)(nil)

type testResolver struct {
	lookupIPData     sync.Map
	lookupIPAddrData sync.Map
}

func (t *testResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	if v, ok := t.lookupIPData.Load(host); ok {
		return v.([]net.IP), nil
	}
	return nil, fmt.Errorf("ip not found")
}

func (t *testResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	if v, ok := t.lookupIPAddrData.Load(host); ok {
		return v.([]net.IPAddr), nil
	}
	return nil, fmt.Errorf("ipaddr not found")
}

func TestResolverCached(t *testing.T) {
	type testCase struct {
		host          string
		wantIPErr     bool
		wantIP        []net.IP
		wantIPAddrErr bool
		wantIPAddr    []net.IPAddr
		check         func(t *testing.T)
	}
	runTestCases := func(t *testing.T, re Resolver, tests []testCase) {
		for _, tt := range tests {
			t.Run(tt.host, func(t *testing.T) {
				t.Run("LookupIP", func(t *testing.T) {
					got, gotErr := re.LookupIP(context.Background(), "ip", tt.host)
					if tt.wantIPErr != (gotErr != nil) {
						t.Fatalf("wantErr=%v got=%v", tt.wantIP, gotErr)
					}
					if !reflect.DeepEqual(tt.wantIP, got) {
						t.Fatalf("got=%v want=%v", got, tt.wantIP)
					}
				})

				t.Run("LookupIPAddr", func(t *testing.T) {
					got, gotErr := re.LookupIPAddr(context.Background(), tt.host)
					if tt.wantIPAddrErr != (gotErr != nil) {
						t.Fatalf("wantErr=%v got=%v", tt.wantIPAddrErr, gotErr)
					}
					if !reflect.DeepEqual(tt.wantIPAddr, got) {
						t.Fatalf("got=%v want=%v", got, tt.wantIP)
					}
				})
				if tt.check != nil {
					t.Run("check", func(t *testing.T) {
						tt.check(t)
					})
				}
			})
		}
	}

	t.Run("default no hooks", func(t *testing.T) {
		std := &testResolver{}
		std.lookupIPData.Store("def.com", []net.IP{net.ParseIP("127.0.0.1")})
		std.lookupIPAddrData.Store("def.com", []net.IPAddr{
			{
				IP: net.ParseIP("127.0.0.1"),
			},
		})
		tests := []testCase{
			{
				host:          "www.abc.com",
				wantIPErr:     true,
				wantIPAddrErr: true,
			},
			{
				host:      "def.com",
				wantIPErr: false,
				wantIP: []net.IP{
					net.ParseIP("127.0.0.1"),
				},
				wantIPAddrErr: false,
				wantIPAddr: []net.IPAddr{
					{
						IP: net.ParseIP("127.0.0.1"),
					},
				},
			},
		}
		re := &ResolverCached{
			Expiration:  time.Minute,
			StdResolver: std,
		}
		runTestCases(t, re, tests)
	})

	t.Run("with hooks", func(t *testing.T) {
		std := &testResolver{}
		std.lookupIPData.Store("def.com", []net.IP{net.ParseIP("127.0.0.1")})
		std.lookupIPAddrData.Store("def.com", []net.IPAddr{
			{
				IP: net.ParseIP("127.0.0.1"),
			},
		})
		re := &ResolverCached{
			Expiration:  time.Minute,
			StdResolver: std,
		}

		var lookupIPNum testNum
		var lookupIPAddrNum testNum

		re.RegisterHook(&ResolverInterceptor{
			LookupIP: func(ctx context.Context, network, host string, fn LookupIPFunc) ([]net.IP, error) {
				lookupIPNum.Incr()
				return fn(ctx, network, host)
			},
		}, &ResolverInterceptor{
			LookupIPAddr: func(ctx context.Context, host string, fn LookupIPAddrFunc) ([]net.IPAddr, error) {
				lookupIPAddrNum.Incr()
				return fn(ctx, host)
			},
		})

		tests := []testCase{
			{
				host:          "www.abc.com",
				wantIPErr:     true,
				wantIPAddrErr: true,
			},
			{
				host:      "def.com",
				wantIPErr: false,
				wantIP: []net.IP{
					net.ParseIP("127.0.0.1"),
				},
				wantIPAddrErr: false,
				wantIPAddr: []net.IPAddr{
					{
						IP: net.ParseIP("127.0.0.1"),
					},
				},
			},
		}

		runTestCases(t, re, tests)

		lookupIPNum.Check(t, len(tests))
		lookupIPAddrNum.Check(t, len(tests))
	})
}
