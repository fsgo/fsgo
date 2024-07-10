// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsconn

import (
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestNewConn(t *testing.T) {
	t.Run("its", func(t *testing.T) {
		w, r := net.Pipe()
		var readTotal, writeTotal int
		var closeNum int

		var readIndex int

		tr := &Interceptor{
			Read: func(info Info, b []byte, raw func([]byte) (int, error)) (n int, err error) {
				defer func() {
					readTotal += n
				}()

				readIndex++
				fst.Equal(t, 2, readIndex)

				return raw(b)
			},
			Write: func(info Info, b []byte, raw func([]byte) (int, error)) (n int, err error) {
				defer func() {
					writeTotal += n
				}()
				return raw(b)
			},
			RemoteAddr: func(info Info, raw func() net.Addr) net.Addr {
				// return the intercepted addr
				return &net.TCPAddr{}
			},
			Close: func(info Info, raw func() error) error {
				closeNum++
				return raw()
			},
		}

		tr2 := &Interceptor{
			Read: func(info Info, b []byte, raw func([]byte) (int, error)) (int, error) {
				readIndex++
				fst.Equal(t, 1, readIndex)
				return raw(b)
			},
		}

		w1 := Wrap(w, tr, tr2)
		r1 := Wrap(r)

		msg := []byte("hello")
		go func() {
			_, _ = w1.Write(msg)
		}()
		buf := make([]byte, 128)

		n, err := r1.Read(buf)
		fst.Nil(t, err)
		fst.Equal(t, len(msg), n)
		fst.Equal(t, string(msg), string(buf[:n]))

		fst.Equal(t, "pipe", w1.LocalAddr().Network())
		fst.Equal(t, "tcp", w1.RemoteAddr().Network())

		t.Run("Close", func(t *testing.T) {
			fst.Nil(t, w1.Close())

			fst.Equal(t, 1, closeNum)
		})
	})
}

func TestNewConn_merge(t *testing.T) {
	var id int
	hk1 := &Interceptor{
		Read: func(info Info, b []byte, raw func([]byte) (int, error)) (int, error) {
			// 先注册的先执行
			id++
			fst.Equal(t, 1, id)
			return raw(b)
		},
		AfterRead: func(info Info, b []byte, readSize int, err error) {
			id++
			fst.Equal(t, 4, id)
		},
	}
	nc := Wrap(&net.TCPConn{}, hk1)

	hk2 := &Interceptor{
		Read: func(info Info, b []byte, raw func([]byte) (int, error)) (int, error) {
			id++
			fst.Equal(t, 2, id)
			return raw(b)
		},
		AfterRead: func(info Info, b []byte, readSize int, err error) {
			id++
			fst.Equal(t, 5, id)
		},
	}

	hk3 := &Interceptor{
		Read: func(info Info, b []byte, raw func([]byte) (int, error)) (int, error) {
			id++
			fst.Equal(t, 3, id)
			return raw(b)
		},
	}
	hk4 := &Interceptor{
		AfterRead: func(info Info, b []byte, readSize int, err error) {
			id++
			fst.Equal(t, 6, id)
		},
	}
	nc1 := Wrap(nc, hk2, hk3, hk4)
	fst.NotEqual(t, nc, nc1)
	bf := make([]byte, 1)
	_, _ = nc1.Read(bf)
	fst.Equal(t, 6, id)
}

func TestOriginConn(t *testing.T) {
	c1 := &net.TCPConn{}
	c2 := Wrap(c1)

	fst.SamePtr(t, c1, OriginConn(c2))
	fst.SamePtr(t, c1, OriginConn(c1))
}

func Test_connInterceptors_CallSetDeadline(t *testing.T) {
	want := time.Now()

	var num int32
	RegisterInterceptor(&Interceptor{
		SetDeadline: func(info Info, tm time.Time, raw func(tm time.Time) error) error {
			fst.Equal(t, int32(1), atomic.AddInt32(&num, 1))
			fst.Equal(t, want, tm)
			return raw(tm)
		},
	})
	RegisterInterceptor(&Interceptor{
		SetDeadline: func(info Info, tm time.Time, raw func(t time.Time) error) error {
			fst.Equal(t, int32(2), atomic.AddInt32(&num, 1))
			fst.Equal(t, want, tm)
			return raw(tm)
		},
	})

	c1 := &net.TCPConn{}
	c2 := Wrap(c1, &Interceptor{
		SetDeadline: func(info Info, tm time.Time, raw func(t time.Time) error) error {
			fst.Equal(t, int32(3), atomic.AddInt32(&num, 1))
			fst.Equal(t, want, tm)
			return raw(tm)
		},
	})
	_ = c2.SetDeadline(want)
}

func BenchmarkConnInterceptor_Read(b *testing.B) {
	var id int
	var its []*Interceptor
	for i := 0; i < 10; i++ {
		hk1 := &Interceptor{
			Read: func(info Info, b []byte, raw func([]byte) (int, error)) (int, error) {
				id++
				return raw(b)
			},
		}

		its = append(its, hk1)
	}
	conn := Wrap(&net.TCPConn{}, its...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := make([]byte, 1)
		_, _ = conn.Read(bf)
	}
}

func BenchmarkConnInterceptor_AfterRead(b *testing.B) {
	var id int
	var its []*Interceptor
	for i := 0; i < 10; i++ {
		hk1 := &Interceptor{
			AfterRead: func(info Info, b []byte, readSize int, err error) {
				id++
			},
		}
		its = append(its, hk1)
	}
	conn := Wrap(&net.TCPConn{}, its...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := make([]byte, 1)
		_, _ = conn.Read(bf)
	}
}
