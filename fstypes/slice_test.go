// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/19

package fstypes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSliceMerge(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		var arr []string
		got := SliceMerge(arr)
		require.Nil(t, got)
	})

	t.Run("1", func(t *testing.T) {
		v1 := []string{"1", "10"}
		got := SliceMerge(v1)
		require.Equal(t, v1, got)
		require.Equal(t, &v1, &got)
	})

	t.Run("2", func(t *testing.T) {
		v1 := []string{"1", "10"}
		v2 := []string{"2", "20"}
		got := SliceMerge(v1, v2)
		want := []string{"1", "10", "2", "20"}
		require.Equal(t, want, got)
	})

	t.Run("4", func(t *testing.T) {
		v1 := []string{"1", "10"}
		v2 := []string{"2", "20"}
		v3 := []string{"3", "30"}
		got := SliceMerge(v1, v2, nil, v3)
		want := []string{"1", "10", "2", "20", "3", "30"}
		require.Equal(t, want, got)
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
	require.Equal(t, want, got)
}

func TestSliceHas(t *testing.T) {
	arr := []string{"a", "a", "b"}
	require.True(t, SliceHas(arr, "a"))
	require.True(t, SliceHas(arr, "b"))
	require.False(t, SliceHas(arr, "c"))
}

func TestSliceDelete(t *testing.T) {
	arr := []string{"a", "a", "b"}

	require.Equal(t, []string{"a", "a"}, SliceDelete(arr, "b"))
	require.Equal(t, []string{"a", "a", "b"}, arr)

	require.Equal(t, []string{"b"}, SliceDelete(arr, "a"))
	require.Equal(t, []string{"a", "a", "b"}, arr)

	require.Equal(t, []string{"a", "a", "b"}, SliceDelete(arr, "c"))

	require.Equal(t, []string(nil), SliceDelete(arr, "a", "b"))
	require.Equal(t, []string{"a", "a", "b"}, arr)

	var a2 []string
	require.Empty(t, SliceDelete(a2, "c"))
}

func TestSliceJoin(t *testing.T) {
	arr1 := []string{"a", "b"}
	got1 := SliceJoin(arr1, ",")
	require.Equal(t, "a,b", got1)

	var arr2 []int
	got2 := SliceJoin(arr2, ",")
	require.Equal(t, "", got2)

	arr3 := []int{1, 9}
	got3 := SliceJoin(arr3, ",")
	require.Equal(t, "1,9", got3)
}

func TestSliceValuesAllow(t *testing.T) {
	arr1 := []string{"a", "b"}
	require.NoError(t, SliceValuesAllow(arr1, arr1))
	allow1 := []string{"a", "c"}
	require.Error(t, SliceValuesAllow(arr1, allow1))
}
