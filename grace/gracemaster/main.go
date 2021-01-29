/*
 * Copyright(C) 2021 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/1/29
 */

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsgo/fsnet/grace"
)

var confName = flag.String("conf", "conf/grace.json", "")

func main() {
	flag.Parse()
	c, err := loadConf(*confName)
	if err != nil {
		log.Printf(" load config %q failed, error=%v\n", *confName, err)
		os.Exit(1)
	}

	g := &grace.Grace{
		PIDFilePath: c.PIDFilePath,
		Keep:        c.Keep,
		StopTimeout: time.Duration(c.StopTimeout) * time.Millisecond,
		Log:         nil,
	}

	if c.Cmd != "" {
		// todo  解析 cmd
		g.SubProcessCommand = []string{c.Cmd}
	}

	for _, r := range c.Resource {
		g.RegisterByDSN(r, nil)
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
	log.Println("gracemaster exit:", err)
}

type Config struct {
	Resource    []string
	Cmd         string
	PIDFilePath string
	Keep        bool
	StopTimeout int
}

func (c *Config) Check() error {
	if len(c.Resource) == 0 {
		return fmt.Errorf("empty resource")
	}
	if c.PIDFilePath == "" {
		return fmt.Errorf("empty PIDFilePath")
	}
	return nil
}

func loadConf(name string) (*Config, error) {
	bf, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	var c *Config
	if e := json.Unmarshal(bf, &c); e != nil {
		return nil, e
	}
	return c, c.Check()
}
