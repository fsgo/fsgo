// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/29

package fsfn

import (
	"context"
	"errors"
)

func RunVoid(fn func()) {
	if fn == nil {
		return
	}
	defer Recover()
	fn()
}

func RunVoidError(fn func()) (err error) {
	if fn == nil {
		return
	}
	defer RecoverError(&err)
	fn()
	return nil
}

func RunVoids(fns []func()) {
	defer Recover()
	for _, fn := range fns {
		fn()
	}
}

func RunCtxVoid(ctx context.Context, fn func(context.Context)) {
	if fn == nil {
		return
	}
	defer RecoverCtx(ctx)
	fn(ctx)
}

func RunCtxVoids(ctx context.Context, fns []func(ctx2 context.Context)) {
	if len(fns) == 0 {
		return
	}
	defer RecoverCtx(ctx)
	for _, fn := range fns {
		fn(ctx)
	}
}

func RunError(fn func() error) (err error) {
	if fn == nil {
		return nil
	}
	defer RecoverError(&err)
	return fn()
}

func RunErrors(fns []func() error) (err error) {
	if len(fns) == 0 {
		return nil
	}
	defer RecoverError(&err)
	var errs []error
	for _, fn := range fns {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func RunCtxError(ctx context.Context, fn func(context.Context) error) (err error) {
	if fn == nil {
		return nil
	}
	defer RecoverCtxError(ctx, &err)
	return fn(ctx)
}

func RunCtxErrors(ctx context.Context, fns []func(context.Context) error) (err error) {
	if len(fns) == 0 {
		return nil
	}
	defer RecoverCtxError(ctx, &err)
	var errs []error
	for _, fn := range fns {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
