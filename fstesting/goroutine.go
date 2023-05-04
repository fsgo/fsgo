// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fstesting

import (
	"runtime"
	"testing"
	"time"

	"github.com/fsgo/fsgo/fsruntime"
)

type Goroutine struct {
	T        testing.TB
	startSR  fsruntime.StackRecords
	startNum int
}

func (g *Goroutine) Start() {
	g.startSR = fsruntime.GoroutineStack()
	g.startNum = runtime.NumGoroutine()
}

func (g *Goroutine) WaitFinish(wait time.Duration) {
	w := wait / 100
	for i := 0; i < 100; i++ {
		if runtime.NumGoroutine() == g.startNum {
			return
		}
		time.Sleep(w)
	}
	sr := fsruntime.GoroutineStack()
	diff := fsruntime.StackRecordDiff(sr, g.startSR)
	if !diff.HasDiff() {
		return
	}
	g.T.Fatalf("GoroutineStack has Diff:\n%s", diff.String())
}

func (g *Goroutine) Finish() {
	g.WaitFinish(100 * time.Millisecond)
}
