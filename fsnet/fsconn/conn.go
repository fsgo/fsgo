// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/14

package fsconn

import (
	"context"
	"net"
	"time"
)

// NewWithContext 取出 ctx 里的 Interceptor， 并对 conn 进行封装
func NewWithContext(ctx context.Context, conn net.Conn) net.Conn {
	cks := InterceptorsFromContext(ctx)
	if len(cks) == 0 {
		return conn
	}
	return WithInterceptor(conn, cks...)
}

var _ net.Conn = (*connWithIt)(nil)

// WithInterceptor wrap connWithIt with Interceptor
// its 将倒序执行：后注册的先执行
func WithInterceptor(c net.Conn, its ...*Interceptor) net.Conn {
	if rc, ok := c.(*connWithIt); ok {
		nc := &connWithIt{
			raw:  rc.raw,
			args: append(rc.args, its...),
		}
		nc.allIts = append(globalConnIts, nc.args...)
		return nc
	}

	nc := &connWithIt{
		raw:    c,
		allIts: append(globalConnIts, its...),
		args:   its,
	}
	return nc
}

// HasRaw 有原始的 net.Conn
type HasRaw interface {
	RawConn() net.Conn
}

// OriginConn 获取最底层的 net.Conn
func OriginConn(conn net.Conn) net.Conn {
	for {
		c, ok := conn.(HasRaw)
		if ok {
			conn = c.RawConn()
		} else {
			return conn
		}
	}
}

var _ net.Conn = (*connWithIt)(nil)
var _ HasRaw = (*connWithIt)(nil)

type connWithIt struct {
	raw net.Conn
	// 包好了全局和创建时传入的拦截器
	allIts interceptors

	// 创建时传入的拦截器
	args interceptors
}

func (c *connWithIt) RawConn() net.Conn {
	return c.raw
}

func (c *connWithIt) Read(b []byte) (n int, err error) {
	return c.allIts.CallRead(b, c.raw.Read, 0)
}

func (c *connWithIt) Write(b []byte) (n int, err error) {
	return c.allIts.CallWrite(b, c.raw.Write, 0)
}

func (c *connWithIt) Close() error {
	return c.allIts.CallClose(c.raw.Close, 0)
}

func (c *connWithIt) LocalAddr() net.Addr {
	return c.allIts.CallLocalAddr(c.raw.LocalAddr, 0)
}

func (c *connWithIt) RemoteAddr() net.Addr {
	return c.allIts.CallRemoteAddr(c.raw.RemoteAddr, 0)
}

func (c *connWithIt) SetDeadline(t time.Time) error {
	return c.allIts.CallSetDeadline(t, c.raw.SetDeadline, 0)
}

func (c *connWithIt) SetReadDeadline(t time.Time) error {
	return c.allIts.CallSetReadDeadline(t, c.raw.SetReadDeadline, 0)
}

func (c *connWithIt) SetWriteDeadline(t time.Time) error {
	return c.allIts.CallSetWriteDeadline(t, c.raw.SetWriteDeadline, 0)
}
