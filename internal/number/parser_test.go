// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/4

package number

import (
	"testing"

	"github.com/fsgo/fst"
)

func Test_parseNumber(t *testing.T) {
	t.Run("int-1", func(t *testing.T) {
		got, err := ParseNumber[int](" 123", 0)
		fst.NoError(t, err)
		fst.Equal(t, int(123), got)
	})

	t.Run("int-2-err", func(t *testing.T) {
		got, err := ParseNumber[int]("123.1", 0)
		fst.Error(t, err)
		fst.Equal(t, int(0), got)
	})

	t.Run("int8-err", func(t *testing.T) {
		got, err := ParseNumber[int8]("65535", int8(0))
		fst.Equal(t, int8(127), got)
		fst.Error(t, err)
	})

	t.Run("uint64-1", func(t *testing.T) {
		got, err := ParseNumber[uint64](" 123", uint64(0))
		fst.NoError(t, err)
		fst.Equal(t, uint64(123), got)
	})

	t.Run("float64-1", func(t *testing.T) {
		got, err := ParseNumber[float64](" 123.1", float64(0))
		fst.NoError(t, err)
		fst.Equal(t, float64(123.1), got)
	})
}
