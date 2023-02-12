// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/02/11

package fstypes

func SliceMerge[T any](a []T, b []T) []T {
	cp := make([]T, 0, len(a))
	return append(append(cp, a...), b...)
}
