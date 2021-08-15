// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/11

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fsgo/fsgo/grace"
)

var startTime = time.Now()

func handler(w http.ResponseWriter, r *http.Request) {
	pid := strconv.Itoa(os.Getpid())
	var bf bytes.Buffer
	bf.WriteString("<table>")
	bf.WriteString(`<thead>
<tr><th>Key</th><th>Value</th></tr>
</thead>
`)
	bf.WriteString("\n<tbody>\n")
	bf.WriteString("<tr><th>pid</th><td>" + pid + "</td></tr>\n")
	bf.WriteString("<tr><th>start</th><td>" + startTime.Format("2006-01-02 15:04:05") + "</td></tr>\n")
	bf.WriteString("<tr><th>os.Environ()</th><td>")
	for _, v := range os.Environ() {
		bf.WriteString(v + "<br/>")
	}
	bf.WriteString("</td></tr>\n")
	bf.WriteString("</tbody></table>")

	w.Header().Set("Content-Type", "text/html")
	w.Write(bf.Bytes())
}

func handlerSlow(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
	w.Write([]byte("hello"))
}
func handlerPanic(w http.ResponseWriter, r *http.Request) {
	panic("must panic")
}

var msg = flag.String("msg", "", "")
var config = flag.String("conf", "./conf/grace.toml", "grace config path")

func main() {
	flag.Parse()

	{
		wd, err := os.Getwd()
		log.Println("os.Getwd()=", wd, err)
	}

	http.HandleFunc("/", handler)

	http.HandleFunc("/test", handler)
	http.HandleFunc("/slow", handlerSlow)
	http.HandleFunc("/panic", handlerPanic)
	_ = *msg
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case sig := <-ch:
			log.Println("worker exiting by signal ...", sig)
		}
	}()

	cf, err := grace.LoadConfig(*config)
	if err != nil {
		log.Fatalf(" load config %q failed, error=%v\n", *config, err)
	}

	wcf := cf.Workers["default"]

	worker := grace.NewWorker(nil)
	worker.RegisterServer(wcf.Listen[0], &http.Server{})
	worker.RegisterServer(wcf.Listen[1], &http.Server{})

	g := grace.Grace{
		Option: cf.ToOption(),
	}
	g.Register("default", worker)

	err = g.Start(context.Background())
	log.Println("worker exit", err, "pid=", os.Getpid())
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	log.SetPrefix(fmt.Sprintf("pid=%d ", os.Getpid()))
}
