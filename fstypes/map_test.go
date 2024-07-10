// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/16

package fstypes

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestRingMap(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		rm := NewRingMap[int, int](3)
		rm.Set(1, 10)
		fst.Equal(t, map[int]int{1: 10}, rm.Values())
		fst.Equal(t, 1, rm.Len())

		rm.Set(2, 20)
		fst.Equal(t, map[int]int{1: 10, 2: 20}, rm.Values())
		fst.Equal(t, 2, rm.Len())

		rm.Set(3, 30)
		fst.Equal(t, map[int]int{1: 10, 2: 20, 3: 30}, rm.Values())
		fst.Equal(t, 3, rm.Len())

		rm.Set(4, 40)
		fst.Equal(t, map[int]int{2: 20, 3: 30, 4: 40}, rm.Values())
		fst.Equal(t, 3, rm.Len())

		oldKey1, oldValue1, s1 := rm.SetSwap(4, 50)
		fst.True(t, s1)
		fst.Equal(t, 4, oldKey1)
		fst.Equal(t, 40, oldValue1)

		oldKey2, oldValue2, s2 := rm.SetSwap(5, 50)
		fst.Equal(t, 2, oldKey2)
		fst.Equal(t, 20, oldValue2)
		fst.True(t, s2)
	})
}
