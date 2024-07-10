// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/11/13

package fshttp

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fsgo/fst"
)

func TestTransport_RoundTrip(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	t.Run("http server", func(t *testing.T) {
		ts := httptest.NewServer(mux)
		defer ts.Close()
		cli := &http.Client{
			Transport: &Transport{},
			Timeout:   time.Second,
		}
		resp, err := cli.Get(ts.URL)
		fst.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		fst.NoError(t, err)
		fst.NoError(t, resp.Body.Close())
		fst.Equal(t, "hello", string(body))
	})

	t.Run("https server", func(t *testing.T) {
		ts := httptest.NewTLSServer(mux)
		defer ts.Close()
		cli := &http.Client{
			Transport: &Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: time.Second,
		}
		resp, err := cli.Get(ts.URL)
		fst.Nil(t, err)

		body, err := io.ReadAll(resp.Body)
		fst.Nil(t, err)
		fst.Nil(t, resp.Body.Close())
		fst.Equal(t, "hello", string(body))
	})
}
