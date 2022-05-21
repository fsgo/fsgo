// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package fsconn

import (
	"io"
	"sync"
)

// ConnCopy 实现对网络连接读写数据的复制
type ConnCopy struct {
	interceptor *Interceptor
	once        sync.Once

	// ReadTo 将 Read 到的数据写入此处,比如 os.Stdout
	ReadTo io.Writer

	// WriterTo 将 Writer 的数据写入此处，比如 os.Stdout
	WriterTo io.Writer
}

func (cc *ConnCopy) init() {
	cc.interceptor = &Interceptor{
		AfterRead: func(b []byte, readSize int, err error) {
			if readSize > 0 && cc.ReadTo != nil {
				_, _ = cc.ReadTo.Write(b[:readSize])
			}
		},
		AfterWrite: func(b []byte, wroteSize int, err error) {
			if wroteSize > 0 && cc.ReadTo != nil {
				_, _ = cc.WriterTo.Write(b[:wroteSize])
			}
		},
	}
}

// Interceptor 获取 Interceptor 实例
func (cc *ConnCopy) Interceptor() *Interceptor {
	cc.once.Do(cc.init)
	return cc.interceptor
}
