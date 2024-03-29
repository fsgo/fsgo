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

	"github.com/fsgo/fsgo/fsnet/fsdialer"
)

// DefaultUserAgent default user agent
var DefaultUserAgent = "fsgo/1.0"

var _ http.RoundTripper = (*Transport)(nil)

// Transport  一个简单的 Transport
//
// 不管理连接,可以使用外部连接池
type Transport struct {
	// Proxy 可选, 当前已支持 http_proxy、https_proxy
	// 若 Proxy 为 nil,将使用 http.ProxyFromEnvironment
	Proxy func(*http.Request) (*url.URL, error)

	// DialContext 可选，拨号
	DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

	// DialTLSContext 可选，TLS 拨号
	DialTLSContext func(ctx context.Context, network, addr string) (net.Conn, error)

	// TLSClientConfig 可选
	TLSClientConfig *tls.Config

	// WriteTimeout 可选，设置写超时时间
	WriteTimeout func(req *http.Request, conn net.Conn) error

	// ReadTimeout 可选，设置读超时时间
	ReadTimeout func(req *http.Request, conn net.Conn) error
}

func (t *Transport) dialConn(ctx context.Context, req *http.Request, proxyURL *url.URL, network, addr string) (net.Conn, error) {
	// has proxy
	if proxyURL != nil && (proxyURL.Scheme == "http" || proxyURL.Scheme == "https") {
		return t.dial(ctx, network, addr)
	}

	// no proxy
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
	return fsdialer.DialContext(ctx, network, addr)
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
	if len(hostName) == 0 {
		hostName, _, _ = net.SplitHostPort(addr)
	} else if strings.Contains(hostName, ":") {
		hostName, _, _ = net.SplitHostPort(hostName)
	}

	return t.connAddTLS(conn, hostName), nil
}

func (t *Transport) connAddTLS(conn net.Conn, hostName string) net.Conn {
	cfg := cloneTLSConfig(t.TLSClientConfig)

	if len(cfg.ServerName) == 0 {
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

func (t *Transport) getProxy() func(*http.Request) (*url.URL, error) {
	if t.Proxy != nil {
		return t.Proxy
	}
	return http.ProxyFromEnvironment
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
	if len(proxyAuthVal) != 0 && req.URL.Scheme == "http" {
		req.Header.Set(authKey, proxyAuthVal)
	}
	return address, proxyURL, nil
}

// RoundTrip 发送请求
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	defer closeRequest(req)
	if err = checkRequest(req); err != nil {
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

	defer func() {
		// 只能在异常的时候关闭连接
		// 正常读取到 response 的时候，body 是一个 stream，若关闭将导致 body 读取不出来
		if err != nil {
			_ = conn.Close()
		}
	}()

	if dl, ok := req.Context().Deadline(); ok {
		if err = conn.SetDeadline(dl); err != nil {
			return nil, err
		}
	}

	if conn, err = t.withProxy(proxyURL, conn, req); err != nil {
		return nil, err
	}

	if err = t.setWriteTimeout(req, conn); err != nil {
		return nil, err
	}

	if proxyURL != nil {
		err = req.WriteProxy(conn)
	} else {
		err = req.Write(conn)
	}
	if err != nil {
		return nil, err
	}

	if err = t.setReadTimeout(req, conn); err != nil {
		return nil, err
	}

	bio := bufio.NewReader(conn)
	resp, err = http.ReadResponse(bio, req)
	return resp, err
}

func (t *Transport) withProxy(proxyURL *url.URL, conn net.Conn, req *http.Request) (net.Conn, error) {
	if proxyURL == nil || req.URL.Scheme != "https" {
		return conn, nil
	}
	if err := t.proxyHTTPSConn(conn, req, proxyURL); err != nil {
		return nil, err
	}
	conn = t.connAddTLS(conn, req.URL.Hostname())
	// http_proxy and https_proxy are work fine
	// todo socks 待实现
	return conn, nil
}

func (t *Transport) setWriteTimeout(req *http.Request, conn net.Conn) error {
	if t.WriteTimeout == nil {
		return nil
	}
	return t.WriteTimeout(req, conn)
}

func (t *Transport) setReadTimeout(req *http.Request, conn net.Conn) error {
	if t.ReadTimeout == nil {
		return nil
	}
	return t.ReadTimeout(req, conn)
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
func (t *Transport) proxyHTTPSConn(conn net.Conn, req *http.Request, proxy *url.URL) error {
	proxyAuthVal := proxyAuth(proxy)
	hdr := make(http.Header)
	if len(proxyAuthVal) != 0 {
		hdr.Set(authKey, proxyAuthVal)
	}
	address := canonicalAddr(req.URL)
	hdr.Set("User-Agent", req.UserAgent())
	connectReq := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: address},
		Host:   address,
		Header: hdr,
	}
	err := connectReq.Write(conn)
	if err != nil {
		return fmt.Errorf("connect to https_proxy(%q) failed: %w", proxy.String(), err)
	}
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, connectReq)
	if err != nil {
		return fmt.Errorf("read connect response from https_proxy(%q) failed: %w", proxy.String(), err)
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
