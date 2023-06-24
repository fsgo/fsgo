// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/fsgo/fsgo/fsserver"
	"github.com/fsgo/fsgo/fssync"
)

type Server struct {
	ser      *fsserver.AnyServer
	initOnce sync.Once

	// Router 路由
	Router RouterFinder

	// OnConn 创建新链接后的回调,可选
	OnConn func(ctx context.Context, conn net.Conn, err error) (context.Context, net.Conn, error)

	OnError func(ctx context.Context, conn net.Conn, err error)
}

func (s *Server) init() {
	s.ser = &fsserver.AnyServer{
		Handler: fsserver.HandleFunc(s.handle),
		OnConn:  s.OnConn,
	}
}

func (s *Server) Serve(l net.Listener) error {
	s.initOnce.Do(s.init)
	return s.ser.Serve(l)
}

func (s *Server) callOnError(ctx context.Context, conn net.Conn, err error) {
	if s.OnError != nil {
		s.OnError(ctx, conn, err)
		return
	}
	log.Printf("Handler error, remote=%s err=%s", conn.RemoteAddr(), err.Error())
}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	connBR := bufio.NewReader(conn)

	err1 := ReadProtocol(connBR)
	if err1 != nil {
		s.callOnError(ctx, conn, err1)
		return
	}

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(ErrCanceledByDefer)

	writeQueue := newBufferQueue(1024)
	defer writeQueue.Close()

	go func() {
		writeQueue.startWrite(conn)
	}()

	rw := newResponseWriter(writeQueue)

	requests := &fssync.Map[uint64, *reqReader]{}

	for i := uint64(0); ; i++ {
		if err3 := s.readOnePackage(ctx, connBR, requests, rw); err3 != nil {
			s.callOnError(ctx, conn, err3)
			return
		}
	}
}

func (s *Server) readOnePackage(ctx context.Context, rd io.Reader, requests *fssync.Map[uint64, *reqReader], rw *respWriter) error {
	header, err1 := ReadHeader(rd)
	if err1 != nil {
		return err1
	}

	switch header.Type {
	default:
		return fmt.Errorf("%w, got=%d", ErrInvalidHeaderType, header.Type)
	case HeaderTypeRequest:
		req, err2 := readMessage(rd, int(header.Length), &Request{})
		if err2 != nil {
			return err2
		}
		method := req.GetMethod()
		hd := s.Router.HandlerFunc(method)
		if hd == nil {
			return fmt.Errorf("%w: %q", ErrMethodNotFound, method)
		}
		id := req.GetID()
		requestReader := newReqReader(req)
		requests.Store(id, requestReader)
		go func() {
			defer requests.Delete(id)
			hd(ctx, requestReader, rw)
		}()
	case HeaderTypePayload:
		pl := readPayload(rd, int(header.Length))
		if pl.Err != nil {
			return pl.Err
		}
		rid := pl.Meta.GetRID()
		requestReader, ok := requests.Load(rid)
		if !ok {
			return fmt.Errorf("request id not found:%d", rid)
		}
		requestReader.sendPayload(pl)
		if !pl.Meta.More {
			requestReader.Close()
		}
	}
	return nil
}

type RouterFinder interface {
	HandlerFunc(method string) HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		handler: map[string]HandlerFunc{},
	}
}

type Router struct {
	handler map[string]HandlerFunc
}

func (rt *Router) Register(method string, h HandlerFunc) {
	if _, has := rt.handler[method]; has {
		panic(fmt.Sprintf("cannot register handler %q twice", method))
	}
	rt.handler[method] = func(ctx context.Context, req RequestReader, w ResponseWriter) {
		defer func() {
			if re := recover(); re != nil {
				log.Println("panic:", re)
			}
		}()
		h(ctx, req, w)
	}
}

func (rt *Router) HandlerFunc(method string) HandlerFunc {
	return rt.handler[method]
}

type HandlerFunc func(ctx context.Context, rr RequestReader, rw ResponseWriter)
