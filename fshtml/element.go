// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/2

package fshtml

import (
	"errors"
	"html"
)

// Element 所有 HTML 组件的基础定义
type Element interface {
	HTML() ([]byte, error)
}

// ToElements 转换为 Elements 类型
func ToElements(es ...Element) Elements {
	return es
}

// Elements alias of []Element
type Elements []Element

// HTML 实现 Element 接口
func (hs Elements) HTML() ([]byte, error) {
	if len(hs) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	for i := 0; i < len(hs); i++ {
		bw.Write(hs[i])
	}
	return bw.HTML()
}

// InsertFront  to the frontend
func (hs *Elements) InsertFront(values ...Element) {
	*hs = append(values, *hs...)
}

// Add append to the end
func (hs *Elements) Add(values ...Element) {
	*hs = append(*hs, values...)
}

// Set 设置内容
func (hs *Elements) Set(values ...Element) {
	*hs = values
}

// ErrEmptyTagName tag 值为空的错误
var ErrEmptyTagName = errors.New("empty tag name")

var _ Element = (*Any)(nil)
var _ AttrsMapper = (*Any)(nil)

// NewAny 创建任意的 tag
func NewAny(tag string) *Any {
	return &Any{
		Tag: tag,
	}
}

// Any 一块 HTML 内容
type Any struct {
	// Tag 标签名称，必填，如 div
	Tag string

	// WithAttrs 可选，属性信息
	WithAttrs

	// Body 内容，可选
	Body Elements

	// SelfClose 当前标签是否自关闭,默认为 false
	// 如 img 标签就是自关闭的：<img src="/a.jpg"/>
	SelfClose bool
}

// HTML 实现 Element 接口
func (c *Any) HTML() ([]byte, error) {
	if len(c.Tag) == 0 {
		return nil, ErrEmptyTagName
	}
	bw := newBufWriter()
	bw.Write("<", c.Tag)
	bw.WriteWithSep(" ", c.Attrs)
	if c.SelfClose {
		bw.Write("/>")
	} else {
		bw.Write(">")
		bw.Write(c.Body)
		bw.Write("</", c.Tag, ">")
	}
	return bw.HTML()
}

// Comment 注释
type Comment string

// HTML 转换为 HTML
func (c Comment) HTML() ([]byte, error) {
	if len(c) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	bw.Write("<!-- ", html.EscapeString(string(c)), " -->\n")
	return bw.HTML()
}
