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
)

func ctxWithServerConnSession(ctx context.Context, session *ServerConnSession) context.Context {
	return context.WithValue(ctx, ctxKeyServerConnSession, session)
}

func ServerConnSessionFromCtx(ctx context.Context) *ServerConnSession {
	val, _ := ctx.Value(ctxKeyServerConnSession).(*ServerConnSession)
	return val
}
