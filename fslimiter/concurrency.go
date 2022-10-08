// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/3

package fslimiter

import (
	"context"
	"sync"
)

// Concurrency 并发度限制器
type Concurrency struct {
	sem chan struct{}

	// Max 最大并发度
	Max int

	once sync.Once
}

func (c *Concurrency) init() {
	c.sem = make(chan struct{}, c.Max)
}

// Wait 获取锁
//
// 返回的 func() 类型的值用于释放锁
func (c *Concurrency) Wait() func() {
	release, _ := c.WaitContext(context.Background())
	return release
}

// WaitContext 获取锁，若失败会返回 error
//
// 返回的 第一个func() 类型的值用于释放锁
func (c *Concurrency) WaitContext(ctx context.Context) (func(), error) {
	if c.Max < 1 {
		return empty, nil
	}

	c.once.Do(c.init)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case c.sem <- struct{}{}:
		return c.release, nil
	}
}

func (c *Concurrency) release() {
	<-c.sem
}

func empty() {}
