// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/29

package fsio

import (
	"io"
	"sync"
)

type ResetWriter interface {
	io.Writer
	Reset(w io.Writer)
}

func TryFlush(w io.Writer) error {
	if fw, ok := w.(interface{ Flush() error }); ok {
		return fw.Flush()
	}
	return nil
}

func NewResetWriter(w io.Writer) ResetWriter {
	if rw, ok := w.(*resetWriter); ok {
		return rw
	}
	return &resetWriter{
		raw: w,
	}
}

type resetWriter struct {
	raw io.Writer
	mux sync.RWMutex
}

func (w *resetWriter) Write(p []byte) (n int, err error) {
	w.mux.RLock()
	raw := w.raw
	w.mux.RUnlock()
	return raw.Write(p)
}

func (w *resetWriter) Reset(raw io.Writer) {
	w.mux.Lock()
	w.raw = raw
	w.mux.Unlock()
}
