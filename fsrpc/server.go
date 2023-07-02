// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
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
	// connReader := bufio.NewReader(conn)
	connReader := conn

	err1 := ReadProtocol(connReader)
	if err1 != nil {
		s.callOnError(ctx, conn, err1)
		return
	}

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(ErrCanceledByDefer)

	writeQueue := newBufferQueue(1024)
	defer writeQueue.CloseWithErr(ErrCanceledByDefer)

	go func() {
		writeQueue.startWrite(conn)
	}()

	rw := newResponseWriter(writeQueue)
	hp := &handlerParam{}

	for {
		err3 := s.readOnePackage(ctx, connReader, rw, hp)
		if err3 != nil {
			s.callOnError(ctx, conn, err3)
			return
		}
	}
}

type handlerParam struct {
	Handlers fssync.Map[string, *requestReader]
	Requests fssync.Map[uint64, *requestReader]
	Payloads fssync.Map[uint64, payloadChan]
}

func (s *Server) readOnePackage(ctx context.Context, rd io.Reader, rw *respWriter, hp *handlerParam) error {
	header, err1 := ReadHeader(rd)
	if err1 != nil {
		return fmt.Errorf("read Header: %w", err1)
	}

	switch header.Type {
	default:
		return fmt.Errorf("%w, got=%d", ErrInvalidHeader, header.Type)
	case HeaderTypeRequest:
		req, err2 := readMessage(rd, int(header.Length), &Request{})
		if err2 != nil {
			return fmt.Errorf("read Request: %w", err2)
		}
		method := req.GetMethod()
		handler := s.Router.HandlerFunc(method)
		if handler == nil {
			return fmt.Errorf("%w: %q", ErrMethodNotFound, method)
		}

		hr, ok := hp.Handlers.Load(method)
		if !ok {
			reader := newRequestReader()
			reader.requests <- req
			if !req.GetHasPayload() {
				reader.payloads <- emptyPayloadChan
			} else {
				hp.Requests.Store(req.GetID(), reader)
				plc := make(payloadChan, 1)
				reader.payloads <- plc
				hp.Payloads.Store(req.GetID(), plc)
			}
			hp.Handlers.Store(method, reader)
			go func() {
				defer hp.Handlers.Delete(method)
				handler(ctx, reader, rw)
			}()
		} else {
			hr.requests <- req
			if !req.GetHasPayload() {
				hr.payloads <- emptyPayloadChan
			} else {
				plc := make(payloadChan, 1)
				hr.payloads <- plc
				hp.Payloads.Store(req.GetID(), plc)
			}
		}
	case HeaderTypePayload:
		pl, err := readPayload(rd, int(header.Length))
		if err != nil {
			return fmt.Errorf("read Payload: %w", err)
		}
		rid := pl.Meta.GetRID()
		plc, ok := hp.Payloads.Load(rid)
		if !ok {
			return fmt.Errorf("request not found: %d", rid)
		}
		plc <- pl
		if !pl.Meta.More {
			close(plc)
			hp.Requests.Delete(rid)
			hp.Payloads.Delete(rid)
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
	rt.handler[method] = func(ctx context.Context, req RequestReader, w ResponseWriter) error {
		defer func() {
			if re := recover(); re != nil {
				log.Println("panic:", re)
			}
		}()
		return h(ctx, req, w)
	}
}

func (rt *Router) HandlerFunc(method string) HandlerFunc {
	return rt.handler[method]
}

type HandlerFunc func(ctx context.Context, rr RequestReader, rw ResponseWriter) error
