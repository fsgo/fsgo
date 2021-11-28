// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/29

package grace

import (
	"context"
	"net"
)

// Server server 类型
type Server interface {
	Serve(l net.Listener) error
	Shutdown(ctx context.Context) error
}

// NewServerConsumer 创建一个新的消费者
func NewServerConsumer(ser Server, dsn Resource) Consumer {
	return &serverConsumer{
		Server: ser,
		res:    dsn,
	}
}

type serverConsumer struct {
	Server Server
	res    Resource
}

func (sc *serverConsumer) Start(ctx context.Context) error {
	l, err := sc.res.Listener(ctx)
	if err != nil {
		return err
	}
	return sc.Server.Serve(l)
}

func (sc *serverConsumer) Stop(ctx context.Context) error {
	return sc.Server.Shutdown(ctx)
}

func (sc *serverConsumer) String() string {
	return "server consumer"
}

var _ Consumer = (*serverConsumer)(nil)
