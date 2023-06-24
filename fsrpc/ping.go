// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/24

package fsrpc

import (
	"context"
	"fmt"
	"log"
	"time"
)

// PingSender 发送 ping 消息并校验响应 pong 是否正确
func PingSender(method string) ClientHandlerFunc {
	return func(ctx context.Context, rw RequestWriter) (ret error) {
		id := uint64(0)
		defer func() {
			log.Println("Ping exit:", ret)
		}()

		tk := time.NewTicker(time.Second)
		defer tk.Stop()

		ping := &PingPong{
			Message: "ping",
		}

		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			req := NewRequest(method)
			ping.ID = id
			rr, err := QuickWriteRequest(ctx, rw, req, ping)
			if err != nil {
				return err
			}

			resp, pong, err := QuickReadResponse(rr, &PingPong{})
			log.Println("QuickReadResponse", err)
			if err != nil {
				return err
			}
			if resp.GetCode() != ErrCode_Success {
				return fmt.Errorf("%w, got=%d", ErrInvalidCode, resp.GetCode())
			}
			if pong.GetID() != id {
				return fmt.Errorf("invalid Pong.ID=%d, want=%d", pong.GetID(), id)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-tk.C:
			}

			id++
		}
	}
}

func PingReceiver(ctx context.Context, rr RequestReader, rw ResponseWriter) (ret error) {
	defer func() {
		log.Println("Ping exit:", ret)
	}()

	pong := &PingPong{
		Message: "pong",
	}

	for {
		req, ping, err := QuickReadRequest(rr, &PingPong{})
		log.Println("QuickReadRequest", err, ping)
		if err != nil {
			return err
		}
		pong.ID = ping.ID

		resp := NewResponseSuccess(req.GetID())
		if err = QuickWriteResponse(ctx, rw, resp, pong); err != nil {
			return err
		}
	}
}
