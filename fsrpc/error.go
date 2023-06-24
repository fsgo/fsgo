// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidHeaderType = errors.New("invalid Header type")
	ErrInvalidProtocol   = fmt.Errorf("invalid Protocol Header, expect is %s", Protocol)
	ErrMissWriteMeta     = errors.New("should wirte meta first")
	ErrMethodNotFound    = errors.New("method not found")
	ErrClosed            = errors.New("already closed")

	ErrCanceledByDefer = errors.New("canceled by defer")
)
