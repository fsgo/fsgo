// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package fstypes

import (
	"sync/atomic"
)

// EnableStatus enable and disable status, default value is enable
type EnableStatus int32

// IsEnable return is enable
func (es *EnableStatus) IsEnable() bool {
	return atomic.LoadInt32((*int32)(es)) == 0
}

// SetEnable set status
func (es *EnableStatus) SetEnable(enable bool) {
	if enable {
		atomic.StoreInt32((*int32)(es), 0)
	} else {
		atomic.StoreInt32((*int32)(es), 1)
	}
}

// Enable set status enable
func (es *EnableStatus) Enable() {
	es.SetEnable(true)
}

// Disable set status disable
func (es *EnableStatus) Disable() {
	es.SetEnable(false)
}

// String return status desc
func (es *EnableStatus) String() string {
	if es.IsEnable() {
		return "enable"
	}
	return "disable"
}
