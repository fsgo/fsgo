// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/24

package fsrpc

import (
	"context"
)

type ctxKey uint8

const (
	ctxKeyServerConnSession ctxKey = iota
	ctxKeyHandlerMethod
)

func ctxWithServerConnSession(ctx context.Context, session *ConnSession) context.Context {
	return context.WithValue(ctx, ctxKeyServerConnSession, session)
}

func ConnSessionFromCtx(ctx context.Context) *ConnSession {
	val, _ := ctx.Value(ctxKeyServerConnSession).(*ConnSession)
	return val
}

func ctxWithServerMethod(ctx context.Context, method string) context.Context {
	return context.WithValue(ctx, ctxKeyHandlerMethod, method)
}

func HandlerMethod(ctx context.Context) string {
	val, _ := ctx.Value(ctxKeyHandlerMethod).(string)
	return val
}
