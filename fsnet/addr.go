// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/1

package fsnet

import (
	"net"
)

// parseIPZone parses s as an IP address, return it and its associated zone
// identifier (IPv6 only).
func parseIPZone(s string) net.IP {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.', ':':
			return net.ParseIP(s)
		}
	}
	return nil
}
