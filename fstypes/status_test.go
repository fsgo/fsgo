// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package fstypes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnableStatus(t *testing.T) {
	var s EnableStatus
	require.True(t, s.IsEnable())
	require.Equal(t, "enable", s.String())

	s.Disable()
	require.False(t, s.IsEnable())
	require.Equal(t, "disable", s.String())

	s.Enable()
	require.True(t, s.IsEnable())
}

func TestGroupEnableStatus(t *testing.T) {
	t.Run("no default", func(t *testing.T) {
		var g GroupEnableStatus
		require.False(t, g.IsEnable("test"))
		g.SetEnable("test", false)

		require.False(t, g.IsEnable("test"))
		g.SetEnable("test", true)
		require.True(t, g.IsEnable("test"))
		got := g.String()
		want := "test:true,*other*:false"
		require.Equal(t, want, got)

		g.SetAllEnable(true)
		require.True(t, g.IsEnable("test"))
		require.True(t, g.IsEnable("other_key"))
		got = g.String()
		want = "test:true,*other*:true"
		require.Equal(t, want, got)
	})

	t.Run("default true", func(t *testing.T) {
		g := &GroupEnableStatus{
			Default: true,
		}
		require.True(t, g.IsEnable("test"))

		g.SetEnable("test", false)
		require.False(t, g.IsEnable("test"))

		g.SetEnable("test", true)
		require.True(t, g.IsEnable("test"))

		got := g.String()
		want := "test:true,*other*:true"
		require.Equal(t, want, got)

		g.SetAllEnable(false)
		require.False(t, g.IsEnable("test"))
		require.False(t, g.IsEnable("other_key"))
		got = g.String()
		want = "test:false,*other*:false"
		require.Equal(t, want, got)
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
		require.Equal(t, want, got)
	})
}
