// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttribute_HTML(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		attr := NewAttributes()
		SetID(attr, "#abc")
		SetName(attr, "hello")
		DeleteClass(attr, "c0")

		attr.MustAttr("data").Set("a")
		attr.DeleteAttr("data")

		SetClass(attr, "c1", "c2")
		SetClass(attr, "c3", "c4")
		AddClass(attr, "c5")
		DeleteClass(attr, "c4", "c6")

		bf, err := attr.HTML()
		require.NoError(t, err)
		want := `id="#abc" name="hello" class="c3 c5"`
		require.Equal(t, want, string(bf))

		wantKeys := []string{"id", "name", "class"}
		require.Equal(t, wantKeys, attr.AttrKeys())
	})
}
