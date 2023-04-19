// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/15

package fsatomic

import (
	"sync/atomic"
	"time"
)

// TimeStamp atomic store for time stamp
// without time Location
type TimeStamp int64

// Load atomic load time
func (at *TimeStamp) Load() time.Time {
	v := atomic.LoadInt64((*int64)(at))
	if v == 0 {
		return time.Time{}
	}
	return time.Unix(v/1e9, v%1e9)
}

// Store atomic store time stamp
func (at *TimeStamp) Store(n time.Time) {
	atomic.StoreInt64((*int64)(at), n.UnixNano())
}

// Sub returns the duration t-n
func (at *TimeStamp) Sub(n time.Time) time.Duration {
	v := atomic.LoadInt64((*int64)(at))
	return time.Duration(v - n.UnixNano())
}

// Since returns the time elapsed since n.
func (at *TimeStamp) Since(n time.Time) time.Duration {
	v := atomic.LoadInt64((*int64)(at))
	return time.Duration(n.UnixNano() - v)
}

// Before reports whether the time instant t is before u.
func (at *TimeStamp) Before(n time.Time) bool {
	v := atomic.LoadInt64((*int64)(at))
	return v < n.UnixNano()
}

// After reports whether the time instant t is after u.
func (at *TimeStamp) After(n time.Time) bool {
	v := atomic.LoadInt64((*int64)(at))
	return v > n.UnixNano()
}

type TimeDuration = NumberInt64[time.Duration]
