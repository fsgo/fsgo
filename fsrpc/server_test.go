// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"context"
	"testing"
)

type testHandler1 struct {
}

func (t *testHandler1) Echo(ctx context.Context, req RequestReader, w ResponseWriter) error {
	return nil
}

func (t *testHandler1) Hello(ctx context.Context, req RequestReader, w ResponseWriter) error {
	return nil
}

func TestRouterRegister(t *testing.T) {
	rt := NewRouter()
	th1 := &testHandler1{}
	rt.Register("demo.echo", th1.Echo)
	rt.Register("demo.hello", th1.Hello)
}
