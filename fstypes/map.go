// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/15

package fstypes

import "sync"

func NewRingMap[K comparable, V any](caption int) *RingMap[K, V] {
	return &RingMap[K, V]{
		caption: caption,
		keys:    NewRingUnique[K](caption),
	}
}

type RingMap[K comparable, V any] struct {
	keys    *RingUnique[K]
	values  sync.Map
	caption int
}

func (rm *RingMap[K, V]) Set(key K, value V) {
	old, swapped := rm.keys.AddSwap(key)
	if swapped {
		rm.values.Delete(old)
	}
	rm.values.Store(key, value)
}

func (rm *RingMap[K, V]) SetSwap(key K, value V) (oldKey K, oldValue V, swapped bool) {
	oldKey, swapped = rm.keys.AddSwap(key)
	if swapped {
		if ol, loaded := rm.values.LoadAndDelete(oldKey); loaded {
			oldValue = ol.(V)
		}
	}
	rm.values.Store(key, value)
	return oldKey, oldValue, swapped
}

func (rm *RingMap[K, V]) Len() int {
	return rm.keys.Len()
}

func (rm *RingMap[K, V]) Delete(key K) {
	rm.keys.Delete(key)
	rm.values.Delete(key)
}

func (rm *RingMap[K, V]) Get(key K) (value V) {
	v, ok := rm.values.Load(key)
	if !ok {
		return value
	}
	return v.(V)
}

func (rm *RingMap[K, V]) GetV2(key K) (value V, found bool) {
	v, ok := rm.values.Load(key)
	if !ok {
		return value, false
	}
	return v.(V), true
}

func (rm *RingMap[K, V]) Range(fn func(key K, value V) bool) {
	rm.values.Range(func(key, value any) bool {
		return fn(key.(K), value.(V))
	})
}

func (rm *RingMap[K, V]) Values() map[K]V {
	length := rm.Len()
	if length == 0 {
		return nil
	}
	vs := make(map[K]V, length)
	rm.Range(func(k K, v V) bool {
		vs[k] = v
		return true
	})
	return vs
}
