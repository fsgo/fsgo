// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsfs

import (
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestKeepFile(t *testing.T) {
	fp := "testdata/tmp/keep.txt"
	defer os.Remove(fp)
	ci := 10 * time.Millisecond
	kp := &Keeper{
		FilePath: func() string {
			return fp
		},
		CheckInterval: ci,
	}
	var changeNum int32
	kp.AfterChange(func(f *os.File) {
		atomic.AddInt32(&changeNum, 1)
	})

	fst.NoError(t, kp.Start())

	t.Run("after start", func(t *testing.T) {
		fst.Equal(t, int32(1), atomic.LoadInt32(&changeNum))
		fst.NotNil(t, kp.File())
	})

	defer kp.Stop()

	checkExists := func() {
		info, err := os.Stat(fp)
		fst.NoError(t, err)
		fst.NotNil(t, info)
	}

	t.Run("same file not change", func(t *testing.T) {
		stat1, err := kp.File().Stat()
		fst.NoError(t, err)
		time.Sleep(ci * 2)

		stat2, err := kp.File().Stat()
		fst.NoError(t, err)

		fst.True(t, os.SameFile(stat1, stat2))
	})

	t.Run("rm and create it auto", func(t *testing.T) {
		checkExists()
		fst.Nil(t, os.Remove(fp))
		time.Sleep(ci * 2)
		checkExists()
		fst.Equal(t, int32(2), atomic.LoadInt32(&changeNum))
	})

	t.Run("stopped", func(t *testing.T) {
		kp.Stop()
		time.Sleep(ci * 2)
		fst.NoError(t, os.Remove(fp))

		// check not exists
		_, err := os.Stat(fp)
		fst.Error(t, err)
	})
}
