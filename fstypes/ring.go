// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/9/15

package fstypes

import (
	"fmt"
	"sync"
)

func NewRing[T any](caption int) *Ring[T] {
	if caption <= 0 {
		panic(fmt.Errorf("invalid Ring caption %d", caption))
	}
	return &Ring[T]{
		caption: caption,
		values:  make([]T, caption),
	}
}

type Ring[T any] struct {
	values  []T
	caption int
	length  int
	index   int
	mux     sync.RWMutex
}

func (r *Ring[T]) Add(v T) {
	r.mux.Lock()
	r.values[r.index] = v
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}
	r.mux.Unlock()
}

// AddSwap 添加并返回被替换的值
func (r *Ring[T]) AddSwap(v T) (old T, swapped bool) {
	r.mux.Lock()
	if r.length > r.index {
		old = r.values[r.index]
		swapped = true
	}
	r.values[r.index] = v
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}
	r.mux.Unlock()
	return old, swapped
}

func (r *Ring[T]) Len() int {
	r.mux.RLock()
	val := r.length
	r.mux.RUnlock()
	return val
}

// Range 遍历，先加入的会先遍历
func (r *Ring[T]) Range(fn func(v T) bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if r.length == 0 {
		return
	}

	if r.length != r.caption {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下

	for i := r.index; i < r.caption; i++ {
		if !fn(r.values[i]) {
			return
		}
	}

	for i := 0; i < r.index; i++ {
		if !fn(r.values[i]) {
			return
		}
	}
}

// Values 返回所有值，先加入的排在前面
func (r *Ring[T]) Values() []T {
	r.mux.RLock()
	defer r.mux.RUnlock()
	length := r.length
	if length == 0 {
		return nil
	}
	vs := make([]T, 0, length)
	if length != r.caption {
		vs = append(vs, r.values[:length]...)
		return vs
	}
	// 容量满的情况下
	vs = append(vs, r.values[r.index:]...)
	vs = append(vs, r.values[:r.index]...)
	return vs
}

func NewRingUnique[T comparable](caption int) *RingUnique[T] {
	if caption <= 0 {
		panic(fmt.Errorf("invalid Ring caption %d", caption))
	}
	return &RingUnique[T]{
		caption:    caption,
		values:     make([]T, caption),
		valueIndex: make(map[T]int, caption),
	}
}

// RingUnique 具有唯一值的 ring list
type RingUnique[T comparable] struct {
	values     []T
	valueIndex map[T]int
	caption    int
	length     int
	index      int
	mux        sync.RWMutex
}

func (r *RingUnique[T]) Add(v T) {
	r.mux.Lock()
	defer r.mux.Unlock()

	oldIndex, has := r.valueIndex[v]
	if has {
		r.values[oldIndex] = v
		return
	}

	r.values[r.index] = v
	r.valueIndex[v] = r.index
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}
}

// AddSwap 添加并返回被替换的值
func (r *RingUnique[T]) AddSwap(v T) (old T, swapped bool) {
	r.mux.Lock()
	defer r.mux.Unlock()

	oldIndex, has := r.valueIndex[v]
	if has {
		old = r.values[oldIndex]
		r.values[oldIndex] = v
		return old, true
	}

	if r.length > r.index {
		old = r.values[r.index]
		swapped = true
	}
	r.values[r.index] = v
	r.valueIndex[v] = r.index
	r.index++
	if r.index == r.caption {
		r.index = 0
	}
	if r.length < r.caption {
		r.length++
	}

	return old, swapped
}

func (r *RingUnique[T]) Len() int {
	r.mux.RLock()
	val := r.length
	r.mux.RUnlock()
	return val
}

// Range 遍历，先加入的会先遍历
func (r *RingUnique[T]) Range(fn func(v T) bool) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if r.length == 0 {
		return
	}

	if r.length != r.caption {
		for i := 0; i < r.length; i++ {
			if !fn(r.values[i]) {
				return
			}
		}
		return
	}

	// 容量满的情况下

	for i := r.index; i < r.caption; i++ {
		if !fn(r.values[i]) {
			return
		}
	}

	for i := 0; i < r.index; i++ {
		if !fn(r.values[i]) {
			return
		}
	}
}

// Values 返回所有值，先加入的排在前面
func (r *RingUnique[T]) Values() []T {
	r.mux.RLock()
	defer r.mux.RUnlock()
	length := r.length
	if length == 0 {
		return nil
	}
	vs := make([]T, 0, length)
	if length != r.caption {
		vs = append(vs, r.values[:length]...)
		return vs
	}
	// 容量满的情况下
	vs = append(vs, r.values[r.index:]...)
	vs = append(vs, r.values[:r.index]...)
	return vs
}

func (r *RingUnique[T]) Delete(key T) (found bool) {
	// todo
	return found
}
