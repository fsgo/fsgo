// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/02/11

package fstypes

import (
	"fmt"
	"strings"
)

// SliceMerge merge 多个 slice 为一个，并最终返回一个新的 slice
func SliceMerge[S ~[]T, T any](items ...S) S {
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
func SliceCopy[S ~[]T, T any](a S) S {
	if a == nil {
		return nil
	}
	cp := make([]T, 0, len(a))
	return append(cp, a...)
}

func SliceUnique[S ~[]T, T comparable](arr S) S {
	if len(arr) < 2 {
		return arr
	}
	c := make(map[T]struct{}, len(arr))
	ret := make([]T, 0, len(arr))
	for i := 0; i < len(arr); i++ {
		v := arr[i]
		if _, ok := c[v]; ok {
			continue
		}
		c[v] = struct{}{}
		ret = append(ret, v)
	}
	return ret
}

func SliceHas[S ~[]T, T comparable](arr S, values ...T) bool {
	if len(values) == 0 {
		return false
	}
	for i := 0; i < len(arr); i++ {
		for j := 0; j < len(values); j++ {
			if arr[i] == values[j] {
				return true
			}
		}
	}
	return false
}

func SliceIndex[S ~[]T, T comparable](arr S, values ...T) int {
	if len(values) == 0 {
		return -1
	}
	for i := 0; i < len(arr); i++ {
		for j := 0; j < len(values); j++ {
			if arr[i] == values[j] {
				return i
			}
		}
	}
	return -1
}

func SliceDelete[S ~[]T, T comparable](arr S, values ...T) S {
	if len(arr) == 0 || len(values) == 0 {
		return arr
	}
	index := SliceIndex(arr, values...)
	if index == -1 {
		return arr
	}
	cp := append([]T(nil), arr[:index]...)
	shadow := arr[index+1:]
	for {
		index = SliceIndex(shadow, values...)
		if index == -1 {
			cp = append(cp, shadow...)
			return cp
		}
		cp = append(cp, shadow[:index]...)
		shadow = shadow[index+1:]
	}
}

func SliceJoin[T any](arr []T, sep string) string {
	return SliceJoinFormat(arr, func(val T) string {
		return fmt.Sprint(val)
	}, sep)
}

func SliceJoinFormat[T any](arr []T, format func(val T) string, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	elems := make([]string, len(arr))
	for i := 0; i < len(arr); i++ {
		elems[i] = format(arr[i])
	}
	return strings.Join(elems, sep)
}

func SliceValuesAllow[T comparable](values []T, allow []T) error {
	if len(values) == 0 || len(allow) == 0 {
		return nil
	}
	allowMap := make(map[T]struct{}, len(allow))
	for i := 0; i < len(allow); i++ {
		allowMap[allow[i]] = struct{}{}
	}
	var notAllow []T
	for i := 0; i < len(values); i++ {
		val := values[i]
		if _, ok := allowMap[val]; !ok {
			notAllow = append(notAllow, val)
		}
	}
	if len(notAllow) == 0 {
		return nil
	}
	str := SliceJoin(notAllow, ",")
	str = SubStr(str, 0, 10)
	return fmt.Errorf("%d values (%s) are not allowed", len(notAllow), str)
}
