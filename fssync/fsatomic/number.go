// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/19

package fsatomic

import (
	"sync/atomic"
)

type NumberInt64[T ~int64] int64

func (n *NumberInt64[T]) Load() T {
	v := atomic.LoadInt64((*int64)(n))
	return T(v)
}

func (n *NumberInt64[T]) Add(v T) {
	atomic.AddInt64((*int64)(n), int64(v))
}

func (n *NumberInt64[T]) Store(v T) {
	atomic.StoreInt64((*int64)(n), int64(v))
}

func (n *NumberInt64[T]) Swap(v T) (old T) {
	o := atomic.SwapInt64((*int64)(n), int64(v))
	return T(o)
}

func (n *NumberInt64[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return atomic.CompareAndSwapInt64((*int64)(n), int64(old), int64(new))
}

// -------------------------------------------------------------------------------

type NumberInt32[T ~int32] int32

func (n *NumberInt32[T]) Load() T {
	v := atomic.LoadInt32((*int32)(n))
	return T(v)
}

func (n *NumberInt32[T]) Add(v T) {
	atomic.AddInt32((*int32)(n), int32(v))
}

func (n *NumberInt32[T]) Store(v T) {
	atomic.StoreInt32((*int32)(n), int32(v))
}

func (n *NumberInt32[T]) Swap(v T) (old T) {
	o := atomic.SwapInt32((*int32)(n), int32(v))
	return T(o)
}

func (n *NumberInt32[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return atomic.CompareAndSwapInt32((*int32)(n), int32(old), int32(new))
}

// -------------------------------------------------------------------------------

type NumberUint64[T ~uint64] uint64

func (n *NumberUint64[T]) Load() T {
	v := atomic.LoadUint64((*uint64)(n))
	return T(v)
}

func (n *NumberUint64[T]) Add(v T) {
	atomic.AddUint64((*uint64)(n), uint64(v))
}

func (n *NumberUint64[T]) Store(v T) {
	atomic.StoreUint64((*uint64)(n), uint64(v))
}

func (n *NumberUint64[T]) Swap(v T) (old T) {
	o := atomic.SwapUint64((*uint64)(n), uint64(v))
	return T(o)
}

func (n *NumberUint64[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return atomic.CompareAndSwapUint64((*uint64)(n), uint64(old), uint64(new))
}

// -------------------------------------------------------------------------------

type NumberUint32[T ~uint32] uint32

func (n *NumberUint32[T]) Load() T {
	v := atomic.LoadUint32((*uint32)(n))
	return T(v)
}

func (n *NumberUint32[T]) Add(v T) {
	atomic.AddUint32((*uint32)(n), uint32(v))
}

func (n *NumberUint32[T]) Store(v T) {
	atomic.StoreUint32((*uint32)(n), uint32(v))
}

func (n *NumberUint32[T]) Swap(v T) (old T) {
	o := atomic.SwapUint32((*uint32)(n), uint32(v))
	return T(o)
}

func (n *NumberUint32[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return atomic.CompareAndSwapUint32((*uint32)(n), uint32(old), uint32(new))
}
