// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/25

package fsconn

import (
	"net"
	"sync"
	"time"
)

// Listener 支持自定义回调的 Listener
type Listener struct {
	net.Listener

	AfterAccepts []func(conn net.Conn) (net.Conn, error)

	AcceptTimeout time.Duration

	once                sync.Once
	raw                 net.Listener
	catSetAcceptTimeout bool
}

type hasRawListener interface {
	RawListener() net.Listener
}

// RawListener 返回当前对象的底层 Listener
func (l *Listener) RawListener() net.Listener {
	return l.Listener
}

type canSetDeadline interface {
	SetDeadline(t time.Time) error
}

// Accept 获得一个连接
func (l *Listener) Accept() (net.Conn, error) {
	l.once.Do(func() {
		l.raw = RawListener(l.Listener)
		_, l.catSetAcceptTimeout = l.raw.(canSetDeadline)
	})

	if l.AcceptTimeout > 0 && l.catSetAcceptTimeout {
		l.raw.(canSetDeadline).SetDeadline(time.Now().Add(l.AcceptTimeout))
	}

	conn, err := l.Listener.Accept()
	if len(l.AfterAccepts) == 0 || err != nil {
		return conn, err
	}
	for i := 0; i < len(l.AfterAccepts); i++ {
		conn, err = l.AfterAccepts[i](conn)
		if err != nil {
			return conn, err
		}
	}
	return conn, err
}

// RawListener 返回原始连接
func RawListener(l net.Listener) net.Listener {
	for {
		if hl, ok := l.(hasRawListener); ok {
			l = hl.RawListener()
		} else {
			return l
		}
	}
}
