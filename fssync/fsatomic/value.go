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

type data[T any] struct {
	Data T
}

// Load atomically loads
func (a *Value[T]) Load() (v T) {
	val, ok := a.storage.Load().(data[T])
	if ok {
		return val.Data
	}
	return v
}

// Store atomically store
func (a *Value[T]) Store(v T) {
	a.storage.Store(data[T]{Data: v})
}

// Swap atomically swap
func (a *Value[T]) Swap(v T) (o T) {
	old, ok := a.storage.Swap(data[T]{Data: v}).(data[T])
	if ok {
		return old.Data
	}
	return o
}

// CompareAndSwap atomically compare and swap
func (a *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	return a.storage.CompareAndSwap(data[T]{Data: old}, data[T]{Data: new})
}
