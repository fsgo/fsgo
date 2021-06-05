// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/29

package grace

import (
	"context"
	"fmt"
	"net"
)

// Server server 类型
type Server interface {
	Serve(l net.Listener) error
	Shutdown(ctx context.Context) error
}

func NewServerConsumer(ser Server) Consumer {
	return &serverConsumer{
		Server: ser,
	}
}

type serverConsumer struct {
	Server Server
	res    Resource
}

func (sc *serverConsumer) Bind(res Resource) {
	sc.res = res
}

func (sc *serverConsumer) getListener() (net.Listener, error) {
	if sc.res == nil {
		return nil, fmt.Errorf("no resource found")
	}
	f, err := sc.res.File()
	if err != nil {
		return nil, err
	}
	return net.FileListener(f)
}

func (sc *serverConsumer) Start(ctx context.Context) error {
	l, err := sc.getListener()
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
