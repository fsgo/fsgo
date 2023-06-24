// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/15

package fsatomic

import (
	"sync/atomic"

	"github.com/fsgo/fsgo/fssync/internal"
)

type (
	Error = Value[error]

	String = Value[string]

	Func = Value[func() error]

	FuncVoid = Value[func()]
)

// Value 存储值类型
type Value[T any] struct {
	_       internal.NoCopy
	storage atomic.Value
}

// Load atomically loads
func (a *Value[T]) Load() T {
	v, _ := a.storage.Load().(T)
	return v
}

// Store atomically store
func (a *Value[T]) Store(v T) {
	a.storage.Store(v)
}

// Swap atomically swap
func (a *Value[T]) Swap(v T) T {
	old, _ := a.storage.Swap(v).(T)
	return old
}

// CompareAndSwap atomically compare and swap
func (a *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	return a.storage.CompareAndSwap(old, new)
}
