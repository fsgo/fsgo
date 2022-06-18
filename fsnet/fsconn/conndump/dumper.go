// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package conndump

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"

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
		AfterWrite: d.dumpWrite,
		AfterRead:  d.dumpRead,
		AfterClose: d.dumpClose,
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

// dumpWrite dump conn 里写出的数据(request)
func (d *Dumper) dumpWrite(conn fsconn.Info, b []byte, size int, err error) {
	if !d.enableRequest.IsEnable() {
		return
	}
	d.doDump(conn, b, size, MessageAction_Write)
}

// dumpRead dump conn 里收到的数据(response)
func (d *Dumper) dumpRead(conn fsconn.Info, b []byte, size int, err error) {
	if !d.enableResponse.IsEnable() {
		return
	}
	d.doDump(conn, b, size, MessageAction_Read)
}

func (d *Dumper) doDump(conn fsconn.Info, b []byte, size int, tp MessageAction) {
	ci := d.getConnInfo(conn)
	msg := ci.newMessage(b, size, tp)
	d.writeMessage(msg)
}

func (d *Dumper) writeMessage(msg proto.Message) {
	bf, err := proto.Marshal(msg)
	if err != nil {
		return
	}
	b1 := make([]byte, len(bf)+4)
	binary.LittleEndian.PutUint32(b1, uint32(len(bf)))
	copy(b1[4:], bf)
	_, _ = d.outFile.Write(b1)
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
		groupID: d.nextGID(),
		Conn:    conn,
	}
	d.conns[conn] = info
	return info
}

func (d *Dumper) dumpClose(info fsconn.Info, _ error) {
	ci := d.getConnInfo(info)
	msg := ci.newMessage(nil, 0, MessageAction_Close)
	d.writeMessage(msg)
	d.connsMux.Lock()
	defer d.connsMux.Unlock()
	delete(d.conns, info)
}

func (d *Dumper) Stop() {
	d.EnableRequest(false)
	d.EnableResponse(false)
	_ = d.outFile.Close()
}

type connInfo struct {
	Conn      fsconn.Info
	groupID   int64
	idInGroup int64
}

var msgID int64

func (in *connInfo) newMessage(b []byte, size int, tp MessageAction) *Message {
	msg := &Message{
		ID:      atomic.AddInt64(&msgID, 1),
		Service: in.service(),
		Group:   in.groupID,
		GID:     atomic.AddInt64(&in.idInGroup, 1),
		Addr:    in.Conn.RemoteAddr().String(),
		Time:    time.Now().UnixNano(),
		Action:  tp,
	}
	if size > 0 {
		msg.Payload = b[:size]
	}
	return msg
}

func (in *connInfo) service() string {
	if ws, ok := in.Conn.(fsconn.HasService); ok {
		service := ws.Service()
		switch v := service.(type) {
		case string:
			return v
		default:
			return fmt.Sprint(v)
		}
	}
	return ""
}
