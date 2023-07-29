// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/29

package fsfn

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

var ErrPanic = errors.New("panic")

func Recover() {
	re := recover()
	if re == nil {
		return
	}
	if fns := loadOnRecoverFns(); len(fns) > 0 {
		data := RecoverData{
			Re:    re,
			Stack: debug.Stack(),
		}
		runOnRecoverFns(data, fns)
	}
}

func RecoverError(err *error) {
	re := recover()
	if re == nil {
		return
	}
	*err = fmt.Errorf("%w: %v", ErrPanic, re)
	if fns := loadOnRecoverFns(); len(fns) > 0 {
		data := RecoverData{
			Re:    re,
			Stack: debug.Stack(),
		}
		runOnRecoverFns(data, fns)
	}
}

func RecoverCtx(ctx context.Context) {
	re := recover()
	if re == nil {
		return
	}
	if fns := loadOnRecoverFns(); len(fns) > 0 {
		data := RecoverData{
			Re:    re,
			Ctx:   ctx,
			Stack: debug.Stack(),
		}
		runOnRecoverFns(data, fns)
	}
}

func RecoverCtxError(ctx context.Context, err *error) {
	re := recover()
	if re == nil {
		return
	}
	*err = fmt.Errorf("%w: %v", ErrPanic, re)
	if fns := loadOnRecoverFns(); len(fns) > 0 {
		data := RecoverData{
			Re:    re,
			Ctx:   ctx,
			Stack: debug.Stack(),
		}
		runOnRecoverFns(data, fns)
	}
}

func runOnRecoverFns(data RecoverData, fns []func(RecoverData)) {
	defer func() {
		recover()
	}()
	for _, fn := range fns {
		fn(data)
	}
}

type RecoverData struct {
	Ctx   context.Context
	Re    any
	Stack []byte
}

func loadOnRecoverFns() []func(RecoverData) {
	onRecoverFnsMux.RLock()
	fns := onRecoverFns
	onRecoverFnsMux.RUnlock()
	return fns
}

var onRecoverFnsMux sync.RWMutex
var onRecoverFns []func(RecoverData)

// RegisterOnRecover recover callback
func RegisterOnRecover(fn func(RecoverData)) {
	onRecoverFnsMux.Lock()
	onRecoverFns = append(onRecoverFns, fn)
	onRecoverFnsMux.Unlock()
}
