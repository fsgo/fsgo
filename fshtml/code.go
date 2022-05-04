// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/2

package fshtml

import (
	"errors"
)

// Code 所有 HTML 组件的基础定义
type Code interface {
	HTML() ([]byte, error)
}

// Bytes 将 []byte 转换为 Code类型
type Bytes []byte

// HTML 实现 Code 接口
func (b Bytes) HTML() ([]byte, error) {
	return b, nil
}

// String 将 String 转换为 Code类型
type String string

// HTML 实现 Code 接口
func (s String) HTML() ([]byte, error) {
	return []byte(s), nil
}

// StringSlice 将 []string 转换为 Code类型
type StringSlice []string

// Blocks 转换为 字段 tag 的 []Block
func (ss StringSlice) Blocks(tag string) Blocks {
	if len(ss) == 0 {
		return nil
	}
	cs := make([]*Block, len(ss))
	for i := 0; i < len(ss); i++ {
		cs[i] = &Block{
			Tag:  tag,
			Body: String(ss[i]),
		}
	}
	return cs
}

// Codes 转换为 字段 tag 的 []Code
func (ss StringSlice) Codes(tag string) Codes {
	if len(ss) == 0 {
		return nil
	}
	cs := make([]Code, len(ss))
	for i := 0; i < len(ss); i++ {
		cs[i] = &Block{
			Tag:  tag,
			Body: String(ss[i]),
		}
	}
	return cs
}

// Codes alias of []Code
type Codes []Code

// HTML 实现 Code 接口
func (hs Codes) HTML() ([]byte, error) {
	bw := newBufWriter()
	for i := 0; i < len(hs); i++ {
		bw.Write(hs[i], "\n")
	}
	return bw.HTML()
}

// Insert insert to the frontend
func (hs *Codes) Insert(values ...Code) {
	*hs = append(values, *hs...)
}

// Add append to the end
func (hs *Codes) Add(values ...Code) {
	*hs = append(*hs, values...)
}

// ErrEmptyTagName tag 值为空的错误
var ErrEmptyTagName = errors.New("empty tag name")

var _ Code = (*Block)(nil)

// Block 一块 HTML 内容
type Block struct {
	Tag  string
	Attr Attributes
	Body Code
}

// MustAttr 获取属性集合，会自动车初始化
func (c *Block) MustAttr() Attributes {
	if c.Attr == nil {
		c.Attr = NewAttributes()
	}
	return c.Attr
}

// HTML 实现 Code 接口
func (c *Block) HTML() ([]byte, error) {
	if len(c.Tag) == 0 {
		return nil, ErrEmptyTagName
	}
	bw := newBufWriter()
	bw.Write("<", c.Tag)
	bw.WriteWithSep(" ", c.Attr)
	bw.Write(">")
	bw.Write(c.Body)
	bw.Write("</", c.Tag, ">")
	return bw.HTML()
}

var _ Code = (Blocks)(nil)

// Blocks alias of []*Block
type Blocks []*Block

// HTML 实现 Code 接口
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

// Codes 转换为 []Code
func (bs Blocks) Codes() Codes {
	if len(bs) == 0 {
		return nil
	}
	cs := make(Codes, len(bs))
	for i := 0; i < len(bs); i++ {
		cs[i] = bs
	}
	return cs
}
