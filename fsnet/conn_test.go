// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/14

package fsnet

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConn(t *testing.T) {
	t.Run("hooks", func(t *testing.T) {
		w, r := net.Pipe()
		var readTotal, writeTotal int
		var closeNum int

		var readIndex int

		tr := &ConnInterceptor{
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
				// return the hooked addr
				return &net.TCPAddr{}
			},
			Close: func(raw func() error) error {
				closeNum++
				return raw()
			},
		}

		tr2 := &ConnInterceptor{
			Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
				readIndex++
				assert.Equal(t, 1, readIndex)
				return raw(b)
			},
		}

		stHook := NewConnStatInterceptor()

		w1 := NewConn(w, tr, tr2, stHook.ConnInterceptor())
		r1 := NewConn(r)

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

		t.Run("StatHook", func(t *testing.T) {
			assert.Greater(t, stHook.WriteSize(), int64(0))

			assert.Greater(t, int(stHook.WriteCost()), 0)

			stHook.Reset()
			assert.Equal(t, int64(0), stHook.WriteSize())
		})
	})

}

func TestNewConn_merge(t *testing.T) {
	var id int
	hk1 := &ConnInterceptor{
		Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
			// 先注册的后执行
			id++
			assert.Equal(t, 2, id)
			return raw(b)
		},
	}
	nc := NewConn(&net.TCPConn{}, hk1)

	hk2 := &ConnInterceptor{
		Read: func(b []byte, raw func([]byte) (int, error)) (int, error) {
			id++
			assert.Equal(t, 1, id)
			return raw(b)
		},
	}
	nc1 := NewConn(nc, hk2)
	assert.Equal(t, nc, nc1)
	bf := make([]byte, 1)
	_, _ = nc1.Read(bf)
}

func TestOriginConn(t *testing.T) {
	c1 := &net.TCPConn{}
	c2 := NewConn(c1)

	assert.Equal(t, c1, OriginConn(c2))
	assert.Equal(t, c1, OriginConn(c1))
}
