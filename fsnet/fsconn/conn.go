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
	return c.allIts.CallRead(c.raw, b, c.raw.Read, 0)
}

func (c *connWithIt) Write(b []byte) (n int, err error) {
	return c.allIts.CallWrite(c.raw, b, c.raw.Write, 0)
}

func (c *connWithIt) Close() error {
	return c.allIts.CallClose(c.raw, c.raw.Close, 0)
}

func (c *connWithIt) LocalAddr() net.Addr {
	return c.allIts.CallLocalAddr(c.raw, c.raw.LocalAddr, 0)
}

func (c *connWithIt) RemoteAddr() net.Addr {
	return c.allIts.CallRemoteAddr(c.raw, c.raw.RemoteAddr, 0)
}

func (c *connWithIt) SetDeadline(t time.Time) error {
	return c.allIts.CallSetDeadline(c.raw, t, c.raw.SetDeadline, 0)
}

func (c *connWithIt) SetReadDeadline(t time.Time) error {
	return c.allIts.CallSetReadDeadline(c.raw, t, c.raw.SetReadDeadline, 0)
}

func (c *connWithIt) SetWriteDeadline(t time.Time) error {
	return c.allIts.CallSetWriteDeadline(c.raw, t, c.raw.SetWriteDeadline, 0)
}

// HasService 用于判断是否有服务名
type HasService interface {
	// Service 服务名
	Service() any
}

func WithService(service any, c net.Conn) net.Conn {
	if ws, ok := c.(HasService); ok && ws.Service() == service {
		return c
	}
	return &withService{
		service: service,
		Conn:    c,
	}
}

var _ HasRaw = (*withService)(nil)

type withService struct {
	service any
	net.Conn
}

func (c *withService) RawConn() net.Conn {
	return c.Conn
}

func (c *withService) Service() any {
	return c.service
}

// Service 读取连接的 service 属性
func Service(c net.Conn) any {
	for {
		if c == nil {
			return nil
		}
		if ws, ok := c.(HasService); ok {
			return ws.Service()
		}
		if c1, ok := c.(HasRaw); ok {
			c = c1.RawConn()
		} else {
			return nil
		}
	}
}
