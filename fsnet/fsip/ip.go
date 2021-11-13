// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/11/13

package fsip

import (
	"net"
)

// IsIPv4only 是否 ipv4
func IsIPv4only(ip net.IP) bool {
	return ip.To4() != nil
}

// IsIPv6only reports whether addr is an IPv6 address except IPv4-mapped IPv6 address.
func IsIPv6only(ip net.IP) bool {
	return len(ip) == net.IPv6len && ip.To4() == nil
}

// FilterList 对 ip 地址进行过滤
func FilterList(ips []net.IP, filter func(ip net.IP) bool) []net.IP {
	var result []net.IP
	for _, ip := range ips {
		if filter == nil || filter(ip) {
			result = append(result, ip)
		}
	}
	return result
}

// IsLoopback is Loopback ip
// 	ipv4: 127.*
func IsLoopback(ip net.IP) bool {
	return ip.IsLoopback()
}

// IsPrivate is private ip
// 	ipv4: 10/8、172.16/12 、192.168/16 prefix
// 	ipv6: FC00::/7 prefix
func IsPrivate(ip net.IP) bool {
	return ip.IsPrivate()
}

// IsLinkLocalUnicast 本地单播地址
// 	ipv4: 169.*、254.*
// 	ipv6: fe80::/7 prefix
func IsLinkLocalUnicast(ip net.IP) bool {
	return ip.IsLinkLocalUnicast()
}

// NewIsFnsAnd create new ip list filter func
func NewIsFnsAnd(fns ...func(ip net.IP) bool) func(ip net.IP) bool {
	return func(ip net.IP) bool {
		for i := 0; i < len(fns); i++ {
			if !fns[i](ip) {
				return false
			}
		}
		return true
	}
}

// NewIsIPFnsOr create new ip list filter func
func NewIsIPFnsOr(fns ...func(ip net.IP) bool) func(ip net.IP) bool {
	return func(ip net.IP) bool {
		for i := 0; i < len(fns); i++ {
			if fns[i](ip) {
				return true
			}
		}
		return false
	}
}

// NotFilter not filter
func NotFilter(fn func(ip net.IP) bool) func(ip net.IP) bool {
	return func(ip net.IP) bool {
		return !fn(ip)
	}
}
