// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/25

package fsserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsgo/fsgo/fssync"
)

type Server interface {
	Serve(l net.Listener) error
}

// CanShutdown 支持优雅关闭
type CanShutdown interface {
	Shutdown(ctx context.Context) error
}

// GracefulServer 支持优雅关闭的 server
type GracefulServer interface {
	Server
	CanShutdown
}

var _ GracefulServer = (*AnyServer)(nil)

// AnyServer 一个通用的 server
type AnyServer struct {
	// Handler 处理请求的 Handler，必填
	Handler Handler

	// BeforeAccept Accept 之前的回调，可选
	BeforeAccept func(l net.Listener) error

	// OnConn 创建新链接后的回调，可选
	OnConn func(ctx context.Context, conn net.Conn, err error) (context.Context, net.Conn, error)

	closeCancel context.CancelFunc
	serverExit  chan bool

	connections fssync.Map[net.Conn, struct{}]

	status atomic.Int32
}

const (
	statusInit    int32 = iota // server 状态，初始状态
	statusRunning              // 已经调用 Serve 方法，处于运行中
	statusClosed               // 已经调用 Shutdown 方法，server 已经关闭
)

func statusTxt(s int32) string {
	switch s {
	case statusInit:
		return "init"
	case statusRunning:
		return "running"
	case statusClosed:
		return "closed"
	default:
		return "invalid status"
	}
}

var (
	ErrShutdown = errors.New("server shutdown")
)

type temporary interface {
	Temporary() bool
}

func (as *AnyServer) Serve(l net.Listener) error {
	if as.Handler == nil {
		return errors.New("handler is nil")
	}
	if !as.status.CompareAndSwap(statusInit, statusRunning) {
		s := as.status.Load()
		return fmt.Errorf("invalid status (%s) for Serve", statusTxt(s))
	}
	ctx, cancel := context.WithCancel(context.Background())
	as.closeCancel = cancel
	as.serverExit = make(chan bool, 1)

	var errResult error
	var wg sync.WaitGroup

	loopAccept := func() error {
		if as.status.Load() != statusRunning {
			return ErrShutdown
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
			if errors.As(err, &ne) && ne.Temporary() {
				return nil
			}

			if strings.Contains(err.Error(), "i/o timeout") {
				return nil
			}
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			as.handleConn(ctxConn, conn)
		}()
		return nil
	}

	for {
		if errResult = loopAccept(); errResult != nil {
			break
		}
	}

	wg.Wait()
	as.serverExit <- true
	as.status.Store(statusClosed)
	return errResult
}

func (as *AnyServer) handleConn(ctx context.Context, conn net.Conn) {
	as.connections.Store(conn, struct{}{})
	defer as.connections.Delete(conn)

	ctx = ContextWithConn(ctx, conn)
	as.Handler.Handle(ctx, conn)
}

func (as *AnyServer) closeAllConn() {
	as.connections.Range(func(c net.Conn, value struct{}) bool {
		_ = c.Close()
		as.connections.Delete(c)
		return true
	})
}

func (as *AnyServer) Shutdown(ctx context.Context) error {
	switch as.status.Load() {
	case statusClosed,
		statusInit:
		return nil
	}
	if !as.status.CompareAndSwap(statusRunning, statusClosed) {
		s := as.status.Load()
		return fmt.Errorf("invalid status (%s) for Shutdown", statusTxt(s))
	}
	select {
	case <-ctx.Done():
		as.closeAllConn()
	case <-as.serverExit:
	}
	as.closeCancel()
	return nil
}
