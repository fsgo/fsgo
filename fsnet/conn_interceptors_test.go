// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/20

package fsnet

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnReadBytesHook(t *testing.T) {
	t.Run("read fail", func(t *testing.T) {
		c1 := &net.TCPConn{}
		ch := NewConnReadBytesHook()
		c2 := WrapConn(c1, ch.ConnInterceptor())
		bf := make([]byte, 1024)
		_, err := c2.Read(bf)
		assert.NotNil(t, t, err)
		assert.Equal(t, len(ch.ReadBytes()), 0)
	})

	t.Run("read success", func(t *testing.T) {
		w, r := net.Pipe()
		defer w.Close()
		defer r.Close()

		ch := NewConnReadBytesHook()
		c2 := WrapConn(r, ch.ConnInterceptor())

		want := []byte("hello")
		go func() {
			if _, err := w.Write(want); err != nil {
				panic(err)
			}
		}()
		bf := make([]byte, 1024)
		n, err := c2.Read(bf)
		assert.Nil(t, err)
		assert.Equal(t, want, bf[:n])

		assert.Equal(t, ch.ReadBytes(), want)

		ch.Reset()
		assert.Len(t, ch.ReadBytes(), 0)
	})
}

func TestConnWriteBytesHook(t *testing.T) {
	t.Run("write fail", func(t *testing.T) {
		c1 := &net.TCPConn{}
		ch := NewConnWriteBytesInterceptor()
		c2 := WrapConn(c1, ch.ConnInterceptor())
		_, err := c2.Write([]byte("hello"))
		assert.NotNil(t, err)
		assert.Len(t, ch.WriteBytes(), 0)
	})

	t.Run("write success", func(t *testing.T) {
		w, r := net.Pipe()
		defer w.Close()
		defer r.Close()

		ch := NewConnWriteBytesInterceptor()
		c2 := WrapConn(r, ch.ConnInterceptor())

		go func() {
			bf := make([]byte, 1024)
			if _, err := w.Read(bf); err != nil {
				panic(err)
			}
		}()

		want := []byte("hello")
		_, err := c2.Write(want)
		assert.Nil(t, err)
		assert.Equal(t, ch.WriteBytes(), want)

		ch.Reset()
		assert.Len(t, ch.WriteBytes(), 0)
	})
}

func Test_hooks(t *testing.T) {
	DefaultDialer = &Dialer{}
	defer func() {
		DefaultDialer = &Dialer{}
	}()

	rt := http.NewServeMux()
	want := []byte("HelloFsNet")
	rt.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err1 := writer.Write(want)
		assert.Nil(t, err1)
	})
	ts := httptest.NewServer(rt)
	defer ts.Close()

	statHK := NewConnStatInterceptor()
	readHK := NewConnReadBytesHook()
	globalHook := NewConnDialerInterceptor(readHK.ConnInterceptor())
	MustRegisterDialerInterceptor(statHK.DialerInterceptor(), globalHook)

	tr := &http.Transport{
		DialContext:           DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	defer tr.Clone()

	req := httptest.NewRequest("get", ts.URL, nil)
	resp, err := tr.RoundTrip(req)
	assert.Nil(t, err)

	defer resp.Body.Close()

	t.Run("body", func(t *testing.T) {
		got, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("statHK", func(t *testing.T) {
		assert.NotEqual(t, 0, statHK.ReadCost())
		assert.NotEqual(t, 0, statHK.WriteCost())
	})

	t.Run("ReadBytes", func(t *testing.T) {
		got := readHK.ReadBytes()
		assert.Contains(t, string(got), string(want))
	})

}
