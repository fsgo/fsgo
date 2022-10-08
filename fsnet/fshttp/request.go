// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/5/15

package fshttp

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"unicode/utf8"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/idna"
)

func closeRequest(req *http.Request) {
	if req.Body == nil {
		return
	}
	req.Body.Close()
}

func checkRequest(req *http.Request) error {
	if req.URL == nil {
		return errors.New("http: nil Request.URL")
	}
	if req.Header == nil {
		return errors.New("http: nil Request.Header")
	}
	scheme := req.URL.Scheme
	isHTTP := scheme == "http" || scheme == "https"
	if isHTTP {
		for k, vv := range req.Header {
			if !httpguts.ValidHeaderFieldName(k) {
				return fmt.Errorf("net/http: invalid header field name %q", k)
			}
			for _, v := range vv {
				if !httpguts.ValidHeaderFieldValue(v) {
					return fmt.Errorf("net/http: invalid header field value %q for key %v", v, k)
				}
			}
		}
	}
	return nil
}

var portMap = map[string]string{
	"http":   "80",
	"https":  "443",
	"socks5": "1080",
}

// canonicalAddr returns url.Host but always with a ":port" suffix
func canonicalAddr(url *url.URL) string {
	addr := url.Hostname()
	if v, err := idnaASCII(addr); err == nil {
		addr = v
	}
	port := url.Port()
	if len(port) == 0 {
		port = portMap[url.Scheme]
	}
	return net.JoinHostPort(addr, port)
}

func idnaASCII(v string) (string, error) {
	if isASCII(v) {
		return v, nil
	}
	return idna.Lookup.ToASCII(v)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
