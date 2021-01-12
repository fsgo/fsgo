/*
 * Copyright(C) 2021 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/1/11
 */

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/fsgo/fsnet/grace"
)

func handler(w http.ResponseWriter, r *http.Request) {
	pid := strconv.Itoa(os.Getpid())
	w.Write([]byte("pid=" + pid))
}

func handlerSlow(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
	w.Write([]byte("hello"))
}

func main() {
	http.HandleFunc("/test", handler)
	http.HandleFunc("/slow", handlerSlow)

	g := &grace.Grace{
		PIDFilePath:     "./ss.pid",
		ShutdownTimeout: 30 * time.Second,
	}

	// server 1
	{
		res := &grace.ServerResource{
			Server:  &http.Server{},
			NetWork: "tcp",
			Address: "127.0.0.1:8909",
		}
		g.Register(res)
	}

	// server 2
	{
		res := &grace.ServerResource{
			Server:  &http.Server{},
			NetWork: "tcp",
			Address: "127.0.0.1:8910",
		}
		g.Register(res)
	}

	err := g.Start(context.Background())
	log.Println("process exit", err, "pid=", os.Getpid())
}
