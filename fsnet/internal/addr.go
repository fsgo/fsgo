// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/1

package internal

import (
	"net"
)

// NewAddr new addr
func NewAddr(network, host string) net.Addr {
	return &addr{
		network: network,
		host:    host,
	}
}

var _ net.Addr = (*addr)(nil)

type addr struct {
	network string
	host    string
}

func (a *addr) Network() string {
	return a.network
}

func (a *addr) String() string {
	return a.host
}

// ParseIPZone parses s as an IP address, return it and its associated zone
// identifier (IPv6 only).
func ParseIPZone(s string) (net.IP, string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return net.ParseIP(s), ""
		case ':':
			return parseIPv6Zone(s)
		}
	}
	return nil, ""
}

// parseIPv6Zone parses s as a literal IPv6 address and its associated zone
// identifier which is described in RFC 4007.
func parseIPv6Zone(s string) (net.IP, string) {
	s, zone := splitHostZone(s)
	return net.ParseIP(s), zone
}

func splitHostZone(s string) (host, zone string) {
	// The IPv6 scoped addressing zone identifier starts after the
	// last percent sign.
	if i := last(s, '%'); i > 0 {
		host, zone = s[:i], s[i+1:]
	} else {
		host = s
	}
	return
}

func last(s string, b byte) int {
	i := len(s)
	for i--; i >= 0; i-- {
		if s[i] == b {
			break
		}
	}
	return i
}
