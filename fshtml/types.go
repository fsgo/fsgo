// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/28

package fshtml

// Bytes 将 []byte 转换为 Code类型
type Bytes []byte

// HTML 实现 Element 接口
func (b Bytes) HTML() ([]byte, error) {
	return b, nil
}

// String 将 String 转换为 Code类型
type String string

// HTML 实现 Element 接口
func (s String) HTML() ([]byte, error) {
	return []byte(s), nil
}

// StringSlice 将 []string 转换为 Code类型
type StringSlice []string

// ToElements 转换为 字段 tag 的 []Element
func (ss StringSlice) ToElements(tag string, fn func(b *Any)) Elements {
	if len(ss) == 0 {
		return nil
	}
	cs := make([]Element, len(ss))
	for i := 0; i < len(ss); i++ {
		b := &Any{
			Tag:  tag,
			Body: ToElements(String(ss[i])),
		}
		if fn != nil {
			fn(b)
		}
		cs[i] = b
	}
	return cs
}
