// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/12/25

package fscmd

import (
	"context"
	"flag"
	"io"
)

type Actuator interface {
	Name() string
	Setup(fs *Config)
	FlagSet(fs *flag.FlagSet)
	Run(ctx context.Context, args []string) error
	String() string
}

type Config struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}
