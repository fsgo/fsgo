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
func PingSender(ctx context.Context, rw RequestProtoWriter, method string) (ret error) {
	log.Println("PingSender start")
	defer func() {
		log.Println("PingSender exit:", ret)
	}()

	tk := time.NewTicker(time.Second)
	defer tk.Stop()

	ping := &PingPong{
		Message: "ping",
	}

	id := uint64(0)
	send := func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		req := NewRequest(method)
		ping.ID = id
		rr, err := rw.Write(ctx, req, ping)
		if err != nil {
			return err
		}
		resp, pong, err := ReadProtoResponse(ctx, rr, &PingPong{})
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

	for {
		err := send()
		log.Println("ping:", err)
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tk.C:
		}
		id++
	}
}

func PingReceiver(ctx context.Context, rr RequestReader, rw ResponseWriter) (ret error) {
	log.Println("PingReceiver start")
	defer func() {
		log.Println("PingReceiver exit:", ret)
	}()

	pong := &PingPong{
		Message: "pong",
	}

	for {
		req, ping, err := ReadProtoRequest(ctx, rr, &PingPong{})
		if err != nil {
			return err
		}
		pong.ID = ping.ID

		resp := NewResponseSuccess(req.GetID())
		log.Println("ping resp:", resp)

		if err = rw.Write(ctx, resp, pong); err != nil {
			return err
		}
	}
}
