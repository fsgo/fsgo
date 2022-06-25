// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/25

package fsserver

import (
	"context"
	"net"
)

type Server interface {
	Serve(l net.Listener) error
}

// CanShutdown 支持优雅关闭
type CanShutdown interface {
	Shutdown(ctx context.Context) error
}

type CanShutdownServer interface {
	Server
	CanShutdown
}
