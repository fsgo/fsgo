/*
 * Copyright(C) 2021 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/1/11
 */

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/fsgo/fsnet/grace"
)

func handler(w http.ResponseWriter, r *http.Request) {
	pid := strconv.Itoa(os.Getpid())
	w.Write([]byte("pid=" + pid))
}

func main() {
	log.SetPrefix(fmt.Sprintf("pid=%d ", os.Getpid()))

	http.HandleFunc("/test", handler)

	g := &grace.Grace{
		PIDFilePath: "./ss.pid",
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
	log.Println("exit", err)
}
