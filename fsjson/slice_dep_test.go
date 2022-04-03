// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/3

package fsjson

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseNumber(t *testing.T) {
	t.Run("int-1", func(t *testing.T) {
		got, err := parseNumber[int](" 123", numberTypeSigned)
		require.NoError(t, err)
		require.Equal(t, int(123), got)
	})

	t.Run("int-2-err", func(t *testing.T) {
		got, err := parseNumber[int]("123.1", numberTypeSigned)
		require.Error(t, err)
		require.Equal(t, int(0), got)
	})

	t.Run("uint64-1", func(t *testing.T) {
		got, err := parseNumber[uint64](" 123", numberTypeFloat)
		require.NoError(t, err)
		require.Equal(t, uint64(123), got)
	})

	t.Run("float64-1", func(t *testing.T) {
		got, err := parseNumber[float64](" 123.1", numberTypeFloat)
		require.NoError(t, err)
		require.Equal(t, float64(123.1), got)
	})

}
