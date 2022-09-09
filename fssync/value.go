// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/9/9

package fssync

import (
	"sync/atomic"
)

// AtomicValue 存储值类型
type AtomicValue[T any] struct {
	_ noCopy
	v atomic.Value
}

// Load atomically loads
func (a *AtomicValue[T]) Load() T {
	v, _ := a.v.Load().(T)
	return v
}

// Store atomically store
func (a *AtomicValue[T]) Store(v T) {
	a.v.Store(v)
}

// Swap atomically swap
func (a *AtomicValue[T]) Swap(v T) T {
	old, _ := a.v.Swap(v).(T)
	return old
}

// CompareAndSwap atomically compare and swap
func (a *AtomicValue[T]) CompareAndSwap(old, new T) (swapped bool) {
	return a.v.CompareAndSwap(old, new)
}
