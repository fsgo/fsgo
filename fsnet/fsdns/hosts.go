// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsdns

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/fsgo/fsgo/fsnet/fsip"
	"github.com/fsgo/fsgo/fsnet/fsresolver"
)

// DefaultHostsFile os default hosts file
var DefaultHostsFile fsresolver.LookupIPer = NewHostsFile("")

// LookupIPFromHosts find domain in hosts file
func LookupIPFromHosts(ctx context.Context, network, host string) ([]net.IP, error) {
	return DefaultHostsFile.LookupIP(ctx, network, host)
}

// NewHostsFile create new HostsFile instance
//
//	if start fail,then panic
//	already start watch the hostPath
//	when hostPath is empty, use the default path:
//	on unix like system,the default path is /etc/hosts.
//	the default hostsPath can set by Environment Variables 'FSGO_HOSTS',
//	eg: export FSGO_HOSTS=./my_hosts
func NewHostsFile(hostsPath string) *HostsFile {
	hf := &HostsFile{}
	hf.FileName = hf.getPath(hostsPath)
	hf.Parser = hf.parse
	if err := hf.Start(); err != nil {
		panic(err)
	}
	return hf
}

// HostsFile hosts file parser
type HostsFile struct {
	fsfs.WatchFile

	domains map[string][]net.IP
	mux     sync.RWMutex
}

func (hf *HostsFile) getPath(fileName string) string {
	if len(fileName) == 0 {
		return getDefaultHostsPath()
	}
	return fileName
}

func getDefaultHostsPath() string {
	hostPath := os.Getenv("FSGO_HOSTS")
	if len(hostPath) == 0 {
		// todo other system,eg windows
		return "/etc/hosts"
	}
	return hostPath
}

func (hf *HostsFile) parse(content []byte) error {
	domains := ParseHosts(content)
	hf.mux.Lock()
	hf.domains = domains
	hf.mux.Unlock()
	return nil
}

var errNotFoundInHosts = errors.New("not found in hosts")

// LookupIP lookup ip from hosts file
func (hf *HostsFile) LookupIP(_ context.Context, network, host string) ([]net.IP, error) {
	host = strings.ToLower(host)
	hf.mux.RLock()
	defer hf.mux.RUnlock()
	if len(hf.domains) == 0 {
		return nil, errNotFoundInHosts
	}
	ips := hf.domains[host]
	if len(ips) == 0 {
		return nil, errNotFoundInHosts
	}
	switch network {
	case "ip":
		return ips, nil
	case "ip4":
		ips = fsip.FilterList(ips, fsip.IsIPv4only)
	case "ip6":
		ips = fsip.FilterList(ips, fsip.IsIPv6only)
	default:
		return nil, fmt.Errorf("not support network=%q", network)
	}
	if len(ips) == 0 {
		return nil, errNotFoundInHosts
	}
	return ips, nil
}

// ParseHosts  hosts content
func ParseHosts(content []byte) map[string][]net.IP {
	lines := bytes.Split(content, []byte("\n"))
	domains := make(map[string][]net.IP)
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		line = bytes.ToLower(line)
		fields := strings.Fields(string(line))
		if len(fields) < 2 { // 异常数据
			continue
		}
		ip := net.ParseIP(fields[0])
		if ip == nil {
			continue
		}
		for _, h := range fields[1:] {
			domains[h] = append(domains[h], ip)
		}
	}
	return domains
}
