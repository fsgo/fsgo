// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/15

package internal

type NoCopy struct{}

func (*NoCopy) Lock() {}

func (*NoCopy) Unlock() {}
