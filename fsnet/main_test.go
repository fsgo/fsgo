// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsnet

import (
	"testing"
)

type testNum int

func (tn *testNum) Check(t *testing.T, want int) {
	if int(*tn) != want {
		t.Fatalf("num=%d want=%d", *tn, want)
	}
	*tn++
}

func (tn *testNum) Incr() {
	*tn++
}
