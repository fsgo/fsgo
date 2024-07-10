// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/15

package fsatomic

import (
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestTimeStamp(t *testing.T) {
	var a TimeStamp
	got1 := a.Load()
	fst.True(t, got1.IsZero())
	now := time.Now()
	a.Store(now)
	got2 := a.Load()
	fst.Equal(t, now.UnixNano(), got2.UnixNano())

	t2 := now.Add(time.Second)
	fst.True(t, a.Before(t2))

	t3 := now.Add(-1 * time.Second)
	fst.True(t, a.After(t3))

	fst.Equal(t, time.Second, a.Sub(t3))

	fst.Greater(t, a.Since(time.Now()), time.Duration(0))
}

func BenchmarkTimeStamp_Load(b *testing.B) {
	var a TimeStamp
	a.Store(time.Now())
	b.ResetTimer()
	var tm time.Time
	for i := 0; i < b.N; i++ {
		tm = a.Load()
	}
	_ = tm
}
