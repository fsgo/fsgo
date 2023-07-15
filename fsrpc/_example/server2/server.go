// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/15

package main

import (
	"context"
	"flag"
	"log"

	"github.com/fsgo/fsgo/fsrpc"
)

var listenAddr = flag.String("addr", "127.0.0.1:8002", "")

func main() {
	flag.Parse()
	rt := fsrpc.NewRouter()
	rt.Register("hello", fsrpc.HandlerFunc(hello))
	log.Println("listen add:", *listenAddr)
	log.Println(fsrpc.ListenAndServe(*listenAddr, rt))
}

func hello(ctx context.Context, rr fsrpc.RequestReader, rw fsrpc.ResponseWriter) error {
	req, pl := rr.Request()
	log.Println("req", req.GetID(), len(pl))
	fsrpc.RangeParserPayloads[*fsrpc.Echo](ctx, pl, func() *fsrpc.Echo {
		return &fsrpc.Echo{}
	}, func(data *fsrpc.Echo) error {
		log.Println(data.GetID(), data.GetMessage())
		return nil
	})
	return nil
}
