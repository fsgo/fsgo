// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/24

package fsrpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

type PingHandler struct {
	Method string
	id     atomic.Uint64
}

func (pp *PingHandler) getMethod() string {
	if pp.Method != "" {
		return pp.Method
	}
	return "sys_ping"
}

func (pp *PingHandler) Send(ctx context.Context, w RequestWriter) (ret error) {
	id := pp.id.Add(1)
	defer func() {
		log.Println("Ping, id=", id, ret)
	}()
	data := &Echo{
		Message: "ping",
		ID:      id,
	}
	req := NewRequest(pp.getMethod())
	rr, err := WriteRequestProto(ctx, w, req, data)
	if err != nil {
		return err
	}
	resp, pong, err := ReadResponseProto(ctx, rr, &Echo{})
	if err != nil {
		return err
	}
	if resp.GetCode() != ErrCode_Success {
		return fmt.Errorf("%w, got=%d", ErrInvalidCode, resp.GetCode())
	}
	if pong.GetID() != id {
		return fmt.Errorf("invalid Pong.ID=%d, want=%d", pong.GetID(), id)
	}
	return nil
}

func (pp *PingHandler) SendMany(ctx context.Context, w RequestWriter, interval time.Duration) error {
	tk := time.NewTimer(0)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tk.C:
		}
		err := pp.Send(ctx, w)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}
		tk.Reset(interval)
	}
	return nil
}

func (pp *PingHandler) RegisterTo(rt RouteRegister) {
	rt.Register(pp.getMethod(), HandlerFunc(pp.Receiver))
}

func (pp *PingHandler) Receiver(ctx context.Context, r RequestReader, w ResponseWriter) (ret error) {
	pong := &Echo{
		Message: "pong",
	}

	for {
		req, ping, err := ReadRequestProto(ctx, r, &Echo{})
		if err != nil {
			return err
		}
		pong.ID = ping.ID

		resp := NewResponseSuccess(req.GetID())
		log.Println("ping resp:", resp)

		if err = WriteResponseProto(ctx, w, resp, pong); err != nil {
			return err
		}
	}
}
