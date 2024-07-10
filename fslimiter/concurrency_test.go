// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/3

package fslimiter

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestConcurrency(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var c Concurrency
		fn := c.Wait()
		fst.Equal(t, fmt.Sprintf("%p", empty), fmt.Sprintf("%p", fn))
	})

	t.Run("limit 1", func(t *testing.T) {
		c := &Concurrency{
			Max: 1,
		}

		done := make(chan bool)
		go func() {
			re := c.Wait()
			time.AfterFunc(3*time.Millisecond, re)
			done <- true
		}()

		<-done
		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			fn, err := c.WaitContext(ctx)
			fst.Error(t, err)
			fst.Nil(t, fn)
		}

		time.Sleep(3 * time.Millisecond)

		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			fn, err := c.WaitContext(ctx)
			fst.NoError(t, err)
			fst.NotNil(t, fn)
			fn()
		}
	})
	t.Run("limit 10", func(t *testing.T) {
		c := &Concurrency{
			Max: 10,
		}
		start := time.Now()
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fn := c.Wait()
				time.AfterFunc(time.Millisecond, fn)
			}()
		}
		wg.Wait()
		cost := time.Since(start)
		fst.GreaterOrEqual(t, int(cost/time.Millisecond), 10)
	})
}
