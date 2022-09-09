// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/9/9

package fssync

import (
	"sync/atomic"
	"time"
)

// AtomicTimeStamp atomic store for time stamp
// without time Location
type AtomicTimeStamp int64

// Load atomic load time
func (at *AtomicTimeStamp) Load() time.Time {
	v := atomic.LoadInt64((*int64)(at))
	if v == 0 {
		return time.Time{}
	}
	return time.Unix(v/1e9, v%1e9)
}

// Store atomic store time stamp
func (at *AtomicTimeStamp) Store(n time.Time) {
	atomic.StoreInt64((*int64)(at), n.UnixNano())
}

// Sub returns the duration t-n
func (at *AtomicTimeStamp) Sub(n time.Time) time.Duration {
	v := atomic.LoadInt64((*int64)(at))
	return time.Duration(v - n.UnixNano())
}

// Since returns the time elapsed since n.
func (at *AtomicTimeStamp) Since(n time.Time) time.Duration {
	v := atomic.LoadInt64((*int64)(at))
	return time.Duration(n.UnixNano() - v)
}

// Before reports whether the time instant t is before u.
func (at *AtomicTimeStamp) Before(n time.Time) bool {
	v := atomic.LoadInt64((*int64)(at))
	return v < n.UnixNano()
}

// After reports whether the time instant t is after u.
func (at *AtomicTimeStamp) After(n time.Time) bool {
	v := atomic.LoadInt64((*int64)(at))
	return v > n.UnixNano()
}
