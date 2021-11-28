// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/11/28

package grace

import (
	"context"
)

// Consumer 资源消费者
type Consumer interface {
	// Start 开始运行 同步、阻塞
	Start(ctx context.Context) error

	// Stop 关闭
	Stop(ctx context.Context) error

	// String 资源的描述
	String() string
}
