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
	"path/filepath"

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

	logger, closeFn := getLogger(cf.LogDir)
	defer closeFn()

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

	err = g.Start(context.Background())
	logger.Println("grace_master exit:", err)
	if err!=nil{
		os.Exit(1)
	}
}

func getLogger(logDir string) (*log.Logger, func()) {
	logger := log.Default()
	logger.SetPrefix(fmt.Sprintf("pid=%d ppid=%d ", os.Getpid(), os.Getppid()))
	logger.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)
	
	// reload,stop 等子命令，直接将日志输出到 stderr 即可
	if isSubCmd(){
		return logger, func() {}
	}
	
	lg := &fsfs.Rotator{
		Path:    filepath.Join(logDir, "grace", "grace.log"),
		ExtRule: "1hour",
	}

	if err := lg.Init(); err != nil {
		log.Fatalf("init logger failed, error=%v\n", err)
	}
	
	logger.SetOutput(lg)
	return logger, func() {
		_ = lg.Close()
	}
}

func isSubCmd()bool{
	if len(os.Args) > 1{
		return os.Args[1]=="reload"|| os.Args[1]=="stop"
	}
	return false
}
