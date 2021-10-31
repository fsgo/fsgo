// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/31

package envfile

import (
	"context"
	"reflect"
	"testing"
)

func TestParserEvnFile(t *testing.T) {
	type args struct {
		fp string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				fp: "testdata/envfile/ok1.txt",
			},
			want: []string{
				"key1=1",
				"key2=2",
				"key3=",
				"key4=4",
			},
			wantErr: false,
		},
		{
			name: "case 2",
			args: args{
				fp: "testdata/envfile/err1.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "case 3",
			args: args{
				fp: "testdata/envfile/err2.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "case 4",
			args: args{
				fp: "testdata/envfile/ok2.sh",
			},
			want: []string{
				"key1=1",
				"key2=2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParserFile(context.Background(), tt.args.fp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParserFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParserFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
