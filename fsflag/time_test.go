// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fsflag

import (
	"flag"
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestDuration(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var d1 Duration
	fs.Var(&d1, "d1", "")

	fst.NoError(t, fs.Parse([]string{"-d1", "1h"}))
	fst.Equal(t, time.Hour, d1.Duration())
	fst.Equal(t, time.Hour.String(), d1.String())

	fst.NoError(t, fs.Parse([]string{"-d1", "2"}))
	fst.Equal(t, 2*time.Millisecond, d1.Duration())

	fst.Error(t, fs.Parse([]string{"-d1", "abc"}))
	fst.Equal(t, time.Duration(0), d1.Duration())
}
