// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/31

package fsdns

import (
	"net"
	"reflect"
	"testing"
)

func TestParseHosts(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name string
		args args
		want map[string][]net.IP
	}{
		{
			name: "default",
			args: args{
				content: []byte(`
127.0.0.1 a b.com
127.0.0.2 b.com
`),
			},
			want: map[string][]net.IP{
				"a": {
					net.ParseIP("127.0.0.1"),
				},
				"b.com": {
					net.ParseIP("127.0.0.1"),
					net.ParseIP("127.0.0.2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseHosts(tt.args.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseHosts() = %v, want %v", got, tt.want)
			}
		})
	}
}
