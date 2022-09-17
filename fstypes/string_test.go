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
				sep: ",",
			},
			want:    []int64{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3 ,",
			args: args{
				sep: ",",
			},
			want:    []int64{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "case 5",
			s:    "8589934592",
			args: args{
				sep: ",",
			},
			want:    []int64{8589934592},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Int64Slice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int64s() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int64s() got = %v, want %v", got, tt.want)
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
				sep: ",",
			},
			want:    []int{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3 ,",
			args: args{
				sep: ",",
			},
			want:    []int{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.IntSlice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ints() got = %v, want %v", got, tt.want)
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
				sep: ",",
			},
			want:    []int32{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3 ,",
			args: args{
				sep: ",",
			},
			want:    []int32{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "case 5",
			s:    "8589934592",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Int32Slice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int32s() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int32s() got = %v, want %v", got, tt.want)
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
				sep: ",",
			},
			want:    []uint64{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3, ",
			args: args{
				sep: ",",
			},
			want:    []uint64{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Uint64Slice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint64s() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint64s() got = %v, want %v", got, tt.want)
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
				sep: ",",
			},
			want:    []uint32{1},
			wantErr: false,
		},
		{
			name: "case 2",
			s:    "1 , 2 , 3 ,",
			args: args{
				sep: ",",
			},
			want:    []uint32{1, 2, 3},
			wantErr: false,
		},
		{
			name: "case 3",
			s:    "",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "case 4",
			s:    "abc",
			args: args{
				sep: ",",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Uint32Slice(tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint32s() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uint32s() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_Split(t *testing.T) {
	type args struct {
		sep string
	}
	tests := []struct {
		name string
		s    String
		args args
		want []string
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
		{
			name: "case 3",
			s:    " ",
			args: args{
				sep: ",",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Split(tt.args.sep)
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
		xi, _ = xs.Int64Slice(",")
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

func TestStringSlice_Unique(t *testing.T) {
	tests := []struct {
		name string
		ss   StringSlice
		want StringSlice
	}{
		{
			name: "case 1",
			ss:   []string{"a", "b", "b"},
			want: []string{"a", "b"},
		},
		{
			name: "case 2",
			ss:   []string{"a"},
			want: []string{"a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ss.Unique(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}
