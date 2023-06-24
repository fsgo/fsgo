// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package fsio

import "io"

type FlushWriter interface {
	Flusher
	io.Writer
}
