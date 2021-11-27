// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/11/27

package fsio

import (
	"io"
)

// NopWriteCloser nop closer for writer
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{Writer: w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}
