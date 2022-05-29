// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fsgo/fsgo/fshtml"
)

func TestNewDiv(t *testing.T) {
	t.Run("div_empty", func(t *testing.T) {
		div := fshtml.NewDiv()
		got, err := div.HTML()
		require.NoError(t, err)
		want := `<div></div>`
		require.Equal(t, want, string(got))
	})

	t.Run("div_p", func(t *testing.T) {
		div := fshtml.NewDiv()
		fshtml.SetID(div, "#abc")
		div.Body = fshtml.ToElements(fshtml.NewP())
		got, err := div.HTML()
		require.NoError(t, err)
		want := `<div id="#abc"><p></p></div>`
		require.Equal(t, want, string(got))
	})

	t.Run("div_attrs", func(t *testing.T) {
		div := fshtml.NewDiv()
		fshtml.SetClass(div, "c1", "c2")
		fshtml.SetID(div, "#abc")
		got, err := div.HTML()
		require.NoError(t, err)
		want := `<div class="c1 c2" id="#abc"></div>`
		require.Equal(t, want, string(got))
	})
}

func TestBody(t *testing.T) {
	t.Run("with children", func(t *testing.T) {
		body := fshtml.NewBody()
		sa := &fshtml.StyleAttr{}
		sa.MaxWidth("100px").Height("200px")
		require.NoError(t, sa.SetTo(body))
		div := fshtml.NewDiv()
		div.Body.Add(fshtml.String("hello"))
		body.Body.Set(div)
		got, err := body.HTML()
		require.NoError(t, err)
		want := `<body style="max-width:100px; height:200px"><div>hello</div></body>`
		require.Equal(t, want, string(got))
	})
}

func TestIMG(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := fshtml.NewIMG("/a.jpg")
		a.ALT("hello")
		got, err := a.HTML()
		require.NoError(t, err)
		want := `<img src="/a.jpg" alt="hello"/>`
		require.Equal(t, want, string(got))
	})

	t.Run("width_height", func(t *testing.T) {
		a := fshtml.NewIMG("/a.jpg")
		fshtml.SetWidth(a, "100px")
		fshtml.SetHeight(a, "110px")
		got, err := a.HTML()
		require.NoError(t, err)
		want := `<img src="/a.jpg" width="100px" height="110px"/>`
		require.Equal(t, want, string(got))
	})
}

func TestA(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := fshtml.NewA("/gogo")
		a.Title("hello")
		got, err := a.HTML()
		require.NoError(t, err)
		want := `<a href="/gogo" title="hello"/>`
		require.Equal(t, want, string(got))
	})
}

func TestMeta(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := fshtml.NewMeta()
		a.Name("robots").Content("all")
		got, err := a.HTML()
		require.NoError(t, err)
		want := `<meta name="robots" content="all"/>`
		require.Equal(t, want, string(got))
	})
}

func TestLink(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := fshtml.NewLink()
		a.Rel("stylesheet").Type("text/css").Href("/a.css")
		got, err := a.HTML()
		require.NoError(t, err)
		want := `<link rel="stylesheet" type="text/css" href="/a.css"/>`
		require.Equal(t, want, string(got))
	})
}

func TestScript(t *testing.T) {
	t.Run("async", func(t *testing.T) {
		a := fshtml.NewScript()
		fshtml.SetAsync(a)
		got, err := a.HTML()
		require.NoError(t, err)
		want := `<script async></script>`
		require.Equal(t, want, string(got))
	})
}

func TestInput(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		a := fshtml.NewInput("text")
		fshtml.SetValue(a, "hello")
		fshtml.SetOnChange(a, `alter("ok")`)
		got, err := a.HTML()
		require.NoError(t, err)
		want := `<input type="text" value="hello" onchange="alter(\"ok\")"/>`
		require.Equal(t, want, string(got))
	})
}
