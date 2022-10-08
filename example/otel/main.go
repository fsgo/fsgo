// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/3/20

package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"

	"github.com/fsgo/fsgo/extra/fsotel"
	"github.com/fsgo/fsgo/fsnet/fsdialer"
	"github.com/fsgo/fsgo/fsnet/fshttp"
	"github.com/fsgo/fsgo/fsnet/fsresolver"
	"github.com/uptrace/opentelemetry-go-extra/otelplay"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var addr = flag.String("addr", "127.0.0.1:8080", "http server addr")

func main() {
	flag.Parse()

	shutdown := otelplay.ConfigureOpentelemetry(context.Background())
	defer shutdown()

	fsdialer.MustRegisterInterceptor(fsotel.DialerTracer)
	fsresolver.MustRegisterInterceptor(fsotel.ResolverTracer)

	ser := &http.Server{
		Addr:    *addr,
		Handler: otelhttp.NewHandler(http.HandlerFunc(handler), "httpapi"),
	}
	log.Println("visit: http://" + *addr + "/")
	log.Fatalln(ser.ListenAndServe())
}

func handlerErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

var tracer = otel.Tracer("example")

func handler(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("url")

	ctx, span := tracer.Start(r.Context(), "httpHandler")
	defer span.End()

	span.SetAttributes(attribute.String("url", u))
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		handlerErr(w, err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		handlerErr(w, err)
		return
	}

	defer resp.Body.Close()
	bf, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		handlerErr(w, err)
		return
	}

	w.Write(bf)
}

var client = &http.Client{
	Transport: &fshttp.Transport{},
}
