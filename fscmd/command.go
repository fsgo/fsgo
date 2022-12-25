// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/12/25

package fscmd

import (
	"context"
	"flag"
)

type Command struct {
	FlagSet flag.FlagSet
}

func (cmd *Command) Run(ctx context.Context, args []string) error {
	return nil
}
