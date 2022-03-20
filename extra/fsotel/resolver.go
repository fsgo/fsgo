// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2022/2/27

package fsotel

import (
	"context"
	"net"

	"github.com/fsgo/fsgo/fsnet"
	"go.opentelemetry.io/otel/attribute"
)

var ResolverTracer = &fsnet.ResolverInterceptor{
	LookupIP: func(ctx context.Context, network, host string, invoker fsnet.LookupIPFunc) (ret []net.IP, err error) {
		ctx, span := tracer.Start(ctx, "LookupIP")
		if span.IsRecording() {
			span.SetAttributes(
				attribute.String("network", network),
				attribute.String("host", host),
			)
		}
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
			if span.IsRecording() {
				ips := make([]string, len(ret))
				for i := 0; i < len(ret); i++ {
					ips[i] = ret[i].String()
				}
				span.SetAttributes(attribute.StringSlice("ips", ips))
			}
			span.End()
		}()
		return invoker(ctx, network, host)
	},
}
