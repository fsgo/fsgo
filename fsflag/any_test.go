// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fsflag

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAny(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	var a1 Any[string]
	fs.Var(&a1, "a1", "")

	a2 := Any[string]{
		Allow: []string{"hello"},
	}
	fs.Var(&a2, "a2", "")

	require.NoError(t, fs.Parse([]string{"-a1", "123"}))
	require.Equal(t, "123", a1.Value())

	require.NoError(t, fs.Parse([]string{"-a2", "hello"}))
	require.Equal(t, "hello", a2.Value())

	require.Error(t, fs.Parse([]string{"-a2", "abc"}))
	require.Equal(t, "", a2.Value())
}
