// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidProtocol = fmt.Errorf("invalid Protocol Header, expect is %s", Protocol)

	ErrInvalidHeader = errors.New("invalid Header")

	ErrInvalidCode = errors.New("invalid error code")

	ErrCannotWritePayload = errors.New("cannot write payload")

	ErrMethodNotFound = errors.New("method not found")

	ErrClosed = errors.New("already closed")

	ErrCanceledByDefer = errors.New("canceled by defer")

	ErrNoPayload = errors.New("no payload")

	ErrAuthFailed = errors.New("auth failed")
)
