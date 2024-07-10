// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fsflag

import (
	"flag"
	"testing"

	"github.com/fsgo/fst"
)

func TestAny(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	var a1 Any[string]
	fs.Var(&a1, "a1", "")

	a2 := Any[string]{
		Allow: []string{"hello"},
	}
	fs.Var(&a2, "a2", "")

	fst.NoError(t, fs.Parse([]string{"-a1", "123"}))
	fst.Equal(t, "123", a1.Value())

	fst.NoError(t, fs.Parse([]string{"-a2", "hello"}))
	fst.Equal(t, "hello", a2.Value())

	fst.Error(t, fs.Parse([]string{"-a2", "abc"}))
	fst.Equal(t, "", a2.Value())
}
