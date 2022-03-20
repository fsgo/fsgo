// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/3/18

package fsotel

import (
	"context"
	"net"

	"github.com/fsgo/fsgo/fsnet"
	"go.opentelemetry.io/otel/attribute"
)

var DialerTracer = &fsnet.DialerInterceptor{
	DialContext: func(ctx context.Context, network string, address string, invoker fsnet.DialContextFunc) (conn net.Conn, err error) {
		ctx, span := tracer.Start(ctx, "Dial")
		if span.IsRecording() {
			span.SetAttributes(
				attribute.String("network", network),
				attribute.String("address", address),
			)
		}
		defer func() {
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
		}()
		return invoker(ctx, network, address)
	},
}
