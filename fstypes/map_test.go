// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/16

package fstypes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRingMap(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		rm := NewRingMap[int, int](3)
		rm.Set(1, 10)
		require.Equal(t, map[int]int{1: 10}, rm.Values())
		require.Equal(t, 1, rm.Len())

		rm.Set(2, 20)
		require.Equal(t, map[int]int{1: 10, 2: 20}, rm.Values())
		require.Equal(t, 2, rm.Len())

		rm.Set(3, 30)
		require.Equal(t, map[int]int{1: 10, 2: 20, 3: 30}, rm.Values())
		require.Equal(t, 3, rm.Len())

		rm.Set(4, 40)
		require.Equal(t, map[int]int{2: 20, 3: 30, 4: 40}, rm.Values())
		require.Equal(t, 3, rm.Len())

		oldKey1, oldValue1, s1 := rm.SetSwap(4, 50)
		require.True(t, s1)
		require.Equal(t, 4, oldKey1)
		require.Equal(t, 40, oldValue1)

		oldKey2, oldValue2, s2 := rm.SetSwap(5, 50)
		require.Equal(t, 2, oldKey2)
		require.Equal(t, 20, oldValue2)
		require.True(t, s2)
	})
}
