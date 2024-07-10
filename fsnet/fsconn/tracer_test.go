// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsconn

import (
	"net"
	"testing"

	"github.com/fsgo/fst"
)

func TestReadTracer(t *testing.T) {
	t.Run("read fail", func(t *testing.T) {
		c1 := &net.TCPConn{}
		ch := &ReadTracer{}
		c2 := Wrap(c1, ch.ConnInterceptor())
		bf := make([]byte, 1024)
		_, err := c2.Read(bf)
		fst.NotNil(t, err)
		fst.Equal(t, len(ch.ReadBytes()), 0)
	})

	t.Run("read success", func(t *testing.T) {
		w, r := net.Pipe()
		defer w.Close()
		defer r.Close()

		ch := &ReadTracer{}
		c2 := Wrap(r, ch.ConnInterceptor())

		want := []byte("hello")
		go func() {
			if _, err := w.Write(want); err != nil {
				panic(err)
			}
		}()
		bf := make([]byte, 1024)
		n, err := c2.Read(bf)
		fst.Nil(t, err)
		fst.Equal(t, want, bf[:n])

		fst.Equal(t, ch.ReadBytes(), want)

		ch.Reset()
		fst.Len(t, ch.ReadBytes(), 0)
	})
}

func TestWriteTracer(t *testing.T) {
	t.Run("write fail", func(t *testing.T) {
		c1 := &net.TCPConn{}
		ch := &WriteTracer{}
		c2 := Wrap(c1, ch.Interceptor())
		_, err := c2.Write([]byte("hello"))
		fst.NotNil(t, err)
		fst.Len(t, ch.WriteBytes(), 0)
	})

	t.Run("write success", func(t *testing.T) {
		w, r := net.Pipe()
		defer w.Close()
		defer r.Close()

		ch := &WriteTracer{}
		c2 := Wrap(r, ch.Interceptor())

		go func() {
			bf := make([]byte, 1024)
			if _, err := w.Read(bf); err != nil {
				panic(err)
			}
		}()

		want := []byte("hello")
		_, err := c2.Write(want)
		fst.Nil(t, err)
		fst.Equal(t, ch.WriteBytes(), want)

		ch.Reset()
		fst.Len(t, ch.WriteBytes(), 0)
	})
}
