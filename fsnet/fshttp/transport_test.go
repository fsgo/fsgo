// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/11/13

package fshttp

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
		require.Nil(t, err)

		body, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err)
		require.Nil(t, resp.Body.Close())
		require.Equal(t, "hello", string(body))
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
		require.Nil(t, err)

		body, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err)
		require.Nil(t, resp.Body.Close())
		require.Equal(t, "hello", string(body))
	})
}
