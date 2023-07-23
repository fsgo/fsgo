// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/fsgo/fsgo/fsserver"
	"github.com/fsgo/fsgo/fssync"
	"github.com/fsgo/fsgo/fssync/fsatomic"
)

type Server struct {
	ser      *fsserver.AnyServer
	initOnce sync.Once

	// Router 路由
	Router RouteFinder

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
	log.Printf("Handler closedErr, remote=%s err=%s", conn.RemoteAddr(), err.Error())
}

var errCanceledByDefer = errors.New("canceled by Server.handle defer")

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	connReader := conn

	session := &ConnSession{
		RemoteAddr: conn.RemoteAddr(),
		LocalAddr:  conn.LocalAddr(),
	}
	ctx = ctxWithServerConnSession(ctx, session)

	err1 := ReadProtocol(connReader)
	if err1 != nil {
		s.callOnError(ctx, conn, err1)
		return
	}

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errCanceledByDefer)

	writeQueue := newBufferQueue(1024)
	defer writeQueue.CloseWithErr(errCanceledByDefer)

	go func() {
		err := writeQueue.startWrite(conn)
		cancel(err)
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
	Handlers fssync.Map[string, *reqReader]
	Requests fssync.Map[uint64, *reqReader]
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
		req, err2 := readProtoMessage(rd, int(header.Length), &Request{})
		if err2 != nil {
			return fmt.Errorf("read Request: %w", err2)
		}
		method := req.GetMethod()
		handler := s.Router.Handler(method)
		if handler == nil {
			handler = s.Router.NotFound()
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
				ctx = ctxWithServerMethod(ctx, method)
				_ = handler.Handle(ctx, reader, rw)
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

type (
	RouteFinder interface {
		Handler(method string) Handler
		NotFound() Handler
	}

	RouteRegister interface {
		Register(method string, h Handler)
	}
)

func NewRouter() *Router {
	return &Router{
		handlers: map[string]Handler{},
	}
}

var _ RouteFinder = (*Router)(nil)
var _ RouteRegister = (*Router)(nil)

var _ Handler = (*NotFoundHandler)(nil)

type NotFoundHandler struct{}

func (nh *NotFoundHandler) Handle(ctx context.Context, rr RequestReader, rw ResponseWriter) error {
	req, _ := rr.Request()
	resp := NewResponse(req.GetID(), ErrCode_NoMethod, fmt.Sprintf("method %q not found", req.GetMethod()))
	_ = WriteResponseProto(ctx, rw, resp)
	return ErrMethodNotFound
}

type Router struct {
	handlers        map[string]Handler
	notFoundHandler Handler
}

var defaultNotFound = &NotFoundHandler{}

func (rt *Router) NotFound() Handler {
	if rt.notFoundHandler != nil {
		return rt.notFoundHandler
	}
	return defaultNotFound
}

func (rt *Router) SetNotFound(h Handler) {
	rt.notFoundHandler = h
}

func (rt *Router) Register(method string, h Handler) {
	if _, has := rt.handlers[method]; has {
		panic(fmt.Sprintf("cannot register handler %q twice", method))
	}
	rt.handlers[method] = h
}

func (rt *Router) Handler(method string) Handler {
	return rt.handlers[method]
}

type Handler interface {
	Handle(ctx context.Context, rr RequestReader, rw ResponseWriter) error
}

type HandlerFunc func(ctx context.Context, rr RequestReader, rw ResponseWriter) error

func (h HandlerFunc) Handle(ctx context.Context, rr RequestReader, rw ResponseWriter) error {
	return h(ctx, rr, rw)
}

// ConnSession server 连接的信息
type ConnSession struct {
	LoggedIn   atomic.Bool
	User       fsatomic.String
	RemoteAddr net.Addr
	LocalAddr  net.Addr
	Data       sync.Map
}

var _ Handler = (*Interceptor)(nil)

type Interceptor struct {
	// Name  名称，可选
	Name string

	// Before 在 Handler 前执行，可选
	// 若 返回的 closedErr != nil,则 Handler 不会执行
	Before func(ctx context.Context, rr RequestReader, rw ResponseWriter) (context.Context, RequestReader, ResponseWriter, error)

	// After 在 Handler 后执行，可选
	After func(ctx context.Context, rr RequestReader, rw ResponseWriter, err error) error

	// Handler 业务逻辑 Handler，必填
	Handler Handler
}

func (it *Interceptor) Handle(ctx context.Context, rr RequestReader, rw ResponseWriter) (err error) {
	if it.After != nil {
		defer func() {
			err = it.After(ctx, rr, rw, err)
		}()
	}
	if it.Before != nil {
		ctx, rr, rw, err = it.Before(ctx, rr, rw)
		if err != nil {
			return err
		}
	}
	return it.Handler.Handle(ctx, rr, rw)
}

func ListenAndServe(addr string, router RouteFinder) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	ser := &Server{
		Router: router,
	}
	return ser.Serve(l)
}
