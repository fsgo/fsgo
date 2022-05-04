// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

// Table1 一个简单的表格
type Table1 struct {
	attr Attributes
	head []Code
	rows [][]Code
	foot []Code
}

// Attr 表格的属性集合
func (t *Table1) Attr() Attributes {
	if t.attr == nil {
		t.attr = NewAttributes()
	}
	return t.attr
}

// SetHeader 设置表头
func (t *Table1) SetHeader(cells ...Code) {
	t.head = cells
}

// AddRow 添加一行内容
func (t *Table1) AddRow(cells ...Code) {
	t.rows = append(t.rows, cells)
}

// AddRows 添加多行内容
func (t *Table1) AddRows(rows ...[]Code) {
	t.rows = append(t.rows, rows...)
}

// SetFooter 设置表格的页脚
func (t *Table1) SetFooter(cells ...Code) {
	t.foot = cells
}

// HTML 实现 Code 接口
func (t *Table1) HTML() ([]byte, error) {
	bw := newBufWriter()
	bw.Write("<table")
	bw.WriteWithSep(" ", t.attr)
	bw.Write(">\n")
	bw.Write("<thead>\n<tr>")
	for i := 0; i < len(t.head); i++ {
		bw.Write(t.head[i])
	}
	bw.Write("</tr>\n</thead>\n<tbody>\n")
	for i := 0; i < len(t.rows); i++ {
		row := t.rows[i]
		bw.Write("<tr>")
		for j := 0; j < len(row); j++ {
			bw.Write(row[j])
		}
		bw.Write("</tr>\n")
	}
	bw.Write("</tbody>\n")
	if len(t.foot) > 0 {
		bw.Write("<tfoot>\n<tr>")
		for i := 0; i < len(t.foot); i++ {
			bw.Write(t.foot[i])
		}
		bw.Write("</tr>\n</tfoot>\n")
	}
	return bw.HTML()
}

// NewTd 创建一个新的 td
func NewTd(val Code) *Block {
	return &Block{
		Tag:  "td",
		Body: val,
	}
}

// NewTh 创建一个新的 th
func NewTh(val Code) *Block {
	return &Block{
		Tag:  "th",
		Body: val,
	}
}
