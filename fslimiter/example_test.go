// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/3

package fslimiter_test

import (
	"log"
	"sync"

	"github.com/fsgo/fsgo/fslimiter"
)

func ExampleConcurrency_Wait() {
	limiter := fslimiter.Concurrency{
		Max: 2, // 限制最大并发为 2
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 从并发限制器获取令牌
			release := limiter.Wait()
			defer release() // 释放令牌

			log.Println("id=", id)
		}(i)
	}
	wg.Wait()
}
