// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/1/12

package grace

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

// Resource 支持 grace 的资源
//
// 由于在 unix 上所有资源都可以用文件来表示
// 所以这里就使用最底层的文件类型
// 当 我们需要 net.Listen 的时候，也可以将文件转换为 net.Listen
// 也就是下面的  ListenerResource
type Resource interface {
	// Open 打开文件，执行后立即返回
	Open(ctx context.Context) error

	// File 资源的文件，父进程使用，以将file传递给子进程
	File(ctx context.Context) (*os.File, error)

	// Listener 获取 Listener
	Listener(ctx context.Context) (net.Listener, error)

	// String 资源的描述
	String() string
}

// TrySetListener 给 Resource 重新设置新的 Listener
func TrySetListener(res Resource, l net.Listener) error {
	if rl, ok := res.(canSetListener); ok {
		rl.SetListener(l)
		return nil
	}
	return errors.New("cannot Set Listener")
}

var _ Resource = (*listenDSN)(nil)

type listenDSN struct {
	listener net.Listener

	file *os.File
	DSN  string

	Index  int
	opened bool
}

type filer interface {
	File() (*os.File, error)
}

func (d *listenDSN) Open(ctx context.Context) error {
	if d.opened {
		return nil
	}

	d.opened = true

	if IsSubProcess() {
		return d.openFileListener(ctx)
	}
	return d.openNetListener(ctx)
}

func (d *listenDSN) openNetListener(ctx context.Context) error {
	network, address, err := d.parser()
	if err != nil {
		return err
	}
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, network, address)
	if err != nil {
		return err
	}
	d.listener = l
	if ff, ok := l.(filer); ok {
		f, err1 := ff.File()
		if err1 != nil {
			return err1
		}
		d.file = f
		return nil
	}
	return fmt.Errorf("listener（%T） has not implement File()(*os.File,error)", l)
}

func (d *listenDSN) openFileListener(_ context.Context) error {
	d.file = os.NewFile(uintptr(3+d.Index), "")
	l, err := net.FileListener(d.file)
	d.listener = l
	return err
}

func (d *listenDSN) File(ctx context.Context) (*os.File, error) {
	if err := d.Open(ctx); err != nil {
		return nil, err
	}
	if d.file != nil {
		return d.file, nil
	}
	return nil, errors.New("file not exists")
}

func (d *listenDSN) String() string {
	return d.DSN
}

func (d *listenDSN) Listener(ctx context.Context) (net.Listener, error) {
	if err := d.Open(ctx); err != nil {
		return nil, err
	}
	if d.listener != nil {
		return d.listener, nil
	}
	return nil, errors.New("listener not exists")
}

type canSetListener interface {
	SetListener(l net.Listener)
}

var _ canSetListener = (*listenDSN)(nil)

func (d *listenDSN) SetListener(l net.Listener) {
	d.listener = l
}

func (d *listenDSN) parser() (network, address string, err error) {
	arr := strings.SplitN(d.DSN, "@", 2)
	if len(arr) != 2 {
		return "", "", fmt.Errorf("wrong dsn format: %q", d.DSN)
	}
	return arr[0], arr[1], nil
}

var drivers = map[string]ResourceDriverFunc{}

// ResourceDriverFunc 解析 DSN 配置
// dsn like "tcp@127.0.0.1:8080"
type ResourceDriverFunc func(index int, dsn string) (Resource, error)

// RegisterResourceDriver 注册新的资源解析协议
func RegisterResourceDriver(scheme string, fn ResourceDriverFunc) {
	drivers[scheme] = fn
}

func init() {
	// 将所有的网络类型全部注册
	// 也可以注册其他自定义类型的
	RegisterResourceDriver("tcp", netResourceDrive)
	RegisterResourceDriver("tcp4", netResourceDrive)
	RegisterResourceDriver("tcp6", netResourceDrive)
	RegisterResourceDriver("udp", netResourceDrive)
	RegisterResourceDriver("udp4", netResourceDrive)
	RegisterResourceDriver("udp6", netResourceDrive)
	RegisterResourceDriver("unix", netResourceDrive)
	RegisterResourceDriver("unixpacket", netResourceDrive)
}

func netResourceDrive(index int, dsn string) (Resource, error) {
	ds := &listenDSN{
		Index: index,
		DSN:   dsn,
	}
	return ds, nil
}

// ParserListenDSN 通过 DSN 获取一个 Resource
func ParserListenDSN(index int, dsn string) (Resource, error) {
	if index < 0 {
		return nil, fmt.Errorf("index should >=0, got=%d", index)
	}
	arr := strings.SplitN(dsn, "@", 2)
	if len(arr) != 2 {
		return nil, fmt.Errorf("wrong dsn=%q format", dsn)
	}
	scheme := arr[0]
	driverFunc, has := drivers[scheme]
	if !has {
		return nil, fmt.Errorf("scheme=%q not support", scheme)
	}
	return driverFunc(index, dsn)
}

var globalResourceID int

// NextResource 获取下一个资源
func NextResource(listen string) Resource {
	res, err := ParserListenDSN(globalResourceID, listen)
	if err != nil {
		panic("parser dsn failed:" + err.Error())
	}
	globalResourceID++
	return res
}
