// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fsflag

import (
	"fmt"
	"strings"

	"github.com/fsgo/fsgo/fstypes"
)

type Any[T Types] struct {
	// Allow 允许出现的值,可选
	// 当为空时不会校验，若有值，
	Allow []T

	// Check 检查值是否正常，可选，在 Parser 的时候会自动调用
	Check func(value T) error

	// value 解析获取到的值
	value T
}

func (a *Any[T]) String() string {
	return fmt.Sprint(a.value)
}

func (a *Any[T]) Set(s string) error {
	var zero T
	a.value = zero
	parser := getParserFunc(zero)
	s = strings.TrimSpace(s)
	v, err := parser(s)
	if err != nil {
		return err
	}
	val := v.(T)
	err1 := fstypes.SliceValuesAllow([]T{val}, a.Allow)
	if err1 != nil {
		return err1
	}
	if a.Check != nil {
		if err2 := a.Check(val); err2 != nil {
			return err2
		}
	}
	a.value = val
	return nil
}

func (a *Any[T]) Value() T {
	return a.value
}
