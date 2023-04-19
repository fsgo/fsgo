// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package conndump

import (
	"encoding/binary"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsgo/fsenv"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/fsgo/fsgo/fsnet/fsconn"
	"github.com/fsgo/fsgo/fstypes"
)

// Dumper 流量 dump
type Dumper struct {
	clientOutFile *fsfs.Rotator
	serverOutFile *fsfs.Rotator

	clientIt *fsconn.Interceptor
	serverIt *fsconn.Interceptor

	conns map[fsconn.Info]*connInfo

	// RotatorConfig 可选，用于配置 dump 的 Rotator
	RotatorConfig func(client bool, r *fsfs.Rotator)

	// DataDir 数据存放目录，可选
	DataDir string

	serverReadStatus  fstypes.GroupEnableStatus
	serverWriteStatus fstypes.GroupEnableStatus

	clientWriteStatus fstypes.GroupEnableStatus

	clientReadStatus fstypes.GroupEnableStatus

	connID   int64
	connsMux sync.RWMutex // 只用于管理连接

	once sync.Once
}

// Setup 初始化配置
func (d *Dumper) initOnce() error {
	var err error
	d.once.Do(func() {
		d.conns = make(map[fsconn.Info]*connInfo)

		d.clientIt = d.newInterceptor(true)
		d.serverIt = d.newInterceptor(false)

		if d.serverOutFile, err = d.initOutFile(false); err != nil {
			return
		}

		if d.clientOutFile, err = d.initOutFile(true); err != nil {
			return
		}
	})
	return err
}

func (d *Dumper) newInterceptor(isClient bool) *fsconn.Interceptor {
	it := &fsconn.Interceptor{
		AfterWrite: func(info fsconn.Info, b []byte, wroteSize int, err error) {
			d.dumpWrite(isClient, info, b, wroteSize, err)
		},
		AfterRead: func(info fsconn.Info, b []byte, readSize int, err error) {
			d.dumpRead(isClient, info, b, readSize, err)
		},
		AfterClose: func(info fsconn.Info, err error) {
			d.dumpClose(isClient, info, err)
		},
	}
	return it
}

func (d *Dumper) initOutFile(client bool) (*fsfs.Rotator, error) {
	subDir := "server"
	if client {
		subDir = "client"
	}
	f := &fsfs.Rotator{
		Path:     filepath.Join(d.getDataDir(), subDir, "dump.pb"),
		ExtRule:  "10minute",
		MaxFiles: 24,
	}

	if d.RotatorConfig != nil {
		d.RotatorConfig(client, f)
	}

	if err := f.Init(); err != nil {
		return nil, err
	}
	return f, nil
}

func (d *Dumper) getDataDir() string {
	if len(d.DataDir) != 0 {
		return d.DataDir
	}
	return filepath.Join(fsenv.DataRootDir(), "rpcdump")
}

// ClientConnInterceptor 返回 client 的 conn Interceptor
func (d *Dumper) ClientConnInterceptor() *fsconn.Interceptor {
	if err := d.initOnce(); err != nil {
		panic(err)
	}
	return d.clientIt
}

// ServerConnInterceptor 返回 server 的 conn Interceptor
//
// 对于 server，建议使用 WrapListener 方法，而不是直接使用这个方法
func (d *Dumper) ServerConnInterceptor() *fsconn.Interceptor {
	if err := d.initOnce(); err != nil {
		panic(err)
	}
	return d.serverIt
}

// DumpClientRead 设置是否允许 dump Read 的数据
func (d *Dumper) DumpClientRead(name string, enable bool) {
	d.clientReadStatus.SetEnable(name, enable)
}

// DumpClientWrite 设置是否允许 dump Write 的数据
func (d *Dumper) DumpClientWrite(name string, enable bool) {
	d.clientWriteStatus.SetEnable(name, enable)
}

// DumpAllClientRead 设置所有的 client 是否允许 dump Read 的数据
func (d *Dumper) DumpAllClientRead(enable bool) {
	d.clientReadStatus.SetAllEnable(enable)
}

// DumpAllClientWrite 设置所有的 client 是否允许 dump Write 的数据
func (d *Dumper) DumpAllClientWrite(enable bool) {
	d.clientReadStatus.SetAllEnable(enable)
}

// DumpServerRead 设置所有 server 是否都允许 dump Read 的数据
func (d *Dumper) DumpServerRead(name string, enable bool) {
	d.serverReadStatus.SetEnable(name, enable)
}

// DumpServerWrite 设置所有 server 是否都允许 dump Write 的数据
func (d *Dumper) DumpServerWrite(name string, enable bool) {
	d.serverWriteStatus.SetEnable(name, enable)
}

// DumpAllServerRead 设置所有 server 是否都允许 dump Read 的数据
func (d *Dumper) DumpAllServerRead(enable bool) {
	d.serverReadStatus.SetAllEnable(enable)
}

// DumpAllServerWrite 设置所有 server 是否都允许 dump Write 的数据
func (d *Dumper) DumpAllServerWrite(enable bool) {
	d.serverWriteStatus.SetAllEnable(enable)
}

