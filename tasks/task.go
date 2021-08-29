package tasks

import (
	"context"
	"flag"
	"os"

	"golang.org/x/sync/errgroup"
)

// Task 任务
type Task interface {
	Name() string
	FlagSet(fg *flag.FlagSet)
	Run(ctx context.Context) error
	TearDown(err error)
}

func RunWorker(ctx context.Context, fn func(ctx context.Context) error, workerNum int) error {
	g1 := errgroup.Group{}
	for i := 0; i < workerNum; i++ {
		g1.Go(func() error {
			return fn(ctx)
		})
	}
	return g1.Wait()
}

func Run(ctx context.Context, task Task) (err error) {
	defer func() {
		task.TearDown(err)
	}()
	fg := flag.NewFlagSet(task.Name(), flag.ExitOnError)
	task.FlagSet(fg)

	if err = fg.Parse(os.Args[3:]); err != nil {
		return err
	}

	return task.Run(ctx)
}
