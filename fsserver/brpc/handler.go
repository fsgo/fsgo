// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/17

package brpc

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
)

type Handler interface {
	Handle()
}

type SimpleHandler interface {
	Handler
	HandleSimple(ctx context.Context, msg *Message) *Message
}

type StreamHandler interface {
	Handler
	HandleStream(ctx context.Context, msg *Message, rw ReadWriter) error
}

type ReadWriter interface {
	Next(ctx context.Context) (*Message, error)
	Write(ctx context.Context, msg *Message) error
}

func newReadWriter(conn net.Conn, r *Reader) *readWriter {
	return &readWriter{
		r:    r,
		rd:   bufio.NewReader(conn),
		w:    conn,
		conn: conn,
	}
}

var _ ReadWriter = (*readWriter)(nil)

type readWriter struct {
	r    *Reader
	rd   io.Reader
	w    io.Writer
	conn net.Conn
}

func (rw *readWriter) Next(ctx context.Context) (*Message, error) {
	_, msg, err := rw.r.ReadPackage(rw.rd)
	return msg, err
}

func (rw *readWriter) Write(ctx context.Context, msg *Message) error {
	_, err := msg.WroteTo(rw.w)
	return err
}

func invokeHandler(ctx context.Context, msg *Message, rw ReadWriter, h Handler) error {
	switch hv := h.(type) {
	case SimpleHandler:
		resp := hv.HandleSimple(ctx, msg)
		return rw.Write(ctx, resp)
	case StreamHandler:
		return hv.HandleStream(ctx, msg, rw)
		// todo
	}
	return errors.New("not support")
}
