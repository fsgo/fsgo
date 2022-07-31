// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/3/18

package fsotel

import (
	"context"
	"net"

	"github.com/fsgo/fsgo/fsnet/fsdialer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DialerTracer 对拨号提供 otel 支持
var DialerTracer = &fsdialer.Interceptor{
	BeforeDialContext: func(ctx context.Context, net string, addr string) (c context.Context, n string, a string) {
		ctx, span := tracer.Start(ctx, "Dial")
		if span.IsRecording() {
			span.SetAttributes(
				attribute.String("network", net),
				attribute.String("address", addr),
			)
		}
		return ctx, net, addr
	},
	AfterDialContext: func(ctx context.Context, net string, addr string, conn net.Conn, err error) (net.Conn, error) {
		span := trace.SpanFromContext(ctx)
		if err != nil {
			span.RecordError(err)
		}
		if span.IsRecording() && conn != nil {
			span.SetAttributes(
				attribute.String("LocalAddr", conn.LocalAddr().String()),
				attribute.String("RemoteAddr", conn.RemoteAddr().String()),
			)
		}
		span.End()
		return conn, err
	},
}
