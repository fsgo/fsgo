// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/12

package grace

import (
	"context"
	"fmt"
	"net"
	"os"
)

// Resource 支持 grace 的资源
type Resource interface {
	// Open 打开文件，执行后立即返回
	Open(ctx context.Context) error

	// File 资源的文件，父进程使用，以将file传递给子进程
	File() (*os.File, error)

	// SetFile 设置文件,子进程使用
	SetFile(file *os.File) error

	// String 资源的描述
	String() string
}

// Consumer 资源消费者
type Consumer interface {
	Bind(res Resource)

	// Start 开始运行 同步、阻塞
	Start(ctx context.Context) error

	// Stop 关闭
	Stop(ctx context.Context) error

	// String 资源的描述
	String() string
}

type filer interface {
	File() (*os.File, error)
}

// ListenerResource server 类型的资源
type ListenerResource struct {
	NetWork string
	Address string

	file *os.File
}

// Open 启动
func (s *ListenerResource) Open(ctx context.Context) error {
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, s.NetWork, s.Address)
	if err != nil {
		return err
	}
	if ff, ok := l.(filer); ok {
		f, err1 := ff.File()
		if err1 != nil {
			return err1
		}
		s.file = f
	}
	return nil
}

// File 获取文件句柄
func (s *ListenerResource) File() (*os.File, error) {
	if s.file == nil {
		return nil, fmt.Errorf("no file, Open or SetFile first")
	}
	return s.file, nil
}

// SetFile 设置文件句柄
func (s *ListenerResource) SetFile(file *os.File) error {
	if file == nil {
		return fmt.Errorf("file is nil")
	}
	s.file = file

	return nil
}

// String 格式化的描述信息，打印日志使用
func (s *ListenerResource) String() string {
	return fmt.Sprintf("NetWork=%q Address=%q", s.NetWork, s.Address)
}

var _ Resource = (*ListenerResource)(nil)
