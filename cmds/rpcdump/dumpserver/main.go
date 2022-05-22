// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package main

import (
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/fsgo/fsgo/fsnet/fsconn"
	"github.com/fsgo/fsgo/fsnet/fsconn/conndump"
)

var addr = flag.String("l", ":8090", "server listen addr")
var out = flag.String("o", "./dump_data/", "dump data dir")
var maxFiles = flag.Int("m", 24, "max dump files total")

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln("listen failed:", err)
	}
	log.Println("dump server listen at:", l.Addr().String())

	log.Fatalln("server exit:", startDumpServer(l))
}

func startDumpServer(l net.Listener) error {
	dm := &conndump.Dumper{
		DataDir: *out,
		RotatorConfig: func(r *fsfs.Rotator) {
			r.MaxFiles = *maxFiles
		},
	}

	handler := func(conn net.Conn) {
		log.Println("connect:", conn.RemoteAddr())
		conn = fsconn.WithInterceptor(conn, dm.Interceptor())
		n, err := io.Copy(ioutil.Discard, conn)
		log.Println("disconnect:", conn.RemoteAddr(), "read=", n, "err=", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Temporary() {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return err
		}
		go handler(conn)
	}
}
