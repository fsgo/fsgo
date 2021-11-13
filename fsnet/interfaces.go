// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/11/13

package fsnet

import (
	"net"

	"github.com/fsgo/fsgo/fsnet/fsip"
)

// Interfaces system's  network interface helper
type Interfaces struct {
}

var netInterfaces = net.Interfaces

// Addrs a list of the system's all network interface's addrs
func (itf *Interfaces) Addrs() ([]net.Addr, error) {
	ifs, err := netInterfaces()
	if err != nil {
		return nil, err
	}
	var result []net.Addr
	for _, it := range ifs {
		addrs, err := it.Addrs()
		if err != nil {
			return nil, err
		}
		result = append(result, addrs...)
	}
	return result, nil
}

// IPs a list of  system's all network ips
func (itf *Interfaces) IPs() ([]net.IP, error) {
	return itf.ips(nil)
}

// IPv4s a list of  system's all network ipv4s
func (itf *Interfaces) IPv4s() ([]net.IP, error) {
	return itf.ips(fsip.IsIPv4only)
}

// IPv6s a list of  system's all network ipv6s
func (itf *Interfaces) IPv6s() ([]net.IP, error) {
	return itf.ips(fsip.IsIPv6only)
}

func (itf *Interfaces) ips(filter func(ip net.IP) bool) ([]net.IP, error) {
	ifs, err := netInterfaces()
	if err != nil {
		return nil, err
	}
	var result []net.IP
	for _, it := range ifs {
		addrs, err := it.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, err
			}
			if filter == nil || filter(ip) {
				result = append(result, ip)
			}
		}
	}
	return result, nil
}
