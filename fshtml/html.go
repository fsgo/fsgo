// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

// NewDiv 创建一个 div
func NewDiv() *Block {
	return &Block{
		Tag: "div",
	}
}

// NewP 创建一个 p
func NewP() *Block {
	return &Block{
		Tag: "p",
	}
}

// NewBody 创建一个 body
func NewBody() *Block {
	return &Block{
		Tag: "body",
	}
}

// NewUl 转换为 ul
func NewUl(values StringSlice) *Block {
	return &Block{
		Tag:  "ul",
		Body: values.Blocks("li", nil),
	}
}

// NewOl 转换为 ol
func NewOl(values StringSlice) *Block {
	return &Block{
		Tag:  "ol",
		Body: values.Blocks("li", nil),
	}
}
