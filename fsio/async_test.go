// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package fsio

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAsyncWriter(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		b := &bytes.Buffer{}
		aw := &AsyncWriter{
			Writer:     b,
			ChanSize:   100,
			NeedStatus: true,
		}

		for i := 0; i < 1000; i++ {
			_, err := aw.Write([]byte("H"))
			require.NoError(t, err)
		}
		require.NoError(t, aw.Close())
		want := WriteStatus{
			Wrote: 1,
		}
		require.Equal(t, want, aw.LastWriteStatus())
		require.Equal(t, 1000, b.Len())
	})

	t.Run("no write", func(t *testing.T) {
		b := &bytes.Buffer{}
		aw := &AsyncWriter{
			Writer:     b,
			ChanSize:   100,
			NeedStatus: true,
		}
		require.NoError(t, aw.Close())
		want := WriteStatus{}
		require.Equal(t, want, aw.LastWriteStatus())
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