// DumpAll 设置所有 server 和 client 是否都允许 dump
func (d *Dumper) DumpAll(enable bool) {
	d.DumpAllClientRead(enable)
	d.DumpAllClientWrite(enable)
	d.DumpAllServerRead(enable)
	d.DumpAllServerWrite(enable)
}

// WrapListener 封装一个 Listener，使得使用这个 Listener 的所有流量都支持 dump
func (d *Dumper) WrapListener(name string, l net.Listener) net.Listener {
	nl := &fsconn.Listener{
		Listener: l,
		AfterAccepts: []func(conn net.Conn) (net.Conn, error){
			func(conn net.Conn) (net.Conn, error) {
				c := fsconn.WithService(name, conn)
				return fsconn.WithInterceptor(c, d.ServerConnInterceptor()), nil
			},
		},
	}
	return nl
}

// dumpWrite dump conn 里写出的数据
func (d *Dumper) dumpWrite(isClient bool, conn fsconn.Info, b []byte, size int, err error) {
	if err != nil {
		return
	}
	name := service(conn)
	if isClient {
		if !d.clientWriteStatus.IsEnable(name) {
			return
		}
	} else {
		if !d.serverWriteStatus.IsEnable(name) {
			return
		}
	}
	d.doDumpReadWrite(isClient, conn, b, size, MessageAction_Write)
}

// dumpRead dump conn 里收到的数据
func (d *Dumper) dumpRead(isClient bool, conn fsconn.Info, b []byte, size int, err error) {
	if err != nil {
		return
	}
	name := service(conn)
	if isClient {
		if !d.clientReadStatus.IsEnable(name) {
			return
		}
	} else {
		if !d.serverReadStatus.IsEnable(name) {
			return
		}
	}
	d.doDumpReadWrite(isClient, conn, b, size, MessageAction_Read)
}

func (d *Dumper) doDumpReadWrite(isClient bool, conn fsconn.Info, b []byte, size int, tp MessageAction) {
	ci := d.getConnInfo(conn, true)
	msg := ci.newMessage(b, size, tp)
	d.writeMessage(isClient, msg)
}

const dumpMaxSize = 32 * 1024 * 1024

func (d *Dumper) writeMessage(isClient bool, msg proto.Message) {
	bf, err := proto.Marshal(msg)
	if err != nil || len(bf) > dumpMaxSize {
		return
	}
	b1 := make([]byte, len(bf)+4)
	binary.LittleEndian.PutUint32(b1, uint32(len(bf)))
	copy(b1[4:], bf)
	if isClient {
		_, _ = d.clientOutFile.Write(b1)
	} else {
		_, _ = d.serverOutFile.Write(b1)
	}
}

func (d *Dumper) nextConnID() int64 {
	return atomic.AddInt64(&d.connID, 1)
}

func (d *Dumper) getConnInfo(conn fsconn.Info, create bool) *connInfo {
	d.connsMux.RLock()
	info := d.conns[conn]
	d.connsMux.RUnlock()

	if info != nil || !create {
		return info
	}

	d.connsMux.Lock()
	defer d.connsMux.Unlock()

	info = d.conns[conn]
	if info != nil {
		return info
	}
	info = &connInfo{
		connID: d.nextConnID(),
		Conn:   conn,
	}
	d.conns[conn] = info
	return info
}

func (d *Dumper) dumpClose(isClient bool, info fsconn.Info, _ error) {
	ci := d.getConnInfo(info, false)
	if ci == nil {
		// 在此之前没有 Read 和 Write，直接 Close 的情况
		return
	}
	name := service(info)
	if isClient {
		if !d.clientReadStatus.IsEnable(name) && !d.clientWriteStatus.IsEnable(name) {
			return
		}
	} else {
		if !d.serverReadStatus.IsEnable(name) && !d.serverWriteStatus.IsEnable(name) {
			return
		}
	}

	msg := ci.newMessage(nil, 0, MessageAction_Close)
	d.writeMessage(isClient, msg)

	d.connsMux.Lock()
	defer d.connsMux.Unlock()
	delete(d.conns, info)
}

// Stop 停止
func (d *Dumper) Stop() {
	d.DumpAll(false)

	_ = d.clientOutFile.Close()
	_ = d.serverOutFile.Close()
}

type connInfo struct {
	Conn     fsconn.Info
	connID   int64
	subGroup int64
}

var msgID int64

func (in *connInfo) newMessage(b []byte, size int, tp MessageAction) *Message {
	msg := &Message{
		ID:      atomic.AddInt64(&msgID, 1),
		Service: in.service(),
		ConnID:  in.connID,
		SubID:   atomic.AddInt64(&in.subGroup, 1),
		Addr:    in.Conn.RemoteAddr().String(),
		Time:    timestamppb.New(time.Now()),
		Action:  tp,
	}
	if size > 0 {
		msg.Payload = b[:size]
	}
	return msg
}

func (in *connInfo) service() string {
	return service(in.Conn)
}

func service(conn fsconn.Info) string {
	name := fsconn.Service(conn)
	if name == nil {
		return ""
	}
	switch v := name.(type) {
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}
