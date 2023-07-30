// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/30

package fsatomic

import (
	"sync/atomic"

	"github.com/fsgo/fsgo/fssync/internal"
)

type Once struct {
	_   internal.NoCopy
	val int32
}

func (one *Once) DoOnce() bool {
	return atomic.CompareAndSwapInt32(&one.val, 0, 1)
}

func (one *Once) Not() bool {
	return atomic.LoadInt32(&one.val) == 0
}

func (one *Once) Done() bool {
	return atomic.LoadInt32(&one.val) == 1
}
