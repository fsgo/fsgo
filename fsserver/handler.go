// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/17

package fsserver

import (
	"context"
	"net"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
}

type HandleFunc func(ctx context.Context, conn net.Conn)

func (hf HandleFunc) Handle(ctx context.Context, conn net.Conn) {
	hf(ctx, conn)
}
