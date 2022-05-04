// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

import (
	"bytes"
	"fmt"
)

func newBufWriter() *bufWriter {
	return &bufWriter{
		bf: &bytes.Buffer{},
	}
}

type bufWriter struct {
	err error
	bf  *bytes.Buffer
}

func (w *bufWriter) WriteWithSep(sep string, values ...any) {
	if w.err != nil {
		return
	}

	for i := 0; i < len(values); i++ {
		switch v := values[i].(type) {
		case []byte:
			w.writeBytes(sep, v)
		case string:
			w.writeString(sep, v)
		case HTML:
			h, e1 := v.HTML()
			if e1 != nil {
				w.err = e1
			} else {
				w.writeBytes(sep, h)
			}
		case error:
			w.err = v
		default:
			panic(fmt.Sprintf("not support type:%T", v))
		}

		if w.err != nil {
			return
		}
	}
}

func (w *bufWriter) writeBytes(sep string, bf []byte) {
	if len(sep) > 0 && len(bf) > 0 {
		_, w.err = w.bf.WriteString(sep)
		if w.err != nil {
			return
		}
	}

	if len(bf) > 0 {
		_, w.err = w.bf.Write(bf)
	}
}
func (w *bufWriter) writeString(sep string, bf string) {
	if len(sep) > 0 && len(bf) > 0 {
		_, w.err = w.bf.WriteString(sep)
		if w.err != nil {
			return
		}
	}

	if len(bf) > 0 {
		_, w.err = w.bf.WriteString(bf)
	}
}

func (w *bufWriter) Write(values ...any) {
	w.WriteWithSep("", values...)
}

func (w *bufWriter) HTML() ([]byte, error) {
	return w.bf.Bytes(), w.err
}
