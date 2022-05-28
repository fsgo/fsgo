// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/2

package fshtml

import (
	"errors"
)

// Element 所有 HTML 组件的基础定义
type Element interface {
	HTML() ([]byte, error)
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
		bw.Write(hs[i], "\n")
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

// ErrEmptyTagName tag 值为空的错误
var ErrEmptyTagName = errors.New("empty tag name")

var _ Element = (*Block)(nil)

// Block 一块 HTML 内容
type Block struct {
	Tag   string
	Attrs *Attributes
	Body  Element
}

// MustAttrs 返回 Attributes,若为 nil，则初始化一个并返回
func (c *Block) MustAttrs() *Attributes {
	if c.Attrs == nil {
		c.Attrs = &Attributes{}
	}
	return c.Attrs
}

// HTML 实现 Element 接口
func (c *Block) HTML() ([]byte, error) {
	if len(c.Tag) == 0 {
		return nil, ErrEmptyTagName
	}
	bw := newBufWriter()
	bw.Write("<", c.Tag)
	bw.WriteWithSep(" ", c.Attrs)
	bw.Write(">")
	bw.Write(c.Body)
	bw.Write("</", c.Tag, ">")
	return bw.HTML()
}

var _ Element = (Blocks)(nil)

// Blocks alias of []*Block
type Blocks []*Block

// HTML 实现 Element 接口
func (bs Blocks) HTML() ([]byte, error) {
	if len(bs) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	for i := 0; i < len(bs); i++ {
		bw.Write(bs[i], "\n")
	}
	return bw.HTML()
}

// Elements 转换为 []Element
func (bs Blocks) Elements() Elements {
	if len(bs) == 0 {
		return nil
	}
	cs := make(Elements, len(bs))
	for i := 0; i < len(bs); i++ {
		cs[i] = bs
	}
	return cs
}
