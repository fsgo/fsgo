// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/16

package fstypes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRing(t *testing.T) {
	t.Run("cap-0", func(t *testing.T) {
		defer func() {
			require.NotNil(t, recover())
		}()
		_ = NewRing[int](0)
	})
}

func TestRing(t *testing.T) {
	t.Run("Add-3", func(t *testing.T) {
		r1 := NewRing[int](3)
		require.Nil(t, r1.Values())
		for i := 0; i < 10; i++ {
			r1.Add(i)

			switch i {
			case 0:
				require.Equal(t, []int{0}, r1.Values())
			case 1:
				require.Equal(t, []int{0, 1}, r1.Values())
			case 2:
				require.Equal(t, []int{0, 1, 2}, r1.Values())
			case 3:
				require.Equal(t, []int{1, 2, 3}, r1.Values())
			case 4:
				require.Equal(t, []int{2, 3, 4}, r1.Values())
			}

			if i < 2 {
				require.Equal(t, i+1, r1.Len())
			} else {
				require.Equal(t, 3, r1.Len())
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		want1 := []int{9, 7, 8}
		require.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		require.Equal(t, want2, r1.Values())
	})

	t.Run("AddSwap", func(t *testing.T) {
		r1 := NewRing[int](3)
		for i := 0; i < 10; i++ {
			old, swapped := r1.AddSwap(i)

			switch i {
			case 0:
				require.Equal(t, []int{0}, r1.Values())
				require.Equal(t, 0, old)
				require.False(t, swapped)
			case 1:
				require.Equal(t, []int{0, 1}, r1.Values())
				require.Equal(t, 0, old)
				require.False(t, swapped)
			case 2:
				require.Equal(t, []int{0, 1, 2}, r1.Values())
				require.Equal(t, 0, old)
				require.False(t, swapped)
			case 3:
				require.Equal(t, []int{1, 2, 3}, r1.Values())
				require.Equal(t, 0, old)
				require.True(t, swapped)
			case 4:
				require.Equal(t, []int{2, 3, 4}, r1.Values())
				require.Equal(t, 1, old)
				require.True(t, swapped)
			}

			if i < 2 {
				require.Equal(t, i+1, r1.Len())
			} else {
				require.Equal(t, 3, r1.Len())
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		want1 := []int{9, 7, 8}
		require.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		require.Equal(t, want2, r1.Values())
	})
}

func TestNewRingUnique(t *testing.T) {
	t.Run("cap-0", func(t *testing.T) {
		defer func() {
			require.NotNil(t, recover())
		}()
		_ = NewRingUnique[int](0)
	})
}

func TestRingUnique(t *testing.T) {
	t.Run("Add-3", func(t *testing.T) {
		r1 := NewRingUnique[int](3)
		require.Nil(t, r1.Values())
		for i := 0; i < 10; i++ {
			r1.Add(i)

			switch i {
			case 0:
				require.Equal(t, []int{0}, r1.Values())
			case 1:
				require.Equal(t, []int{0, 1}, r1.Values())
			case 2:
				require.Equal(t, []int{0, 1, 2}, r1.Values())
			case 3:
				require.Equal(t, []int{1, 2, 3}, r1.Values())
			case 4:
				require.Equal(t, []int{2, 3, 4}, r1.Values())
			}

			if i < 2 {
				require.Equal(t, i+1, r1.Len())
			} else {
				require.Equal(t, 3, r1.Len())
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		want1 := []int{9, 7, 8}
		require.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		require.Equal(t, want2, r1.Values())
	})

	t.Run("AddSwap", func(t *testing.T) {
		r1 := NewRingUnique[int](3)
		for i := 0; i < 10; i++ {
			old, swapped := r1.AddSwap(i)

			switch i {
			case 0:
				require.Equal(t, []int{0}, r1.Values())
				require.Equal(t, 0, old)
				require.False(t, swapped)
			case 1:
				require.Equal(t, []int{0, 1}, r1.Values())
				require.Equal(t, 0, old)
				require.False(t, swapped)
			case 2:
				require.Equal(t, []int{0, 1, 2}, r1.Values())
				require.Equal(t, 0, old)
				require.False(t, swapped)
			case 3:
				require.Equal(t, []int{1, 2, 3}, r1.Values())
				require.Equal(t, 0, old)
				require.True(t, swapped)
			case 4:
				require.Equal(t, []int{2, 3, 4}, r1.Values())
				require.Equal(t, 1, old)
				require.True(t, swapped)
			}

			if i < 2 {
				require.Equal(t, i+1, r1.Len())
			} else {
				require.Equal(t, 3, r1.Len())
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		want1 := []int{9, 7, 8}
		require.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		require.Equal(t, want2, r1.Values())
	})

	t.Run("add dup", func(t *testing.T) {
		r1 := NewRingUnique[int](3)
		for i := 0; i < 10; i++ {
			r1.Add(1)
			require.Equal(t, []int{1}, r1.Values())
		}
		for i := 0; i < 10; i++ {
			old, swapped := r1.AddSwap(1)
			require.Equal(t, []int{1}, r1.Values())
			require.Equal(t, 1, old)
			require.True(t, swapped)
			require.Equal(t, []int{1}, r1.Values())
		}

		{
			old, swapped := r1.AddSwap(2)
			require.Equal(t, 0, old)
			require.Equal(t, 2, r1.Len())
			require.False(t, swapped)
			require.Equal(t, []int{1, 2}, r1.Values())
		}

		{
			old, swapped := r1.AddSwap(3)
			require.Equal(t, 0, old)
			require.False(t, swapped)
			require.Equal(t, []int{1, 2, 3}, r1.Values())
		}
		{
			old, swapped := r1.AddSwap(4)
			require.Equal(t, []int{2, 3, 4}, r1.Values())
			require.Equal(t, 1, old)
			require.True(t, swapped)
		}
	})
}
