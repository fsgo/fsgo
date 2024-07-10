// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package fstypes

import (
	"testing"

	"github.com/fsgo/fst"
)

func TestEnableStatus(t *testing.T) {
	var s EnableStatus
	fst.True(t, s.IsEnable())
	fst.Equal(t, "enable", s.String())

	s.Disable()
	fst.False(t, s.IsEnable())
	fst.Equal(t, "disable", s.String())

	s.Enable()
	fst.True(t, s.IsEnable())
}

func TestGroupEnableStatus(t *testing.T) {
	t.Run("no default", func(t *testing.T) {
		var g GroupEnableStatus
		fst.False(t, g.IsEnable("test"))
		g.SetEnable("test", false)

		fst.False(t, g.IsEnable("test"))
		g.SetEnable("test", true)
		fst.True(t, g.IsEnable("test"))
		got := g.String()
		want := "test:true,*other*:false"
		fst.Equal(t, want, got)

		g.SetAllEnable(true)
		fst.True(t, g.IsEnable("test"))
		fst.True(t, g.IsEnable("other_key"))
		got = g.String()
		want = "test:true,*other*:true"
		fst.Equal(t, want, got)
	})

	t.Run("default true", func(t *testing.T) {
		g := &GroupEnableStatus{
			Default: true,
		}
		fst.True(t, g.IsEnable("test"))

		g.SetEnable("test", false)
		fst.False(t, g.IsEnable("test"))

		g.SetEnable("test", true)
		fst.True(t, g.IsEnable("test"))

		got := g.String()
		want := "test:true,*other*:true"
		fst.Equal(t, want, got)

		g.SetAllEnable(false)
		fst.False(t, g.IsEnable("test"))
		fst.False(t, g.IsEnable("other_key"))
		got = g.String()
		want = "test:false,*other*:false"
		fst.Equal(t, want, got)
	})

	t.Run("range", func(t *testing.T) {
		g := &GroupEnableStatus{
			Default: true,
		}
		g.SetEnable("a", true)
		g.SetEnable("b", false)
		got := map[string]bool{}
		g.Range(func(key any, enable bool) bool {
			got[key.(string)] = enable
			return true
		})
		want := map[string]bool{
			"a": true,
			"b": false,
		}
		fst.Equal(t, want, got)
	})
}
