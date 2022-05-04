// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestElement_HTML(t *testing.T) {
	t.Run("div_empty", func(t *testing.T) {
		div := NewElement("div")
		got, err := div.HTML()
		require.NoError(t, err)
		want := `<div></div>`
		require.Equal(t, want, string(got))
	})

	t.Run("div_attrs", func(t *testing.T) {
		div := NewElement("div")
		div.MustAttr("id").Set("#123")
		div.MustAttr("name").Set("hello")
		div.AddChild(String("<p></p>"))
		got, err := div.HTML()
		require.NoError(t, err)
		want := `<div id="#123" name="hello"><p></p></div>`
		require.Equal(t, want, string(got))
	})
}
