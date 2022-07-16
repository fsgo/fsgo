// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/25

package fsserver

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

type Server interface {
	Serve(l net.Listener) error
}

// CanShutdown 支持优雅关闭
type CanShutdown interface {
	Shutdown(ctx context.Context) error
}

type CanShutdownServer interface {
	Server
	CanShutdown
}

var _ CanShutdownServer = (*AnyServer)(nil)

type AnyServer struct {
	BeforeAccept func(l net.Listener) error

	OnConn func(ctx context.Context, conn net.Conn, err error) (context.Context, net.Conn, error)

	Handler func(ctx context.Context, conn net.Conn)

	status      int64
	closeCancel context.CancelFunc
	serverExit  chan bool
}

const (
	statusInit    int64 = 0
	statusRunning int64 = 1
	statusClosed  int64 = 2
)

var ErrShutdown = errors.New("server shutdown")

type temporary interface {
	Temporary() bool
}

func (as *AnyServer) Serve(l net.Listener) error {
	atomic.StoreInt64(&as.status, statusRunning)
	ctx, cancel := context.WithCancel(context.Background())
	as.closeCancel = cancel
	as.serverExit = make(chan bool, 1)

	var errResult error
	var wg sync.WaitGroup
	for {
		if atomic.LoadInt64(&as.status) == statusClosed {
			errResult = ErrShutdown
			break
		}

		if as.BeforeAccept != nil {
			if err := as.BeforeAccept(l); err != nil {
				return err
			}
		}

		conn, err := l.Accept()
		ctxConn := ctx
		if as.OnConn != nil {
			ctxConn, conn, err = as.OnConn(ctxConn, conn, err)
		}

		if err != nil {
			var ne temporary
			if errors.As(err, ne) && ne.Temporary() {
				continue
			}

			if strings.Contains(err.Error(), "i/o timeout") {
				continue
			}
			errResult = err
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			as.handleConn(ctxConn, conn)
		}()
	}
	wg.Wait()
	as.serverExit <- true
	atomic.StoreInt64(&as.status, statusClosed)
	return errResult
}

func (as *AnyServer) handleConn(ctx context.Context, conn net.Conn) {
	ctx = ContextWithConn(ctx, conn)
	as.Handler(ctx, conn)
}

func (as *AnyServer) Shutdown(ctx context.Context) error {
	switch atomic.LoadInt64(&as.status) {
	case statusClosed:
		return nil
	case statusInit:
		return errors.New("server not started")
	}
	atomic.StoreInt64(&as.status, statusClosed)
	select {
	case <-ctx.Done():
	case <-as.serverExit:
	}
	as.closeCancel()
	return nil
}
