// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/19

package fstypes

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestSliceMerge(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		var arr []string
		got := SliceMerge(arr)
		fst.Nil(t, got)
	})

	t.Run("1", func(t *testing.T) {
		v1 := []string{"1", "10"}
		got := SliceMerge(v1)
		fst.Equal(t, v1, got)
		fst.Equal(t, &v1, &got)
	})

	t.Run("2", func(t *testing.T) {
		v1 := []string{"1", "10"}
		v2 := []string{"2", "20"}
		got := SliceMerge(v1, v2)
		want := []string{"1", "10", "2", "20"}
		fst.Equal(t, want, got)
	})

	t.Run("4", func(t *testing.T) {
		v1 := []string{"1", "10"}
		v2 := []string{"2", "20"}
		v3 := []string{"3", "30"}
		got := SliceMerge(v1, v2, nil, v3)
		want := []string{"1", "10", "2", "20", "3", "30"}
		fst.Equal(t, want, got)
	})
}

func BenchmarkSliceMerge(b *testing.B) {
	v1 := []string{"1", "10"}
	v2 := []string{"2", "20"}
	for i := 0; i < b.N; i++ {
		_ = SliceMerge(v1, v2)
	}
}

func BenchmarkSliceCopy(b *testing.B) {
	v1 := []string{"1", "10"}
	for i := 0; i < b.N; i++ {
		_ = SliceCopy(v1)
	}
}

func TestSliceUnique(t *testing.T) {
	arr := []string{"a", "a", "b"}
	got := SliceUnique(arr)
	want := []string{"a", "b"}
	fst.Equal(t, want, got)
}

func TestSliceHas(t *testing.T) {
	arr := []string{"a", "a", "b"}
	fst.True(t, SliceHas(arr, "a"))
	fst.True(t, SliceHas(arr, "b"))
	fst.False(t, SliceHas(arr, "c"))
}

func TestSliceDelete(t *testing.T) {
	arr := []string{"a", "a", "b"}

	fst.Equal(t, []string{"a", "a"}, SliceDelete(arr, "b"))
	fst.Equal(t, []string{"a", "a", "b"}, arr)

	fst.Equal(t, []string{"b"}, SliceDelete(arr, "a"))
	fst.Equal(t, []string{"a", "a", "b"}, arr)

	fst.Equal(t, []string{"a", "a", "b"}, SliceDelete(arr, "c"))

	fst.Equal(t, []string(nil), SliceDelete(arr, "a", "b"))
	fst.Equal(t, []string{"a", "a", "b"}, arr)

	var a2 []string
	fst.Empty(t, SliceDelete(a2, "c"))
}

func TestSliceJoin(t *testing.T) {
	arr1 := []string{"a", "b"}
	got1 := SliceJoin(arr1, ",")
	fst.Equal(t, "a,b", got1)

	var arr2 []int
	got2 := SliceJoin(arr2, ",")
	fst.Equal(t, "", got2)

	arr3 := []int{1, 9}
	got3 := SliceJoin(arr3, ",")
	fst.Equal(t, "1,9", got3)
}

func TestSliceValuesAllow(t *testing.T) {
	arr1 := []string{"a", "b"}
	fst.NoError(t, SliceValuesAllow(arr1, arr1))
	allow1 := []string{"a", "c"}
	fst.Error(t, SliceValuesAllow(arr1, allow1))
}
