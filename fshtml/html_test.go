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
		fshtml.SetID(div.MustAttrs(), "#abc")
		div.Body = fshtml.NewP()
		got, err := div.HTML()
		require.NoError(t, err)
		want := `<div id="#abc"><p></p></div>`
		require.Equal(t, want, string(got))
	})

	t.Run("div_attrs", func(t *testing.T) {
		div := fshtml.NewDiv()
		fshtml.SetClass(div.MustAttrs(), "c1", "c2")
		fshtml.SetID(div.MustAttrs(), "#abc")
		got, err := div.HTML()
		require.NoError(t, err)
		want := `<div class="c1 c2" id="#abc"></div>`
		require.Equal(t, want, string(got))
	})
}
