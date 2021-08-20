// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/20

package fsnet

import (
	"bytes"
	"net"
	"testing"
)

func TestConnReadBytesHook(t *testing.T) {
	t.Run("read fail", func(t *testing.T) {
		c1 := &net.TCPConn{}
		ch := NewConnReadBytesHook()
		c2 := NewConn(c1, ch.ConnHook())
		bf := make([]byte, 1024)
		_, err := c2.Read(bf)
		if err == nil {
			t.Fatalf("expect error")
		}
		if got := ch.ReadBytes(); len(got) != 0 {
			t.Fatalf("ch.ReadBytes()=%q want 0 bytes", got)
		}
	})

	t.Run("read success", func(t *testing.T) {
		w, r := net.Pipe()
		defer w.Close()
		defer r.Close()

		ch := NewConnReadBytesHook()
		c2 := NewConn(r, ch.ConnHook())

		want := []byte("hello")
		go func() {
			if _, err := w.Write(want); err != nil {
				panic(err)
			}
		}()
		bf := make([]byte, 1024)
		n, err := c2.Read(bf)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !bytes.Equal(want, bf[:n]) {
			t.Fatalf("got=%v want=%v", bf[:n], want)
		}

		if got := ch.ReadBytes(); !bytes.Equal(want, got) {
			t.Fatalf("ch.ReadBytes()=%q want=%q ", got, want)
		}

		ch.Reset()
		if got := ch.ReadBytes(); len(got) != 0 {
			t.Fatalf("ch.ReadBytes()=%q want 0 bytes ", got)
		}
	})
}

func TestConnWriteBytesHook(t *testing.T) {
	t.Run("write fail", func(t *testing.T) {
		c1 := &net.TCPConn{}
		ch := NewConnWriteBytesHook()
		c2 := NewConn(c1, ch.ConnHook())
		_, err := c2.Write([]byte("hello"))
		if err == nil {
			t.Fatalf("expect error")
		}
		if got := ch.WriteBytes(); len(got) != 0 {
			t.Fatalf("ch.WriteBytes()=%q want 0 bytes", got)
		}
	})

	t.Run("write success", func(t *testing.T) {
		w, r := net.Pipe()
		defer w.Close()
		defer r.Close()

		ch := NewConnWriteBytesHook()
		c2 := NewConn(r, ch.ConnHook())

		go func() {
			bf := make([]byte, 1024)
			if _, err := w.Read(bf); err != nil {
				panic(err)
			}
		}()

		want := []byte("hello")
		_, err := c2.Write(want)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if got := ch.WriteBytes(); !bytes.Equal(want, got) {
			t.Fatalf("ch.WriteBytes()=%q want=%q ", got, want)
		}
		ch.Reset()
		if got := ch.WriteBytes(); len(got) != 0 {
			t.Fatalf("ch.WriteBytes()=%q want 0 bytes ", got)
		}
	})
}
