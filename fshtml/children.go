// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

type (
	ChildReadWriter interface {
		ChildReader
		ChildWriter
	}

	ChildReader interface {
		Children() []HTML
	}

	ChildWriter interface {
		SetChild(b ...HTML)
		AddChild(b ...HTML)
	}
)

var _ ChildReadWriter = (*Children)(nil)
var _ HTML = (*Children)(nil)

type Children struct {
	elements []HTML
}

func (c *Children) Children() []HTML {
	return c.elements
}

func (c *Children) SetChild(b ...HTML) {
	c.elements = b
}

func (c *Children) AddChild(b ...HTML) {
	c.elements = append(c.elements, b...)
}

func (c *Children) HTML() ([]byte, error) {
	if len(c.elements) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	for i := 0; i < len(c.elements); i++ {
		bw.Write(c.elements[i])
	}
	return bw.HTML()
}
