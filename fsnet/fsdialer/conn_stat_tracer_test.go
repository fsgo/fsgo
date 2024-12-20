// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsdialer

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fsgo/fst"

	"github.com/fsgo/fsgo/fsnet/fsconn"
)

func TestTraces(t *testing.T) {
	Default = &Simple{}
	defer func() {
		Default = &Simple{}
	}()

	rt := http.NewServeMux()
	want := []byte("HelloFsNet")
	rt.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err1 := writer.Write(want)
		fst.Nil(t, err1)
	})
	ts := httptest.NewServer(rt)
	defer ts.Close()

	statHK := &ConnStatTracer{}
	readHK := &fsconn.ReadTracer{}
	globalHook := TransConnInterceptor(readHK.ConnInterceptor())
	MustRegisterInterceptor(statHK.DialerInterceptor(), globalHook)

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
	fst.Nil(t, err)

	defer resp.Body.Close()

	t.Run("body", func(t *testing.T) {
		got, err := io.ReadAll(resp.Body)
		fst.Nil(t, err)
		fst.Equal(t, want, got)
	})

	t.Run("statHK", func(t *testing.T) {
		fst.NotEqual(t, 0, statHK.ReadCost())
		fst.NotEqual(t, 0, statHK.WriteCost())
	})

	t.Run("ReadBytes", func(t *testing.T) {
		got := readHK.ReadBytes()
		fst.Contains(t, string(got), string(want))
	})
}
