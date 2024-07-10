// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fsflag

import (
	"flag"
	"testing"

	"github.com/fsgo/fst"
)

func TestSliceFlag(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	var s1 Slice[int]
	fs.Var(&s1, "s1", "")

	s2 := &Slice[int]{Sep: ";"}
	fs.Var(s2, "s2", "")

	var s3 Slice[int8]
	fs.Var(&s3, "s3", "")

	var s4 Slice[bool]
	fs.Var(&s4, "s4", "")

	var s5 Slice[string]
	fs.Var(&s5, "s5", "")

	var s6 Slice[uint16]
	fs.Var(&s6, "s6", "")

	var s7 Slice[float32]
	fs.Var(&s7, "s7", "")

	fst.NoError(t, fs.Parse([]string{"-s1", "123"}))
	fst.Equal(t, []int{123}, s1.Value())

	fst.NoError(t, fs.Parse([]string{"-s1", "123,456,,7"}))
	fst.Equal(t, []int{123, 456, 7}, s1.Value())

	fst.Error(t, fs.Parse([]string{"-s1", "123;456;"}))
	fst.Empty(t, s1.Value())

	fst.NoError(t, fs.Parse([]string{"-s2", "123;456;"}))
	fst.Equal(t, []int{123, 456}, s2.Value())

	fst.NoError(t, fs.Parse([]string{"-s3=123,45"}))
	fst.Equal(t, []int8{123, 45}, s3.Value())

	fst.Error(t, fs.Parse([]string{"-s3", "123,456"}))
	fst.Equal(t, []int8(nil), s3.Value())

	fst.NoError(t, fs.Parse([]string{"-s4=true,false,"}))
	fst.Equal(t, []bool{true, false}, s4.Value())

	fst.NoError(t, fs.Parse([]string{"-s5=true,false,"}))
	fst.Equal(t, []string{"true", "false"}, s5.Value())
	fst.NotEmpty(t, s5.String())

	fst.NoError(t, fs.Parse([]string{"-s6", "123,456"}))
	fst.Equal(t, []uint16{123, 456}, s6.Value())
	fst.NotEmpty(t, s6.String())

	fst.NoError(t, fs.Parse([]string{"-s7", "123,456,0.1"}))
	fst.Equal(t, []float32{123, 456, 0.1}, s7.Value())
	fst.NotEmpty(t, s7.String())
}
