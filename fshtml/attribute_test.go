// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fsgo/fsgo/fshtml"
)

func TestAttributes(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		attr := &fshtml.WithAttrs{}
		fshtml.SetID(attr, "#abc")
		fshtml.SetName(attr, "hello")
		fshtml.DeleteClass(attr, "c0")

		fshtml.SetClass(attr, "c1", "c2")
		fshtml.SetClass(attr, "c3", "c4")
		fshtml.AddClass(attr, "c5")
		fshtml.DeleteClass(attr, "c4", "c6")

		attrs := attr.FindAttrs()
		bf, err := attrs.HTML()
		require.NoError(t, err)
		want := `id="#abc" name="hello" class="c3 c5"`
		require.Equal(t, want, string(bf))

		wantKeys := []string{"id", "name", "class"}
		require.Equal(t, wantKeys, attrs.Keys())
	})
}
