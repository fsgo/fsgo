// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/1

package fsnet

// Network dialer and resolver network types
type Network string

// 所有的 Network 定义
const (
	// NetworkTCP tcp network
	NetworkTCP        = "tcp"
	NetworkTCP4       = "tcp4"
	NetworkTCP6       = "tcp6"
	NetworkUDP        = "udp"
	NetworkUDP4       = "udp4"
	NetworkUDP6       = "udp6"
	NetworkIP         = "ip"
	NetworkIP4        = "ip4"
	NetworkIP6        = "ip6"
	NetworkUnix       = "unix"
	NetworkUnixGram   = "unixgram"
	NetworkUnixPacket = "unixpacket"
)

// Resolver 转换为 Resolver 所需要的类型
//
// 	tcp->ip，tcp4->ip4, tcp6->ip6
// 	udp->ip，udp4->ip4, udp6->ip6
// 	其他类型原样返回
func (nt Network) Resolver() Network {
	switch nt {
	case NetworkTCP, NetworkUDP:
		return NetworkIP
	case NetworkTCP4, NetworkUDP4:
		return NetworkIP4
	case NetworkTCP6, NetworkUDP6:
		return NetworkIP6
	}
	return nt
}

// IsIP 是否 ip 类型的（ip、ip4、ip6）
func (nt Network) IsIP() bool {
	switch nt {
	case NetworkIP, NetworkIP4, NetworkIP6:
		return true
	}
	return false
}

// String 字符串类型
func (nt Network) String() string {
	return string(nt)
}
