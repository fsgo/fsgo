// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/15

package fsatomic

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValue_Load(t *testing.T) {
	var val Value[time.Time]
	require.True(t, val.Load().IsZero())
	now := time.Now()
	val.Store(now)
	got := val.Load()
	require.Equal(t, now, got)

	n2 := time.Now()
	got1 := val.Swap(n2)
	require.Equal(t, got, got1)
	require.False(t, val.CompareAndSwap(time.Now(), time.Now()))
	require.True(t, val.CompareAndSwap(n2, time.Now()))
}

func TestError(t *testing.T) {
	var val Error
	require.Nil(t, val.Load())
	var e1 error
	val.Store(e1)
	require.Nil(t, val.Load())
	require.Nil(t, val.Swap(io.EOF))
	require.Error(t, val.Load())
	require.Error(t, val.Swap(nil))
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
