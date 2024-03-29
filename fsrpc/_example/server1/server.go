// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/9

package main

import (
	"context"
	"flag"
	"log"
	"net"

	"google.golang.org/protobuf/proto"

	"github.com/fsgo/fsgo/fsrpc"
)

var listenAddr = flag.String("addr", "127.0.0.1:8001", "")

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	l, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Listen At:", l.Addr().String())

	rt := fsrpc.NewRouter()
	rt.Register("hello", fsrpc.HandlerFunc(hello))

	ph := fsrpc.PingHandler{}
	ph.RegisterTo(rt)

	ser := &fsrpc.Server{
		Router: rt,
	}
	log.Println(ser.Serve(l))
}

func hello(ctx context.Context, rr fsrpc.RequestReader, rw fsrpc.ResponseWriter) error {
	req, pl := rr.Request()
	for item := range pl {
		log.Println("pl:", item.Meta)
		bf, err := item.Bytes()
		if err != nil {
			log.Println("pl.Bytes().err=", err)
			continue
		}
		echo := &fsrpc.Echo{}
		proto.Unmarshal(bf, echo)
		log.Println("pl.Echo:", echo.String())
	}
	log.Println("request", req.String())
	resp := fsrpc.NewResponse(req.GetID(), 0, "success")
	err1 := rw.Write(ctx, resp, nil)
	log.Println("WriteResponse", err1)
	return nil
}
