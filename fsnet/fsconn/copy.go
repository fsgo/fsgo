// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsconn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsgo/fsgo/fsnet/fsaddr"
)

// Copy 实现对网络连接读写数据的复制
type Copy struct {
	// ReadTo 将 Read 到的数据写入此处,比如 os.Stdout
	ReadTo io.Writer

	// WriterTo 将 writer 的数据写入此处，比如 os.Stdout
	WriterTo io.Writer

	interceptor *Interceptor
	once        sync.Once

	disableRead  atomic.Bool
	disableWrite atomic.Bool
}

func (cc *Copy) init() {
	cc.interceptor = &Interceptor{
		AfterRead: func(_ Info, b []byte, readSize int, err error) {
			if cc.disableRead.Load() {
				return
			}
			if readSize > 0 && cc.ReadTo != nil {
				_, _ = cc.ReadTo.Write(b[:readSize])
			}
		},
		AfterWrite: func(_ Info, b []byte, wroteSize int, err error) {
			if cc.disableWrite.Load() {
				return
			}
			if wroteSize > 0 && cc.ReadTo != nil {
				_, _ = cc.WriterTo.Write(b[:wroteSize])
			}
		},
	}
}

// Interceptor 获取 Interceptor 实例
func (cc *Copy) Interceptor() *Interceptor {
	cc.once.Do(cc.init)
	return cc.interceptor
}

// EnableRead 设置是否允许 copy read 流量
func (cc *Copy) EnableRead(enable bool) {
	cc.disableRead.Store(!enable)
}

// EnableWrite 设置是否允许 copy write 流量
func (cc *Copy) EnableWrite(enable bool) {
	cc.disableWrite.Store(!enable)
}

var _ io.ReadWriteCloser = (*StreamConn)(nil)

// StreamConn 一个总是保持连接的 WriterReader
type StreamConn struct {
	// Addr 要连接的网络地址，必填
	Addr net.Addr

	// Logger 可选，打印日志的 writer
	Logger io.Writer

	conn net.Conn

	// Dial 可选，拨号函数
	Dial func(ctx context.Context, addr net.Addr) (net.Conn, error)

	// Wrap 可选
	Wrap func(conn net.Conn) net.Conn

	// DialTimeout 拨号超时时间，可选，默认为 3s
	DialTimeout time.Duration

	// Retry 可选，重试次数，默认为 0 （不重试）
	// -1 :无限重试
	Retry int

	// RetryWait 可选，重试等待间隔时间
	// 默认值为 1s
	RetryWait time.Duration

	mux sync.Mutex

	// 关闭状态，0-正常，1-已关闭
	closed atomic.Bool
}

var zeroDialer = &net.Dialer{}

var errClosed = errors.New("already closed")

func (sc *StreamConn) checkConn() error {
	if sc.isClosed() {
		return errClosed
	}

	if sc.conn != nil {
		return nil
	}
	var try int
	for {
		err := sc.connect()
		if err == nil {
			return nil
		}
		if sc.canLog() {
			sc.log("connect to ", sc.Addr.String(), " failed, try=", try, " err=", err.Error())
		}
		if !sc.canTry(try) {
			return err
		}
		try++
		time.Sleep(sc.getRetryWait())
	}
}

func (sc *StreamConn) canLog() bool {
	return sc.Logger != nil
}

func (sc *StreamConn) getDialTimeout() time.Duration {
	if sc.DialTimeout > 0 {
		return sc.DialTimeout
	}
	return 3 * time.Second
}

func (sc *StreamConn) getRetryWait() time.Duration {
	if sc.RetryWait > 0 {
		return sc.RetryWait
	}
	return time.Second
}

func (sc *StreamConn) connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), sc.getDialTimeout())
	defer cancel()
	var conn net.Conn
	var err error
	if sc.Dial != nil {
		conn, err = sc.Dial(ctx, sc.Addr)
	} else {
		conn, err = zeroDialer.DialContext(ctx, sc.Addr.Network(), sc.Addr.String())
	}
	if conn != nil {
		if sc.Wrap != nil {
			conn = sc.Wrap(conn)
		}
		sc.conn = conn
	}
	return err
}

func (sc *StreamConn) log(args ...any) {
	if sc.Logger == nil {
		return
	}
	prefix := "[stream_conn] " + time.Now().Format("2006-01-02 15:04:05.9999") + " "
	_, _ = sc.Logger.Write([]byte(prefix + fmt.Sprint(args...) + "\n"))
}

// RemoteAddr 当前连接的远端地址
func (sc *StreamConn) RemoteAddr() net.Addr {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	if sc.conn != nil {
		return sc.conn.RemoteAddr()
	}
	return fsaddr.Empty
}

func (sc *StreamConn) canTry(index int) bool {
	if sc.isClosed() {
		return false
	}
	if sc.Retry == -1 || index < sc.Retry {
		return true
	}
	return false
}

func (sc *StreamConn) Write(b []byte) (int, error) {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	if err := sc.checkConn(); err != nil {
		return 0, err
	}
	var try int
	for {
		n, err := sc.conn.Write(b)
		if err == nil {
			return n, err
		}
		if sc.canLog() {
			sc.log("write to ", sc.conn.RemoteAddr().String(), " failed, try=", try, " err=", err.Error())
		}
		if !sc.canTry(try) {
			return n, err
		}
		var ne net.Error
		if errors.As(err, &ne) && !ne.Timeout() {
			// 可能是连接已经关闭了，所以需要重新连接
			_ = sc.close()
			if err = sc.checkConn(); err != nil {
				return 0, err
			}
		}
		try++
	}
}

func (sc *StreamConn) Read(b []byte) (int, error) {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	if err := sc.checkConn(); err != nil {
		return 0, err
	}

	var try int
	for {
		n, err := sc.conn.Read(b)
		if err == nil {
			return n, err
		}

		if sc.canLog() {
			sc.log("read from ", sc.conn.RemoteAddr().String(), " failed, try=", try, " err=", err.Error(), "\n")
		}

		if !sc.canTry(try) {
			return n, err
		}
		var ne net.Error
		if errors.As(err, &ne) && !ne.Timeout() {
			// 可能是连接已经关闭了，所以需要重新连接
			_ = sc.close()
			if err = sc.checkConn(); err != nil {
				return 0, err
			}
		}
	}
}

func (sc *StreamConn) isClosed() bool {
	return sc.closed.Load()
}

func (sc *StreamConn) Close() error {
	if !sc.closed.CompareAndSwap(false, true) {
		return nil
	}

	sc.mux.Lock()
	defer sc.mux.Unlock()
	return sc.close()
}

func (sc *StreamConn) close() error {
	if sc.conn == nil {
		return nil
	}
	err := sc.conn.Close()
	sc.conn = nil
	return err
}
