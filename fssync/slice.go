// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/19

package fssync

import "sync"

type Slice[T any] struct {
	items []T
	mux   sync.RWMutex
}

func (s *Slice[T]) Add(v ...T) {
	s.mux.Lock()
	s.items = append(s.items, v...)
	s.mux.Unlock()
}

func (s *Slice[T]) Load() []T {
	s.mux.RLock()
	val := s.items
	s.mux.RUnlock()
	return val
}

func (s *Slice[T]) Purge() {
	s.mux.Lock()
	s.items = nil
	s.mux.Unlock()
}

func (s *Slice[T]) Store(all []T) {
	s.mux.Lock()
	s.items = append(make([]T, 0, len(all)), all...)
	s.mux.Unlock()
}

func (s *Slice[T]) Swap(all []T) []T {
	s.mux.Lock()
	old := s.items
	s.items = append(make([]T, 0, len(all)), all...)
	s.mux.Unlock()
	return old
}
