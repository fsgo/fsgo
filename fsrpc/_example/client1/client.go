// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package main

import (
	"context"
	"flag"
	"log"
	"sync"
	"time"

	"github.com/fsgo/fsgo/fsnet/fsconn"
	"github.com/fsgo/fsgo/fsnet/fsdialer"
	"github.com/fsgo/fsgo/fsrpc"
)

var serverAddr = flag.String("addr", "127.0.0.1:8000", "")
var debug = flag.Bool("d", false, "enable debug")
var wait = flag.Int("wait", 1, "wait seconds")

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

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

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)
	stream := client.MustOpen(ctx)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			log.Println("i=", i)
			req := fsrpc.NewRequest("hello")
			rr, err2 := stream.WriteChan(ctx, req, nil)
			log.Println("WriteChan=", err2)

			resp, _, err3 := rr.Response()
			log.Println("resp:", resp.String(), err3)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	ph := &fsrpc.PingHandler{}
	go func() {
		defer wg.Done()
		err = ph.ClientSendMany(ctx, stream, time.Duration(*wait)*time.Second)
		log.Println("Ping.Err=", err)
	}()
	wg.Wait()

	log.Println("Client.Err=", client.LastError())

	log.Println("Close()", client.Close())
}
