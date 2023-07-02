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
	return Wrap(conn, cks...)
}

var _ net.Conn = (*connWithIt)(nil)

// Wrap  conn WithIt with Interceptor
// its 将倒序执行：后注册的先执行
func Wrap(c net.Conn, its ...*Interceptor) net.Conn {
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
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].Read != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		n, err = c.raw.Read(b)
	} else {
		n, err = c.allIts.CallRead(c.raw, b, c.raw.Read, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterRead != nil {
			c.allIts[i].AfterRead(c.raw, b, n, err)
		}
	}
	return n, err
}

func (c *connWithIt) Write(b []byte) (n int, err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].Write != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		n, err = c.raw.Write(b)
	} else {
		n, err = c.allIts.CallWrite(c.raw, b, c.raw.Write, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterWrite != nil {
			c.allIts[i].AfterWrite(c.raw, b, n, err)
		}
	}
	return n, err
}

func (c *connWithIt) Close() (err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].Close != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		err = c.raw.Close()
	} else {
		err = c.allIts.CallClose(c.raw, c.raw.Close, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterClose != nil {
			c.allIts[i].AfterClose(c.raw, err)
		}
	}
	return err
}

func (c *connWithIt) LocalAddr() net.Addr {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].LocalAddr != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		return c.raw.LocalAddr()
	}
	return c.allIts.CallLocalAddr(c.raw, c.raw.LocalAddr, idx)
}

func (c *connWithIt) RemoteAddr() net.Addr {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].RemoteAddr != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		return c.raw.RemoteAddr()
	}
	return c.allIts.CallRemoteAddr(c.raw, c.raw.RemoteAddr, idx)
}

func (c *connWithIt) SetDeadline(t time.Time) (err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].SetDeadline != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		err = c.raw.SetDeadline(t)
	} else {
		err = c.allIts.CallSetDeadline(c.raw, t, c.raw.SetDeadline, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterSetDeadline != nil {
			c.allIts[i].AfterSetDeadline(c.raw, t, err)
		}
	}
	return err
}

func (c *connWithIt) SetReadDeadline(t time.Time) (err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].SetReadDeadline != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		err = c.raw.SetReadDeadline(t)
	} else {
		err = c.allIts.CallSetReadDeadline(c.raw, t, c.raw.SetReadDeadline, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterSetReadDeadline != nil {
			c.allIts[i].AfterSetReadDeadline(c.raw, t, err)
		}
	}
	return err
}

func (c *connWithIt) SetWriteDeadline(t time.Time) (err error) {
	idx := -1
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].SetWriteDeadline != nil {
			idx = i
			break
		}
	}
	if idx == -1 {
		err = c.raw.SetWriteDeadline(t)
	} else {
		err = c.allIts.CallSetWriteDeadline(c.raw, t, c.raw.SetWriteDeadline, idx)
	}
	for i := 0; i < len(c.allIts); i++ {
		if c.allIts[i].AfterSetWriteDeadline != nil {
			c.allIts[i].AfterSetWriteDeadline(c.raw, t, err)
		}
	}
	return err
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
func Service(c Info) any {
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
