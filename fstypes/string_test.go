// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/6/6

package fstypes

import (
	"reflect"
	"testing"
)

func TestString_ToInt64Slice(t *testing.T) {
	type args struct {
		sep string
	}
	tests := []struct {
		name    string
		s       String
		args    args
		want    []int64
		wantErr bool
	}{
		{
			name: "case 1",
			s:    "1",
			args: args{
				",",
			},
			want:    []int64{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3 ,",
			args: args{
				",",
			},
			want:    []int64{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				",",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "case 5",
			s:    "8589934592",
			args: args{
				",",
			},
			want:    []int64{8589934592},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.ToInt64Slice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInt64Slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToInt64Slice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_ToIntSlice(t *testing.T) {
	type args struct {
		sep string
	}
	tests := []struct {
		name    string
		s       String
		args    args
		want    []int
		wantErr bool
	}{
		{
			name: "case 1",
			s:    "1",
			args: args{
				",",
			},
			want:    []int{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3 ,",
			args: args{
				",",
			},
			want:    []int{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				",",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.ToIntSlice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToIntSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToIntSlice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_ToInt32Slice(t *testing.T) {
	type args struct {
		sep string
	}
	tests := []struct {
		name    string
		s       String
		args    args
		want    []int32
		wantErr bool
	}{
		{
			name: "case 1",
			s:    "1",
			args: args{
				",",
			},
			want:    []int32{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3 ,",
			args: args{
				",",
			},
			want:    []int32{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				",",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "case 5",
			s:    "8589934592",
			args: args{
				",",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.ToInt32Slice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInt32Slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToInt32Slice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_ToUint64Slice(t *testing.T) {
	type args struct {
		sep string
	}
	tests := []struct {
		name    string
		s       String
		args    args
		want    []uint64
		wantErr bool
	}{
		{
			name: "case 1",
			s:    "1",
			args: args{
				",",
			},
			want:    []uint64{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3, ",
			args: args{
				",",
			},
			want:    []uint64{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				",",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.ToUint64Slice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToUint64Slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToUint64Slice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_ToUint32Slice(t *testing.T) {
	type args struct {
		sep string
	}
	tests := []struct {
		name    string
		s       String
		args    args
		want    []uint32
		wantErr bool
	}{
		{
			name: "case 1",
			s:    "1",
			args: args{
				",",
			},
			want:    []uint32{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3 ,",
			args: args{
				",",
			},
			want:    []uint32{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				",",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.ToUint32Slice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToUint32Slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToUint32Slice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_ToStrSlice(t *testing.T) {
	type args struct {
		sep string
	}
	tests := []struct {
		name    string
		s       String
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "case 1",
			s:    "a,b, c, , e , ",
			args: args{
				sep: ",",
			},
			want: []string{"a", "b", "c", "e"},
		},
		{
			name: "case 2",
			s:    "",
			args: args{
				sep: ",",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.ToStrSlice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToStrSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToStrSlice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkString_ToInt64Slice(b *testing.B) {
	xs := String("1, 2, 3 ,4 ,")
	var xi []int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		xi, _ = xs.ToInt64Slice(",")
	}
	_ = xi
}

func BenchmarkString_split(b *testing.B) {
	xs := String("1, 2, 3 ,4 ,")
	var xa []string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		xa = xs.split(",")
	}
	_ = xa
}
