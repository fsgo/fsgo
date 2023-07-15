// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/15

package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/fsgo/fsgo/fsrpc"
)

var serverAddr = flag.String("addr", "127.0.0.1:8002", "")

func main() {
	flag.Parse()

	client, err := fsrpc.DialTimeout("tcp", *serverAddr, time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	rw := client.MustOpen(context.Background())
	req := fsrpc.NewRequest("hello")
	pl := &fsrpc.PayloadChan[*fsrpc.Echo]{
		RID: req.GetID(),
	}
	go func() {
		for i := 0; i < 10000; i++ {
			msg := &fsrpc.Echo{
				ID: uint64(i),
			}
			err2 := pl.Write(context.Background(), msg, i < 10000-1)
			log.Println("Write msg:", i, err2)
		}
	}()
	_, err = rw.Write(context.Background(), req, pl.Chan())
	log.Println("WriteChan:", err)

	time.Sleep(time.Second)
}
