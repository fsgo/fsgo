// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/31

package fsdns

import (
	"reflect"
	"testing"
)

func TestParseResolv(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name string
		args args
		want *ResolverConfig
	}{
		{
			name: "default",
			args: args{
				content: []byte(`
#
nameserver 192.168.1.1
nameserver 127.0.0.1
nameserver 127.0.0.1
#nameserver 114.114.114.114

search qq.com abc.com
search qq.com abc.com
options timeout:1
options timeout:1
options no-check-names

sortlist 127.0.0.0
sortlist 127.0.0.0
`),
			},
			want: &ResolverConfig{
				Nameserver: []string{"192.168.1.1", "127.0.0.1"},
				Options:    []string{"timeout:1", "no-check-names"},
				Search:     []string{"qq.com", "abc.com"},
				SortList:   []string{"127.0.0.0"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseResolv(tt.args.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseResolv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolvFile_Nameserver(t *testing.T) {
	type fields struct {
		Path string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "default",
			fields: fields{
				Path: "testdata/resolv.conf",
			},
			want: []string{"192.168.1.1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := NewResolvConf(tt.fields.Path)
			defer rf.Stop()
			if got := rf.Nameserver(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Nameserver() = %v, want %v", got, tt.want)
			}
		})
	}
}
