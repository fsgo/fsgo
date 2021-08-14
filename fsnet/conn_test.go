// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/14

package fsnet

import (
	"bytes"
	"net"
	"testing"
)

func TestNewConn(t *testing.T) {
	t.Run("hooks", func(t *testing.T) {
		w, r := net.Pipe()
		var readTotal, writeTotal int
		var closeNum int

		var readIndex int

		tr := &ConnHook{
			Read: func(b []byte, raw func([]byte) (int, error)) (n int, err error) {
				defer func() {
					readTotal += n
				}()

				readIndex++
				if readIndex != 2 {
					t.Fatalf("readIndex=%d want=2", readIndex)
				}

				return raw(b)
			},
			Write: func(b []byte, raw func([]byte) (int, error)) (n int, err error) {
				defer func() {
					writeTotal += n
				}()
				return raw(b)
			},
			RemoteAddr: func(raw func() net.Addr) net.Addr {
				// return the hooked addr
				return &net.TCPAddr{}
			},
			Close: func(raw func() error) error {
				closeNum++
				return raw()
			},
		}

		tr2 := &ConnHook{
			Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
				readIndex++
				if readIndex != 1 {
					t.Fatalf("readIndex=%d want=1", readIndex)
				}
				return raw(b)
			},
		}

		w1 := NewConn(w, tr, tr2)
		r1 := NewConn(r)

		msg := []byte("hello")
		go func() {
			_, _ = w1.Write(msg)
		}()
		buf := make([]byte, 128)
		if n, err := r1.Read(buf); err != nil || n != len(msg) {
			t.Fatalf("read faild, err=%v n=%v", err, n)
		} else if !bytes.Equal(msg, buf[:n]) {
			t.Fatalf("read msg not expect, got=%q", buf[:n])
		}

		if got := w1.LocalAddr().Network(); got != "pipe" {
			t.Fatalf("w1.LocalAddr().Network()=%v want=%v", got, "pipe")
		}

		if got := w1.RemoteAddr().Network(); got != "tcp" {
			t.Fatalf("w1.RemoteAddr().Network()=%v want=%v", got, "tcp")
		}

		if err := w1.Close(); err != nil {
			t.Fatalf("w1.close failed: %v", err)
		}

		if err := r1.Close(); err != nil {
			t.Fatalf("r1.close failed: %v", err)
		}
		if closeNum != 1 {
			t.Fatalf("closeNum=%v want=%v", closeNum, 1)
		}
	})
}
