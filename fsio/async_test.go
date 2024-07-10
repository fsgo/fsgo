// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package fsio

import (
	"bytes"
	"io"
	"sync"
	"testing"

	"github.com/fsgo/fst"
)

func TestAsyncWriter(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		b := &bytes.Buffer{}
		mw := NewMutexWriter(b)
		aw := &AsyncWriter{
			Writer:     mw,
			ChanSize:   100,
			NeedStatus: true,
		}

		for i := 0; i < 1000; i++ {
			_, err := aw.Write([]byte("H"))
			fst.NoError(t, err)
		}
		fst.NoError(t, aw.Close())
		want := WriteStatus{
			Wrote: 1,
		}
		fst.Equal(t, want, aw.LastWriteStatus())

		mw.WithRLock(func(_ io.Writer) {
			fst.Equal(t, 1000, b.Len())
		})
	})

	t.Run("no write", func(t *testing.T) {
		b := &bytes.Buffer{}
		aw := &AsyncWriter{
			Writer:     b,
			ChanSize:   100,
			NeedStatus: true,
		}
		fst.NoError(t, aw.Close())
		want := WriteStatus{}
		fst.Equal(t, want, aw.LastWriteStatus())
	})

	t.Run("with gor", func(t *testing.T) {
		b := &bytes.Buffer{}
		aw := &AsyncWriter{
			Writer:   b,
			ChanSize: 100,
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				_, _ = aw.Write([]byte("abc"))
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				_ = aw.Close()
			}
		}()
		wg.Wait()
	})
}
