// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fsgo/fsgo/fshtml"
)

func TestStringSlice_Codes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var a fshtml.StringSlice
		got, err := a.ToElements("li", nil).HTML()
		require.NoError(t, err)
		want := ``
		require.Equal(t, want, string(got))
	})

	t.Run("1 value", func(t *testing.T) {
		a := fshtml.StringSlice{"123"}
		got, err := a.ToElements("li", func(b *fshtml.Any) {
			fshtml.SetClass(b, "red")
		}).HTML()
		require.NoError(t, err)
		want := `<li class="red">123</li>`
		require.Equal(t, want, string(got))
	})

	t.Run("2 value", func(t *testing.T) {
		a := fshtml.StringSlice{"123", "456"}
		got, err := a.ToElements("li", nil).HTML()
		require.NoError(t, err)
		want := "<li>123</li><li>456</li>"
		require.Equal(t, want, string(got))
	})
}
