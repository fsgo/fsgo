// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/16

package fstypes

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestNewRing(t *testing.T) {
	t.Run("cap-0", func(t *testing.T) {
		defer func() {
			fst.NotNil(t, recover())
		}()
		_ = NewRing[int](0)
	})
}

func TestRing(t *testing.T) {
	t.Run("Add-3", func(t *testing.T) {
		r1 := NewRing[int](3)
		fst.Nil(t, r1.Values())
		for i := 0; i < 10; i++ {
			r1.Add(i)

			switch i {
			case 0:
				fst.Equal(t, []int{0}, r1.Values())
			case 1:
				fst.Equal(t, []int{0, 1}, r1.Values())
			case 2:
				fst.Equal(t, []int{0, 1, 2}, r1.Values())
			case 3:
				fst.Equal(t, []int{1, 2, 3}, r1.Values())
			case 4:
				fst.Equal(t, []int{2, 3, 4}, r1.Values())
			}

			if i < 2 {
				fst.Equal(t, i+1, r1.Len())
			} else {
				fst.Equal(t, 3, r1.Len())
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		want1 := []int{9, 7, 8}
		fst.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		fst.Equal(t, want2, r1.Values())
	})

	t.Run("AddSwap", func(t *testing.T) {
		r1 := NewRing[int](3)
		for i := 0; i < 10; i++ {
			old, swapped := r1.AddSwap(i)

			switch i {
			case 0:
				fst.Equal(t, []int{0}, r1.Values())
				fst.Equal(t, 0, old)
				fst.False(t, swapped)
			case 1:
				fst.Equal(t, []int{0, 1}, r1.Values())
				fst.Equal(t, 0, old)
				fst.False(t, swapped)
			case 2:
				fst.Equal(t, []int{0, 1, 2}, r1.Values())
				fst.Equal(t, 0, old)
				fst.False(t, swapped)
			case 3:
				fst.Equal(t, []int{1, 2, 3}, r1.Values())
				fst.Equal(t, 0, old)
				fst.True(t, swapped)
			case 4:
				fst.Equal(t, []int{2, 3, 4}, r1.Values())
				fst.Equal(t, 1, old)
				fst.True(t, swapped)
			}

			if i < 2 {
				fst.Equal(t, i+1, r1.Len())
			} else {
				fst.Equal(t, 3, r1.Len())
			}
		}
		// 0,1,2 | 3,4,5 | 6,7,8 | 9
		want1 := []int{9, 7, 8}
		fst.Equal(t, want1, r1.values)

		want2 := []int{7, 8, 9}
		fst.Equal(t, want2, r1.Values())
	})
}

func TestNewRingUnique(t *testing.T) {
	t.Run("cap-0", func(t *testing.T) {
		defer func() {
			fst.NotNil(t, recover())
		}()
		_ = NewRingUnique[int](0)
	})
}

func TestRingUnique1(t *testing.T) {
	r1 := NewRingUnique[int](3)
	fst.Nil(t, r1.Values())
	for i := 0; i < 10; i++ {
		r1.Add(i)

		switch i {
		case 0:
			fst.Equal(t, []int{0}, r1.Values())
		case 1:
			fst.Equal(t, []int{0, 1}, r1.Values())
		case 2:
			fst.Equal(t, []int{0, 1, 2}, r1.Values())
		case 3:
			fst.Equal(t, []int{1, 2, 3}, r1.Values())
		case 4:
			fst.Equal(t, []int{2, 3, 4}, r1.Values())
		}

		if i < 2 {
			fst.Equal(t, i+1, r1.Len())
		} else {
			fst.Equal(t, 3, r1.Len())
		}
	}
	// 0,1,2 | 3,4,5 | 6,7,8 | 9
	want1 := []int{9, 7, 8}
	fst.Equal(t, want1, r1.values)

	want2 := []int{7, 8, 9}
	fst.Equal(t, want2, r1.Values())
}

func TestRingUnique2(t *testing.T) {
	r1 := NewRingUnique[int](3)
	for i := 0; i < 10; i++ {
		old, swapped := r1.AddSwap(i)

		switch i {
		case 0:
			fst.Equal(t, []int{0}, r1.Values())
			fst.Equal(t, 0, old)
			fst.False(t, swapped)
		case 1:
			fst.Equal(t, []int{0, 1}, r1.Values())
			fst.Equal(t, 0, old)
			fst.False(t, swapped)
		case 2:
			fst.Equal(t, []int{0, 1, 2}, r1.Values())
			fst.Equal(t, 0, old)
			fst.False(t, swapped)
		case 3:
			fst.Equal(t, []int{1, 2, 3}, r1.Values())
			fst.Equal(t, 0, old)
			fst.True(t, swapped)
		case 4:
			fst.Equal(t, []int{2, 3, 4}, r1.Values())
			fst.Equal(t, 1, old)
			fst.True(t, swapped)
		}

		if i < 2 {
			fst.Equal(t, i+1, r1.Len())
		} else {
			fst.Equal(t, 3, r1.Len())
		}
	}
	// 0,1,2 | 3,4,5 | 6,7,8 | 9
	want1 := []int{9, 7, 8}
	fst.Equal(t, want1, r1.values)

	want2 := []int{7, 8, 9}
	fst.Equal(t, want2, r1.Values())
}

func TestRingUnique3(t *testing.T) {
	r1 := NewRingUnique[int](3)
	for i := 0; i < 10; i++ {
		r1.Add(1)
		fst.Equal(t, []int{1}, r1.Values())
	}
	for i := 0; i < 10; i++ {
		old, swapped := r1.AddSwap(1)
		fst.Equal(t, []int{1}, r1.Values())
		fst.Equal(t, 1, old)
		fst.True(t, swapped)
		fst.Equal(t, []int{1}, r1.Values())
	}

	{
		old, swapped := r1.AddSwap(2)
		fst.Equal(t, 0, old)
		fst.Equal(t, 2, r1.Len())
		fst.False(t, swapped)
		fst.Equal(t, []int{1, 2}, r1.Values())
	}

	{
		old, swapped := r1.AddSwap(3)
		fst.Equal(t, 0, old)
		fst.False(t, swapped)
		fst.Equal(t, []int{1, 2, 3}, r1.Values())
	}
	{
		old, swapped := r1.AddSwap(4)
		fst.Equal(t, []int{2, 3, 4}, r1.Values())
		fst.Equal(t, 1, old)
		fst.True(t, swapped)
	}
}
