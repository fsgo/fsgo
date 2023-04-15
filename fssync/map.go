// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/02/11

package fssync

import (
	"sync"

	"github.com/fsgo/fsgo/fssync/internal"
)

type Map[K comparable, V any] struct {
	_       internal.NoCopy
	storage sync.Map
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	v1, ok1 := m.storage.Load(key)
	if !ok1 {
		return value, false
	}
	v2, ok2 := v1.(V)
	return v2, ok2
}

func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, ok := m.storage.LoadAndDelete(key)
	if !ok {
		return value, false
	}
	return v.(V), true
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	v, ok := m.storage.LoadOrStore(key, value)
	return v.(V), ok
}

func (m *Map[K, V]) Store(key K, value V) {
	m.storage.Store(key, value)
}

func (m *Map[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	p, ok := m.storage.Swap(key, value)
	if !ok {
		return previous, false
	}
	return p.(V), true
}

func (m *Map[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.storage.CompareAndDelete(key, old)
}

func (m *Map[K, V]) CompareAndSwap(key K, old V, new V) bool {
	return m.storage.CompareAndSwap(key, old, new)
}

func (m *Map[K, V]) Delete(key K) {
	m.storage.Delete(key)
}

func (m *Map[K, V]) Range(fn func(key K, value V) bool) {
	m.storage.Range(func(key, value any) bool {
		return fn(key.(K), value.(V))
	})
}

func (m *Map[K, V]) Count() int {
	var c int
	m.storage.Range(func(_, _ any) bool {
		c++
		return true
	})
	return c
}

func (m *Map[K, V]) Purge() {
	m.storage.Range(func(key any, _ any) bool {
		m.storage.Delete(key)
		return true
	})
}
