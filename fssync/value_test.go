// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/9/9

package fssync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAtomicValue_Load(t *testing.T) {
	var val AtomicValue[time.Time]
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

func BenchmarkAtomicValue_Load(b *testing.B) {
	var val AtomicValue[time.Time]
	val.Store(time.Now())
	b.ResetTimer()
	var tm time.Time
	for i := 0; i < b.N; i++ {
		tm = val.Load()
	}
	_ = tm
}
