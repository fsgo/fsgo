// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/29

package grace

import (
	"fmt"
	"strings"
)

var drivers = map[string]ResourceDriverFunc{}

// ResourceDriverFunc 解析 DNS 配置
// dsn like "tcp@127.0.0.1:8080"
type ResourceDriverFunc func(dsn string) (Resource, error)

// RegisterResourceDriver 注册新的资源解析协议
func RegisterResourceDriver(scheme string, fn ResourceDriverFunc) {
	drivers[scheme] = fn
}

func init() {
	RegisterResourceDriver("tcp", netResourceDrive)
	RegisterResourceDriver("tcp4", netResourceDrive)
	RegisterResourceDriver("tcp6", netResourceDrive)
	RegisterResourceDriver("udp", netResourceDrive)
	RegisterResourceDriver("udp4", netResourceDrive)
	RegisterResourceDriver("udp6", netResourceDrive)
	RegisterResourceDriver("unix", netResourceDrive)
	RegisterResourceDriver("unixpacket", netResourceDrive)
}

func netResourceDrive(dsn string) (Resource, error) {
	arr := strings.SplitN(dsn, "@", 2)
	if len(arr) != 2 {
		return nil, fmt.Errorf("wrong dsn format")
	}
	return &ListenerResource{
		NetWork: arr[0],
		Address: strings.TrimSpace(arr[1]),
	}, nil
}

// GenResourceByDSN 通过 DNS 获取一个 Resource
func GenResourceByDSN(dsn string) (Resource, error) {
	arr := strings.SplitN(dsn, "@", 2)
	if len(arr) != 2 {
		return nil, fmt.Errorf("wrong dsn format")
	}
	scheme := arr[0]
	driverFunc, has := drivers[scheme]
	if !has {
		return nil, fmt.Errorf("scheme=%q not support", scheme)
	}
	return driverFunc(dsn)
}
