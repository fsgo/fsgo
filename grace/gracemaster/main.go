/*
 * Copyright(C) 2021 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/1/29
 */

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsgo/fsnet/grace"
)

var confName = flag.String("conf", "conf/grace.json", "")

func main() {
	flag.Parse()
	cf, g, err := grace.NewWithConfigName(*confName)
	if err != nil {
		log.Printf(" load config %q failed, error=%v\n", *confName, err)
		os.Exit(1)
	}

	for name, wcf := range cf.Workers {
		opt := &grace.WorkerOption{
			Cmd:         wcf.Cmd,
			CmdArgs:     wcf.CmdArgs,
			StopTimeout: wcf.StopTimeout,
		}
		group := grace.NewWorker(opt)
		for _, dsn := range wcf.Listen {
			if err := group.Register(dsn, nil); err != nil {
				panic(err.Error())
			}
		}
		g.Register(name, group)
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case <-ch:
			log.Println("signal exiting...")
		}
	}()

	err = g.Start(context.Background())
	log.Println("grace_master exit:", err)
}
