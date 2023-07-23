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

	ErrInvalidCode = errors.New("invalid closedErr code")

	ErrCannotWritePayload = errors.New("cannot write payload")

	ErrMethodNotFound = errors.New("method not found")

	ErrClosed = errors.New("already closed")

	ErrNoPayload = errors.New("no payload")

	// ErrMorePayload 还有更多的 payload 数据需要读取
	ErrMorePayload = errors.New("has more payload need to read")

	ErrInvalidEncodingType = errors.New("invalid encoding type")

	ErrAuthFailed = errors.New("auth failed")
)

type stringError string

func (e stringError) Error() string {
	return string(e)
}
