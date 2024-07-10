// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/15

package fsatomic

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestValue_Load(t *testing.T) {
	var val Value[time.Time]
	fst.True(t, val.Load().IsZero())
	now := time.Now()
	val.Store(now)
	got := val.Load()
	fst.Equal(t, now, got)

	n2 := time.Now()
	got1 := val.Swap(n2)
	fst.Equal(t, got, got1)
	fst.False(t, val.CompareAndSwap(time.Now(), time.Now()))
	fst.True(t, val.CompareAndSwap(n2, time.Now()))
}

func TestError(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var val Error
		fst.Nil(t, val.Load())
		var e1 error
		val.Store(e1)
		fst.Nil(t, val.Load())
		fst.Nil(t, val.Swap(io.EOF))
		fst.Error(t, val.Load())
		fst.Error(t, val.Swap(nil))
		fst.NoError(t, val.Swap(nil))
		fst.True(t, val.CompareAndSwap(nil, io.EOF))
		fst.False(t, val.CompareAndSwap(nil, io.EOF))
		err2 := errors.New("some err")
		fst.True(t, val.CompareAndSwap(io.EOF, err2))
	})
	t.Run("case 2", func(t *testing.T) {
		var val2 Error
		fst.True(t, val2.CompareAndSwap(nil, io.EOF))
	})
	t.Run("case 3", func(t *testing.T) {
		var val2 Error
		fst.False(t, val2.CompareAndSwap(io.EOF, nil))
	})
	t.Run("case 4", func(t *testing.T) {
		var val2 Error
		val2.Store(io.EOF)
		fst.True(t, val2.CompareAndSwap(io.EOF, nil))
	})
	t.Run("case 5", func(t *testing.T) {
		var val2 Error
		fst.NoError(t, val2.Swap(io.EOF))
		fst.True(t, val2.CompareAndSwap(io.EOF, nil))
	})
}

func TestFuncVoid(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		var val FuncVoid
		fst.Nil(t, val.Load())
		fst.Nil(t, val.Swap(func() {}))
		fst.NotNil(t, val.Swap(func() {}))
	})
}

func BenchmarkValue_Load(b *testing.B) {
	var val Value[time.Time]
	val.Store(time.Now())
	b.ResetTimer()
	var tm time.Time
	for i := 0; i < b.N; i++ {
		tm = val.Load()
	}
	_ = tm
}
