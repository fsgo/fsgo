// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fstesting

type TF interface {
	Fatalf(format string, args ...any)
	Fatal(args ...any)
}

type TE interface {
	Error(args ...any)
	Errorf(format string, args ...any)
}
