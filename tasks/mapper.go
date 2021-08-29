// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/6/30

package tasks

import (
	"context"
	"flag"
	"fmt"
	"os"
)

var ErrTaskNotFound = fmt.Errorf("task not found")

type Mapper struct {
	tasks map[string]Task
}

// Register 注册
func (m *Mapper) Register(task Task) error {
	if m.tasks == nil {
		m.tasks = map[string]Task{}
	}
	if _, has := m.tasks[task.Name()]; has {
		return fmt.Errorf("task.Name=%q already exists", task.Name())
	}
	m.tasks[task.Name()] = task
	return nil
}

func (m *Mapper) Find(name string) Task {
	if len(m.tasks) == 0 {
		return nil
	}
	return m.tasks[name]
}

func (m *Mapper) Run(ctx context.Context, name string) error {
	task := m.Find(name)
	if task == nil {
		return fmt.Errorf("%w, name=%q", ErrTaskNotFound, name)
	}
	return Run(ctx, task)
}

func (m *Mapper) Execute(ctx context.Context) error {
	var name string
	fs := flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	fs.StringVar(&name, "name", "", "task name")
	var help bool
	fs.BoolVar(&help, "help", false, "help")
	if len(os.Args) < 3 {
		fs.PrintDefaults()
		return fmt.Errorf("os.Args too short")
	}
	if err := fs.Parse(os.Args[1:3]); err != nil {
		return err
	}
	if name == "" {
		fs.PrintDefaults()
		return fmt.Errorf("-name is empty")
	}
	return m.Run(ctx, name)
}

func (m *Mapper) Tasks() map[string]Task {
	return m.tasks
}
