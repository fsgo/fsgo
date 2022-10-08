// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/25

package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/fsgo/fsgo/fsnet/fsconn/conndump"
)

var addr = flag.String("l", ":8080", "server listen addr")
var out = flag.String("o", "./dump_data/", "dump data dir")
var maxFiles = flag.Int("m", 24, "max dump files total")

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln("listen failed:", err)
	}

	log.Println("HTTP server listen at:", l.Addr().String(), "dump data dir:", *out)
	err = startHTTPServer(l)
	log.Println("server exit:", err)
}

func startHTTPServer(l net.Listener) error {
	dm := &conndump.Dumper{
		DataDir: *out,
		RotatorConfig: func(client bool, r *fsfs.Rotator) {
			r.MaxFiles = *maxFiles
		},
	}
	dm.DumpAll(true)

	l = dm.WrapListener("http_server", l)
	hs := &http.Server{
		Handler: http.HandlerFunc(hello),
	}
	return hs.Serve(l)
}

func hello(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(r.RequestURI))
}
