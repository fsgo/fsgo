// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/17

package brpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/fsgo/fsgo/fsserver"
)

var _ fsserver.Handler = (*Router)(nil)

type Router struct {
	handlers map[string]Handler
	Reader   *Reader
	once     sync.Once
}

func (r *Router) initOnce() {
	r.once.Do(func() {
		r.handlers = make(map[string]Handler, 10)
		if r.Reader == nil {
			r.Reader = &Reader{}
		}
	})
}

func (r *Router) Handle(ctx context.Context, conn net.Conn) {
	r.initOnce()
	rw := newReadWriter(conn, r.Reader)
	for {
		msg, err := rw.Next(ctx)
		if err != nil {
			return
		}
		if err = r.invokeMessage(ctx, msg, rw); err != nil {
			return
		}
	}
}

func (r *Router) Register(service, method string, handler Handler) {
	r.initOnce()
	key := serviceMethod(service, method)
	r.handlers[key] = handler
}

var ErrMethodNotFound = errors.New("service or method not found")

func (r *Router) invokeMessage(ctx context.Context, msg *Message, rw ReadWriter) error {
	key := serviceMethod(msg.ServiceName(), msg.MethodName())
	h, ok := r.handlers[key]
	if ok {
		return invokeHandler(ctx, msg, rw, h)
	}
	return fmt.Errorf("%w, service=%q, method=%q", ErrMethodNotFound, msg.ServiceName(), msg.MethodName())
}

func serviceMethod(service, method string) string {
	return service + "/" + method
}
