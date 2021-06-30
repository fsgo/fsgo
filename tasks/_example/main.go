// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/6/30

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/fsgo/fsnet/tasks"
)

func main() {
	mp := &tasks.Mapper{}
	mp.Register(newMyTask())
	err := mp.Execute(context.Background())
	log.Println("err=", err)
}

func newMyTask() *myTask {
	return &myTask{
		ids: make(chan int),
	}
}

var _ tasks.Task = (*myTask)(nil)

type myTask struct {
	msg string
	num int
	ids chan int
}

func (m *myTask) FlagSet(fg *flag.FlagSet) {
	fg.StringVar(&m.msg, "msg", "", "message")
	fg.IntVar(&m.num, "num", 10, "total ids")
}

func (m *myTask) Name() string {
	return "hello"
}

func (m *myTask) Run(ctx context.Context) error {
	go func() {
		tasks.RunWorker(ctx, m.producer, 10)
		close(m.ids)
	}()
	fmt.Println("Hello World, msg=", m.msg)
	for id := range m.ids {
		fmt.Println("get id=", id)
	}
	return nil
}

func (m *myTask) producer(ctx context.Context) error {
	for i := 0; i < m.num; i++ {
		m.ids <- i
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (m *myTask) TearDown(err error) {
	fmt.Println("TearDown, got err=", err)
}
