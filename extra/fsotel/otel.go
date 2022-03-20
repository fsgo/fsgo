// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/3/18

package fsotel

import (
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/fsgo/fsgo/fsnet")
