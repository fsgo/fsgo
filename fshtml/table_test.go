// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/29

package fshtml_test

import (
	"testing"

	"github.com/fsgo/fst"

	"github.com/fsgo/fsgo/fshtml"
)

func TestTable1(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		tb := &fshtml.Table1{}
		tb.SetHeader(fshtml.NewTh(fshtml.String("name")), fshtml.NewTh(fshtml.String("age")))
		tb.AddRow(fshtml.NewTd(fshtml.String("lilei")), fshtml.NewTd(fshtml.String("18")))
		tb.AddRow(fshtml.NewTd(fshtml.String("hanmeimei")), fshtml.NewTd(fshtml.String("15")))
		tb.SetFooter(fshtml.NewTd(fshtml.String("f1")), fshtml.NewTd(fshtml.String("f2")))

		fshtml.SetID(tb, "#abc")

		got, err := tb.HTML()
		fst.NoError(t, err)
		want := `<table id="#abc">` + "\n" +
			"<thead>\n<tr><th>name</th><th>age</th></tr>\n</thead>\n" +
			"<tbody>\n" +
			"<tr><td>lilei</td><td>18</td></tr>\n" +
			"<tr><td>hanmeimei</td><td>15</td></tr>\n" +
			"</tbody>\n" +
			"<tfoot>\n<tr><td>f1</td><td>f2</td></tr>\n</tfoot>\n" +
			"</table>\n"
		fst.Equal(t, want, string(got))
	})
}
