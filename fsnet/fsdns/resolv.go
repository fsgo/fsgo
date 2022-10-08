// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/31

package fsdns

import (
	"bytes"
	"net"
	"strings"
	"sync"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/fsgo/fsgo/fstypes"
)

// DefaultResolvConf os default ResolvConf
var DefaultResolvConf = NewResolvConf("")

// ResolvNameserver get nameserver list from DefaultResolvConf
func ResolvNameserver() []string {
	return DefaultResolvConf.Nameserver()
}

// NewResolvConf create new ResolvConf instance
//
//	already start watch the confPath
func NewResolvConf(confPath string) *ResolvConf {
	rf := &ResolvConf{}
	rf.Parser = rf.parse
	rf.FileName = rf.getPath(confPath)

	if err := rf.Start(); err != nil {
		panic(err)
	}
	return rf
}

// ResolvConf parser for resolv.conf
type ResolvConf struct {
	re     *ResolverConfig
	onStop func()
	fsfs.WatchFile
	mux sync.RWMutex
}

func (rf *ResolvConf) getPath(fileName string) string {
	if fileName == "" {
		return "/etc/resolv.conf"
	}
	return fileName
}

func (rf *ResolvConf) parse(content []byte) error {
	re := ParseResolv(content)

	rf.mux.Lock()
	rf.re = re
	rf.mux.Unlock()
	return nil
}

// Config 读取 ResolverConfig
func (rf *ResolvConf) Config() *ResolverConfig {
	rf.mux.RLock()
	defer rf.mux.RUnlock()
	return rf.re
}

// Nameserver get nameserver list from file
func (rf *ResolvConf) Nameserver() []string {
	rf.mux.RLock()
	defer rf.mux.RUnlock()
	if rf.re == nil {
		return nil
	}
	return rf.re.Nameserver
}

// ResolverConfig resolv.conf
// see https://man7.org/linux/man-pages/man5/resolv.conf.5.html
type ResolverConfig struct {
	Nameserver []string
	Search     []string
	SortList   []string
	Options    []string
}

func (re *ResolverConfig) parserNameserver(info []string) {
	if len(info) != 1 {
		return
	}
	ipStr := info[0]
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return
	}
	re.Nameserver = append(re.Nameserver, ipStr)
}

func (re *ResolverConfig) parserSearch(info []string) {
	re.Search = append(re.Search, info...)
}

func (re *ResolverConfig) parserOptions(info []string) {
	re.Options = append(re.Options, info...)
}

func (re *ResolverConfig) parserSortList(info []string) {
	re.SortList = append(re.SortList, info...)
}

func (re *ResolverConfig) unique() {
	re.Nameserver = fstypes.StringSlice(re.Nameserver).Unique()
	re.Search = fstypes.StringSlice(re.Search).Unique()
	re.Options = fstypes.StringSlice(re.Options).Unique()
	re.SortList = fstypes.StringSlice(re.SortList).Unique()
}

// ParseResolv parse  resolv.conf content
func ParseResolv(content []byte) *ResolverConfig {
	lines := bytes.Split(content, []byte("\n"))
	re := &ResolverConfig{}
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		fields := strings.Fields(string(line))
		switch fields[0] {
		case "nameserver":
			re.parserNameserver(fields[1:])
		case "search":
			re.parserSearch(fields[1:])
		case "options":
			re.parserOptions(fields[1:])
		case "sortlist":
			re.parserSortList(fields[1:])
		}
	}
	re.unique()
	return re
}
