// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/3

package main

import (
	"go/format"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(tplHead))
	builder.WriteString("\n")

	lines := make([]string, 0, len(types))
	for _, item := range types {
		lines = append(lines, render(item))
	}
	body := strings.Join(lines, "\n// "+strings.Repeat("-", 80)+"\n\n")
	builder.WriteString(body)

	out, err := format.Source([]byte(builder.String()))
	if err != nil {
		log.Fatalln(err)
	}
	ioutil.WriteFile("slice_number.go", out, 0644)
}

var types = []*tplData{
	// signed int
	{
		SliceName:         "IntSlice",
		NumberType:        "int",
		TypeCommentValues: tplCommentSigned,
	},
	{
		SliceName:         "Int8Slice",
		NumberType:        "int8",
		TypeCommentValues: tplCommentSigned,
	},
	{
		SliceName:         "Int16Slice",
		NumberType:        "int16",
		TypeCommentValues: tplCommentSigned,
	},
	{
		SliceName:         "Int32Slice",
		NumberType:        "int32",
		TypeCommentValues: tplCommentSigned,
	},
	{
		SliceName:         "Int64Slice",
		NumberType:        "int64",
		TypeCommentValues: tplCommentSigned,
	},
	// unsigned int
	{
		SliceName:         "UintSlice",
		NumberType:        "uint",
		TypeCommentValues: tplCommentUnsigned,
	},
	{
		SliceName:         "Uint8Slice",
		NumberType:        "uint8",
		TypeCommentValues: tplCommentUnsigned,
	},
	{
		SliceName:         "Uint16Slice",
		NumberType:        "uint16",
		TypeCommentValues: tplCommentUnsigned,
	},
	{
		SliceName:         "Uint32Slice",
		NumberType:        "uint32",
		TypeCommentValues: tplCommentUnsigned,
	},
	{
		SliceName:         "Uint64Slice",
		NumberType:        "uint64",
		TypeCommentValues: tplCommentUnsigned,
	},
	// float number
	{
		SliceName:         "Float32Slice",
		NumberType:        "float32",
		TypeCommentValues: tplCommentFloat,
	},
	{
		SliceName:         "Float64Slice",
		NumberType:        "float64",
		TypeCommentValues: tplCommentFloat,
	},
}

func render(td *tplData) string {
	txt := strings.ReplaceAll(tplType, "{SliceName}", td.SliceName)
	txt = strings.ReplaceAll(txt, "{numberType}", td.NumberType)
	txt = strings.ReplaceAll(txt, "{TypeCommentValues}", strings.TrimSpace(td.TypeCommentValues))
	return strings.TrimSpace(txt)
}

type tplData struct {
	SliceName         string
	NumberType        string
	TypeCommentValues string
}

var tplHead = `
// Code generated by cmd/slice_number_go.go. DO NOT EDIT.

package fsjson

import (
	"encoding/json"
)

`

var tplType = `
var _ json.Unmarshaler = (*{SliceName})(nil)

// {SliceName} 扩展支持 JSON 的 []{numberType} 类型
// 其实际值可以是多种格式，比如:
{TypeCommentValues}
type {SliceName} []{numberType}

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *{SliceName}) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[{numberType}](bf, {numberType}(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []{numberType} 的值
func (ns {SliceName}) Slice() []{numberType} {
	return ns
}
`

var tplCommentUnsigned = `
// 	value: ""
// 	value: "123,456,-1"
// 	value: [123,"456",1,-1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
`
var tplCommentSigned = `
// 	value: ""
// 	value: "123,456"
// 	value: [123,"456",1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
`

var tplCommentFloat = `
// 	value: ""
// 	value: "123,456,-1,1.2"
// 	value: [123,"456",1,"2.1",2.3]
// 	value: null
// 	value: 123
// 	value: 123.1
`
