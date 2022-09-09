// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/9/9

package fssync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAtomicTimeStamp(t *testing.T) {
	var a AtomicTimeStamp
	got1 := a.Load()
	require.True(t, got1.IsZero(), got1.String())
	now := time.Now()
	a.Store(now)
	got2 := a.Load()
	require.Equal(t, now.UnixNano(), got2.UnixNano())

	t2 := now.Add(time.Second)
	require.True(t, a.Before(t2))

	t3 := now.Add(-1 * time.Second)
	require.True(t, a.After(t3))

	require.Equal(t, time.Second, a.Sub(t3))

	require.Greater(t, a.Since(time.Now()), time.Duration(0))
}

func BenchmarkAtomicTimeStamp_Load(b *testing.B) {
	var a AtomicTimeStamp
	a.Store(time.Now())
	b.ResetTimer()
	var tm time.Time
	for i := 0; i < b.N; i++ {
		tm = a.Load()
	}
	_ = tm
}
