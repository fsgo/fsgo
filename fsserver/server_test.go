// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/16

package fsserver_test

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/fsgo/fsgo/fsserver"
)

func TestAnyServer(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	fst.NoError(t, err)
	fst.NotNil(t, l)
	defer l.Close()

	ser := &fsserver.AnyServer{
		Handler: fsserver.HandleFunc(echoHandler),
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = ser.Serve(l)
	}()
	conn, err := net.DialTimeout("tcp", l.Addr().String(), 100*time.Millisecond)
	fst.NoError(t, err)
	rd := bufio.NewReader(conn)
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("loop=%d", i), func(t *testing.T) {
			_, err = conn.Write([]byte("hello\n"))
			fst.NoError(t, err)
			line, _, err := rd.ReadLine()
			fst.NoError(t, err)
			fst.Equal(t, `resp:"hello"`, string(line))
		})
	}
	fst.NoError(t, conn.Close())
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	fst.NoError(t, ser.Shutdown(ctx))
	fst.NoError(t, l.Close())
	wg.Wait()
}

func echoHandler(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	rd := bufio.NewReader(conn)
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			return
		}
		resp := fmt.Sprintf("resp:%q\n", line)
		_, err = conn.Write([]byte(resp))
		if err != nil {
			return
		}
	}
}
