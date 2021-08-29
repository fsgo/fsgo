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
	"path/filepath"
	"syscall"

	"github.com/fsgo/fsgo/fsos"
	"github.com/fsgo/fsgo/grace"
)

var confName = flag.String("conf", "./conf/grace.toml", "")

func main() {
	flag.Parse()

	cf, err := grace.LoadConfig(*confName)
	if err != nil {
		log.Fatalf(" load config %q failed, error=%v\n", *confName, err)
	}

	lg := &fsos.RotateFile{
		Path:    filepath.Join(cf.LogDir, "grace", "grace.log"),
		ExtRule: "1hour",
	}
	if err := lg.Init(); err != nil {
		log.Fatalf("init logger failed, error=%v\n", err)
	}

	defer lg.Close()

	logger := log.Default()
	logger.SetOutput(lg)

	{
		fn, err := filepath.Abs(*confName)
		if err != nil {
			logger.Fatalf("filepath.Abs(%q) failed, err=%v", *confName, err)
		}
		wd := filepath.Dir(filepath.Dir(fn))
		logger.Println("[grace][master] working dir=", wd)
	}

	g := grace.Grace{
		Option: cf.ToOption(),
		Logger: logger,
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
		logger.Printf("received signal %v, exiting...", sig)
	}()

	err = g.Start(context.Background())
	logger.Println("grace_master exit:", err)
}
