// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/1

package fsnet

import (
	"net"

	"github.com/fsgo/fsgo/fsnet/internal"
)

// NewAddr new addr
func NewAddr(network, host string) net.Addr {
	return internal.NewAddr(network, host)
}
