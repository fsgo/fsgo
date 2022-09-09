// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package fstypes

import (
	"fmt"
	"strings"
	"sync"
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

// GroupEnableStatus 一组状态，默认未设置会返回 false
type GroupEnableStatus struct {
	all atomic.Value

	detail sync.Map
	// Default 未设置时的默认值
	Default bool
}

// IsEnable 获取状态
func (gs *GroupEnableStatus) IsEnable(key any) bool {
	value, ok := gs.detail.Load(key)
	if !ok {
		if a, o1 := gs.all.Load().(bool); o1 {
			return a
		}
		return gs.Default
	}
	return value == true
}

// SetEnable 设置状态
func (gs *GroupEnableStatus) SetEnable(key any, enable bool) {
	gs.detail.Store(key, enable)
}

// SetAllEnable 设置所有的状态,设置后，也会调整默认状态为此值
//
//	如 默认值为 false，若 key="key123" 未设置，调用 IsEnable("key123") 会返回 false,
//	当 SetAllEnable(true) 之后，再次调用 IsEnable("key123") 会返回 true
func (gs *GroupEnableStatus) SetAllEnable(enable bool) {
	gs.detail.Range(func(key, _ any) bool {
		gs.detail.Store(key, enable)
		return true
	})
	gs.all.Store(enable)
}

// Range 遍历所有已设置的值
func (gs *GroupEnableStatus) Range(fn func(key any, enable bool) bool) {
	gs.detail.Range(func(key, value any) bool {
		return fn(key, value.(bool))
	})
}

type specKey uint8

const notFoundStatusKey = specKey(0)

// String 打印出所有已经设置的状态信息
func (gs *GroupEnableStatus) String() string {
	var b strings.Builder
	gs.detail.Range(func(key, value any) bool {
		b.WriteString(fmt.Sprint(key))
		b.WriteString(":")
		if value == true {
			b.WriteString("true,")
		} else {
			b.WriteString("false,")
		}
		return true
	})
	if gs.IsEnable(notFoundStatusKey) {
		b.WriteString("*other*:true")
	} else {
		b.WriteString("*other*:false")
	}
	return b.String()
}
