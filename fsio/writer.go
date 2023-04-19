// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/29

package fsio

import (
	"io"
	"sync"
)

// ResetWriter writer can reset
type ResetWriter interface {
	io.Writer
	Reset(w io.Writer)
}

// Flusher can flush
type Flusher interface {
	Flush() error
}

// TryFlush try flush
func TryFlush(w io.Writer) error {
	if fw, ok := w.(Flusher); ok {
		return fw.Flush()
	}
	return nil
}

// NewResetWriter wrap writer to ResetWriter
func NewResetWriter(w io.Writer) ResetWriter {
	if rw, ok := w.(ResetWriter); ok {
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

// NewMutexWriter wrap a writer with a mutex
func NewMutexWriter(w io.Writer) *MutexWriter {
	return &MutexWriter{
		Writer: w,
	}
}

type MutexWriter struct {
	Writer io.Writer
	mu     sync.RWMutex
}

func (w *MutexWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Writer.Write(b)
}

func (w *MutexWriter) WithRLock(fn func(w io.Writer)) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	fn(w.Writer)
}

func (w *MutexWriter) WithLock(fn func(w io.Writer)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fn(w.Writer)
}

// WriteStatus status for Write
type WriteStatus struct {
	Err   error
	Wrote int
}
