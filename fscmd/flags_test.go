// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/12/25

package fscmd

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSliceFlag_Set(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	var s1 SliceFlag[int]
	fs.Var(&s1, "s1", "")

	s2 := &SliceFlag[int]{Sep: ";"}
	fs.Var(s2, "s2", "")

	var s3 SliceFlag[int8]
	fs.Var(&s3, "s3", "")

	var s4 SliceFlag[bool]
	fs.Var(&s4, "s4", "")

	var s5 SliceFlag[string]
	fs.Var(&s5, "s5", "")

	var s6 SliceFlag[uint16]
	fs.Var(&s6, "s6", "")

	var s7 SliceFlag[float32]
	fs.Var(&s7, "s7", "")

	require.NoError(t, fs.Parse([]string{"-s1", "123"}))
	require.Equal(t, []int{123}, s1.Values)

	require.NoError(t, fs.Parse([]string{"-s1", "123,456,,7"}))
	require.Equal(t, []int{123, 456, 7}, s1.Values)

	require.Error(t, fs.Parse([]string{"-s1", "123;456;"}))
	require.Empty(t, s1.Values)

	require.NoError(t, fs.Parse([]string{"-s2", "123;456;"}))
	require.Equal(t, []int{123, 456}, s2.Values)

	require.NoError(t, fs.Parse([]string{"-s3=123,45"}))
	require.Equal(t, []int8{123, 45}, s3.Values)

	require.Error(t, fs.Parse([]string{"-s3", "123,456"}))
	require.Equal(t, []int8{123}, s3.Values)

	require.NoError(t, fs.Parse([]string{"-s4=true,false,"}))
	require.Equal(t, []bool{true, false}, s4.Values)

	require.NoError(t, fs.Parse([]string{"-s5=true,false,"}))
	require.Equal(t, []string{"true", "false"}, s5.Values)
	require.NotEmpty(t, s5.String())

	require.NoError(t, fs.Parse([]string{"-s6", "123,456"}))
	require.Equal(t, []uint16{123, 456}, s6.Values)
	require.NotEmpty(t, s6.String())

	require.NoError(t, fs.Parse([]string{"-s7", "123,456,0.1"}))
	require.Equal(t, []float32{123, 456, 0.1}, s7.Values)
	require.NotEmpty(t, s7.String())
}
