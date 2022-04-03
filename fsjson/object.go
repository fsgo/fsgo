// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/3

package fsjson

import (
	"bytes"
	"encoding/json"
)

var _ json.Unmarshaler = (*Object)(nil)
var _ json.Marshaler = (*Object)(nil)

// NewStruct 创建一个新的 Object
func NewStruct(value any) *Object {
	return &Object{
		Value: value,
	}
}

// Object 兼容多种空值格式的 JSON Object，一般是用在解析 JSON 数据时。
//
// 正常情况下，JSON 的 object 的空值可以是：null 和 {} 这两种，
// 但是对于 PHP 程序，由于使用 PHP 的 array 类型而不是 PHP 的 class 类型，导致 PHP 程序数据的
// JSON 数据可能 {"User":[]} 这样，而 User 数据实际应该是一个 Object， 导致 Go 程序解析失败。
//
// 这个 Object 类型即为解决该问题而定义。
// 可以允许 JSON 值为 [],""
type Object struct {
	Value any
	bf    []byte
}

// MarshalJSON 实现了自定义的 json.Marshaler
// 返回的是 Value 的 json 内容
func (obj *Object) MarshalJSON() ([]byte, error) {
	if obj.Value == nil {
		return null, nil
	}
	return json.Marshal(obj.Value)
}

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (obj *Object) UnmarshalJSON(bf []byte) error {
	if bytes.Equal(bf, emptyString) || bytes.Equal(bf, emptyArray) {
		return nil
	}
	obj.bf = append([]byte(nil), bf...)
	if obj.Value != nil {
		return json.Unmarshal(bf, &obj.Value)
	}
	return nil
}

// UnmarshalTo 将数据解析到指定的类型上
//
// 若 JSON 内容为空，将直接返回 nil，不会修改 value 的值
func (obj *Object) UnmarshalTo(value any) error {
	if len(obj.bf) == 0 {
		return nil
	}
	return json.Unmarshal(obj.bf, value)
}
