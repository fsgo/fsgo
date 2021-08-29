// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/29

package fsos

import (
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKeepFile(t *testing.T) {
	fp := "testdata/tmp/keep.txt"
	defer os.Remove(fp)
	ci := 10 * time.Millisecond
	kp := &KeepFile{
		FilePath: func() string {
			return fp
		},
		CheckInterval: ci,
	}
	var changeNum int32
	kp.AfterChange(func(f *os.File) {
		atomic.AddInt32(&changeNum, 1)
	})

	require.Nil(t, kp.Start())

	t.Run("after start", func(t *testing.T) {
		require.Equal(t, int32(1), atomic.LoadInt32(&changeNum))
		require.NotNil(t, kp.File())
	})

	defer func() {
		require.Nil(t, kp.Stop())
	}()

	checkExists := func() {
		info, err := os.Stat(fp)
		require.Nil(t, err)
		require.NotNil(t, info)
	}

	t.Run("same file not change", func(t *testing.T) {
		stat1, err := kp.File().Stat()
		require.Nil(t, err)
		time.Sleep(ci * 2)

		stat2, err := kp.File().Stat()
		require.Nil(t, err)

		require.True(t, os.SameFile(stat1, stat2))
	})

	t.Run("rm and create it auto", func(t *testing.T) {
		checkExists()
		require.Nil(t, os.Remove(fp))
		time.Sleep(ci * 2)
		checkExists()
		require.Equal(t, int32(2), atomic.LoadInt32(&changeNum))
	})

	t.Run("stopped", func(t *testing.T) {
		require.Nil(t, kp.Stop())
		require.Nil(t, os.Remove(fp))

		time.Sleep(ci * 2)
		// check not exists
		_, err := os.Stat(fp)
		require.NotNil(t, err)
	})
}
