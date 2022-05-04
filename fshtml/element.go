// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/2

package fshtml

func NewElement(name string) *Element {
	return &Element{
		TagName:         name,
		Attributes:      NewAttributes(),
		ChildReadWriter: &Children{},
	}
}

var _ Attributes = (*Element)(nil)
var _ ChildReadWriter = (*Element)(nil)

type Element struct {
	TagName string
	Attributes
	ChildReadWriter
}

func (e *Element) HTML() ([]byte, error) {
	bw := newBufWriter()
	bw.Write("<", e.TagName)
	bw.WriteWithSep(" ", e.Attributes)
	bw.Write(">")
	bw.Write(e.ChildReadWriter)
	bw.Write("</", e.TagName, ">")
	return bw.HTML()
}
