// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/29

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsgo/fsgo/grace"
)

var confName = flag.String("conf", "./conf/grace.json", "")

func main() {
	flag.Parse()

	cf, err := grace.LoadConfig(*confName)
	if err != nil {
		log.Fatalf(" load config %q failed, error=%v\n", *confName, err)
	}

	g := grace.Grace{
		Option: cf.ToOption(),
	}

	for name, wcf := range cf.Workers {
		group := grace.NewWorker(wcf)
		for _, dsn := range wcf.Listen {
			if err = group.Register(dsn, nil); err != nil {
				panic(err.Error())
			}
		}
		_ = g.Register(name, group)
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-ch
		log.Printf("received signal %v, exiting...", sig)
	}()

	err = g.Start(context.Background())
	log.Println("grace_master exit:", err)
}
