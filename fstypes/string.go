// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/6/6

package fstypes

import (
	"fmt"
	"strings"

	"github.com/fsgo/fsgo/internal/number"
)

// String 字符串
type String string

// IntSlice 转换为 []int
func (s String) IntSlice(sep string) ([]int, error) {
	return toNumberSlice[int](s, sep, int(0))
}

// Int8Slice 转换为 []int8
func (s String) Int8Slice(sep string) ([]int8, error) {
	return toNumberSlice[int8](s, sep, int8(0))
}

// Int16Slice 转换为 []int16
func (s String) Int16Slice(sep string) ([]int16, error) {
	return toNumberSlice[int16](s, sep, int16(0))
}

// Int32Slice 转换为 []int32
func (s String) Int32Slice(sep string) ([]int32, error) {
	return toNumberSlice[int32](s, sep, int32(0))
}

// Int64Slice 转换为 []int32
func (s String) Int64Slice(sep string) ([]int64, error) {
	return toNumberSlice[int64](s, sep, int64(0))
}

// UintSlice 转换为 []uint
func (s String) UintSlice(sep string) ([]uint, error) {
	return toNumberSlice[uint](s, sep, uint(0))
}

// Uint8Slice 转换为 []uint8
func (s String) Uint8Slice(sep string) ([]uint8, error) {
	return toNumberSlice[uint8](s, sep, uint8(0))
}

// Uint16Slice 转换为 []uint16
func (s String) Uint16Slice(sep string) ([]uint16, error) {
	return toNumberSlice[uint16](s, sep, uint16(0))
}

// Uint32Slice 转换为 []uint32
func (s String) Uint32Slice(sep string) ([]uint32, error) {
	return toNumberSlice[uint32](s, sep, uint32(0))
}

// Uint64Slice 转换为 []uint64
func (s String) Uint64Slice(sep string) ([]uint64, error) {
	return toNumberSlice[uint64](s, sep, uint64(0))
}

// Split 转换为 []string
// 会剔除掉空字符串，如 `a, c,` -> []string{"a","c"}
func (s String) Split(sep string) []string {
	vs := s.split(sep)
	if len(vs) == 0 {
		return nil
	}
	result := make([]string, 0, len(vs))
	for i := 0; i < len(vs); i++ {
		v := strings.TrimSpace(vs[i])
		if len(v) == 0 {
			continue
		}
		result = append(result, v)
	}
	return result
}

func (s String) split(sep string) []string {
	if len(s) == 0 {
		return nil
	}
	ts := strings.TrimSpace(string(s))
	if len(ts) == 0 {
		return nil
	}
	return strings.Split(string(s), sep)
}

func toNumberSlice[T number.Number](s String, sep string, zero any) ([]T, error) {
	vs := s.split(sep)
	if len(vs) == 0 {
		return nil, nil
	}
	result := make([]T, 0, len(vs))
	for i := 0; i < len(vs); i++ {
		v := strings.TrimSpace(vs[i])
		if len(v) == 0 {
			continue
		}
		vi, err := number.ParseNumber[T](v, zero)
		if err != nil {
			return nil, fmt.Errorf("strconv.Atoi([%d]=%q) failed: %w", i, vs[i], err)
		}
		result = append(result, vi)
	}
	return result, nil
}

// StringSlice alias off []string
type StringSlice []string

// Unique uniq
func (ss StringSlice) Unique() StringSlice {
	return SliceUnique(ss)
}

// Has 是否包含指定的值
func (ss StringSlice) Has(value string) bool {
	return SliceHas(ss, value)
}

// Delete 删除对应的值
func (ss *StringSlice) Delete(values ...string) {
	if len(values) == 0 || len(*ss) == 0 {
		return
	}
	s1 := SliceDelete(*ss, values...)
	*ss = s1
}
