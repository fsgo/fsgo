// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/29

package fsos

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHookStderr(t *testing.T) {
	fp := "testdata/tmp/hook_stderr.txt"
	defer os.Remove(fp)
	kp := &KeepFile{
		FilePath: func() string {
			return fp
		},
	}
	require.Nil(t, kp.Start())
	defer func() {
		require.Nil(t, kp.Stop())
	}()

	require.Nil(t, HookStderr(kp.File()))

	checkFile := func(want string) {
		bf, err := os.ReadFile(fp)
		require.Nil(t, err)
		require.Contains(t, string(bf), want)
	}

	println("hello")
	checkFile("hello")

	require.Nil(t, kp.File().Truncate(0))

	log.Println("fsgo")
	checkFile("fsgo")
}

func TestHookStdout(t *testing.T) {
	fp := "testdata/tmp/hook_stdout.txt"
	defer os.Remove(fp)
	kp := &KeepFile{
		FilePath: func() string {
			return fp
		},
	}
	require.Nil(t, kp.Start())
	defer func() {
		require.Nil(t, kp.Stop())
	}()

	require.Nil(t, HookStdout(kp.File()))

	checkFile := func(want string) {
		bf, err := os.ReadFile(fp)
		require.Nil(t, err)
		require.Contains(t, string(bf), want)
	}
	fmt.Fprintf(os.Stdout, "%s", "hello\n")
	checkFile("hello")
}
