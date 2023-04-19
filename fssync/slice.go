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
	defer s.mux.Unlock()
	s.items = append(s.items, v...)
}

func (s *Slice[T]) Load() []T {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.items
}

func (s *Slice[T]) Purge() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.items = nil
}

func (s *Slice[T]) Store(all []T) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.items = append(make([]T, 0, len(all)), all...)
}

func (s *Slice[T]) Swap(all []T) []T {
	s.mux.Lock()
	defer s.mux.Unlock()
	old := s.items
	s.items = append(make([]T, 0, len(all)), all...)
	return old
}
