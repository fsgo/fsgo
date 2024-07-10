// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/29

package fsfn

import (
	"errors"
	"io"
	"sync/atomic"
	"testing"

	"github.com/fsgo/fst"
)

func TestRunVoid(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var num atomic.Int32
		RunVoid(func() {
			num.Add(1)
		})
		fst.Equal(t, int32(1), num.Load())
	})
	t.Run("case 2", func(t *testing.T) {
		RunVoid(func() {
			panic("hello")
		})
	})
	t.Run("case 3", func(t *testing.T) {
		RunVoid(nil)
	})
}

func TestRunError(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		err := RunError(func() error {
			return io.EOF
		})
		fst.SamePtr(t, io.EOF, err)
	})
	t.Run("case 2", func(t *testing.T) {
		err := RunError(func() error {
			panic("hello")
		})
		fst.Error(t, err)
		fst.True(t, errors.Is(err, ErrPanic))
	})
}
