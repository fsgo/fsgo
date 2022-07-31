// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/31

package fstime_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fsgo/fsgo/fstime"
)

func TestInterval(t *testing.T) {
	it := fstime.Interval{}
	defer it.Stop()
	var num int32
	it.Add(func() {
		atomic.AddInt32(&num, 1)
	})
	var f1 int32
	it.Add(func() {
		if it.Running() {
			atomic.AddInt32(&f1, 1)
		}
	})
	var f2 int32
	var wg2 sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg2.Add(1)
		it.Add(func() {
			defer wg2.Done()
			select {
			case <-it.Done():
				atomic.AddInt32(&f2, 1)
			case <-time.After(2 * time.Millisecond):
				return
			}
		})
	}
	it.Start(3 * time.Millisecond)
	time.Sleep(time.Millisecond)
	it.Stop()
	wg2.Wait()

	it.Reset(time.Millisecond)
	require.Equal(t, int32(1), atomic.LoadInt32(&num))
	require.Equal(t, int32(1), atomic.LoadInt32(&f1))
	require.Equal(t, int32(2), atomic.LoadInt32(&f2))
}
