// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/2

package fsio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var _ io.Writer = (*PrintByteWriter)(nil)

type PrintByteWriter struct {
	Name string

	// Out 实际输出目标
	Out io.Writer

	// LineMax 单行最大字符数，可选，默认 40
	LineMax int

	id  atomic.Int64
	mux sync.Mutex
}

func (pb *PrintByteWriter) getLineMax() int {
	if pb.LineMax > 0 {
		return pb.LineMax
	}
	return 40
}

func (pb *PrintByteWriter) getOut() io.Writer {
	if pb.Out != nil {
		return pb.Out
	}
	return os.Stdout
}

func (pb *PrintByteWriter) Write(p []byte) (n int, err error) {
	return pb.WriteWithMeta(p, "")
}

func (pb *PrintByteWriter) WriteWithMeta(p []byte, meta string) (n int, err error) {
	pb.mux.Lock()
	defer pb.mux.Unlock()

	total := len(p)
	maxLen := pb.getLineMax()

	bf := &bytes.Buffer{}
	fmt.Fprintf(bf, "[%s][%d][Len=%d] %s %s\n", pb.Name, pb.id.Add(1), len(p), meta, time.Now().Format(time.DateTime+".99999"))
	lineNo := -1
	startIndex := 0
	for len(p) > 0 {
		lineNo++
		end := len(p)
		if len(p) > maxLen {
			end = maxLen
		}

		fmt.Fprintf(bf, "%3d\t%q [Pos: %d - %d]\n", lineNo, p[:end], startIndex, startIndex+end)
		fmt.Fprintf(bf, "\t%v\n", pb.format(p[:end]))
		p = p[end:]
		startIndex += end
	}
	_, err = pb.getOut().Write(bf.Bytes())
	return total, err
}

func (pb *PrintByteWriter) format(bf []byte) []string {
	ss := make([]string, 0, len(bf))
	for _, b := range bf {
		s := fmt.Sprint(b)
		ss = append(ss, fmt.Sprintf("%3s", s))
	}
	return ss
}
