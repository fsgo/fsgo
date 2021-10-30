// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/29

package fsnet

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var _ Resolver = (*DNSClient)(nil)

// DNSClient prue dns client
type DNSClient struct {
	// Servers nameserver list,eg 114.114.114.114:53
	Servers []net.Addr

	// LookupIPHook after query dns success, hook the result
	LookupIPHook func(ctx context.Context, network, host string, ns net.Addr, result []net.IP, err error) ([]net.IP, error)
}

// ServersFromEnv parser nameserver list from os.env 'FSNET_NAMESERVER'
// 	eg: export FSNET_NAMESERVER=1.1.1.1,8.8.8.8:53
func (client *DNSClient) ServersFromEnv() []net.Addr {
	ev := os.Getenv("FSNET_NAMESERVER")
	if len(ev) == 0 {
		return nil
	}
	var list []net.Addr
	for _, host := range strings.Split(ev, ",") {
		host = strings.TrimSpace(host)
		if len(host) == 0 {
			continue
		}
		if ip := net.ParseIP(host); ip != nil {
			list = append(list, NewAddr("udp", host+":53"))
			continue
		}
		if _, _, err := net.SplitHostPort(host); err == nil {
			list = append(list, NewAddr("udp", host))
		}
	}
	return list
}

// SetServersAuto use ServersFromEnv or use def list
func (client *DNSClient) SetServersAuto(def []net.Addr) {
	list := client.ServersFromEnv()
	if len(list) > 0 {
		client.Servers = list
		return
	}
	client.Servers = def
}

// LookupIP lookup ip
func (client *DNSClient) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	switch network {
	case "ip", "ip4", "ip6":
		return client.lookupIP(ctx, network, host)
	default:
		return nil, fmt.Errorf("not support network %q", network)
	}
}

func (client *DNSClient) lookupIP(ctx context.Context, network, host string) (ret []net.IP, err error) {
	if len(client.Servers) == 0 {
		return nil, errors.New("no nameserver")
	}
	for _, ns := range client.Servers {
		ret, err = client.LookupIPByNS(ctx, network, host, ns)
		if err == nil {
			return ret, nil
		}
		if ctx.Err() != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("query all nameserver faild, last err: %w", err)
}

var errDNSEmptyResult = fmt.Errorf("dns empty result")

var dnsExchange = func(ctx context.Context, host string, t uint16, ns net.Addr) (*dns.Msg, error) {
	c := new(dns.Client)
	m1 := new(dns.Msg)
	m1.SetQuestion(host+".", t)
	msg, _, err := c.ExchangeContext(ctx, m1, ns.String())
	return msg, err
}

// LookupIPByNS LookupIP by nameserver
func (client *DNSClient) LookupIPByNS(ctx context.Context, network, host string, ns net.Addr) ([]net.IP, error) {
	var t uint16
	switch network {
	case "ip4":
		t = dns.TypeA
	case "ip6":
		t = dns.TypeAAAA
	case "ip":
		ret4, err4 := client.LookupIPByNS(ctx, "ip4", host, ns)
		if ctx.Err() != nil {
			return ret4, err4
		}
		ret6, err6 := client.LookupIPByNS(ctx, "ip6", host, ns)
		if err4 != nil && err6 != nil {
			return nil, fmt.Errorf("query ip4 err: %w; ip6 err: %v", err4, err6)
		}
		return append(ret4, ret6...), nil
	}
	msg, err := dnsExchange(ctx, host, t, ns)
	if err != nil {
		return nil, err
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
	var errRet error
	if len(result) == 0 {
		errRet = errDNSEmptyResult
	}
	if client.LookupIPHook != nil {
		result, errRet = client.LookupIPHook(ctx, network, host, ns, result, errRet)
	}
	if errRet != nil {
		return nil, errRet
	}
	return result, nil
}

// LookupIPAddr lookup ipaddr
func (client *DNSClient) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	ips, err := client.lookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	result := make([]net.IPAddr, len(ips))
	for i := 0; i < len(ips); i++ {
		ip, zone := parseIPZone(ips[i].String())
		result[i] = net.IPAddr{
			IP:   ip,
			Zone: zone,
		}
	}
	return result, nil
}

// ResolverHook to ResolverHook
func (client *DNSClient) ResolverHook() *ResolverHook {
	return &ResolverHook{
		LookupIP: func(ctx context.Context, network, host string, fn LookupIPFunc) ([]net.IP, error) {
			return client.LookupIP(ctx, network, host)
		},
		LookupIPAddr: func(ctx context.Context, host string, fn LookupIPAddrFunc) ([]net.IPAddr, error) {
			return client.LookupIPAddr(ctx, host)
		},
	}
}
