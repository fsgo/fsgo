// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/miekg/dns"

	"github.com/fsgo/fsgo/fsnet"
	"github.com/fsgo/fsgo/fsnet/internal"
)

var errEmptyResult = fmt.Errorf("dns empty result")

var dnsExchange = func(ctx context.Context, host string, t uint16, ns net.Addr) (*dns.Msg, error) {
	c := new(dns.Client)
	m1 := new(dns.Msg)
	m1.SetQuestion(host+".", t)
	msg, _, err := c.ExchangeContext(ctx, m1, ns.String())
	return msg, err
}

// LookupIPByNS LookupIP with nameserver
func LookupIPByNS(ctx context.Context, network, host string, ns net.Addr) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	var t uint16
	switch network {
	case "ip4":
		t = dns.TypeA
	case "ip6":
		t = dns.TypeAAAA
	case "ip":
		ret4, err4 := LookupIPByNS(ctx, "ip4", host, ns)
		if ctx.Err() != nil {
			return ret4, err4
		}
		ret6, err6 := LookupIPByNS(ctx, "ip6", host, ns)
		if err4 != nil && err6 != nil {
			return nil, fmt.Errorf("query ip4 err: %w; query ip6 err: %v", err4, err6)
		}
		return append(ret4, ret6...), nil
	}
	msg, err := dnsExchange(ctx, host, t, ns)
	if err != nil {
		return nil, fmt.Errorf("lookup %s faild: %w", host, err)
	}
	result := make([]net.IP, 0, len(msg.Answer))
	for _, item := range msg.Answer {
		switch t {
		case dns.TypeA:
			if ra, ok := item.(*dns.A); ok {
				result = append(result, ra.A)
			}
		case dns.TypeAAAA:
			if ra, ok := item.(*dns.AAAA); ok {
				result = append(result, ra.AAAA)
			}
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("lookup %s faild: %w", host, errEmptyResult)
	}
	return result, nil
}

var _ fsnet.Resolver = (*Client)(nil)

// Client prue dns client
type Client struct {
	// Servers nameserver list,eg 114.114.114.114:53
	Servers []net.Addr

	HostsFile fsnet.HasLookupIP

	mux sync.RWMutex

	// LookupIPHook after query dns success, hook the result
	LookupIPHook func(ctx context.Context, network, host string, ns net.Addr, result []net.IP) ([]net.IP, error)
}

// SetServers set servers
func (client *Client) SetServers(servers []net.Addr) {
	client.mux.Lock()
	client.Servers = servers
	client.mux.Unlock()
}

// GetServers get servers
func (client *Client) GetServers() []net.Addr {
	client.mux.RLock()
	client.mux.RUnlock()
	return client.Servers
}

// LookupIP lookup ip
func (client *Client) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	if client.HostsFile != nil {
		// 先尝试读取
		if ips, _ := client.HostsFile.LookupIP(ctx, network, host); len(ips) > 0 {
			return ips, nil
		}
	}
	switch network {
	case "ip", "ip4", "ip6":
		return client.lookupIP(ctx, network, host)
	default:
		return nil, fmt.Errorf("not support network %q", network)
	}
}

func (client *Client) callLookupIPHook(ctx context.Context, network, host string, ns net.Addr, result []net.IP) ([]net.IP, error) {
	if client.LookupIPHook == nil {
		return result, nil
	}
	ret, err := client.LookupIPHook(ctx, network, host, ns, result)
	if err != nil {
		return nil, err
	}
	if len(ret) == 0 {
		return nil, errEmptyResult
	}
	return ret, nil
}

func (client *Client) lookupIP(ctx context.Context, network, host string) (ret []net.IP, err error) {
	servers := client.GetServers()
	if len(servers) == 0 {
		return nil, errors.New("no nameserver")
	}
	for _, ns := range servers {
		ret, err = LookupIPByNS(ctx, network, host, ns)
		if err == nil && len(ret) > 0 {
			ret, err = client.callLookupIPHook(ctx, network, host, ns, ret)
		}
		if err == nil {
			return ret, nil
		}
		if ctx.Err() != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("query all nameserver faild, last err: %w", err)
}

// LookupIPAddr lookup ipaddr
func (client *Client) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	ips, err := client.lookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	result := make([]net.IPAddr, len(ips))
	for i := 0; i < len(ips); i++ {
		ip, zone := internal.ParseIPZone(ips[i].String())
		result[i] = net.IPAddr{
			IP:   ip,
			Zone: zone,
		}
	}
	return result, nil
}

// ResolverHook to ResolverHook
func (client *Client) ResolverHook() *fsnet.ResolverHook {
	return &fsnet.ResolverHook{
		LookupIP: func(ctx context.Context, network, host string, fn fsnet.LookupIPFunc) ([]net.IP, error) {
			return client.LookupIP(ctx, network, host)
		},
		LookupIPAddr: func(ctx context.Context, host string, fn fsnet.LookupIPAddrFunc) ([]net.IPAddr, error) {
			return client.LookupIPAddr(ctx, host)
		},
	}
}
