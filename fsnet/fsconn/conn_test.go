// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsconn

import (
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConn(t *testing.T) {
	t.Run("its", func(t *testing.T) {
		w, r := net.Pipe()
		var readTotal, writeTotal int
		var closeNum int

		var readIndex int

		tr := &Interceptor{
			Read: func(b []byte, raw func([]byte) (int, error)) (n int, err error) {
				defer func() {
					readTotal += n
				}()

				readIndex++
				assert.Equal(t, 2, readIndex)

				return raw(b)
			},
			Write: func(b []byte, raw func([]byte) (int, error)) (n int, err error) {
				defer func() {
					writeTotal += n
				}()
				return raw(b)
			},
			RemoteAddr: func(raw func() net.Addr) net.Addr {
				// return the intercepted addr
				return &net.TCPAddr{}
			},
			Close: func(raw func() error) error {
				closeNum++
				return raw()
			},
		}

		tr2 := &Interceptor{
			Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
				readIndex++
				assert.Equal(t, 1, readIndex)
				return raw(b)
			},
		}

		w1 := WithInterceptor(w, tr, tr2)
		r1 := WithInterceptor(r)

		msg := []byte("hello")
		go func() {
			_, _ = w1.Write(msg)
		}()
		buf := make([]byte, 128)

		n, err := r1.Read(buf)
		assert.Nil(t, err)
		assert.Equal(t, len(msg), n)
		assert.Equal(t, string(msg), string(buf[:n]))

		assert.Equal(t, "pipe", w1.LocalAddr().Network())
		assert.Equal(t, "tcp", w1.RemoteAddr().Network())

		t.Run("Close", func(t *testing.T) {
			assert.Nil(t, w1.Close())

			assert.Equal(t, 1, closeNum)
		})

	})
}

func TestNewConn_merge(t *testing.T) {
	var id int
	hk1 := &Interceptor{
		Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
			// 先注册的先执行
			id++
			assert.Equal(t, 1, id)
			return raw(b)
		},
		AfterRead: func(b []byte, readSize int, err error) {
			id++
			assert.Equal(t, 4, id)
		},
	}
	nc := WithInterceptor(&net.TCPConn{}, hk1)

	hk2 := &Interceptor{
		Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
			id++
			assert.Equal(t, 2, id)
			return raw(b)
		},
		AfterRead: func(b []byte, readSize int, err error) {
			id++
			assert.Equal(t, 5, id)
		},
	}

	hk3 := &Interceptor{
		Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
			id++
			assert.Equal(t, 3, id)
			return raw(b)
		},
	}
	hk4 := &Interceptor{
		AfterRead: func(b []byte, readSize int, err error) {
			id++
			assert.Equal(t, 6, id)
		},
	}
	nc1 := WithInterceptor(nc, hk2, hk3, hk4)
	assert.NotEqual(t, nc, nc1)
	bf := make([]byte, 1)
	_, _ = nc1.Read(bf)
	assert.Equal(t, 6, id)
}

func TestOriginConn(t *testing.T) {
	c1 := &net.TCPConn{}
	c2 := WithInterceptor(c1)

	assert.Equal(t, c1, OriginConn(c2))
	assert.Equal(t, c1, OriginConn(c1))
}

func Test_connInterceptors_CallSetDeadline(t *testing.T) {
	want := time.Now()

	var num int32
	RegisterInterceptor(&Interceptor{
		SetDeadline: func(tm time.Time, raw func(tm time.Time) error) error {
			require.Equal(t, int32(1), atomic.AddInt32(&num, 1))
			require.Equal(t, want, tm)
			return raw(tm)
		},
	})
	RegisterInterceptor(&Interceptor{
		SetDeadline: func(tm time.Time, raw func(t time.Time) error) error {
			require.Equal(t, int32(2), atomic.AddInt32(&num, 1))
			require.Equal(t, want, tm)
			return raw(tm)
		},
	})

	c1 := &net.TCPConn{}
	c2 := WithInterceptor(c1, &Interceptor{
		SetDeadline: func(tm time.Time, raw func(t time.Time) error) error {
			require.Equal(t, int32(3), atomic.AddInt32(&num, 1))
			require.Equal(t, want, tm)
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
			Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
				id++
				return raw(b)
			},
		}

		its = append(its, hk1)
	}
	conn := WithInterceptor(&net.TCPConn{}, its...)
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
			AfterRead: func(b []byte, readSize int, err error) {
				id++
			},
		}
		its = append(its, hk1)
	}
	conn := WithInterceptor(&net.TCPConn{}, its...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := make([]byte, 1)
		_, _ = conn.Read(bf)
	}
}
