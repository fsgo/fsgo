// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/29

package fsfn

import (
	"errors"
	"io"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunVoid(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var num atomic.Int32
		RunVoid(func() {
			num.Add(1)
		})
		require.Equal(t, int32(1), num.Load())
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
		require.Same(t, io.EOF, err)
	})
	t.Run("case 2", func(t *testing.T) {
		err := RunError(func() error {
			panic("hello")
		})
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrPanic))
	})
}
