// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/02/11

package fstypes

func SliceMerge[T any](items ...[]T) []T {
	switch len(items) {
	case 0:
		return nil
	case 1:
		return items[0]
	}

	var n int
	for i := 0; i < len(items); i++ {
		n += len(items[i])
	}
	cp := make([]T, n)
	var id int
	for i := 0; i < len(items); i++ {
		v1 := items[i]
		for j := 0; j < len(v1); j++ {
			cp[id] = v1[j]
			id++
		}
	}
	return cp
}
