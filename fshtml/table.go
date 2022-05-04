// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

type TD string

func (t TD) HTML() ([]byte, error) {
	bw := newBufWriter()
	bw.Write("<td>")
	bw.Write(string(t))
	bw.Write("</td>")
	return bw.HTML()
}

type TH string

func (t TH) HTML() ([]byte, error) {
	bw := newBufWriter()
	bw.Write("<th>")
	bw.Write(string(t))
	bw.Write("</th>")
	return bw.HTML()
}

type Table1 struct {
	attr Attributes
	head []string
	rows [][]string
	foot []string
}

func (t *Table1) Attr() Attributes {
	if t.attr == nil {
		t.attr = NewAttributes()
	}
	return t.attr
}

func (t *Table1) SetHeader(values ...string) {
	t.head = values
}

func (t *Table1) AddRow(row ...string) {
	t.rows = append(t.rows, row)
}

func (t *Table1) AddRows(rows ...[]string) {
	t.rows = append(t.rows, rows...)
}

func (t *Table1) SetFooter(values ...string) {
	t.foot = values
}

func (t *Table1) HTML() ([]byte, error) {
	bw := newBufWriter()
	bw.Write("<table")
	bw.WriteWithSep(" ", t.attr)
	bw.Write(">\n", "<thead>\n<tr>")
	for i := 0; i < len(t.head); i++ {
		bw.Write(TH(t.head[i]))
	}
	bw.Write("</tr>\n</thead>\n<tbody>\n")
	for i := 0; i < len(t.rows); i++ {
		row := t.rows[i]
		bw.Write("<tr>")
		for j := 0; j < len(row); j++ {
			bw.Write(TD(row[j]))
		}
		bw.Write("</tr>\n")
	}
	bw.Write("</tbody>\n")
	if len(t.foot) > 0 {
		bw.Write("<tfoot>\n<tr>")
		for i := 0; i < len(t.foot); i++ {
			bw.Write(TD(t.foot[i]))
		}
		bw.Write("</tr>\n</tfoot>\n")
	}
	return bw.HTML()
}
