// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	"github.com/fsgo/fsgo/fsrpc"
)

var serverAddr = flag.String("addr", "127.0.0.1:8128", "")

func main() {
	flag.Parse()
	conn, err := net.DialTimeout("tcp", *serverAddr, time.Second)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	client := fsrpc.NewClientConn(conn)
	client.SetBeforeReadLoop(func() {
		conn.SetReadDeadline(time.Now().Add(time.Second))
	})

	err = client.Open(context.Background(), func(ctx context.Context, rw fsrpc.RequestWriter) error {
		for i := 0; i < 10; i++ {
			log.Println("i=", i)
			req := fsrpc.NewRequest("hello")
			req.HasPayload = true
			req.LogID = time.Now().String()
			pw, rr, err2 := rw.WriteRequest(req)
			log.Println("WriteRequest=", err2, rr)

			err3 := pw.WritePayload([]byte("hello"), false)
			log.Println("WritePayload=", err3)

			resp := rr.Response()
			log.Println("resp:", resp.String())
		}
		return nil
	})
	log.Println("Open.Err=", err)
	log.Println("lastErr=", client.LastError())
	log.Println("Close()", client.Close())
}
