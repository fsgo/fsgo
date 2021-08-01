// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/1

package fsnet

import (
	"testing"
)

func TestNetwork_Resolver(t *testing.T) {
	tests := []struct {
		name string
		nt   Network
		want Network
	}{
		{
			name: "tcp-ip",
			nt:   NetworkTCP,
			want: NetworkIP,
		},
		{
			name: "udp-ip",
			nt:   NetworkUDP,
			want: NetworkIP,
		},
		{
			name: "unix-unix",
			nt:   NetworkUnix,
			want: NetworkUnix,
		},
		{
			name: "other",
			nt:   "other",
			want: "other",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.nt.Resolver(); got != tt.want {
				t.Errorf("Resolver() = %v, want %v", got, tt.want)
			}
		})
	}
}
