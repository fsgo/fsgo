/*
 * Copyright(C) 2021 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/1/12
 */

package grace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type resourceServer struct {
	Resource Resource
	Consumer Consumer
}

// subProcess 子进程的逻辑
type subProcess struct {
	group *Worker
}

func (sp *subProcess) logit(msgs ...interface{}) {
	msg := fmt.Sprintf("[grace][worker.process] pid=%d ppid=%d %s", os.Getpid(), os.Getppid(), fmt.Sprint(msgs...))
	_ = sp.group.main.Logger.Output(1, msg)
}

// Start 子进程的启动逻辑
func (sp *subProcess) Start(ctx context.Context) (errLast error) {
	sp.logit("Start() start")
	start := time.Now()
	defer func() {
		cost := time.Since(start)
		sp.logit("Start() finish, error=", errLast,
			", start_at=", start.Format("2006-01-02 15:04:05"),
			", duration=", cost,
		)
	}()

	errChan := make(chan error, len(sp.group.resources))
	for idx, s := range sp.group.resources {
		f := os.NewFile(uintptr(3+idx), "")
		_ = s.Resource.SetFile(f)

		go func(c Consumer) {
			errChan <- c.Start(ctx)
		}(s.Consumer)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	var err error

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case e1 := <-errChan:
		err = e1
	case sig := <-ch:
		sp.logit(fmt.Sprintf("receive signal(%v)", sig))

		err = fmt.Errorf("exit by signal(%v)", sig)
	}

	ctx, cancel := context.WithTimeout(ctx, sp.group.getStopTimeout())
	defer cancel()

	_ = sp.Stop(ctx)

	return err
}

func (sp *subProcess) Stop(ctx context.Context) (errStop error) {
	sp.logit("Stop() start")
	defer func() {
		sp.logit("Stop() finish, error=", errStop)
	}()

	var wg sync.WaitGroup
	errChans := make(chan error, len(sp.group.resources))
	for idx, s := range sp.group.resources {
		wg.Add(1)

		go func(idx int, res Consumer) {
			defer wg.Done()

			if err := res.Stop(ctx); err != nil {
				errChans <- fmt.Errorf("resource[%d] (%s) Stop error: %w", idx, res.String(), err)
			} else {
				errChans <- nil
			}
		}(idx, s.Consumer)
	}
	wg.Wait()

	close(errChans)

	var bd strings.Builder
	for err := range errChans {
		if err != nil {
			bd.WriteString(err.Error())
			bd.WriteString(";")
		}
	}
	if bd.Len() == 0 {
		return nil
	}
	return errors.New(bd.String())
}
