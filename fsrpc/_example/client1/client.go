// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/23

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsgo/fsgo/fsnet/fsconn"
	"github.com/fsgo/fsgo/fsnet/fsdialer"
	"github.com/fsgo/fsgo/fsrpc"
)

var serverAddr = flag.String("addr", "127.0.0.1:8001", "")
var debug = flag.Bool("d", false, "enable debug")
var wait = flag.Int("wait", 1, "wait seconds")

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
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

	client := fsrpc.NewClient(conn)
	// client.SetBeforeReadLoop(func() {
	// 	conn.SetReadDeadline(time.Now().Add(time.Second))
	// })

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	stream := client.OpenStream()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		req2 := fsrpc.NewRequest("hello")
		pc := &fsrpc.PayloadChan[*fsrpc.Echo]{
			EncodingType: fsrpc.EncodingType_Protobuf,
			RID:          req2.GetID(),
		}
		done := make(chan bool)
		go func() {
			defer close(done)
			for i := 0; i < 10; i++ {
				more := i < 9
				err1 := pc.Write(ctx, &fsrpc.Echo{Message: fmt.Sprintf("PayloadChan:%d", i)}, more)
				log.Println("go1 pc.Write, i=", i, "more=", more, err1)
			}
		}()
		rr, err2 := stream.Write(ctx, req2, pc.Chan())
		<-done
		log.Println("PayloadChan err2:", err2)
		if rr != nil {
			resp, pl, err3 := rr.Response()
			log.Println("go1 response:", resp, err3)
			if pl != nil {
				fsrpc.PayloadsDiscard(ctx, pl)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			log.Println("go2 100_i=", i)
			req := fsrpc.NewRequest("hello")
			rr, err2 := stream.Write(ctx, req, nil)
			log.Println("go2 WriteChan=", err2)
			if err2 != nil {
				continue
			}

			resp, pl, err3 := rr.Response()
			log.Println("go2 rr.Response()", resp, pl, err3)
			if pl != nil {
				fsrpc.PayloadsDiscard(ctx, pl)
			}
			log.Println("resp:", resp, err3)
			time.Sleep(100 * time.Millisecond)
		}
		log.Println("go2 exit")
	}()

	// wg.Add(1)
	// ph := &fsrpc.PingHandler{}
	// go func() {
	// 	defer wg.Done()
	// 	err = ph.ClientSendMany(ctx, stream, time.Duration(*wait)*time.Second)
	// 	log.Println("Ping.Err=", err)
	// }()
	wg.Wait()

	log.Println("Client.Err=", client.LastError())

	log.Println("Close()", client.Close())
}
