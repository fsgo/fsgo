// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/11/28

package grace_test

import (
	"context"
	"log"
	"path/filepath"

	"github.com/fsgo/fsgo/grace"
)

func ExampleNewSimpleConfig() {
	do:= func() {
		cfg := grace.NewSimpleConfig()
		g := cfg.NewGrace()

		workerCfg := &grace.WorkerConfig{
			LogDir:  filepath.Join(cfg.LogDir, "echo"),
			Cmd:     "echo",
			CmdArgs: []string{"hello"},
		}
		worker := grace.NewWorker(workerCfg)
		g.MustRegister("echo", worker)
		err := g.Start(context.Background())
		log.Println("grace server exit:", err)
	}
	_=do
}
