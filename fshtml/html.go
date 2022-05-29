// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

// NewDiv 创建一个 div
func NewDiv() *Any {
	return &Any{
		Tag: "div",
	}
}

// NewP 创建一个 p
func NewP() *Any {
	return &Any{
		Tag: "p",
	}
}

// NewBody 创建一个 body
func NewBody() *Any {
	return &Any{
		Tag: "body",
	}
}

// NewUl 转换为 ul
func NewUl(values StringSlice) *Any {
	return &Any{
		Tag:  "ul",
		Body: values.ToElements("li", nil),
	}
}

// NewOl 转换为 ol
func NewOl(values StringSlice) *Any {
	return &Any{
		Tag:  "ol",
		Body: values.ToElements("li", nil),
	}
}

// NewIMG 创建一个  img  标签
func NewIMG(src string) *IMG {
	return (&IMG{}).SRC(src)
}

// IMG 图片 img 标签
type IMG struct {
	WithAttrs
}

func (m *IMG) set(key string, value string) *IMG {
	attr := &Attr{
		Key:    key,
		Values: []string{value},
	}
	m.MustAttrs().Set(attr)
	return m
}

// SRC 设置 src 属性
func (m *IMG) SRC(src string) *IMG {
	return m.set("src", src)
}

// ALT 设置 alt 属性
func (m *IMG) ALT(alt string) *IMG {
	return m.set("alt", alt)
}

// HTML 转换为 html
func (m *IMG) HTML() ([]byte, error) {
	bw := newBufWriter()
	bw.Write("<img")
	bw.WriteWithSep(" ", m.Attrs)
	bw.Write("/>")
	return bw.HTML()
}
