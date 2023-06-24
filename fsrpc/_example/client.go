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

var serverAddr = flag.String("addr", "127.0.0.1:8013", "")

func main() {
	flag.Parse()
	conn, err := net.DialTimeout("tcp", *serverAddr, time.Second)
	if err != nil {
		log.Fatalln(err)
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
			rr, err2 := rw.WriteRequest(ctx, req, nil)
			log.Println("WriteRequest=", err2, rr)

			resp := rr.Response()
			log.Println("resp:", resp.String())
		}
		return nil
	})
	log.Println("Open.Err=", err)
	log.Println("lastErr=", client.LastError())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Open(ctx, fsrpc.PingSender("sys_ping"))
	log.Println("Ping.Err=", err)

	log.Println("Close()", client.Close())
}
