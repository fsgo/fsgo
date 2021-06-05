// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/11

package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
func handlerPanic(w http.ResponseWriter, r *http.Request) {
	panic("must panic")
}

var msg = flag.String("msg", "", "")

func main() {
	flag.Parse()

	http.HandleFunc("/test", handler)
	http.HandleFunc("/slow", handlerSlow)
	http.HandleFunc("/panic", handlerPanic)
	_ = *msg
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case <-ch:
			log.Println("worker exiting...")
		}
	}()

	cf, err := LoadConfig("conf/grace.json")
	if err != nil {
		log.Fatalf(" load config %q failed, error=%v\n", "conf/grace.json", err)
	}

	if err = cf.Parser(); err != nil {
		log.Fatalf(" parser config %q failed, error=%v\n", "conf/grace.json", err)
	}

	wcf := cf.Workers["default"]

	worker := grace.NewWorker(nil)
	worker.Register(wcf.Listen[0], grace.NewServerConsumer(&http.Server{}))
	worker.Register(wcf.Listen[1], grace.NewServerConsumer(&http.Server{}))

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

func LoadConfig(name string) (*grace.Config, error) {
	bf, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	var c *grace.Config
	if e := json.Unmarshal(bf, &c); e != nil {
		return nil, e
	}

	return c, c.Parser()
}
