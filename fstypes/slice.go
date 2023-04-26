// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/02/11

package fstypes

// SliceMerge merge 多个 slice 为一个，并最终返回一个新的 slice
func SliceMerge[T any](items ...[]T) []T {
	switch len(items) {
	case 0:
		return nil
	case 1:
		return SliceCopy(items[0])
	}

	var n int
	for i := 0; i < len(items); i++ {
		n += len(items[i])
	}
	cp := make([]T, 0, n)
	for i := 0; i < len(items); i++ {
		cp = append(cp, items[i]...)
	}
	return cp
}

// SliceCopy 复制一个 slice
//
//	若原 slice == nil 会返回 nil
//	其他情况总是返回一个新的 slice，及时 len == 0
func SliceCopy[T any](a []T) []T {
	if a == nil {
		return nil
	}
	cp := make([]T, 0, len(a))
	return append(cp, a...)
}
