// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package conndump

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/fsgo/fsgo/fsnet/fsconn"
	"github.com/fsgo/fsgo/fstypes"
)

// Dumper 流量 dump
type Dumper struct {
	// DataDir 数据存放目录，必填
	DataDir string

	// RotatorConfig 可选，用于配置 dump 的 Rotator
	RotatorConfig func(r *fsfs.Rotator)

	it             *fsconn.Interceptor
	enableRequest  fstypes.EnableStatus
	enableResponse fstypes.EnableStatus

	gid     int64
	outFile *fsfs.Rotator

	conns    map[fsconn.Info]*connInfo
	connsMux sync.RWMutex
}

// Setup 初始化配置
func (d *Dumper) init() error {
	if d.it != nil {
		return nil
	}
	d.conns = make(map[fsconn.Info]*connInfo)

	d.it = &fsconn.Interceptor{
		AfterWrite: d.dumpRequest,
		AfterRead:  d.dumpResponse,
		AfterClose: d.dumpFinish,
	}

	d.outFile = &fsfs.Rotator{
		Path:     filepath.Join(d.DataDir, "dump.pb"),
		ExtRule:  "10minute",
		MaxFiles: 24,
	}

	if d.RotatorConfig != nil {
		d.RotatorConfig(d.outFile)
	}

	if err := d.outFile.Init(); err != nil {
		return err
	}

	return nil
}

func (d *Dumper) Interceptor() *fsconn.Interceptor {
	if err := d.init(); err != nil {
		panic(err)
	}
	return d.it
}

func (d *Dumper) EnableRequest(enable bool) {
	d.enableRequest.SetEnable(enable)
}

func (d *Dumper) EnableResponse(enable bool) {
	d.enableResponse.SetEnable(enable)
}

// dumpRequest dump conn 里写出的数据(request)
func (d *Dumper) dumpRequest(conn fsconn.Info, b []byte, size int, err error) {
	if !d.enableRequest.IsEnable() {
		return
	}
	d.doDump(conn, b, size, Message_Request)
}

// dumpResponse dump conn 里收到的数据(response)
func (d *Dumper) dumpResponse(conn fsconn.Info, b []byte, size int, err error) {
	if !d.enableResponse.IsEnable() {
		return
	}
	d.doDump(conn, b, size, Message_Response)
}

func (d *Dumper) doDump(conn fsconn.Info, b []byte, size int, tp Message_Type) {
	ci := d.getConnInfo(conn)
	msg := ci.newMessage(b, size, tp)
	d.writeMessage(msg)
}

func (d *Dumper) writeMessage(msg proto.Message) {
	bf, err := proto.Marshal(msg)
	if err != nil {
		return
	}
	_, _ = d.outFile.Write(bf)
}

func (d *Dumper) nextGID() int64 {
	return atomic.AddInt64(&d.gid, 1)
}

func (d *Dumper) getConnInfo(conn fsconn.Info) *connInfo {
	d.connsMux.RLock()
	info := d.conns[conn]
	d.connsMux.RUnlock()
	if info != nil {
		return info
	}

	d.connsMux.Lock()
	defer d.connsMux.Unlock()
	info = d.conns[conn]
	if info != nil {
		return info
	}
	info = &connInfo{
		gid:  d.nextGID(),
		Conn: conn,
	}
	d.conns[conn] = info
	return info
}

func (d *Dumper) dumpFinish(conn fsconn.Info, _ error) {
	ci := d.getConnInfo(conn)
	msg := ci.newMessage(nil, 0, Message_Close)
	d.writeMessage(msg)
	d.connsMux.Lock()
	defer d.connsMux.Unlock()
	delete(d.conns, conn)
}

func (d *Dumper) Stop() {
	d.EnableRequest(false)
	d.EnableResponse(false)
	_ = d.outFile.Close()
}

type connInfo struct {
	Conn fsconn.Info
	gid  int64
	id   int64
}

func (in *connInfo) newMessage(b []byte, size int, tp Message_Type) *Message {
	msg := &Message{
		ID:   atomic.AddInt64(&in.id, 1),
		GID:  in.gid,
		Addr: in.Conn.RemoteAddr().String(),
		Time: time.Now().UnixNano(),
		Type: tp,
	}
	if size > 0 {
		msg.Payload = b[:size]
	}
	return msg
}
