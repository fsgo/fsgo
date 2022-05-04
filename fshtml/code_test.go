// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringSlice_Codes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var a StringSlice
		got, err := a.Codes("li").HTML()
		require.NoError(t, err)
		want := ``
		require.Equal(t, want, string(got))
	})

	t.Run("1 value", func(t *testing.T) {
		a := StringSlice{"123"}
		got, err := a.Codes("li").HTML()
		require.NoError(t, err)
		want := "<li>123</li>\n"
		require.Equal(t, want, string(got))
	})

	t.Run("2 value", func(t *testing.T) {
		a := StringSlice{"123", "456"}
		got, err := a.Codes("li").HTML()
		require.NoError(t, err)
		want := "<li>123</li>\n<li>456</li>\n"
		require.Equal(t, want, string(got))
	})
}
