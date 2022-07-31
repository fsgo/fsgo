// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2022/2/27

package fsotel

import (
	"context"
	"net"

	"github.com/fsgo/fsgo/fsnet/fsresolver"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ResolverTracer 对 域名解析提供 otel 支持
var ResolverTracer = &fsresolver.Interceptor{
	BeforeLookupIP: func(ctx context.Context, network, host string) (c context.Context, n, h string) {
		ctx, span := tracer.Start(ctx, "LookupIP")
		if span.IsRecording() {
			span.SetAttributes(
				attribute.String("network", network),
				attribute.String("host", host),
			)
		}
		return ctx, network, host
	},
	AfterLookupIP: func(ctx context.Context, network, host string, ips []net.IP, err error) ([]net.IP, error) {
		span := trace.SpanFromContext(ctx)
		if err != nil {
			span.RecordError(err)
		}
		if span.IsRecording() {
			result := make([]string, len(ips))
			for i := 0; i < len(ips); i++ {
				result[i] = ips[i].String()
			}
			span.SetAttributes(attribute.StringSlice("ips", result))
		}
		span.End()
		return ips, err
	},
}
