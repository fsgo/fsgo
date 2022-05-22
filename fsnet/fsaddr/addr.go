// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsaddr

import (
	"net"

	"github.com/fsgo/fsgo/fsnet/internal"
)

// New new addr
func New(network, host string) net.Addr {
	return internal.NewAddr(network, host)
}

// Empty not network and host
var Empty = New("", "")
