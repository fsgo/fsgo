// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/02/11

package fssync

import (
	"sync"
)

type Map[T any] struct {
	_       noCopy
	storage sync.Map
}

func (m *Map[T]) Load(key any) (value T, ok bool) {
	v1, ok1 := m.storage.Load(key)
	if !ok1 {
		return value, false
	}
	v2, ok2 := v1.(T)
	return v2, ok2
}

func (m *Map[T]) LoadAndDelete(key any) (value T, loaded bool) {
	v, ok := m.storage.LoadAndDelete(key)
	if !ok {
		return value, false
	}
	return v.(T), true
}

func (m *Map[T]) LoadOrStore(key any, value T) (actual T, loaded bool) {
	v, ok := m.storage.LoadOrStore(key, value)
	return v.(T), ok
}

func (m *Map[T]) Store(key any, value T) {
	m.storage.Store(key, value)
}

func (m *Map[T]) Swap(key any, value T) (previous T, loaded bool) {
	p, ok := m.storage.Swap(key, value)
	if !ok {
		return previous, false
	}
	return p.(T), true
}

func (m *Map[T]) CompareAndDelete(key any, old T) (deleted bool) {
	return m.storage.CompareAndDelete(key, old)
}

func (m *Map[T]) CompareAndSwap(key any, old T, new T) bool {
	return m.storage.CompareAndSwap(key, old, new)
}

func (m *Map[T]) Delete(key any) {
	m.storage.Delete(key)
}

func (m *Map[T]) Range(fn func(key any, value T) bool) {
	m.storage.Range(func(key, value any) bool {
		return fn(key, value.(T))
	})
}
