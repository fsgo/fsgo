// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/11/13

package fsip

import (
	"net"
	"reflect"
	"testing"
)

func TestIsIPv4only(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{
			ip:   "127.0.0.1",
			want: true,
		},
		{
			ip:   "::1",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			if got := IsIPv4only(net.ParseIP(tt.ip)); got != tt.want {
				t.Errorf("IsIPv4only() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIPv6only(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{
			ip:   "127.0.0.1",
			want: false,
		},
		{
			ip:   "::1",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			if got := IsIPv6only(net.ParseIP(tt.ip)); got != tt.want {
				t.Errorf("IsIPv6only() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterIPList(t *testing.T) {
	type args struct {
		ips    []net.IP
		filter func(ip net.IP) bool
	}
	tests := []struct {
		name string
		args args
		want []net.IP
	}{
		{
			name: "case 1",
			args: args{
				ips: []net.IP{
					net.ParseIP("127.0.0.1"),
					net.ParseIP("::1"),
					net.ParseIP("10.0.0.1"),
					net.ParseIP("1.1.1.1"),
					net.ParseIP("192.168.1.1"),
				},
				filter: NewIsFnsAnd(IsIPv4only, IsLoopback),
			},
			want: []net.IP{
				net.ParseIP("127.0.0.1"),
			},
		},
		{
			name: "case 2",
			args: args{
				ips: []net.IP{
					net.ParseIP("127.0.0.1"),
					net.ParseIP("::1"),
					net.ParseIP("10.0.0.1"),
					net.ParseIP("1.1.1.1"),
					net.ParseIP("192.168.1.1"),
				},
				filter: NewIsIPFnsOr(IsIPv6only, NewIsFnsAnd(NotFilter(IsPrivate))),
			},
			want: []net.IP{
				net.ParseIP("127.0.0.1"),
				net.ParseIP("::1"),
				net.ParseIP("1.1.1.1"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterList(tt.args.ips, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterList() = %v, want %v", got, tt.want)
			}
		})
	}
}
