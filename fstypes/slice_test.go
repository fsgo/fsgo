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
		got := SliceMerge[string]()
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
