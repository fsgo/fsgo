// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/2/25

package internal

import (
	"testing"

	"github.com/fsgo/fsgo/fsnet/fsconn/conndump"
)

func TestIsAction(t *testing.T) {
	type args struct {
		a  string
		ac conndump.MessageAction
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case 1",
			args: args{
				a:  "",
				ac: conndump.MessageAction_Read,
			},
			want: true,
		},
		{
			name: "case 2",
			args: args{
				a:  "crw",
				ac: conndump.MessageAction_Read,
			},
			want: true,
		},
		{
			name: "case 3",
			args: args{
				a:  "r",
				ac: conndump.MessageAction_Read,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAction(tt.args.a, tt.args.ac); got != tt.want {
				t.Errorf("IsAction() = %v, want %v", got, tt.want)
			}
		})
	}
}
