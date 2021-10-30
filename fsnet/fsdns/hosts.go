// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsdns

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/fsgo/fsgo/fsnet/internal"
	"github.com/fsgo/fsgo/fsos/fsfs"
)

var DefaultHostsFile = NewHostsFile("")

// HostsLookupIP find domain in hosts file
func HostsLookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return DefaultHostsFile.LookupIP(ctx, network, host)
}

func NewHostsFile(hostPath string) *HostsFile {
	hf := &HostsFile{
		Path: hostPath,
	}
	if err := hf.Start(); err != nil {
		panic(err)
	}
	return hf
}

// HostsFile hosts file parser
type HostsFile struct {
	Path string

	domains map[string][]net.IP
	mux     sync.RWMutex

	tk     *time.Ticker
	onStop func()
}

var errNotFoundInHosts = errors.New("not found in hosts")

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
		ips = internal.FilterIPList(internal.IPv4only, ips)
	case "ip6":
		ips = internal.FilterIPList(internal.IPv6only, ips)
	default:
		return nil, fmt.Errorf("not support network=%q", network)
	}
	if len(ips) == 0 {
		return nil, errNotFoundInHosts
	}
	return ips, nil
}

func (hf *HostsFile) getPath() string {
	if hf.Path == "" {
		return "/etc/hosts"
	}
	return hf.Path
}

func (hf *HostsFile) Start() error {
	if err := hf.Load(); err != nil {
		return err
	}
	if hf.onStop != nil {
		return errors.New("already started")
	}
	w := &fsfs.Watcher{
		Interval: time.Second,
	}
	w.Watch(hf.getPath(), func(event *fsfs.WatcherEvent) {
		hf.Load()
	})
	hf.onStop = func() {
		w.Stop()
	}
	return nil
}

func (hf *HostsFile) Stop() {
	if hf.onStop != nil {
		hf.onStop()
		hf.onStop = nil
	}
}

func (hf *HostsFile) Load() error {
	return hf.load(hf.getPath())
}

func (hf *HostsFile) load(file string) error {
	if len(file) == 0 {
		return errors.New("hosts file path is empty")
	}
	bf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	lines := strings.Split(string(bf), "\n")

	domains := make(map[string][]net.IP)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		ip, hosts := hf.parserLine(line)
		if ip == nil {
			continue
		}
		for _, h := range hosts {
			domains[h] = append(domains[h], ip)
		}
	}
	hf.mux.Lock()
	hf.domains = domains
	hf.mux.Unlock()
	return nil
}

func (hf *HostsFile) parserLine(line string) (ip net.IP, hosts []string) {
	line = strings.ToLower(line)
	fields := strings.Fields(line)
	if len(fields) < 2 { // 异常数据
		return nil, nil
	}
	ip = net.ParseIP(fields[0])
	if ip == nil {
		return nil, nil
	}
	return ip, fields[1:]
}
