// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fsflag

import (
	"fmt"
	"strings"

	"github.com/fsgo/fsgo/fstypes"
)

type Slice[T Types] struct {
	// Sep 多个值之间的连接符，默认为英文逗号，可选
	Sep string

	// values 解析获取到的值
	values []T

	// Allow 允许出现的值,可选
	// 当为空时不会校验，若有值，
	Allow []T

	// Check 检查值是否正常，可选，在 Parser 的时候会自动调用
	Check func(values []T) error
}

func (ss *Slice[T]) GetSep() string {
	if ss.Sep == "" {
		return ","
	}
	return ss.Sep
}

func (ss *Slice[T]) String() string {
	list := make([]string, 0, len(ss.values))
	for _, s := range ss.values {
		list = append(list, fmt.Sprintf("%v", s))
	}
	return strings.Join(list, ss.GetSep())
}

func (ss *Slice[T]) Set(s2 string) error {
	var zero T
	ss.values = nil
	parser := getParserFunc(zero)
	arr := strings.Split(s2, ss.GetSep())
	values := make([]T, 0, len(arr))
	for _, s := range arr {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		v, err := parser(s)
		if err != nil {
			return err
		}
		values = append(values, v.(T))
	}
	err1 := fstypes.SliceValuesAllow(values, ss.Allow)
	if err1 != nil {
		return err1
	}
	if ss.Check != nil {
		if err2 := ss.Check(values); err2 != nil {
			return err2
		}
	}
	ss.values = values
	return nil
}

func (ss *Slice[T]) Value() []T {
	return ss.values
}
