// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/02/11

package fssync

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapLoad(t *testing.T) {
	var m1 Map[string, string]
	key1 := "hello"
	v1, ok1 := m1.Load(key1)
	require.Equal(t, "", v1)
	require.False(t, ok1)

	m1.Store(key1, "world")

	v1, ok1 = m1.Load(key1)
	require.Equal(t, "world", v1)
	require.True(t, ok1)

	m1.Delete(key1)

	v1, ok1 = m1.Load(key1)
	require.Equal(t, "", v1)
	require.False(t, ok1)

	v2, ok2 := m1.LoadOrStore(key1, "h1")
	require.Equal(t, "h1", v2)
	require.False(t, ok2)

	v2, ok2 = m1.LoadOrStore(key1, "h2")
	require.Equal(t, "h1", v2)
	require.True(t, ok2)

	var num1 int
	m1.Range(func(key string, value string) bool {
		num1++
		require.Equal(t, key1, key)
		require.Equal(t, "h1", value)
		return true
	})

	require.Equal(t, 1, num1)

	v3, ok3 := m1.LoadAndDelete(key1)
	require.Equal(t, "h1", v3)
	require.True(t, ok3)

	v3, ok3 = m1.LoadAndDelete(key1)
	require.Equal(t, "", v3)
	require.False(t, ok3)
}
