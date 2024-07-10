// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml_test

import (
	"testing"

	"github.com/fsgo/fst"

	"github.com/fsgo/fsgo/fshtml"
)

func TestStringSlice_Codes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var a fshtml.StringSlice
		got, err := a.ToElements("li", nil).HTML()
		fst.NoError(t, err)
		want := ``
		fst.Equal(t, want, string(got))
	})

	t.Run("1 value", func(t *testing.T) {
		a := fshtml.StringSlice{"123"}
		got, err := a.ToElements("li", func(b *fshtml.Any) {
			fshtml.SetClass(b, "red")
		}).HTML()
		fst.NoError(t, err)
		want := `<li class="red">123</li>`
		fst.Equal(t, want, string(got))
	})

	t.Run("2 value", func(t *testing.T) {
		a := fshtml.StringSlice{"123", "456"}
		got, err := a.ToElements("li", nil).HTML()
		fst.NoError(t, err)
		want := "<li>123</li><li>456</li>"
		fst.Equal(t, want, string(got))
	})
}

func TestStringSlice_HTML(t *testing.T) {
	ss := fshtml.StringSlice{"hello", "world"}
	b, err := ss.HTML()
	fst.NoError(t, err)
	want := "<ul><li>hello</li><li>world</li></ul>"
	fst.Equal(t, want, string(b))
}
