// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/19

package fsatomic

import "sync/atomic"

// ErrorCounter 统计成功失败次数
type ErrorCounter struct {
	success atomic.Int64
	fail    atomic.Int64
}

func (ec *ErrorCounter) AddSuccess() int64 {
	return ec.success.Add(1)
}

func (ec *ErrorCounter) AddSuccessN(delta int64) int64 {
	return ec.success.Add(delta)
}

func (ec *ErrorCounter) AddFail() int64 {
	return ec.fail.Add(1)
}

func (ec *ErrorCounter) AddFailN(delta int64) int64 {
	return ec.fail.Add(delta)
}

func (ec *ErrorCounter) AddError(err error) {
	if err == nil {
		ec.success.Add(1)
	} else {
		ec.fail.Add(1)
	}
}

func (ec *ErrorCounter) Success() int64 {
	return ec.success.Load()
}

func (ec *ErrorCounter) Fail() int64 {
	return ec.fail.Load()
}
