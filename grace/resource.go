/*
 * Copyright(C) 2021 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/1/12
 */

package grace

import (
	"context"
	"fmt"
	"net"
	"os"
)

// Resource grace的资源
type Resource interface {
	// Open 打开文件，执行后立即返回
	Open(ctx context.Context) error

	// File 资源的文件，父进程使用，以将file传递给子进程
	File() (*os.File, error)

	// SetFile 设置文件,子进程使用
	SetFile(file *os.File) error

	// Start 开始运行 同步、阻塞
	Start(ctx context.Context) error

	// Stop 关闭
	Stop(ctx context.Context) error

	// String 资源的描述
	String() string
}

// Server server 类型
type Server interface {
	Serve(l net.Listener) error
	Shutdown(ctx context.Context) error
}

type filer interface {
	File() (*os.File, error)
}

// ServerResource server 类型的资源
type ServerResource struct {
	Server Server

	NetWork string
	Address string

	listener net.Listener
	file     *os.File
}

func (s *ServerResource) Open(ctx context.Context) error {
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, s.NetWork, s.Address)
	if err != nil {
		return err
	}
	s.listener = l

	if ff, ok := l.(filer); ok {
		f, err1 := ff.File()
		if err1 != nil {
			return err1
		}
		s.file = f
	}

	return nil
}

func (s *ServerResource) File() (*os.File, error) {
	if s.file == nil {
		return nil, fmt.Errorf("no file, Open or SetFile first")
	}
	return s.file, nil
}

func (s *ServerResource) SetFile(file *os.File) error {
	if file == nil {
		return fmt.Errorf("file is nil")
	}
	s.file = file

	l, err := net.FileListener(file)
	if err != nil {
		return err
	}
	s.listener = l
	return nil
}

func (s *ServerResource) Start(ctx context.Context) error {
	return s.Server.Serve(s.listener)
}

func (s *ServerResource) Stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

func (s *ServerResource) String() string {
	return fmt.Sprintf("NetWork=%q Address=%q", s.NetWork, s.Address)
}

var _ Resource = (*ServerResource)(nil)
