// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/6/6

package fstypes

import (
	"fmt"
	"strconv"
	"strings"
)

// String 字符串
type String string

func (s String) split(sep string) []string {
	if s == "" {
		return nil
	}
	ts := strings.TrimSpace(string(s))
	if ts == "" {
		return nil
	}
	return strings.Split(string(s), sep)
}

// Ints 转换为 []int
func (s String) Ints(sep string) ([]int, error) {
	vs := s.split(sep)
	if len(vs) == 0 {
		return nil, nil
	}
	result := make([]int, 0, len(vs))
	for i := 0; i < len(vs); i++ {
		v := strings.TrimSpace(vs[i])
		if v == "" {
			continue
		}
		vi, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("strconv.Atoi([%d]=%q) failed: %w", i, vs[i], err)
		}
		result = append(result, vi)
	}
	return result, nil
}

// Int32s 转换为 []int32
func (s String) Int32s(sep string) ([]int32, error) {
	vs := s.split(sep)
	if len(vs) == 0 {
		return nil, nil
	}
	result := make([]int32, 0, len(vs))
	for i := 0; i < len(vs); i++ {
		v := strings.TrimSpace(vs[i])
		if v == "" {
			continue
		}
		vi, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("strconv.ParseInt32([%d]=%q) failed: %w", i, vs[i], err)
		}
		result = append(result, int32(vi))
	}
	return result, nil
}

// Uint32s 转换为 []uint32
func (s String) Uint32s(sep string) ([]uint32, error) {
	vs := s.split(sep)
	if len(vs) == 0 {
		return nil, nil
	}
	result := make([]uint32, 0, len(vs))
	for i := 0; i < len(vs); i++ {
		v := strings.TrimSpace(vs[i])
		if v == "" {
			continue
		}
		vi, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("strconv.ParseUint([%d]=%q) failed: %w", i, vs[i], err)
		}
		result = append(result, uint32(vi))
	}
	return result, nil
}

// Int64s 转换为 []int64
func (s String) Int64s(sep string) ([]int64, error) {
	vs := s.split(sep)
	if len(vs) == 0 {
		return nil, nil
	}
	result := make([]int64, 0, len(vs))
	for i := 0; i < len(vs); i++ {
		v := strings.TrimSpace(vs[i])
		if v == "" {
			continue
		}
		vi, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("strconv.ParseInt([%d]=%q) failed: %w", i, vs[i], err)
		}
		result = append(result, vi)
	}
	return result, nil
}

// Uint64s 转换为 []uint64
func (s String) Uint64s(sep string) ([]uint64, error) {
	vs := s.split(sep)
	if len(vs) == 0 {
		return nil, nil
	}
	result := make([]uint64, 0, len(vs))
	for i := 0; i < len(vs); i++ {
		v := strings.TrimSpace(vs[i])
		if v == "" {
			continue
		}
		vi, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("strconv.ParseUint([%d]=%q) failed: %w", i, vs[i], err)
		}
		result = append(result, vi)
	}
	return result, nil
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
		if v == "" {
			continue
		}
		result = append(result, v)
	}
	return result
}

// StringSlice alias off []string
type StringSlice []string

// Unique uniq
func (ss StringSlice) Unique() StringSlice {
	if len(ss) < 2 {
		return ss
	}
	c := make(map[string]bool, len(ss))
	ret := make(StringSlice, 0, len(ss))
	for _, v := range ss {
		if c[v] {
			continue
		}
		c[v] = true
		ret = append(ret, v)
	}
	return ret
}
