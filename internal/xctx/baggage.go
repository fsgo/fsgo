// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/5

package xctx

import (
	"context"

	"github.com/fsgo/fsgo/fstypes"
)

type baggage[K any, V any] struct {
	ctx    context.Context
	values []V
}

func (b *baggage[K, V]) All(key K) []V {
	var vs []V
	if pic, ok := b.ctx.Value(key).(*baggage[K, V]); ok {
		vs = pic.All(key)
	}
	if len(vs) == 0 {
		return b.values
	} else if len(b.values) == 0 {
		return vs
	}
	return fstypes.SliceMerge(vs, b.values)
}

func WithValues[K any, V any](ctx context.Context, key K, vs ...V) context.Context {
	if len(vs) == 0 {
		return ctx
	}
	val := &baggage[K, V]{
		ctx:    ctx,
		values: vs,
	}
	return context.WithValue(ctx, key, val)
}

func Values[K any, V any](ctx context.Context, key K) []V {
	if val, ok := ctx.Value(key).(*baggage[K, V]); ok {
		return val.All(key)
	}
	return nil
}
