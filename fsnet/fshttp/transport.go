// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/5/15

package fshttp

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/http/httpproxy"

	"github.com/fsgo/fsgo/fsnet"
)

var _ http.RoundTripper = (*Transport)(nil)

// Transport  一个简单的 Transport
//
// 不管理连接,可以使用外部连接池
type Transport struct {
	// Proxy 可选
	Proxy func(*http.Request) (*url.URL, error)

	DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

	DialTLSContext func(ctx context.Context, network, addr string) (net.Conn, error)

	TLSClientConfig *tls.Config
}

func (t *Transport) dialConn(ctx context.Context, req *http.Request, proxyURL *url.URL, network, addr string) (net.Conn, error) {
	if proxyURL != nil && (proxyURL.Scheme == "http" || proxyURL.Scheme == "https") {
		return t.dial(ctx, network, addr)
	}

	switch req.URL.Scheme {
	case "https":
		return t.dialTLS(ctx, req, network, addr)
	}
	return t.dial(ctx, network, addr)
}

func (t *Transport) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	if t.DialContext != nil {
		return t.DialContext(ctx, network, addr)
	}
	return fsnet.DialContext(ctx, network, addr)
}

func (t *Transport) dialTLS(ctx context.Context, req *http.Request, network, addr string) (net.Conn, error) {
	if t.DialTLSContext != nil {
		return t.DialTLSContext(ctx, network, addr)
	}
	conn, err := t.dial(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	hostName := req.Host
	if hostName == "" {
		hostName, _, _ = net.SplitHostPort(addr)
	} else if strings.Contains(hostName, ":") {
		hostName, _, _ = net.SplitHostPort(hostName)
	}

	return t.connAddTLS(conn, hostName), nil
}

func (t *Transport) connAddTLS(conn net.Conn, hostName string) net.Conn {
	cfg := cloneTLSConfig(t.TLSClientConfig)

	// todo check user config？
	//  if t.TLSClientConfig==nil && cfg.ServerName == "" {}
	if cfg.ServerName == "" {
		cfg.ServerName = hostName
	}
	return tls.Client(conn, cfg)
}

func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		return &tls.Config{}
	}
	return cfg.Clone()
}

var envProxy = httpproxy.FromEnvironment().ProxyFunc()

func (t *Transport) getProxy() func(*http.Request) (*url.URL, error) {
	if t.Proxy != nil {
		return t.Proxy
	}
	return func(req *http.Request) (*url.URL, error) {
		return envProxy(req.URL)
	}
}

func (t *Transport) getAddress(req *http.Request) (address string, proxyURL *url.URL, err error) {
	proxyURL, err = t.getProxy()(req)
	if err != nil {
		return "", nil, fmt.Errorf("getProxy failed:%w", err)
	}

	if proxyURL == nil {
		return canonicalAddr(req.URL), nil, nil
	}

	address = canonicalAddr(proxyURL)

	proxyAuthVal := proxyAuth(proxyURL)
	if proxyAuthVal != "" && req.URL.Scheme == "http" {
		req.Header.Set(authKey, proxyAuthVal)
	}
	return address, proxyURL, nil
}

// RoundTrip 发送请求
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	defer closeRequest(req)
	if err := checkRequest(req); err != nil {
		return nil, err
	}

	address, proxyURL, err := t.getAddress(req)
	if err != nil {
		return nil, err
	}

	conn, err := t.dialConn(req.Context(), req, proxyURL, "tcp", address)

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if dl, ok := req.Context().Deadline(); ok {
		if err = conn.SetDeadline(dl); err != nil {
			return nil, err
		}
	}

	// https proxy
	if proxyURL != nil && req.URL.Scheme == "https" {
		if err = proxyHTTPSConn(conn, req, proxyURL); err != nil {
			return nil, err
		}
		conn = t.connAddTLS(conn, proxyURL.Hostname())
	}

	// http proxy 已可以工作
	// todo https 和 socks 待实现

	if proxyURL != nil {
		err = req.WriteProxy(conn)
	} else {
		err = req.Write(conn)
	}

	if err != nil {
		return nil, err
	}
	bio := bufio.NewReader(conn)
	return http.ReadResponse(bio, req)
}

const authKey = "Proxy-Authorization"

func proxyAuth(proxy *url.URL) string {
	if u := proxy.User; u != nil {
		username := u.Username()
		password, _ := u.Password()
		return "Basic " + basicAuth(username, password)
	}
	return ""
}

// https://datatracker.ietf.org/doc/html/rfc7230
// https://datatracker.ietf.org/doc/html/rfc7231#section-4.3.6
func proxyHTTPSConn(conn net.Conn, req *http.Request, proxy *url.URL) error {
	proxyAuthVal := proxyAuth(proxy)
	hdr := make(http.Header)
	if proxyAuthVal != "" {
		hdr.Set(authKey, proxyAuthVal)
	}
	connectReq := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: req.URL.Host},
		Host:   req.URL.Host,
		Header: hdr,
	}
	err := connectReq.Write(conn)
	if err != nil {
		return err
	}
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, connectReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("unknow proxy response status code %d", resp.StatusCode)
	}
	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
