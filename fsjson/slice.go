// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/3/26

//go:generate go run cmd/slice_number_gen.go

package fsjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

var errNotSupport = errors.New("fsjson not support")

var _ json.Unmarshaler = (*StringSlice)(nil)

// StringSlice 扩展支持 JSON 的 []string 类型
// 其实际值可以是多种格式，比如:
// 	value: "a"
// 	value: "a,b"
// 	value: ["a","b",123]
// 	value: null
// 	value: 123
type StringSlice []string

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ss *StringSlice) UnmarshalJSON(bf []byte) error {
	if bytes.Equal(bf, emptyString) || bytes.Equal(bf, null) {
		return nil
	}

	head := bf[0]
	tail := bf[len(bf)-1]

	if head == '"' && tail == '"' {
		*ss = strings.Split(string(bf[1:len(bf)-1]), ",")
		return nil
	}

	if head == '[' && tail == ']' {
		list := strings.Split(string(bf[1:len(bf)-1]), ",")
		for i := 0; i < len(list); i++ {
			list[i] = strings.Trim(list[i], `"`)
		}
		*ss = list
		return nil
	}

	// 其他情况，比如：
	// {"Alias":123}
	*ss = append(*ss, string(bf))
	return nil
}

// Slice 返回 []string 的值
func (ss StringSlice) Slice() []string {
	return ss
}
