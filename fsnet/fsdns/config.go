// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsdns

import (
	"net"
	"os"
	"strings"

	"github.com/fsgo/fsgo/fsnet/fsaddr"
)

// ServersFromEnv parser nameserver list from os.env 'FSNET_NAMESERVER'
//
//	eg: export FSNET_NAMESERVER=1.1.1.1,8.8.8.8:53
func ServersFromEnv() []net.Addr {
	ev := os.Getenv("FSNET_NAMESERVER")
	if len(ev) == 0 {
		return nil
	}
	return ParserServers(strings.Split(ev, ","))
}

// ParserServers parser servers addr from lines
func ParserServers(lines []string) []net.Addr {
	var list []net.Addr
	for _, host := range lines {
		host = strings.TrimSpace(host)
		if len(host) == 0 {
			continue
		}
		if ip := net.ParseIP(host); ip != nil {
			list = append(list, fsaddr.New("udp", host+":53"))
			continue
		}
		if _, _, err := net.SplitHostPort(host); err == nil {
			list = append(list, fsaddr.New("udp", host))
		}
	}
	return list
}
