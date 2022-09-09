// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/9/9

package fssync

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
