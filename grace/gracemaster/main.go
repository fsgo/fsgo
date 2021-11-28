// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/29

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/fsgo/fsgo/grace"
)

var confName = flag.String("conf", "./conf/grace.toml", "")

func main() {
	flag.Parse()

	cf, err := grace.LoadConfig(*confName)
	if err != nil {
		log.Fatalf(" load config %q failed, error=%v\n", *confName, err)
	}

	logger, close := getLogger(cf.LogDir)
	defer close()

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
		for i := 0; i < len(wcf.Listen); i++ {
			res := group.Resource(i)
			group.MustRegister(nil, res)
		}
		_ = g.Register(name, group)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-ch
		logger.Printf("received signal %v, exiting...", sig)
	}()

	err = g.Start(context.Background())
	logger.Println("grace_master exit:", err)
}

func getLogger(logDir string) (*log.Logger, func()) {
	lg := &fsfs.Rotator{
		Path:    filepath.Join(logDir, "grace", "grace.log"),
		ExtRule: "1hour",
	}

	if err := lg.Init(); err != nil {
		log.Fatalf("init logger failed, error=%v\n", err)
	}

	logger := log.Default()
	logger.SetOutput(lg)
	logger.SetPrefix(fmt.Sprintf("pid=%d ppid=%d ", os.Getpid(), os.Getppid()))
	logger.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)
	return logger, func() {
		_ = lg.Close()
	}
}
