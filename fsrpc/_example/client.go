// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/fsgo/fsgo/fsnet/fsconn"
	"github.com/fsgo/fsgo/fsnet/fsdialer"
	"github.com/fsgo/fsgo/fsrpc"
)

var serverAddr = flag.String("addr", "127.0.0.1:8000", "")
var debug = flag.Bool("d", false, "enable debug")

func forDebug() {
	pt := &fsconn.PrintByteTracer{}
	fsconn.RegisterInterceptor(pt.Interceptor())
}

func main() {
	flag.Parse()
	if *debug {
		forDebug()
	}
	conn, err := fsdialer.DialTimeout("tcp", *serverAddr, time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	conn = fsconn.Wrap(conn)

	client := fsrpc.NewClientConn(conn)
	// client.SetBeforeReadLoop(func() {
	// 	conn.SetReadDeadline(time.Now().Add(time.Second))
	// })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream := client.MustOpen(ctx)
	// for i := 0; i < 10; i++ {
	// 	log.Println("i=", i)
	// 	req := fsrpc.NewRequest("hello")
	// 	rr, err2 := stream.WriteChan(ctx, req, nil)
	// 	log.Println("WriteRequest=", err2)
	//
	// 	resp, _ := rr.Response()
	// 	log.Println("resp:", resp.String())
	// }

	ph := &fsrpc.PingHandler{}
	err = ph.SendMany(ctx, stream, time.Second)
	log.Println("Ping.Err=", err)
	log.Println("Client.Err=", client.LastError())

	log.Println("Close()", client.Close())
}
